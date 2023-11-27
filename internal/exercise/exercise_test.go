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

func TestRetreiveExerciseWithID(t *testing.T) {
	t.Parallel()

	xs := internal.NewMockSqliteDatastore(t)

	xs.WithUser(t, user.User{ID: 1})

	xs.WithWorkout(t, workout.Workout{ID: 1})

	e1 := exercise.Exercise{ID: 1, Owner: 1, Workout: 1, Name: "Interval", Weight: 100.0, Repetitions: 8, Next: exercise.ExerciseRef{}, Previous: exercise.ExerciseRef{}}
	e2 := exercise.Exercise{ID: 2, Owner: 1, Workout: 1, Name: "Interval", Weight: 80.0, Repetitions: 10, Next: exercise.ExerciseRef{}, Previous: exercise.ExerciseRef{}}

	e1.Next = e2.ToRef()
	e2.Previous = e1.ToRef()

	xs.WithExercise(t, e1)
	xs.WithExercise(t, e2)

	cs := []struct {
		name        string
		exerciseID  int
		expected    exercise.Exercise
		expectedErr error
	}{
		{"ErrorIfNegativeID", -1, exercise.Exercise{}, exercise.ErrInvalidID},
		{"ErrorIfNoID", 0, exercise.Exercise{}, exercise.ErrInvalidID},
		{"ErrorIfNotExist", 3, exercise.Exercise{}, exercise.ErrNotFound},
		{"ExistingExercise", 1, e1, nil},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			x, err := xs.WithID(c.exerciseID)
			if !errors.Is(err, c.expectedErr) {
				t.Errorf("expected error '%v', but got '%v'", c.expectedErr, err)
			}

			if !cmp.Equal(x, c.expected) {
				t.Errorf("expected %v but got %v", c.expected, x)
			}
		})
	}
}

func TestStoringNewExercise(t *testing.T) {
	t.Parallel()

	xs := internal.NewMockSqliteDatastore(t)

	xs.WithUser(t, user.User{ID: 1})

	xs.WithWorkout(t, workout.Workout{ID: 1})

	e1 := exercise.Exercise{Owner: 1, Workout: 1, Name: "First", Weight: 80.0, Repetitions: 10}
	e2 := exercise.Exercise{Owner: 1, Workout: 1, Name: "Second", Weight: 80.0, Repetitions: 10}
	e3 := exercise.Exercise{Owner: 1, Workout: 1, Name: "Third", Weight: 80.0, Repetitions: 10}

	em := exercise.Exercise{Owner: 1, Workout: 1, Name: "Missing"}

	eiw := exercise.Exercise{Owner: 1, Workout: 2, Name: "Invalid Workout", Weight: 80.0, Repetitions: 10}
	eio := exercise.Exercise{Owner: 2, Workout: 1, Name: "Invalid Owner", Weight: 80.0, Repetitions: 10}

	r0 := map[int][]exercise.ExerciseRef{}
	r1 := map[int][]exercise.ExerciseRef{}
	r2 := map[int][]exercise.ExerciseRef{
		1: {exercise.ExerciseRef{}, exercise.ExerciseRef{2, "Second"}},
		2: {exercise.ExerciseRef{1, "First"}, exercise.ExerciseRef{}},
	}
	rn := map[int][]exercise.ExerciseRef{
		1: {exercise.ExerciseRef{}, exercise.ExerciseRef{2, "Second"}},
		2: {exercise.ExerciseRef{1, "First"}, exercise.ExerciseRef{3, "Third"}},
		3: {exercise.ExerciseRef{2, "Second"}, exercise.ExerciseRef{}},
	}

	cs := []struct {
		name          string
		exercise      exercise.Exercise
		expectedNewID bool
		expectedRefs  map[int][]exercise.ExerciseRef
		expectedErr   error
	}{
		{"ErrorIfMissingField", em, false, r0, exercise.ErrMissingFields},
		{"ErrorIfInvalidWorkout", eiw, false, r0, exercise.ErrInvalidWorkout},
		{"ErrorIfInvalidOwner", eio, false, r0, exercise.ErrInvalidOwner},
		{"WorkoutWithOneExercise", e1, true, r1, nil},
		{"WorkoutWithTwoExercises", e2, true, r2, nil},
		{"WorkoutWithMoreExercises", e3, true, rn, nil},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			id, err := xs.Exercises.New(c.exercise.Owner, c.exercise.Workout,
				c.exercise.Name, c.exercise.Weight, c.exercise.Repetitions)

			if !errors.Is(err, c.expectedErr) {
				t.Errorf("expected '%v' but got '%v'", c.expectedErr, err)
			}

			hasID := id > 0
			if hasID != c.expectedNewID {
				t.Errorf("expected new id but got %d", id)
			}

			// compare all workout exercises references
			for k, v := range c.expectedRefs {
				e, err := xs.WithID(k)
				if err != nil {
					t.Errorf("expected no error but got '%v'", err)
				}

				prev := v[0]
				if e.Previous.Name != prev.Name {
					t.Errorf("exercise %d expected previous '%s' but got '%s'", k, prev.Name, e.Previous.Name)
				}

				next := v[1]
				if e.Next.Name != next.Name {
					t.Errorf("exercise %d expected next '%s' but got '%s'", k, next.Name, e.Next.Name)
				}
			}
		})
	}
}

func TestChangingExerciseName(t *testing.T) {
	t.Parallel()

	xs := internal.NewMockSqliteDatastore(t)
	xs.WithUser(t, user.User{ID: 1})
	xs.WithWorkout(t, workout.Workout{ID: 1})
	xs.WithExercise(t, exercise.Exercise{ID: 1, Owner: 1, Workout: 1})

	cs := []struct {
		name         string
		exerciseID   int
		exerciseName string
		expected     string
		expectedErr  error
	}{
		{"ErrorIfNegativeID", -1, "INVALID", "", exercise.ErrInvalidID},
		{"ErrorIfNoID", 0, "INVALID", "", exercise.ErrInvalidID},
		{"ErrorIfNoName", 1, "", "", exercise.ErrMissingFields},
		{"HasNewName", 1, "CHANGED", "CHANGED", nil},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			if err := xs.ChangeName(c.exerciseID, c.exerciseName); !errors.Is(err, c.expectedErr) {
				t.Errorf("expected error '%v' but got '%v'", c.expectedErr, err)
			}

			var name string
			xs.QueryRow("SELECT name FROM exercises WHERE exercise_id = $1", c.exerciseID).Scan(&name)
			if name != c.expected {
				t.Errorf("expected name '%s' but got '%s'", c.expected, name)
			}
		})
	}
}

func TestUpdatingWeight(t *testing.T) {
	t.Parallel()

	xs := internal.NewMockSqliteDatastore(t)
	xs.WithUser(t, user.User{ID: 1})
	xs.WithWorkout(t, workout.Workout{ID: 1})
	xs.WithExercise(t, exercise.Exercise{ID: 1, Owner: 1, Workout: 1})

	cs := []struct {
		name           string
		exerciseID     int
		exerciseWeight float64
		expected       float64
		expectedErr    error
	}{
		{"ErrorIfNegativeID", -1, 100, 0, exercise.ErrInvalidID},
		{"ErrorIfNoID", 0, 100, 0, exercise.ErrInvalidID},
		{"ErrorIfNegativeWeight", 1, -100, 0, exercise.ErrNegativeValue},
		{"HasNewWeight", 1, 100, 100, nil},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			if err := xs.UpdateWeight(c.exerciseID, c.exerciseWeight); !errors.Is(err, c.expectedErr) {
				t.Errorf("expected error '%v' but got '%v'", c.expectedErr, err)
			}

			var weight float64
			xs.QueryRow("SELECT weight FROM exercises WHERE exercise_id = $1", c.exerciseID).Scan(&weight)
			if weight != c.expected {
				t.Errorf("expected weight '%.0f' but got '%.0f'", c.expected, weight)
			}
		})
	}
}

func TestUpdateRepetitions(t *testing.T) {
	t.Parallel()

	xs := internal.NewMockSqliteDatastore(t)
	xs.WithUser(t, user.User{ID: 1})
	xs.WithWorkout(t, workout.Workout{ID: 1})
	xs.WithExercise(t, exercise.Exercise{ID: 1, Owner: 1, Workout: 1})

	cs := []struct {
		name                string
		exerciseID          int
		exerciseRepetitions int
		expected            int
		expectedErr         error
	}{
		{"ErrorIfNegativeID", -1, 100, 0, exercise.ErrInvalidID},
		{"ErrorIfNoID", 0, 100, 0, exercise.ErrInvalidID},
		{"ErrorIfNegativeRepetitions", 1, -100, 0, exercise.ErrNegativeValue},
		{"HasNewRepetitions", 1, 100, 100, nil},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			if err := xs.UpdateRepetitions(c.exerciseID, c.exerciseRepetitions); !errors.Is(err, c.expectedErr) {
				t.Errorf("expected error '%v' but got '%v'", c.expectedErr, err)
			}

			var reps int
			xs.QueryRow("SELECT repetitions FROM exercises WHERE exercise_id = $1", c.exerciseID).Scan(&reps)
			if reps != c.expected {
				t.Errorf("expected repetitions '%d' but got '%d'", c.expected, reps)
			}
		})
	}
}

func TestDeletingExercise(t *testing.T) {
	t.Parallel()

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
		{"ErrorIfNegativeID", -1, false, exercise.ErrInvalidID},
		{"ErrorIfNoID", 0, false, exercise.ErrInvalidID},
		{"ErrorIfNotExist", 2, false, exercise.ErrNotFound},
		{"DeleteOneExercise", 1, false, nil},
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
	t.Parallel()

	// Only swap if i and j belong to same workout
	t.Error("Todo")
}
