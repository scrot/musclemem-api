package workout_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/scrot/musclemem-api/internal"
	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/scrot/musclemem-api/internal/user"
	"github.com/scrot/musclemem-api/internal/workout"
)

func TestRetreivingWorkoutExercises(t *testing.T) {
	t.Parallel()

	xs := internal.NewMockSqliteDatastore(t)

	xs.WithUser(t, user.User{ID: 1})

	xs.WithWorkout(t, workout.Workout{ID: 1})
	xs.WithWorkout(t, workout.Workout{ID: 2})

	e1 := exercise.Exercise{ID: 1, Owner: 1, Workout: 1, Name: "Interval", Weight: 100.0, Repetitions: 8, Next: exercise.ExerciseRef{}, Previous: exercise.ExerciseRef{}}
	e2 := exercise.Exercise{ID: 2, Owner: 1, Workout: 1, Name: "Interval", Weight: 80.0, Repetitions: 10, Next: exercise.ExerciseRef{}, Previous: exercise.ExerciseRef{}}
	e3 := exercise.Exercise{ID: 3, Owner: 1, Workout: 1, Name: "Interval", Weight: 80.0, Repetitions: 10, Next: exercise.ExerciseRef{}, Previous: exercise.ExerciseRef{}}
	e4 := exercise.Exercise{ID: 4, Owner: 1, Workout: 2, Name: "Interval", Weight: 80.0, Repetitions: 10, Next: exercise.ExerciseRef{}, Previous: exercise.ExerciseRef{}}

	e1.Next = e2.ToRef()
	e2.Previous = e1.ToRef()
	e2.Next = e3.ToRef()
	e3.Previous = e2.ToRef()

	xs.WithExercise(t, e1)
	xs.WithExercise(t, e2)
	xs.WithExercise(t, e3)
	xs.WithExercise(t, e4)

	cs := []struct {
		name        string
		workoutID   int
		expected    []exercise.Exercise
		expectedErr error
	}{
		{"ErrorIfNegativeID", -1, []exercise.Exercise{}, workout.ErrInvalidID},
		{"ErrorIfNoID", 0, []exercise.Exercise{}, workout.ErrInvalidID},
		{"ErrorIfNotExist", 3, []exercise.Exercise{}, workout.ErrNotFound},
		{"RetreiveWorkoutExercises", 1, []exercise.Exercise{e1, e2, e3}, nil},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			xs, err := xs.Workouts.Exercises(1, c.workoutID)
			if !errors.Is(err, c.expectedErr) {
				t.Errorf("expected error '%v' but got '%v'", c.expectedErr, err)
			}

			if !cmp.Equal(xs, c.expected) {
				t.Errorf("expected '%v' but got '%v'", c.expected, xs)
			}
		})
	}
}
