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

func TestStoreExercise(t *testing.T) {
	t.Run("ErrorOnEmpty", func(t *testing.T) {
		store := newMockStore(t)
		insertUser(t, User{ID: 1}, store)
		insertWorkout(t, Workout{ID: 1}, store)

		var empty Exercise

		if _, err := store.StoreExercise(empty); !errors.Is(err, ErrMissingFields) {
			t.Errorf("expected %q but got %q", ErrMissingFields, err)
		}
	})

	t.Run("ErrorOnMissingFields", func(t *testing.T) {
		store := newMockStore(t)
		insertUser(t, User{ID: 1}, store)
		insertWorkout(t, Workout{ID: 1}, store)

		missing := Exercise{Owner: 1, Workout: 1}

		if _, err := store.StoreExercise(missing); !errors.Is(err, ErrMissingFields) {
			t.Errorf("expected %q but got %q", ErrMissingFields, err)
		}
	})

	t.Run("InsertAndUpdateRefs", func(t *testing.T) {
		store := newMockStore(t)
		insertUser(t, User{ID: 1}, store)
		insertWorkout(t, Workout{ID: 1}, store)

		e1 := Exercise{1, 1, 1, "Interval", 100.0, 8, ExerciseRef{}, ExerciseRef{}}
		e2 := Exercise{2, 1, 1, "Interval", 80.0, 10, ExerciseRef{}, ExerciseRef{}}

		e1.Next = e2.Ref()
		e2.Previous = e1.Ref()

		insertExercise(t, e1, store)
		insertExercise(t, e2, store)

		newExercise := Exercise{
			Owner:       1,
			Workout:     1,
			Name:        "Interval",
			Weight:      80.0,
			Repetitions: 10,
		}

		newID, err := store.StoreExercise(newExercise)
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

func TestExerciseById(t *testing.T) {
	store := newMockStore(t)

	insertUser(t, User{ID: 1}, store)
	insertWorkout(t, Workout{ID: 1}, store)

	e1 := Exercise{1, 1, 1, "Interval", 100.0, 8, ExerciseRef{}, ExerciseRef{}}
	e2 := Exercise{2, 1, 1, "Interval", 80.0, 10, ExerciseRef{}, ExerciseRef{}}

	e1.Next = e2.Ref()
	e2.Previous = e1.Ref()

	insertExercise(t, e1, store)
	insertExercise(t, e2, store)

	t.Run("ErrorOnNotExist", func(t *testing.T) {
		e, err := store.ExerciseByID(3)
		if !errors.Is(err, ErrExerciseNotFound) {
			t.Errorf("expected %q but got %q", ErrExerciseNotFound, err)
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

func TestExerciseExists(t *testing.T) {
	store := newMockStore(t)

	insertUser(t, User{ID: 1}, store)
	insertExercise(t, Exercise{ID: 1}, store)

	t.Run("TrueIfExists", func(t *testing.T) {
		if !store.ExerciseExists(1) {
			t.Errorf("expected true for existing exercise but got false")
		}
	})

	t.Run("FalseIfNotExists", func(t *testing.T) {
		if store.ExerciseExists(2) {
			t.Errorf("expected false for non existing exercise but got true")
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

func TestNewSqliteDatastore(t *testing.T) {
	log := slog.Default()
	dbURL := fmt.Sprintf("file://%s/%s", t.TempDir(), "test.db")
	config := SqliteDatastoreConfig{DatabaseURL: dbURL, Overwrite: true, Logger: log}

	t.Run("CreateNewIfNotExisting", func(t *testing.T) {
		_, err := NewSqliteDatastore(config)
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}
	})

	t.Run("DeleteExistingIfOverwrite", func(t *testing.T) {
		db, err := NewSqliteDatastore(config)
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}

		db.Exec("DROP TABLE IF EXISTS exercises;")

		config := SqliteDatastoreConfig{DatabaseURL: dbURL, Overwrite: true, Logger: log}
		db, err = NewSqliteDatastore(config)
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}

		wantTables := []string{"exercises", "users", "workouts"}
		if gotTables, eq := tablesEqual(t, db.DB, wantTables); !eq {
			t.Errorf("expected tables %s but got %s", wantTables, gotTables)
		}
	})

	t.Run("OpenExistingIfNotOverwrite", func(t *testing.T) {
		config := SqliteDatastoreConfig{DatabaseURL: dbURL, Overwrite: true, Logger: log}
		db, err := NewSqliteDatastore(config)
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}

		db.Exec("DROP TABLE IF EXISTS exercises;")

		config = SqliteDatastoreConfig{DatabaseURL: dbURL, Overwrite: false, Logger: log}
		db, err = NewSqliteDatastore(config)
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}

		wantTables := []string{"users", "workouts"}
		if gotTables, eq := tablesEqual(t, db.DB, wantTables); !eq {
			t.Errorf("expected tables %s but got %s", wantTables, gotTables)
		}
	})
}

func insertUser(t *testing.T, u User, store *SqliteDatastore) {
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

	if _, err := store.Exec(q, args...); err != nil {
		t.Fatalf("expected no error but got %s", err)
	}
}

func insertWorkout(t *testing.T, w Workout, store *SqliteDatastore) {
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

	if _, err := store.Exec(q, args...); err != nil {
		t.Fatalf("expected no error but got %q", err)
	}
}

func insertExercise(t *testing.T, e Exercise, store *SqliteDatastore) {
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

	if _, err := store.Exec(q, args...); err != nil {
		t.Fatalf("expected no error but got %q", err)
	}
}

func newMockStore(t *testing.T) *SqliteDatastore {
	t.Helper()

	log := slog.Default()
	dbURL := fmt.Sprintf("file://%s/%s", t.TempDir(), "test.db")
	config := SqliteDatastoreConfig{DatabaseURL: dbURL, Overwrite: false, Logger: log}

	store, err := NewSqliteDatastore(config)
	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	return store
}

func tablesEqual(t *testing.T, db *sql.DB, wantTables []string) ([]string, bool) {
	t.Helper()

	stmt := `
  SELECT 
    name
  FROM 
    sqlite_schema
  WHERE 
    type = 'table' AND 
    name != 'schema_migrations' AND
    name NOT LIKE 'sqlite_%';
  `

	rows, err := db.Query(stmt)
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
