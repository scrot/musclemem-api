package exercise

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/VauntDev/tqla"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestExerciseById(t *testing.T) {
	store := newMockDatastore(t)
	store.withUser(t, User{ID: 1})
	store.withWorkout(t, Workout{ID: 1})

	e1 := Exercise{1, 1, 1, "Interval", 100.0, 8, ExerciseRef{}, ExerciseRef{}}
	e2 := Exercise{2, 1, 1, "Interval", 80.0, 10, ExerciseRef{}, ExerciseRef{}}

	e1.Next = e2.Ref()
	e2.Previous = e1.Ref()

	store.withExercise(t, e1)
	store.withExercise(t, e2)

	t.Run("ErrorOnNotExist", func(t *testing.T) {
		e, err := store.ExerciseByID(3)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected %q but got %q", ErrNotFound, err)
		}

		if !cmp.Equal(e, Exercise{}) {
			t.Errorf("expected Exercise{} but got %v", e)
		}
	})

	t.Run("ErrorOnInvalidId", func(t *testing.T) {
		e, err := store.ExerciseByID(-1)
		if !errors.Is(err, ErrInvalidIdFormat) {
			t.Errorf("expected %q but got %q", ErrInvalidIdFormat, err)
		}

		if !cmp.Equal(e, Exercise{}) {
			t.Errorf("expected Exercise{} but got %v", e)
		}
	})

	t.Run("ExerciseWithRefs", func(t *testing.T) {
		e, err := store.ExerciseByID(1)
		if err != nil {
			t.Errorf("expected no error but got %q", err)
		}

		if !cmp.Equal(e, e1) {
			t.Errorf("expected %v but got %v", e1, e)
		}
	})
}

func TestExercisesByWorkout(t *testing.T) {
	store := newMockDatastore(t)
	store.withUser(t, User{ID: 1})
	store.withWorkout(t, Workout{ID: 1})

	e1 := Exercise{1, 1, 1, "Interval", 100.0, 8, ExerciseRef{}, ExerciseRef{}}
	e2 := Exercise{2, 1, 1, "Interval", 80.0, 10, ExerciseRef{}, ExerciseRef{}}
	e3 := Exercise{3, 1, 1, "Interval", 80.0, 10, ExerciseRef{}, ExerciseRef{}}

	e1.Next = e2.Ref()
	e2.Previous = e1.Ref()
	e2.Next = e3.Ref()
	e3.Previous = e2.Ref()

	store.withExercise(t, e1)
	store.withExercise(t, e2)
	store.withExercise(t, e3)

	t.Run("ErrorInvalidID", func(t *testing.T) {
		xs, err := store.ExercisesByWorkoutID(0, 0)
		if err == nil {
			t.Errorf("expected %q but got nil", ErrNoID)
		}
		if len(xs) != 0 {
			t.Errorf("expected empty list but got %v", xs)
		}
	})

	t.Run("ErrorNotFound", func(t *testing.T) {
		xs, err := store.ExercisesByWorkoutID(1, 2)
		if err == nil {
			t.Errorf("expected %q but got nil", ErrNotFound)
		}
		if len(xs) != 0 {
			t.Errorf("expected empty list but got %v", xs)
		}
	})

	t.Run("ListOnValidWorkout", func(t *testing.T) {
		xs, err := store.ExercisesByWorkoutID(1, 1)
		if err != nil {
			t.Errorf("expected no error but got %q", err)
		}
		if len(xs) != 3 {
			t.Errorf("expected %d exercises but got %d", 3, len(xs))
		}
	})
}

func TestCreateLocalDatabase(t *testing.T) {
	t.Run("PingableOnValidPath", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.db")
		db, err := newLocalDatabase(path)
		if err != nil {
			t.Fatalf("expected no error but got %q", err)
		}

		if err := db.Ping(); err != nil {
			t.Fatalf("expected no error but got %q", err)
		}
	})

	t.Run("ErrOnEmptyPath", func(t *testing.T) {
		path := ""
		_, err := newLocalDatabase(path)
		if err == nil {
			t.Fatalf("expected error but got none")
		}
	})

	t.Run("ErrOnInvalidPath", func(t *testing.T) {
		path := "///test.db"
		_, err := newLocalDatabase(path)
		if err == nil {
			t.Fatalf("expected error but got none")
		}
	})

	t.Run("ErrOnURL", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.db")
		url := "sqlite3://" + path
		_, err := newLocalDatabase(url)
		if err == nil {
			t.Fatalf("expected error but got nil")
		}
	})
}

func TestStore(t *testing.T) {
	t.Run("ErrorOnEmpty", func(t *testing.T) {
		store := newMockDatastore(t)
		store.withUser(t, User{ID: 1})
		store.withWorkout(t, Workout{ID: 1})

		var empty Exercise

		if _, err := store.Store(empty); !errors.Is(err, ErrMissingFields) {
			t.Errorf("expected %q but got %q", ErrMissingFields, err)
		}
	})

	t.Run("ErrorOnMissingFields", func(t *testing.T) {
		store := newMockDatastore(t)
		store.withUser(t, User{ID: 1})
		store.withWorkout(t, Workout{ID: 1})

		missing := Exercise{Owner: 1, Workout: 1}

		if _, err := store.Store(missing); !errors.Is(err, ErrMissingFields) {
			t.Errorf("expected %q but got %q", ErrMissingFields, err)
		}
	})

	t.Run("InsertAndUpdateRefs", func(t *testing.T) {
		store := newMockDatastore(t)
		store.withUser(t, User{ID: 1})
		store.withWorkout(t, Workout{ID: 1})

		e1 := Exercise{1, 1, 1, "Interval", 100.0, 8, ExerciseRef{}, ExerciseRef{}}
		e2 := Exercise{2, 1, 1, "Interval", 80.0, 10, ExerciseRef{}, ExerciseRef{}}

		e1.Next = e2.Ref()
		e2.Previous = e1.Ref()

		store.withExercise(t, e1)
		store.withExercise(t, e2)

		newExercise := Exercise{
			Owner:       1,
			Workout:     1,
			Name:        "Interval",
			Weight:      80.0,
			Repetitions: 10,
		}

		newID, err := store.Store(newExercise)
		if err != nil {
			t.Errorf("expected no error but got %q", err)
		}

		if newID == 0 {
			t.Errorf("expected new id but got %d", newID)
		}

		e, err := store.ExerciseByID(newID)
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

		e, err = store.ExerciseByID(e2.ID)
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

type MockDatastore struct {
	Datastore
	db *sql.DB
}

func newMockDatastore(t *testing.T) *MockDatastore {
	t.Helper()

	dbURL := fmt.Sprintf("file://%s/%s", t.TempDir(), "test.db")
	mURL := "file://../migrations"
	config := SqliteDatastoreConfig{slog.Default(), dbURL, mURL, true}

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

func (ds *MockDatastore) withUser(t *testing.T, u User) {
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
