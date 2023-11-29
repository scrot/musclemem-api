package workout

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/VauntDev/tqla"
	"github.com/scrot/musclemem-api/internal/exercise"
)

type SQLWorkouts struct {
	db *sql.DB
}

func NewSQLWorkouts(db *sql.DB) *SQLWorkouts {
	return &SQLWorkouts{db}
}

func (ds *SQLWorkouts) New(owner int, name string) (int, error) {
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

	if !ds.userExists(owner) {
		return 0, fmt.Errorf("New: validate user: %w", ErrNotFound)
	}

	tmpl, err := tqla.New()
	if err != nil {
		return 0, fmt.Errorf("New: %w", err)
	}

	q, args, err := tmpl.Compile(stmt, Workout{Owner: owner, Name: name})
	if err != nil {
		return 0, fmt.Errorf("New: compile insert statement: %w", err)
	}

	res, err := ds.db.Exec(q, args...)
	if err != nil {
		return 0, fmt.Errorf("New: execute insert statement: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("New: get new id: %w", err)
	}

	return int(id), nil
}

func (ds *SQLWorkouts) ByID(workoutID int) (Workout, error) {
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

	tmpl, err := tqla.New()
	if err != nil {
		return Workout{}, fmt.Errorf("ByID: %w", err)
	}

	q, args, err := tmpl.Compile(stmt, workoutID)
	if err != nil {
		return Workout{}, fmt.Errorf("ByID: compile statement: %w", err)
	}

	var w Workout
	if err := ds.db.QueryRow(q, args...).Scan(&w.ID, &w.Owner, &w.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Workout{}, ErrNotFound
		}
		return Workout{}, fmt.Errorf("ByID: query workout: %w", err)
	}

	return w, nil
}

func (ds *SQLWorkouts) WorkoutExercises(workoutID int) ([]exercise.Exercise, error) {
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

	w, err := ds.ByID(workoutID)
	if err != nil {
		return []exercise.Exercise{}, fmt.Errorf("WorkoutExercises: get workout: %w", err)
	}

	tmpl, err := tqla.New()
	if err != nil {
		return []exercise.Exercise{}, fmt.Errorf("WorkoutExercises: %w", err)
	}

	q, args, err := tmpl.Compile(selectStmt, w)
	if err != nil {
		return []exercise.Exercise{}, fmt.Errorf("WorkoutExercises: compile select statement: %w", err)
	}

	rs, err := ds.db.Query(q, args...)
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

	return xs, nil
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

	tmpl, err := tqla.New()
	if err != nil {
		return fmt.Errorf("ChangeName: %w", err)
	}

	patch := struct {
		ID   int
		Name string
	}{id, name}
	q, args, err := tmpl.Compile(stmt, patch)
	if err != nil {
		return fmt.Errorf("ChangeName: compile update statement: %w", err)
	}

	if _, err := ws.db.Exec(q, args...); err != nil {
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

	tmpl, err := tqla.New()
	if err != nil {
		return false
	}

	q, args, err := tmpl.Compile(stmt, userID)
	if err != nil {
		return false
	}

	var exists bool
	if err := ws.db.QueryRow(q, args...).Scan(&exists); err != nil {
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

	tmpl, err := tqla.New()
	if err != nil {
		return false
	}

	q, args, err := tmpl.Compile(stmt, workoutID)
	if err != nil {
		return false
	}

	var exists bool
	if err := ws.db.QueryRow(q, args...).Scan(&exists); err != nil {
		return false
	}

	return exists
}
