package exercise

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/scrot/musclemem-api/internal/storage"
)

type SQLExerciseStore struct {
	*storage.SqliteDatastore
}

func NewSQLExerciseStore(db *storage.SqliteDatastore) *SQLExerciseStore {
	return &SQLExerciseStore{db}
}

// ExerciseByID returns an exercise from the database if exists
// otherwise returns NotFound error
func (xs *SQLExerciseStore) ByID(owner string, workout int, exercise int) (Exercise, error) {
	const (
		stmt = `
    SELECT owner, workout, exercise_index, name, weight, repetitions
    FROM exercises
    WHERE owner = {{ .Owner }} AND workout = {{ .Workout }} AND exercise_index = {{ .Exercise }}
    `
	)

	if owner == "" || workout <= 0 || exercise <= 0 {
		return Exercise{}, ErrInvalidFields
	}

	data := struct {
		Owner    string
		Workout  int
		Exercise int
	}{owner, workout, exercise}

	q, args, err := xs.CompileStatement(stmt, data)
	if err != nil {
		return Exercise{}, fmt.Errorf("WithID: compile: %w", err)
	}

	var e Exercise
	if err := xs.QueryRow(q, args...).Scan(
		&e.Owner,
		&e.Workout,
		&e.Index,
		&e.Name,
		&e.Weight,
		&e.Repetitions,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Exercise{}, nil
		}
		return Exercise{}, fmt.Errorf("WithID: %w", err)
	}

	return e, nil
}

func (xs *SQLExerciseStore) ByWorkout(owner string, workout int) ([]Exercise, error) {
	const (
		selectStmt = `
    SELECT owner, workout, exercise_index, name, weight, repetitions
    FROM exercises
    WHERE owner = {{ .Owner }} AND workout = {{ .Workout }}
    `
	)

	if owner == "" || workout <= 0 {
		return []Exercise{}, fmt.Errorf("ByWorkout: %w", ErrInvalidFields)
	}

	data := struct {
		Owner   string
		Workout int
	}{owner, workout}

	q, args, err := xs.CompileStatement(selectStmt, data)
	if err != nil {
		return []Exercise{}, fmt.Errorf("ByWorkout: compile: %w", err)
	}

	rs, err := xs.Query(q, args...)
	if err != nil {
		return []Exercise{}, fmt.Errorf("ByWorkout: query: %w", err)
	}

	var es []Exercise
	for rs.Next() {
		var e Exercise

		err := rs.Scan(
			&e.Owner,
			&e.Workout,
			&e.Index,
			&e.Name,
			&e.Weight,
			&e.Repetitions,
		)
		if err != nil {
			return []Exercise{}, fmt.Errorf("ByWorkout: scan: %w", err)
		}

		es = append(es, e)
	}

	sort.Sort(ByIndex(es))

	return es, nil
}

func (xs *SQLExerciseStore) New(owner string, workout int, name string, weight float64, repetitions int) (Exercise, error) {
	const stmt = `
  INSERT INTO exercises (owner, workout, exercise_index, name, weight, repetitions)
  VALUES ({{ .Owner }}, {{ .Workout }}, {{ .Index }}, {{ .Name }}, {{ .Weight }}, {{ .Repetitions }})
  `

	if owner == "" || workout <= 0 || name == "" || weight < 0 || repetitions < 0 {
		return Exercise{}, fmt.Errorf("New: %w", ErrInvalidFields)
	}

	if !xs.workoutExists(owner, workout) {
		return Exercise{}, fmt.Errorf("New: check workout %s/%d: %w", owner, workout, ErrNotFound)
	}

	last, err := xs.lastIndex(owner, workout)
	if err != nil {
		return Exercise{}, fmt.Errorf("New: get last index: %w", err)
	}

	x := Exercise{
		Owner:       owner,
		Workout:     workout,
		Index:       last + 1,
		Name:        name,
		Weight:      weight,
		Repetitions: repetitions,
	}

	q, args, err := xs.CompileStatement(stmt, x)
	if err != nil {
		return Exercise{}, fmt.Errorf("New: compile: %w", err)
	}

	if _, err := xs.Exec(q, args...); err != nil {
		return Exercise{}, fmt.Errorf("New: execute: %w", err)
	}

	ne, err := xs.ByID(owner, workout, last+1)
	if err != nil {
		return Exercise{}, fmt.Errorf("New: get exercise %s/%d/%d: %w", owner, workout, last+1, err)
	}

	return ne, nil
}

func (xs *SQLExerciseStore) ChangeName(owner string, workout int, exercise int, name string) (Exercise, error) {
	const stmt = `
  UPDATE exercises
  SET name = {{ .Name }}
  WHERE owner = {{ .Owner }} AND workout = {{ .Workout }} AND exercise_index = {{ .Exercise }}
  `
	if owner == "" || workout <= 0 || exercise <= 0 || name == "" {
		return Exercise{}, ErrInvalidFields
	}

	data := struct {
		Owner    string
		Workout  int
		Exercise int
		Name     string
	}{owner, workout, exercise, name}

	q, args, err := xs.CompileStatement(stmt, data)
	if err != nil {
		return Exercise{}, fmt.Errorf("ChangeName: compile: %w", err)
	}

	_, err = xs.Exec(q, args...)
	if err != nil {
		return Exercise{}, fmt.Errorf("ChangeName: execute: %w", err)
	}

	e, err := xs.ByID(owner, workout, exercise)
	if err != nil {
		return Exercise{}, fmt.Errorf("ChangeName: fetch exercise %s/%d/%d: %w", owner, workout, exercise, err)
	}

	return e, nil
}

func (xs *SQLExerciseStore) UpdateWeight(owner string, workout int, exercise int, weight float64) (Exercise, error) {
	const updateStmt = `
  UPDATE exercises
  SET weight = {{ .Weight }}
  WHERE owner = {{ .Owner }} AND workout = {{ .Workout }} AND exercise_index = {{ .Exercise }}
  `
	if owner == "" || workout <= 0 || exercise <= 0 || weight < 0 {
		return Exercise{}, ErrInvalidFields
	}

	data := struct {
		Owner    string
		Workout  int
		Exercise int
		Weight   float64
	}{owner, workout, exercise, weight}

	q, args, err := xs.CompileStatement(updateStmt, data)
	if err != nil {
		return Exercise{}, fmt.Errorf("UpdateWeights: compile: %w", err)
	}

	_, err = xs.Exec(q, args...)
	if err != nil {
		return Exercise{}, fmt.Errorf("UpdateWeights: execute: %w", err)
	}

	e, err := xs.ByID(owner, workout, exercise)
	if err != nil {
		return Exercise{}, fmt.Errorf("UpdateWeights: fetch exercise: %w", err)
	}

	return e, nil
}

func (xs *SQLExerciseStore) UpdateRepetitions(owner string, workout int, exercise int, repetitions int) (Exercise, error) {
	const updateStmt = `
  UPDATE exercises
  SET repetitions = {{ .Repetitions }}
  WHERE owner = {{ .Owner }} AND workout = {{ .Workout }} AND exercise_index = {{ .Exercise }}
  `
	if owner == "" || workout <= 0 || exercise <= 0 || repetitions < 0 {
		return Exercise{}, ErrInvalidFields
	}

	data := struct {
		Owner       string
		Workout     int
		Exercise    int
		Repetitions int
	}{owner, workout, exercise, repetitions}

	q, args, err := xs.CompileStatement(updateStmt, data)
	if err != nil {
		return Exercise{}, fmt.Errorf("UpdateRepetitions: %w", err)
	}

	_, err = xs.Exec(q, args...)
	if err != nil {
		return Exercise{}, fmt.Errorf("UpdateRepetitions: %w", err)
	}

	e, err := xs.ByID(owner, workout, exercise)
	if err != nil {
		return Exercise{}, fmt.Errorf("UpdateRepetitions: fetch exercise: %w", err)
	}

	return e, nil
}

func (xs *SQLExerciseStore) Delete(owner string, workout int, exercise int) (Exercise, error) {
	const stmt = `
  DELETE FROM exercises
  WHERE owner = {{ .Owner }} AND workout = {{ .Workout }} AND exercise_index = {{ .Index }}
  `

	e, err := xs.ByID(owner, workout, exercise)
	if err != nil {
		return Exercise{}, fmt.Errorf("Delete: fetch exercise: %w", err)
	}

	if owner == "" || workout <= 0 || exercise <= 0 {
		return Exercise{}, fmt.Errorf("Delete: %w", ErrInvalidFields)
	}

	data := struct {
		Owner   string
		Workout int
		Index   int
	}{owner, workout, exercise}
	q, args, err := xs.CompileStatement(stmt, data)
	if err != nil {
		return Exercise{}, fmt.Errorf("Delete: compile: %w", err)
	}

	res, err := xs.Exec(q, args...)
	if err != nil {
		return Exercise{}, fmt.Errorf("Delete: execute: %w", err)
	}

	c, err := res.RowsAffected()
	if err != nil {
		return Exercise{}, fmt.Errorf("Delete: %w", err)
	}

	if c == 0 {
		return Exercise{}, fmt.Errorf("Delete: %w", ErrNotFound)
	}

	return e, nil
}

func (xs *SQLExerciseStore) Swap(owner string, workout int, e1 int, e2 int) error {
	const stmt = `
  UPDATE exercises
  SET exercise_index = {{ .NewIndex }}
  WHERE owner = {{ .Owner }} AND workout = {{ .Workout }} AND exercise_index = {{ .Index }}
  `

	if owner == "" || workout <= 0 || e1 <= 0 || e2 <= 0 {
		return fmt.Errorf("Swap: %w", ErrInvalidFields)
	}

	if _, err := xs.ByID(owner, workout, e1); err != nil {
		return fmt.Errorf("Swap: check index %d: %w", e1, err)
	}

	if _, err := xs.ByID(owner, workout, e2); err != nil {
		return fmt.Errorf("Swap: check index %d: %w", e2, err)
	}

	var (
		temp = struct {
			Owner           string
			Workout         int
			Index, NewIndex int
		}{owner, workout, e2, 0}

		newE1 = struct {
			Owner           string
			Workout         int
			Index, NewIndex int
		}{owner, workout, e1, e2}

		newE2 = struct {
			Owner           string
			Workout         int
			Index, NewIndex int
		}{owner, workout, 0, e1}
	)

	tx, err := xs.Begin()
	if err != nil {
		return fmt.Errorf("Swap: new transaction: %w", err)
	}

	// unique key contraint, first move to other index
	q, args, err := xs.CompileStatement(stmt, temp)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Swap: compile temp: %w", err)
	}

	if _, err = tx.Exec(q, args...); err != nil {
		tx.Rollback()
		return fmt.Errorf("Swap: update temp: %w", err)
	}

	q, args, err = xs.CompileStatement(stmt, newE1)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Swap: compile %d: %w", e1, err)
	}

	if _, err = tx.Exec(q, args...); err != nil {
		tx.Rollback()
		return fmt.Errorf("Swap: update %d: %w", e1, err)
	}

	q, args, err = xs.CompileStatement(stmt, newE2)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("Swap: compile %d: %w", e2, err)
	}

	if _, err = tx.Exec(q, args...); err != nil {
		tx.Rollback()
		return fmt.Errorf("Swap: update %d: %w", e2, err)
	}

	tx.Commit()

	return nil
}

func (xs *SQLExerciseStore) Len(owner string, workout int) (int, error) {
	const stmt = `
  SELECT COUNT(*)
  FROM exercises
  WHERE owner = {{ .Owner }} AND workout = {{.Workout }}
  `
	data := struct {
		Owner   string
		Workout int
	}{owner, workout}

	q, args, err := xs.CompileStatement(stmt, data)
	if err != nil {
		return 0, fmt.Errorf("Len: compile: %w", err)
	}

	var count sql.NullInt32
	err = xs.QueryRow(q, args...).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("Len: query: %w", err)
	}

	if !count.Valid {
		return 0, errors.New("not a valid number")
	}

	return int(count.Int32), nil
}

// lastIndex returns the last exercise index of a workout
// if the index is 0 and no error, then there are no exercises
func (xs *SQLExerciseStore) lastIndex(owner string, workout int) (int, error) {
	const stmt = `
    SELECT MAX(exercise_index)
    FROM exercises
    WHERE owner = {{ .Owner }} AND workout = {{.Workout }}
    `

	data := struct {
		Owner   string
		Workout int
	}{owner, workout}

	q, args, err := xs.CompileStatement(stmt, data)
	if err != nil {
		return 0, fmt.Errorf("lastIndex: compile: %w", err)
	}

	var index sql.NullInt32
	err = xs.QueryRow(q, args...).Scan(&index)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("lastIndex: query: %w", err)
	}

	if !index.Valid {
		return 0, nil
	}

	return int(index.Int32), nil
}

func (xs *SQLExerciseStore) workoutExists(owner string, workout int) bool {
	const stmt = `
  SELECT 1
  FROM workouts
  WHERE owner = {{ .Owner }} AND workout_index = {{ .Workout }}
  `

	data := struct {
		Owner   string
		Workout int
	}{owner, workout}

	q, args, err := xs.CompileStatement(stmt, data)
	if err != nil {
		return false
	}

	var exists bool
	if err := xs.QueryRow(q, args...).Scan(&exists); err != nil {
		return false
	}

	return exists
}
