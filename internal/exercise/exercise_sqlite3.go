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
    SELECT exercise_id, owner, workout, name, weight, repetitions, previous, next
    FROM exercises
    WHERE exercise_id = {{ . }}
    `
		refStmt = `
    SELECT exercise_id, name
    FROM exercises
    WHERE {{ . }} = exercise_id
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

	var (
		e    Exercise
		prev int
		next int
	)

	if err := ds.db.QueryRow(q, args...).Scan(
		&e.ID,
		&e.Owner,
		&e.Workout,
		&e.Name,
		&e.Weight,
		&e.Repetitions,
		&prev,
		&next,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Exercise{}, fmt.Errorf("WithID: %w", ErrNotFound)
		}
		return Exercise{}, fmt.Errorf("WithID: %w", err)
	}

	if prev != 0 {
		var pRef ExerciseRef

		q, args, err = tmpl.Compile(refStmt, prev)
		if err != nil {
			return Exercise{}, fmt.Errorf("WithID: %w", err)
		}

		if err := ds.db.QueryRow(q, args...).Scan(&pRef.ID, &pRef.Name); err != nil {
			return Exercise{}, fmt.Errorf("WithID: %w", err)
		}

		e.Previous = pRef
	}

	if next != 0 {
		var nRef ExerciseRef

		q, args, err = tmpl.Compile(refStmt, next)
		if err != nil {
			return Exercise{}, fmt.Errorf("WithID: %w", err)
		}

		if err := ds.db.QueryRow(q, args...).Scan(&nRef.ID, &nRef.Name); err != nil {
			return Exercise{}, fmt.Errorf("WithID: %w", err)
		}

		e.Next = nRef
	}

	return e, nil
}

// StoreExercise stores the given exrcise in the sqlite database
// If the id already exists it updates the existing record
// ErrMissingFields is returned when fields are missing
func (ds *SqliteExercises) New(owner, workout int, name string,
	weight float64, repetitions int,
) (int, error) {
	const (
		insertStmt = `
    INSERT INTO exercises (owner, workout, name, weight, repetitions, previous, next)
    VALUES ({{.Owner}}, {{.Workout}}, {{.Name}}, {{.Weight}}, {{.Repetitions}}, {{.Previous.ID}}, {{.Next.ID}});
    `
		updateNextStmt = `
    UPDATE exercises
    SET next = {{ .NextID }}
    WHERE exercise_id = {{ .ID }}
    `
		updatePreviousStmt = `
    UPDATE exercises
    SET previous = {{.PreviousID }}   
    WHERE exercise_id = {{ .ID }}
    `
	)

	// required fields
	if owner <= 0 || workout <= 0 || name == "" ||
		weight <= 0 || repetitions <= 0 {
		return 0, fmt.Errorf("New: %w", ErrMissingFields)
	}

	tmpl, err := tqla.New()
	if err != nil {
		return 0, err
	}

	// get last exercise in workout
	tail, err := tail(ds, owner, workout)
	if err != nil {
		return 0, fmt.Errorf("New: %w", err)
	}

	// insert new exercise
	tx, err := ds.db.Begin()
	if err != nil {
		return 0, nil
	}

	x := Exercise{
		Owner:       owner,
		Workout:     workout,
		Name:        name,
		Weight:      weight,
		Repetitions: repetitions,
	}

	insert, args, err := tmpl.Compile(insertStmt, x)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("New: insert new exercise: %w", err)
	}

	res, err := tx.Exec(insert, args...)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("New: insert new exercise: %w", err)
	}

	newID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("New: insert new exercise: %w", err)
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
			return 0, fmt.Errorf("New: patch new exercise: %w", err)
		}

		_, err = tx.Exec(patchNew, args...)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("New: patch new exercise: %w", err)
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
			return 0, fmt.Errorf("New: patch old exercise: %w", err)
		}

		_, err = tx.Exec(patchOld, args...)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("New: patch old exercise: %w", err)
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
  SET name = {{ .Name }},
  WHERE exercise_id = {{ .ID }}
  `

	tmpl, err := tqla.New()
	if err != nil {
		return err
	}

	data := struct {
		ID   int
		Name string
	}{id, name}

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

func (ds *SqliteExercises) UpdateWeight(id int, weight float64) error {
	const updateStmt = `
  UPDATE exercises
  SET weight = {{ .Weight }},
  WHERE exercise_id = {{ .ID }}
  `

	tmpl, err := tqla.New()
	if err != nil {
		return err
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
  SET repetitions = {{ .Repetitions }},
  WHERE exercise_id = {{ .ID }}
  `

	tmpl, err := tqla.New()
	if err != nil {
		return err
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

// TODO: Implement
func (ds *SqliteExercises) Swap(i int, j int) error {
	return nil
}

// tail returns last exercise in linked list or empty exercise if
// no exercises exists for that workout
func tail(ds *SqliteExercises, owner, workout int) (Exercise, error) {
	const (
		tailStmt = `
    SELECT exercise_id   
    FROM exercises
    WHERE owner={{.Owner}} AND workout={{.Workout}} AND next=0
    `
	)

	tmpl, err := tqla.New()
	if err != nil {
		return Exercise{}, err
	}

	data := struct {
		Owner   int
		Workout int
	}{owner, workout}

	// find last exercise if at least one in user workout
	q, args, err := tmpl.Compile(tailStmt, data)
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

func (ds *SqliteExercises) exists(id int) bool {
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
