package exercise

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/VauntDev/tqla"
	"github.com/google/go-cmp/cmp"
)

type SqliteExercises struct {
	db *sql.DB
}

func NewSqliteExercises(db *sql.DB) *SqliteExercises {
	return &SqliteExercises{db}
}

// ExerciseByID returns an exercise from the database if exists
// otherwise returns NotFound error
func (ds *SqliteExercises) WithID(id int) (Exercise, error) {
	const (
		stmt = `
    SELECT exercise_id, workout, name, weight, repetitions, previous, next
    FROM exercises
    WHERE exercise_id = {{ . }}
    `
	)

	if id <= 0 {
		return Exercise{}, ErrInvalidID
	}

	tmpl, err := tqla.New()
	if err != nil {
		return Exercise{}, err
	}

	q, args, err := tmpl.Compile(stmt, id)
	if err != nil {
		return Exercise{}, fmt.Errorf("WithID: %w", err)
	}

	var e Exercise

	if err := ds.db.QueryRow(q, args...).Scan(
		&e.ID,
		&e.Workout,
		&e.Name,
		&e.Weight,
		&e.Repetitions,
		&e.PreviousID,
		&e.NextID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Exercise{}, fmt.Errorf("WithID: %w", ErrNotFound)
		}
		return Exercise{}, fmt.Errorf("WithID: %w", err)
	}

	return e, nil
}

// StoreExercise stores the given exrcise in the sqlite database
// If the id already exists it updates the existing record
// ErrMissingFields is returned when fields are missing
func (ds *SqliteExercises) New(workout int, name string, weight float64, repetitions int) (int, error) {
	const (
		insertStmt = `
    INSERT INTO exercises (workout, name, weight, repetitions, previous, next)
    VALUES ({{.Workout}}, {{.Name}}, {{.Weight}}, {{.Repetitions}}, {{.PreviousID}}, {{.NextID}})
    `
		updatePreviousStmt = `
    UPDATE exercises
    SET previous = {{.PreviousID }}   
    WHERE exercise_id = {{ .ID }}
    `
		updateNextStmt = `
    UPDATE exercises
    SET next = {{ .NextID }}
    WHERE exercise_id = {{ .ID }}
    `
	)

	if !ds.workoutExists(workout) {
		return 0, fmt.Errorf("New: workoutExists: %w", ErrNotFound)
	}

	if name == "" || weight <= 0 || repetitions <= 0 {
		return 0, fmt.Errorf("New: %w", ErrMissingFields)
	}

	tmpl, err := tqla.New()
	if err != nil {
		return 0, fmt.Errorf("New: %w", err)
	}

	// get last exercise in workout
	tail, err := tail(ds, workout)
	if err != nil {
		return 0, fmt.Errorf("New: %w", err)
	}

	// insert new exercise
	tx, err := ds.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("New: begin transaction: %w", err)
	}

	x := Exercise{
		Workout:     workout,
		Name:        name,
		Weight:      weight,
		Repetitions: repetitions,
	}

	insert, args, err := tmpl.Compile(insertStmt, x)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("New: compile insert statement: %w", err)
	}

	res, err := tx.Exec(insert, args...)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("New: execute insert new: %w", err)
	}

	newID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("New: last id: %w", err)
	}

	// update exercise references
	if !cmp.Equal(tail, Exercise{}) {
		newExPatch := struct {
			ID         int
			PreviousID int
		}{
			ID:         int(newID),
			PreviousID: tail.ID,
		}

		patchNew, args, err := tmpl.Compile(updatePreviousStmt, newExPatch)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("New: compile previous statement: %w", err)
		}

		_, err = tx.Exec(patchNew, args...)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("New: execute patch new exercise: %w", err)
		}

		oldExPatch := struct {
			ID     int
			NextID int
		}{
			ID:     tail.ID,
			NextID: int(newID),
		}

		patchOld, args, err := tmpl.Compile(updateNextStmt, oldExPatch)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("New: compile next statement: %w", err)
		}

		_, err = tx.Exec(patchOld, args...)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("New: execute patch last exercise: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("New: commit transaction: %w", err)
	}

	return int(newID), nil
}

func (ds *SqliteExercises) ChangeName(id int, name string) error {
	const updateStmt = `
  UPDATE exercises
  SET name = {{ .Name }}
  WHERE exercise_id = {{ .ID }}
  `
	if id <= 0 {
		return ErrInvalidID
	}

	if name == "" {
		return ErrMissingFields
	}

	tmpl, err := tqla.New()
	if err != nil {
		return fmt.Errorf("ChangeName: %w", err)
	}

	data := struct {
		ID   int
		Name string
	}{id, name}

	q, args, err := tmpl.Compile(updateStmt, data)
	if err != nil {
		return fmt.Errorf("ChangeName: %w", err)
	}

	_, err = ds.db.Exec(q, args...)
	if err != nil {
		return fmt.Errorf("ChangeName: %w", err)
	}

	return nil
}

func (ds *SqliteExercises) UpdateWeight(id int, weight float64) error {
	const updateStmt = `
  UPDATE exercises
  SET weight = {{ .Weight }}
  WHERE exercise_id = {{ .ID }}
  `
	if id <= 0 {
		return ErrInvalidID
	}

	if weight < 0 {
		return ErrNegativeValue
	}

	tmpl, err := tqla.New()
	if err != nil {
		return fmt.Errorf("UpdateWeights: %w", err)
	}

	data := struct {
		ID     int
		Weight float64
	}{id, weight}

	q, args, err := tmpl.Compile(updateStmt, data)
	if err != nil {
		return fmt.Errorf("UpdateWeights: %w", err)
	}

	_, err = ds.db.Exec(q, args...)
	if err != nil {
		return fmt.Errorf("UpdateWeights: %w", err)
	}

	return nil
}

func (ds *SqliteExercises) UpdateRepetitions(id int, repetitions int) error {
	const updateStmt = `
  UPDATE exercises
  SET repetitions = {{ .Repetitions }}
  WHERE exercise_id = {{ .ID }}
  `
	if id <= 0 {
		return ErrInvalidID
	}

	if repetitions < 0 {
		return ErrNegativeValue
	}

	tmpl, err := tqla.New()
	if err != nil {
		return fmt.Errorf("UpdateRepetitions: %w", err)
	}

	data := struct {
		ID          int
		Repetitions int
	}{id, repetitions}

	q, args, err := tmpl.Compile(updateStmt, data)
	if err != nil {
		return fmt.Errorf("UpdateRepetitions: %w", err)
	}

	_, err = ds.db.Exec(q, args...)
	if err != nil {
		return fmt.Errorf("UpdateRepetitions: %w", err)
	}

	return nil
}

func (ds *SqliteExercises) Delete(id int) error {
	deleteStmt := `
  DELETE FROM exercises
  WHERE exercise_id = {{.}}
  `

	if id <= 0 {
		return fmt.Errorf("Delete: %w", ErrInvalidID)
	}

	tmpl, err := tqla.New()
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}

	q, args, err := tmpl.Compile(deleteStmt, id)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	res, err := ds.db.Exec(q, args...)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	c, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}

	if c == 0 {
		return fmt.Errorf("Delete: %w", ErrNotFound)
	}

	return nil
}

// tail returns last exercise in linked list or empty exercise if
// no exercises exists for that workout
func tail(ds *SqliteExercises, workout int) (Exercise, error) {
	const (
		tailStmt = `
    SELECT exercise_id   
    FROM exercises
    WHERE workout={{ . }} AND next=0
    `
	)

	tmpl, err := tqla.New()
	if err != nil {
		return Exercise{}, err
	}

	// find last exercise if at least one in user workout
	q, args, err := tmpl.Compile(tailStmt, workout)
	if err != nil {
		return Exercise{}, fmt.Errorf("tail: %w", err)
	}

	var tailID int

	err = ds.db.QueryRow(q, args...).Scan(&tailID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Exercise{}, nil
		}
		return Exercise{}, fmt.Errorf("tail: %w", err)
	}

	tailExercise, err := ds.WithID(tailID)
	if err != nil {
		return Exercise{}, fmt.Errorf("tail: %w", err)
	}
	return tailExercise, nil
}

func (ds *SqliteExercises) userExists(id int) bool {
	stmt := `
  SELECT 1
  FROM users
  WHERE user_id = {{ . }}
  `

	tmpl, err := tqla.New()
	if err != nil {
		return false
	}

	q, args, err := tmpl.Compile(stmt, id)
	if err != nil {
		return false
	}

	var exists bool
	if err := ds.db.QueryRow(q, args...).Scan(&exists); err != nil {
		return false
	}

	return exists
}

func (ds *SqliteExercises) workoutExists(id int) bool {
	stmt := `
  SELECT 1
  FROM workouts
  WHERE workout_id = {{ . }}
  `

	tmpl, err := tqla.New()
	if err != nil {
		return false
	}

	q, args, err := tmpl.Compile(stmt, id)
	if err != nil {
		return false
	}

	var exists bool
	if err := ds.db.QueryRow(q, args...).Scan(&exists); err != nil {
		return false
	}

	return exists
}

func (ds *SqliteExercises) exerciseExists(id int) bool {
	stmt := `
  SELECT 1
  FROM exercises
  WHERE exercise_id = {{ . }}
  `

	tmpl, err := tqla.New()
	if err != nil {
		return false
	}

	q, args, err := tmpl.Compile(stmt, id)
	if err != nil {
		return false
	}

	var exists bool
	if err := ds.db.QueryRow(q, args...).Scan(&exists); err != nil {
		return false
	}

	return exists
}
