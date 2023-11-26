package workout_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/scrot/musclemem-api/internal"
	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/scrot/musclemem-api/internal/user"
	"github.com/scrot/musclemem-api/internal/workout"
)

func TestRetreivingWorkoutExercises(t *testing.T) {
	xs := internal.NewMockSqliteDatastore(t)
	xs.WithUser(t, user.User{ID: 1})
	xs.WithWorkout(t, workout.Workout{ID: 1})

	e1 := exercise.Exercise{ID: 1, Owner: 1, Workout: 1, Name: "Interval", Weight: 100.0, Repetitions: 8, Next: exercise.ExerciseRef{}, Previous: exercise.ExerciseRef{}}
	e2 := exercise.Exercise{ID: 2, Owner: 1, Workout: 1, Name: "Interval", Weight: 80.0, Repetitions: 10, Next: exercise.ExerciseRef{}, Previous: exercise.ExerciseRef{}}
	e3 := exercise.Exercise{ID: 3, Owner: 1, Workout: 1, Name: "Interval", Weight: 80.0, Repetitions: 10, Next: exercise.ExerciseRef{}, Previous: exercise.ExerciseRef{}}

	e1.Next = e2.ToRef()
	e2.Previous = e1.ToRef()
	e2.Next = e3.ToRef()
	e3.Previous = e2.ToRef()

	xs.WithExercise(t, e1)
	xs.WithExercise(t, e2)
	xs.WithExercise(t, e3)

	t.Run("ErrorInvalidID", func(t *testing.T) {
		xs, err := xs.Workouts.Exercises(0, 0)
		if err == nil {
			t.Errorf("expected %q but got nil", exercise.ErrInvalidID)
		}
		if len(xs) != 0 {
			t.Errorf("expected empty list but got %v", xs)
		}
	})

	t.Run("ErrorNotFound", func(t *testing.T) {
		xs, err := xs.Workouts.Exercises(1, 2)
		if err == nil {
			t.Errorf("expected %q but got nil", exercise.ErrNotFound)
		}
		if len(xs) != 0 {
			t.Errorf("expected empty list but got %v", xs)
		}
	})

	t.Run("ListOnValidWorkout", func(t *testing.T) {
		xs, err := xs.Workouts.Exercises(1, 1)
		if err != nil {
			t.Errorf("expected no error but got %q", err)
		}
		if len(xs) != 3 {
			t.Errorf("expected %d exercises but got %d", 3, len(xs))
		}

		if !cmp.Equal(xs, []exercise.Exercise{e1, e2, e3}) {
			t.Errorf("expected %v but got %v", []exercise.Exercise{e1, e2, e3}, xs)
		}
	})
}
