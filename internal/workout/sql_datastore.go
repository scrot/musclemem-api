package workout

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/VauntDev/tqla"
	"github.com/scrot/musclemem-api/internal/exercise"
)

type SqliteWorkouts struct {
	db *sql.DB
}

func NewSQLWorkouts(db *sql.DB) *SqliteWorkouts {
	return &SqliteWorkouts{db}
}

func (ds *SqliteWorkouts) New(owner int, name string) (int, error) {
	return 0, errors.New("todo")
}

func (ds *SqliteWorkouts) ByID(id int) (Workout, error) {
	const stmt = `
    SELECT workout_id, owner, name
    FROM workouts
    WHERE workout_id = {{ . }}
    `

	tmpl, err := tqla.New()
	if err != nil {
		return Workout{}, fmt.Errorf("ByID: %w", err)
	}

	q, args, err := tmpl.Compile(stmt, id)
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

func (ds *SqliteWorkouts) WorkoutExercises(workoutID int) ([]exercise.Exercise, error) {
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
