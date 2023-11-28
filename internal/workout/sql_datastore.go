package workout

import (
	"database/sql"
	"errors"

	"github.com/VauntDev/tqla"
	"github.com/scrot/musclemem-api/internal/exercise"
)

type SqliteWorkouts struct {
	db *sql.DB
}

func NewSqliteWorkouts(db *sql.DB) *SqliteWorkouts {
	return &SqliteWorkouts{db}
}

func (ds *SqliteWorkouts) New(userID int, name string) (int, error) {
	return 0, errors.New("todo")
}

func (ds *SqliteWorkouts) Exercises(userID, workoutID int) ([]exercise.Exercise, error) {
	const (
		selectStmt = `
    SELECT exercise_id, owner, workout, name, weight, repetitions, previous, next
    FROM exercises
    WHERE owner = {{ .OwnerID }} AND workout = {{ .WorkoutID }}
    `
		selectRefsStmt = `
    SELECT exercise_id, name
    FROM exercises
    WHERE exercise_id = {{.}}
    `
	)

	if userID <= 0 {
		return []exercise.Exercise{}, ErrUserNotExist
	}

	if workoutID <= 0 {
		return []exercise.Exercise{}, ErrWorkoutNotExist
	}

	tmpl, err := tqla.New()
	if err != nil {
		return []exercise.Exercise{}, err
	}

	data := struct {
		OwnerID   int
		WorkoutID int
	}{userID, workoutID}

	q, args, err := tmpl.Compile(selectStmt, data)
	if err != nil {
		return []exercise.Exercise{}, err
	}

	rs, err := ds.db.Query(q, args...)
	if err != nil {
		return []exercise.Exercise{}, err
	}

	var xs []exercise.Exercise
	for rs.Next() {
		var (
			x          exercise.Exercise
			prev, next int
		)

		err := rs.Scan(&x.ID, &x.Owner, &x.Workout, &x.Name, &x.Weight,
			&x.Repetitions, &prev, &next)
		if err != nil {
			return []exercise.Exercise{}, err
		}

		if prev != 0 {
			q, args, err := tmpl.Compile(selectRefsStmt, prev)
			if err != nil {
				return []exercise.Exercise{}, err
			}

			if err := ds.db.QueryRow(q, args...).Scan(&x.Previous.ID, &x.Previous.Name); err != nil {
				return []exercise.Exercise{}, err
			}
		}

		if next != 0 {
			q, args, err := tmpl.Compile(selectRefsStmt, next)
			if err != nil {
				return []exercise.Exercise{}, err
			}

			if err := ds.db.QueryRow(q, args...).Scan(&x.Next.ID, &x.Next.Name); err != nil {
				return []exercise.Exercise{}, err
			}
		}

		xs = append(xs, x)
	}

	if len(xs) == 0 {
		return []exercise.Exercise{}, ErrWorkoutNotExist
	}

	return xs, nil
}
