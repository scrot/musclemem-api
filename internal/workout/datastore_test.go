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

func TestCreatingNewWorkout(t *testing.T) {
	t.Parallel()

	xs := internal.NewMockSqliteDatastore(t)
	xs.WithUser(t, user.User{ID: 1})

	cs := []struct {
		name        string
		userID      int
		workoutName string
		expectedNew bool
		expectedErr error
	}{
		{"ErrorIfNegativeUserID", -1, "name", false, workout.ErrInvalidID},
		{"ErrorIfZeroUserID", 0, "name", false, workout.ErrMissingFields},
		{"ErrorIfUserNotExists", 2, "name", false, workout.ErrNotFound},
		{"ErrorIfNoName", 1, "", false, workout.ErrMissingFields},
		{"Valid", 1, "NEW", true, nil},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			id, err := xs.Workouts.New(c.userID, c.workoutName)
			if !errors.Is(err, c.expectedErr) {
				t.Errorf("expected error '%v' but got '%v'", c.expectedErr, err)
			}

			newID := id > 0
			if newID != c.expectedNew {
				t.Errorf("expected new id")
			}
		})
	}
}

func TestWorkoutByID(t *testing.T) {
	t.Parallel()

	xs := internal.NewMockSqliteDatastore(t)
	xs.WithUser(t, user.User{ID: 1})
	xs.WithWorkout(t, workout.Workout{ID: 1, Owner: 1})

	cs := []struct {
		name        string
		id          int
		expected    workout.Workout
		expectedErr error
	}{
		{"ErrorOnNegativeID", -1, workout.Workout{}, workout.ErrInvalidID},
		{"ErrorOnZeroID", 0, workout.Workout{}, workout.ErrMissingFields},
		{"ErrorOnUserNotExists", 2, workout.Workout{}, workout.ErrNotFound},
		{"WorkoutOnExists", 1, workout.Workout{ID: 1, Owner: 1}, nil},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			w, err := xs.Workouts.ByID(c.id)
			if !errors.Is(err, c.expectedErr) {
				t.Errorf("expected error '%v' but got '%v'", c.expectedErr, err)
			}

			if !cmp.Equal(w, c.expected) {
				t.Errorf("expected %v but got %v", c.expected, w)
			}
		})
	}
}

func TestRetreivingWorkoutExercises(t *testing.T) {
	t.Parallel()

	xs := internal.NewMockSqliteDatastore(t)

	xs.WithUser(t, user.User{ID: 1})

	xs.WithWorkout(t, workout.Workout{ID: 1, Owner: 1})
	xs.WithWorkout(t, workout.Workout{ID: 2, Owner: 1})

	e1 := exercise.Exercise{ID: 1, Workout: 1, Name: "Interval", Weight: 100.0, Repetitions: 8}
	e2 := exercise.Exercise{ID: 2, Workout: 1, Name: "Interval", Weight: 80.0, Repetitions: 10}
	e3 := exercise.Exercise{ID: 3, Workout: 1, Name: "Interval", Weight: 80.0, Repetitions: 10}
	e4 := exercise.Exercise{ID: 4, Workout: 2, Name: "Interval", Weight: 80.0, Repetitions: 10}

	e1.NextID = e2.ID
	e2.PreviousID = e1.ID
	e2.NextID = e3.ID
	e3.PreviousID = e2.ID

	xs.WithExercise(t, e1)
	xs.WithExercise(t, e2)
	xs.WithExercise(t, e3)
	xs.WithExercise(t, e4)

	cs := []struct {
		name        string
		userID      int
		workoutID   int
		expected    []exercise.Exercise
		expectedErr error
	}{
		{"ErrorIfNegativeWorkoutID", 1, -1, []exercise.Exercise{}, workout.ErrInvalidID},
		{"ErrorIfNoWorkoutID", 1, 0, []exercise.Exercise{}, workout.ErrInvalidID},
		{"ErrorIfWorkoutNotExist", 1, 3, []exercise.Exercise{}, workout.ErrNotFound},
		{"RetreiveWorkoutExercises", 1, 1, []exercise.Exercise{e1, e2, e3}, nil},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			xs, err := xs.Workouts.WorkoutExercises(c.workoutID)
			if !errors.Is(err, c.expectedErr) {
				t.Errorf("expected error '%v' but got '%v'", c.expectedErr, err)
			}

			if !cmp.Equal(xs, c.expected) {
				t.Errorf("expected '%v' but got '%v'", c.expected, xs)
			}
		})
	}
}

func TestChangingName(t *testing.T) {
	mock := internal.NewMockSqliteDatastore(t)
	mock.WithUser(t, user.User{ID: 1})
	mock.WithWorkout(t, workout.Workout{ID: 1, Owner: 1})

	if err := mock.Workouts.ChangeName(1, "NEW"); err != nil {
		t.Fatalf("expected no error but got '%v'", err)
	}

	w, err := mock.Workouts.ByID(1)
	if err != nil {
		t.Fatalf("expected no error but got '%v'", err)
	}

	if w.Name != "NEW" {
		t.Errorf("expected updated field but got '%s'", w.Name)
	}
}
