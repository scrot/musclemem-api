package exercise

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/VauntDev/tqla"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/scrot/musclemem-api/internal/user"
)

func TestExercisesWithID(t *testing.T) {
	xs := newMockExercises(t)
	xs.withUser(t, user.User{ID: 1})
	xs.withWorkout(t, Workout{ID: 1})

	e1 := Exercise{ID: 1, Owner: 1, Workout: 1, Name: "Interval", Weight: 100.0, Repetitions: 8, Next: ExerciseRef{}, Previous: ExerciseRef{}}
	e2 := Exercise{ID: 2, Owner: 1, Workout: 1, Name: "Interval", Weight: 80.0, Repetitions: 10, Next: ExerciseRef{}, Previous: ExerciseRef{}}

	e1.Next = e2.ToRef()
	e2.Previous = e1.ToRef()

	xs.withExercise(t, e1)
	xs.withExercise(t, e2)

	t.Run("ErrorOnNotExist", func(t *testing.T) {
		e, err := xs.WithID(3)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected %q but got %q", ErrNotFound, err)
		}

		if !cmp.Equal(e, Exercise{}) {
			t.Errorf("expected Exercise{} but got %v", e)
		}
	})

	t.Run("ErrorOnInvalidId", func(t *testing.T) {
		e, err := xs.WithID(-1)
		if !errors.Is(err, ErrNoID) {
			t.Errorf("expected %q but got %q", ErrNoID, err)
		}

		if !cmp.Equal(e, Exercise{}) {
			t.Errorf("expected Exercise{} but got %v", e)
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

func TestExercisesFromWorkout(t *testing.T) {
	xs := newMockExercises(t)
	xs.withUser(t, user.User{ID: 1})
	xs.withWorkout(t, Workout{ID: 1})

	e1 := Exercise{ID: 1, Owner: 1, Workout: 1, Name: "Interval", Weight: 100.0, Repetitions: 8, Next: ExerciseRef{}, Previous: ExerciseRef{}}
	e2 := Exercise{ID: 2, Owner: 1, Workout: 1, Name: "Interval", Weight: 80.0, Repetitions: 10, Next: ExerciseRef{}, Previous: ExerciseRef{}}
	e3 := Exercise{ID: 3, Owner: 1, Workout: 1, Name: "Interval", Weight: 80.0, Repetitions: 10, Next: ExerciseRef{}, Previous: ExerciseRef{}}

	e1.Next = e2.ToRef()
	e2.Previous = e1.ToRef()
	e2.Next = e3.ToRef()
	e3.Previous = e2.ToRef()

	xs.withExercise(t, e1)
	xs.withExercise(t, e2)
	xs.withExercise(t, e3)

	t.Run("ErrorInvalidID", func(t *testing.T) {
		xs, err := xs.FromWorkout(0, 0)
		if err == nil {
			t.Errorf("expected %q but got nil", ErrNoID)
		}
		if len(xs) != 0 {
			t.Errorf("expected empty list but got %v", xs)
		}
	})

	t.Run("ErrorNotFound", func(t *testing.T) {
		xs, err := xs.FromWorkout(1, 2)
		if err == nil {
			t.Errorf("expected %q but got nil", ErrNotFound)
		}
		if len(xs) != 0 {
			t.Errorf("expected empty list but got %v", xs)
		}
	})

	t.Run("ListOnValidWorkout", func(t *testing.T) {
		xs, err := xs.FromWorkout(1, 1)
		if err != nil {
			t.Errorf("expected no error but got %q", err)
		}
		if len(xs) != 3 {
			t.Errorf("expected %d exercises but got %d", 3, len(xs))
		}

		if !cmp.Equal(xs, []Exercise{e1, e2, e3}) {
			t.Errorf("expected %v but got %v", []Exercise{e1, e2, e3}, xs)
		}
	})
}

func TestStore(t *testing.T) {
	t.Run("ErrorOnEmpty", func(t *testing.T) {
		xs := newMockExercises(t)
		xs.withUser(t, user.User{ID: 1})
		xs.withWorkout(t, Workout{ID: 1})

		var empty Exercise

		if _, err := xs.Store(empty); !errors.Is(err, ErrMissingFields) {
			t.Errorf("expected %q but got %q", ErrMissingFields, err)
		}
	})

	t.Run("ErrorOnMissingFields", func(t *testing.T) {
		xs := newMockExercises(t)
		xs.withUser(t, user.User{ID: 1})
		xs.withWorkout(t, Workout{ID: 1})

		missing := Exercise{Owner: 1, Workout: 1}

		if _, err := xs.Store(missing); !errors.Is(err, ErrMissingFields) {
			t.Errorf("expected %q but got %q", ErrMissingFields, err)
		}
	})

	t.Run("InsertWithInvalidWorkout", func(t *testing.T) {
		xs := newMockExercises(t)
		xs.withUser(t, user.User{ID: 1})
		xs.withWorkout(t, Workout{ID: 1})

		newExercise := Exercise{
			Owner:       1,
			Workout:     2,
			Name:        "Interval",
			Weight:      80.0,
			Repetitions: 10,
		}

		if _, err := xs.Store(newExercise); err == nil {
			t.Errorf("expected error but got %q", err)
		}
	})

	t.Run("InsertFirstExercise", func(t *testing.T) {
		xs := newMockExercises(t)
		xs.withUser(t, user.User{ID: 1})
		xs.withWorkout(t, Workout{ID: 1})

		newExercise := Exercise{
			Owner:       1,
			Workout:     1,
			Name:        "Interval",
			Weight:      80.0,
			Repetitions: 10,
		}

		id, err := xs.Store(newExercise)
		if err != nil {
			t.Errorf("expected no error but got %q", err)
		}

		if id <= 0 {
			t.Errorf("expected new id but got %d", id)
		}
	})

	t.Run("InsertAndUpdateRefs", func(t *testing.T) {
		xs := newMockExercises(t)
		xs.withUser(t, user.User{ID: 1})
		xs.withWorkout(t, Workout{ID: 1})

		e1 := Exercise{ID: 1, Owner: 1, Workout: 1, Name: "Interval", Weight: 100.0, Repetitions: 8, Next: ExerciseRef{}, Previous: ExerciseRef{}}
		e2 := Exercise{ID: 2, Owner: 1, Workout: 1, Name: "Interval", Weight: 80.0, Repetitions: 10, Next: ExerciseRef{}, Previous: ExerciseRef{}}

		e1.Next = e2.ToRef()
		e2.Previous = e1.ToRef()

		xs.withExercise(t, e1)
		xs.withExercise(t, e2)

		newExercise := Exercise{
			Owner:       1,
			Workout:     1,
			Name:        "Interval",
			Weight:      80.0,
			Repetitions: 10,
		}

		newID, err := xs.Store(newExercise)
		if err != nil {
			t.Errorf("expected no error but got %q", err)
		}

		if newID == 0 {
			t.Errorf("expected new id but got %d", newID)
		}

		e, err := xs.WithID(newID)
		if err != nil {
			t.Errorf("expected no error but got %q", err)
		}

		if e.ID != newID {
			t.Errorf("id of new exercise should be %d but got %d", newID, e.ID)
		}

		if e.Previous.ID != e2.ID {
			t.Errorf("previous refers to %d but expected %d", e.Previous.ID, e2.ID)
		}

		if e.Next.ID != 0 {
			t.Errorf("next refers to %d but expected 0", e.Next.ID)
		}

		e, err = xs.WithID(e2.ID)
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}

		if e.Next.ID != newID {
			t.Errorf("next of n-1 refers to %d but expected %d", e.Next.ID, newID)
		}

		if e.Previous.ID != 1 {
			t.Errorf("previous of n-1 refers to %d but expected 1", e.Previous.ID)
		}
	})
}

func TestDelete(t *testing.T) {
	xs := newMockExercises(t)
	xs.withUser(t, user.User{ID: 1})
	xs.withWorkout(t, Workout{ID: 1})
	xs.withExercise(t, Exercise{ID: 1})

	cs := []struct {
		name           string
		input          int
		expectedOutput bool
		expectedErr    error
	}{
		{"ErrorOnInvalidId", -1, false, ErrNoID},
		{"ErrNoID", 0, false, ErrNoID},
		{"ErrNotFound", 2, false, ErrNotFound},
		{"OnExist", 1, false, nil},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			if err := xs.Delete(c.input); err != c.expectedErr {
				t.Errorf("expected no error but got %q", err)
			}

			var found bool
			err := xs.db.QueryRow("SELECT 1 FROM exercises WHERE exercise_id=$1", c.input).Scan(&found)
			if !errors.Is(err, sql.ErrNoRows) {
				t.Errorf("unexpected error got %q", err)
			}

			if found != c.expectedOutput {
				t.Errorf("expected isFound to be %t but got %t", c.expectedOutput, found)
			}
		})
	}
}

type MockDatastore struct {
	Exercises
	db *sql.DB
}

func newMockExercises(t *testing.T) *MockDatastore {
	t.Helper()

	dbURL := fmt.Sprintf("file://%s/%s", t.TempDir(), "test.db")
	mURL := "file://../../migrations"
	config := SqliteDatastoreConfig{
		DatabaseURL:        dbURL,
		MigrationURL:       mURL,
		Overwrite:          true,
		ForeignKeyEnforced: true,
	}

	store, err := NewSqliteDatastore(config)
	if err != nil {
		t.Errorf("expected no error but got %q", err)
	}

	db, err := sql.Open("sqlite3", dbURL)
	if err != nil {
		t.Errorf("expected no error but got %q", err)
	}

	return &MockDatastore{store, db}
}

func (ds *MockDatastore) withUser(t *testing.T, u user.User) {
	t.Helper()

	const stmt = `
    INSERT INTO users (user_id, email, password)
    VALUES ({{.ID}}, {{.Email}}, {{.Password}})
    `

	tmpl, _ := tqla.New()
	q, args, err := tmpl.Compile(stmt, u)
	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	if _, err := ds.db.Exec(q, args...); err != nil {
		t.Fatalf("expected no error but got %s", err)
	}
}

func (ds *MockDatastore) withWorkout(t *testing.T, w Workout) {
	t.Helper()

	const stmt = `
    INSERT INTO workouts (workout_id, name)
    VALUES({{.ID}}, {{.Name}})
    `

	tmpl, _ := tqla.New()
	q, args, err := tmpl.Compile(stmt, w)
	if err != nil {
		t.Fatalf("expected no error but got %q", err)
	}

	if _, err := ds.db.Exec(q, args...); err != nil {
		t.Fatalf("expected no error but got %q", err)
	}
}

func (ds *MockDatastore) withExercise(t *testing.T, e Exercise) {
	t.Helper()

	const stmt = `
    INSERT INTO exercises (exercise_id, owner, workout, name, weight, repetitions, previous, next)
    VALUES (
    {{.ID}},
    {{.Owner}},
    {{.Workout}},
    {{.Name}},
    {{.Weight}},
    {{.Repetitions}},
    {{.Previous.ID}},
    {{.Next.ID}}    
    )
  `

	tmpl, _ := tqla.New()
	q, args, err := tmpl.Compile(stmt, e)
	if err != nil {
		t.Fatalf("expected no error but got %q", err)
	}

	if _, err := ds.db.Exec(q, args...); err != nil {
		t.Fatalf("expected no error but got %q", err)
	}
}

func (ds *MockDatastore) tablesEqual(t *testing.T, wantTables []string) ([]string, bool) {
	t.Helper()

	const stmt = `
  SELECT 
    name
  FROM 
    sqlite_schema
  WHERE 
    type = 'table' AND 
    name != 'schema_migrations' AND
    name NOT LIKE 'sqlite_%';
  `

	rows, err := ds.db.Query(stmt)
	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	var tables []string
	for rows.Next() {
		var table string
		rows.Scan(&table)
		tables = append(tables, table)
	}

	less := func(a, b string) bool { return a < b }
	if cmp.Equal(tables, wantTables, cmpopts.SortSlices(less)) {
		return tables, true
	}

	return tables, false
}
