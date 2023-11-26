package exercise_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/scrot/musclemem-api/internal"
	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/scrot/musclemem-api/internal/user"
	"github.com/scrot/musclemem-api/internal/workout"
)

func TestExercisesWithID(t *testing.T) {
	xs := internal.NewMockSqliteDatastore(t)
	xs.WithUser(t, user.User{ID: 1})
	xs.WithWorkout(t, workout.Workout{ID: 1})

	e1 := exercise.Exercise{ID: 1, Owner: 1, Workout: 1, Name: "Interval", Weight: 100.0, Repetitions: 8, Next: exercise.ExerciseRef{}, Previous: exercise.ExerciseRef{}}
	e2 := exercise.Exercise{ID: 2, Owner: 1, Workout: 1, Name: "Interval", Weight: 80.0, Repetitions: 10, Next: exercise.ExerciseRef{}, Previous: exercise.ExerciseRef{}}

	e1.Next = e2.ToRef()
	e2.Previous = e1.ToRef()

	xs.WithExercise(t, e1)
	xs.WithExercise(t, e2)

	t.Run("ErrorOnNotExist", func(t *testing.T) {
		e, err := xs.WithID(3)
		if !errors.Is(err, exercise.ErrNotFound) {
			t.Errorf("expected %q but got %q", exercise.ErrNotFound, err)
		}

		if !cmp.Equal(e, exercise.Exercise{}) {
			t.Errorf("expected exercise.Exercise{} but got %v", e)
		}
	})

	t.Run("ErrorOnInvalidId", func(t *testing.T) {
		e, err := xs.WithID(-1)
		if !errors.Is(err, exercise.ErrInvalidID) {
			t.Errorf("expected %q but got %q", exercise.ErrInvalidID, err)
		}

		if !cmp.Equal(e, exercise.Exercise{}) {
			t.Errorf("expected exercise.Exercise{} but got %v", e)
		}
	})

	t.Run("ExerciseWithRefs", func(t *testing.T) {
		e, err := xs.WithID(1)
		if err != nil {
			t.Errorf("expected no error but got %q", err)
		}

		if !cmp.Equal(e, e1) {
			t.Errorf("expected %v but got %v", e1, e)
		}
	})
}

func TestStoringNewExercise(t *testing.T) {
	t.Run("ErrorOnEmpty", func(t *testing.T) {
		xs := internal.NewMockSqliteDatastore(t)
		xs.WithUser(t, user.User{ID: 1})
		xs.WithWorkout(t, workout.Workout{ID: 1})

		var e exercise.Exercise
		_, err := xs.Exercises.New(e.Owner, e.Workout, e.Name, e.Weight, e.Repetitions)
		if !errors.Is(err, exercise.ErrMissingFields) {
			t.Errorf("expected %q but got %q", exercise.ErrMissingFields, err)
		}
	})

	t.Run("ErrorOnMissingFields", func(t *testing.T) {
		xs := internal.NewMockSqliteDatastore(t)
		xs.WithUser(t, user.User{ID: 1})
		xs.WithWorkout(t, workout.Workout{ID: 1})

		e := exercise.Exercise{Owner: 1, Workout: 1}

		_, err := xs.Exercises.New(e.Owner, e.Workout, e.Name, e.Weight, e.Repetitions)
		if !errors.Is(err, exercise.ErrMissingFields) {
			t.Errorf("expected %q but got %q", exercise.ErrMissingFields, err)
		}
	})

	t.Run("InsertWithInvalidWorkout", func(t *testing.T) {
		xs := internal.NewMockSqliteDatastore(t)
		xs.WithUser(t, user.User{ID: 1})
		xs.WithWorkout(t, workout.Workout{ID: 1})

		e := exercise.Exercise{
			Owner:       1,
			Workout:     2,
			Name:        "Interval",
			Weight:      80.0,
			Repetitions: 10,
		}

		id, err := xs.Exercises.New(e.Owner, e.Workout, e.Name, e.Weight, e.Repetitions)
		if err == nil {
			t.Errorf("expected error but got %q", err)
		}

		if id != 0 {
			t.Errorf("expected id to be 0 but got %d", id)
		}
	})

	t.Run("InsertFirstExercise", func(t *testing.T) {
		xs := internal.NewMockSqliteDatastore(t)
		xs.WithUser(t, user.User{ID: 1})
		xs.WithWorkout(t, workout.Workout{ID: 1})

		e := exercise.Exercise{
			Owner:       1,
			Workout:     1,
			Name:        "Interval",
			Weight:      80.0,
			Repetitions: 10,
		}

		id, err := xs.Exercises.New(e.Owner, e.Workout, e.Name, e.Weight, e.Repetitions)
		if err != nil {
			t.Errorf("expected no error but got %q", err)
		}

		if id <= 0 {
			t.Errorf("expected new id but got %d", id)
		}
	})

	t.Run("InsertAndUpdateRefs", func(t *testing.T) {
		xs := internal.NewMockSqliteDatastore(t)
		xs.WithUser(t, user.User{ID: 1})
		xs.WithWorkout(t, workout.Workout{ID: 1})

		e1 := exercise.Exercise{ID: 1, Owner: 1, Workout: 1, Name: "Interval", Weight: 100.0, Repetitions: 8, Next: exercise.ExerciseRef{}, Previous: exercise.ExerciseRef{}}
		e2 := exercise.Exercise{ID: 2, Owner: 1, Workout: 1, Name: "Interval", Weight: 80.0, Repetitions: 10, Next: exercise.ExerciseRef{}, Previous: exercise.ExerciseRef{}}

		e1.Next = e2.ToRef()
		e2.Previous = e1.ToRef()

		xs.WithExercise(t, e1)
		xs.WithExercise(t, e2)

		e := exercise.Exercise{
			Owner:       1,
			Workout:     1,
			Name:        "Interval",
			Weight:      80.0,
			Repetitions: 10,
		}

		id, err := xs.Exercises.New(e.Owner, e.Workout, e.Name, e.Weight, e.Repetitions)
		if err != nil {
			t.Errorf("expected no error but got %q", err)
		}

		if id == 0 {
			t.Errorf("expected new id but got %d", id)
		}

		got, err := xs.WithID(id)
		if err != nil {
			t.Errorf("expected no error but got %q", err)
		}

		if got.ID != id {
			t.Errorf("id of new exercise should be %d but got %d", id, got.ID)
		}

		if got.Previous.ID != e2.ID {
			t.Errorf("previous refers to %d but expected %d", got.Previous.ID, e2.ID)
		}

		if got.Next.ID != 0 {
			t.Errorf("next refers to %d but expected 0", got.Next.ID)
		}

		e, err = xs.WithID(e2.ID)
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}

		if e.Next.ID != id {
			t.Errorf("next of n-1 refers to %d but expected %d", e.Next.ID, id)
		}

		if e.Previous.ID != 1 {
			t.Errorf("previous of n-1 refers to %d but expected 1", e.Previous.ID)
		}
	})
}

func TestUpdatingExercise(t *testing.T) {
	t.Run("ChangeName", func(t *testing.T) {
		t.Error("todo")
	})
	t.Run("UpdateWeight", func(t *testing.T) {
		t.Error("todo")
	})
	t.Run("UpdateRepetitions", func(t *testing.T) {
		t.Error("todo")
	})
}

func TestDeletingExercise(t *testing.T) {
	xs := internal.NewMockSqliteDatastore(t)
	xs.WithUser(t, user.User{ID: 1})
	xs.WithWorkout(t, workout.Workout{ID: 1})
	xs.WithExercise(t, exercise.Exercise{ID: 1, Owner: 1, Workout: 1})

	cs := []struct {
		name           string
		input          int
		expectedOutput bool
		expectedErr    error
	}{
		{"ErrorOnInvalidId", -1, false, exercise.ErrInvalidID},
		{"ErrNoID", 0, false, exercise.ErrInvalidID},
		{"ErrNotFound", 2, false, exercise.ErrNotFound},
		{"OnExist", 1, false, nil},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			if err := xs.Delete(c.input); !errors.Is(err, c.expectedErr) {
				t.Errorf("expected %q but got %q", c.expectedErr, err)
			}

			var found bool
			err := xs.QueryRow("SELECT 1 FROM exercises WHERE exercise_id=$1", c.input).Scan(&found)
			if !errors.Is(err, sql.ErrNoRows) {
				t.Errorf("unexpected error got %q", err)
			}

			if found != c.expectedOutput {
				t.Errorf("expected isFound to be %t but got %t", c.expectedOutput, found)
			}
		})
	}
}

func TestSwapExercises(t *testing.T) {
	// Only swap if i and j belong to same workout
	t.Error("Todo")
}
