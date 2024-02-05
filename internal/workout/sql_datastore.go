package workout

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/scrot/musclemem-api/internal/storage"
)

type SQLWorkouts struct {
	*storage.SqliteDatastore
}

func NewSQLWorkouts(db *storage.SqliteDatastore) *SQLWorkouts {
	return &SQLWorkouts{db}
}

func (ws *SQLWorkouts) New(owner int, name string) (int, error) {
	const stmt = `
    INSERT INTO workouts (owner, name)
    VALUES ({{ .Owner }}, {{ .Name }})
    `

	if owner < 0 {
		return 0, fmt.Errorf("New: %w", ErrInvalidID)
	}

	if owner == 0 || name == "" {
		return 0, fmt.Errorf("New: %w", ErrMissingFields)
	}

	if !ws.userExists(owner) {
		return 0, fmt.Errorf("New: validate user: %w", ErrNotFound)
	}

	q, args, err := ws.CompileStatement(stmt, Workout{Owner: owner, Name: name})
	if err != nil {
		return 0, fmt.Errorf("New: compile insert statement: %w", err)
	}

	res, err := ws.Exec(q, args...)
	if err != nil {
		return 0, fmt.Errorf("New: execute insert statement: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("New: get new id: %w", err)
	}

	return int(id), nil
}

func (ws *SQLWorkouts) ByID(workoutID int) (Workout, error) {
	const stmt = `
    SELECT workout_id, owner, name
    FROM workouts
    WHERE workout_id = {{ . }}
    `
	if workoutID < 0 {
		return Workout{}, ErrInvalidID
	}

	if workoutID == 0 {
		return Workout{}, ErrMissingFields
	}

	q, args, err := ws.CompileStatement(stmt, workoutID)
	if err != nil {
		return Workout{}, fmt.Errorf("ByID: compile statement: %w", err)
	}

	var w Workout
	if err := ws.QueryRow(q, args...).Scan(&w.ID, &w.Owner, &w.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Workout{}, ErrNotFound
		}
		return Workout{}, fmt.Errorf("ByID: query workout: %w", err)
	}

	return w, nil
}

func (ws *SQLWorkouts) WorkoutExercises(workoutID int) ([]exercise.Exercise, error) {
	const (
		selectStmt = `
    SELECT exercise_id, workout, name, weight, repetitions, previous, next
    FROM exercises
    WHERE workout = {{ .ID }}
    `
	)

	if workoutID <= 0 {
		return []exercise.Exercise{}, fmt.Errorf("WorkoutExercises: %w", ErrInvalidID)
	}

	w, err := ws.ByID(workoutID)
	if err != nil {
		return []exercise.Exercise{}, fmt.Errorf("WorkoutExercises: get workout: %w", err)
	}

	q, args, err := ws.CompileStatement(selectStmt, w)
	if err != nil {
		return []exercise.Exercise{}, fmt.Errorf("WorkoutExercises: compile select statement: %w", err)
	}

	rs, err := ws.Query(q, args...)
	if err != nil {
		return []exercise.Exercise{}, fmt.Errorf("WorkoutExercises: query workout exercises: %w", err)
	}

	var xs []exercise.Exercise
	for rs.Next() {
		var x exercise.Exercise

		err := rs.Scan(
			&x.ID,
			&x.Workout,
			&x.Name,
			&x.Weight,
			&x.Repetitions,
			&x.PreviousID,
			&x.NextID,
		)
		if err != nil {
			return []exercise.Exercise{}, fmt.Errorf("WorkoutExercises: scan row: %w", err)
		}

		xs = append(xs, x)
	}

	return exercise.Sort(xs), nil
}

func (ws *SQLWorkouts) ChangeName(id int, name string) error {
	const stmt = `
    UPDATE workouts
    SET name = {{ .Name }}
    WHERE workout_id = {{ .ID }}
    `
	if id < 0 {
		return fmt.Errorf("ChangeName: %w", ErrInvalidID)
	}

	if id == 0 || name == "" {
		return fmt.Errorf("ChangeName: %w", ErrMissingFields)
	}

	if !ws.workoutExists(id) {
		return fmt.Errorf("ChangeName: workout %d: %w", id, ErrNotFound)
	}

	patch := struct {
		ID   int
		Name string
	}{id, name}
	q, args, err := ws.CompileStatement(stmt, patch)
	if err != nil {
		return fmt.Errorf("ChangeName: compile update statement: %w", err)
	}

	if _, err := ws.Exec(q, args...); err != nil {
		return fmt.Errorf("ChangeName: execute update name: %w", err)
	}

	return nil
}

func (ws *SQLWorkouts) userExists(userID int) bool {
	const stmt = `
  SELECT 1
  FROM users
  WHERE user_id = {{ . }}
  `

	q, args, err := ws.CompileStatement(stmt, userID)
	if err != nil {
		return false
	}

	var exists bool
	if err := ws.QueryRow(q, args...).Scan(&exists); err != nil {
		return false
	}

	return exists
}

func (ws *SQLWorkouts) workoutExists(workoutID int) bool {
	const stmt = `
  SELECT 1
  FROM workouts
  WHERE workout_id = {{ . }}
  `

	q, args, err := ws.CompileStatement(stmt, workoutID)
	if err != nil {
		return false
	}

	var exists bool
	if err := ws.QueryRow(q, args...).Scan(&exists); err != nil {
		return false
	}

	return exists
}
