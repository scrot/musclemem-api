package workout_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/scrot/musclemem-api/internal/storage"
	"github.com/scrot/musclemem-api/internal/workout"
)

func TestCreatingNewWorkout(t *testing.T) {
	t.Parallel()

	mock := storage.NewMockSqliteDatastore(t)
	mock.WithUser(t, 1, "", "")
	ws := workout.NewSQLWorkoutStore(mock.SqliteDatastore)

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
			id, err := ws.New(c.userID, c.workoutName)
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

	mock := storage.NewMockSqliteDatastore(t)
	mock.WithUser(t, 1, "", "")
	mock.WithWorkout(t, 1, 1, "")
	ws := workout.NewSQLWorkoutStore(mock.SqliteDatastore)

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
			w, err := ws.ByID(c.id)
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

	mock := storage.NewMockSqliteDatastore(t)
	mock.WithUser(t, 1, "u1", "pwd")
	mock.WithWorkout(t, 1, 1, "w1")
	mock.WithWorkout(t, 2, 1, "w2")
	mock.WithExercise(t, 1, 1, "w1e1", 100, 10, 0, 2)
	mock.WithExercise(t, 2, 1, "w1e2", 100, 10, 1, 3)
	mock.WithExercise(t, 3, 1, "w1e3", 100, 10, 2, 0)
	mock.WithExercise(t, 4, 2, "w2e1", 100, 10, 0, 0)

	e1 := exercise.Exercise{ID: 1, Workout: 1, Name: "w1e1", Weight: 100, Repetitions: 10, PreviousID: 0, NextID: 2}
	e2 := exercise.Exercise{ID: 2, Workout: 1, Name: "w1e2", Weight: 100, Repetitions: 10, PreviousID: 1, NextID: 3}
	e3 := exercise.Exercise{ID: 3, Workout: 1, Name: "w1e3", Weight: 100, Repetitions: 10, PreviousID: 2, NextID: 0}

	ws := workout.NewSQLWorkoutStore(mock.SqliteDatastore)

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
			xs, err := ws.WorkoutExercises(c.workoutID)
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
	mock := storage.NewMockSqliteDatastore(t)
	mock.WithUser(t, 1, "u1", "pwd")
	mock.WithWorkout(t, 1, 1, "w1")

	ws := workout.NewSQLWorkoutStore(mock.SqliteDatastore)

	if err := ws.ChangeName(1, "NEW"); err != nil {
		t.Fatalf("expected no error but got '%v'", err)
	}

	w, err := ws.ByID(1)
	if err != nil {
		t.Fatalf("expected no error but got '%v'", err)
	}

	if w.Name != "NEW" {
		t.Errorf("expected updated field but got '%s'", w.Name)
	}
}
