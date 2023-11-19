package musclememapi

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
	store := newStore(t)

	insertUser(t, User{ID: 1}, store)

	e1 := Exercise{1, 1, "Benchpress", "Interval", 100.0, 8, nil, nil}
	e2 := Exercise{2, 1, "Shoulder Press", "Interval", 80.0, 10, nil, nil}

	e1.Next = e2.Ref()
	e2.Previous = e1.Ref()

	insertExercise(t, e1, store)
	insertExercise(t, e2, store)

	t.Run("ErrorOnNotExist", func(t *testing.T) {
		e, err := store.ExerciseByID(2)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected %q but got %v", ErrNotFound, err)
		}

		if !cmp.Equal(e, Exercise{}) {
			t.Errorf("expected Exercise{} but got %v", e)
		}
	})

	t.Run("ErrorOnInvalidId", func(t *testing.T) {
		e, err := store.ExerciseByID(-1)
		t.Log(err)
		if err == nil {
			t.Log(err)
			t.Errorf("expected %q but got %v", ErrInvalidID, err)
		}

		if !cmp.Equal(e, Exercise{}) {
			t.Errorf("expected Exercise{} but got %v", e)
		}
	})

}

func TestExerciseExists(t *testing.T) {
	store := newStore(t)

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
			t.Fatalf("expected no error but got %s", err)
		}

		if err := db.Ping(); err != nil {
			t.Fatalf("expected no error but got %s", err)
		}
	})

	t.Run("ErrOnEmptyPath", func(t *testing.T) {
		path := ""
		_, err := newLocalDatabase(path)
		if err == nil {
			t.Fatalf("expected error but got nil")
		}
	})

	t.Run("ErrOnInvalidPath", func(t *testing.T) {
		path := "///test.db"
		_, err := newLocalDatabase(path)
		if err == nil {
			t.Fatalf("expected error but got nil")
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

	t.Run("CreateNewIfNotExisting", func(t *testing.T) {
		_, err := NewSqliteDatastore(SqliteDatastoreConfig{dbURL, true, log})
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}
	})

	t.Run("DeleteExistingIfOverwrite", func(t *testing.T) {
		db, err := NewSqliteDatastore(SqliteDatastoreConfig{dbURL, true, log})
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}

		db.Exec("DROP TABLE IF EXISTS exercises;")

		db, err = NewSqliteDatastore(SqliteDatastoreConfig{dbURL, true, log})
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}

		wantTables := []string{"exercises", "users"}
		if gotTables, eq := tablesEqual(t, db.DB, wantTables); !eq {
			t.Errorf("expected tables %s but got %s", wantTables, gotTables)
		}
	})

	t.Run("OpenExistingIfNotOverwrite", func(t *testing.T) {
		db, err := NewSqliteDatastore(SqliteDatastoreConfig{dbURL, true, log})
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}

		db.Exec("DROP TABLE IF EXISTS exercises;")

		db, err = NewSqliteDatastore(SqliteDatastoreConfig{dbURL, false, log})
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}

		wantTables := []string{"users"}
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
    {{if .Previous}}{{.Previous.ID}}{{else}}0{{end}},
    {{if .Next}}{{.Next.ID}}{{else}}0{{end}}
    )
  `

	tmpl, _ := tqla.New()
	q, args, err := tmpl.Compile(stmt, e)
	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	if _, err := store.Exec(q, args...); err != nil {
		t.Fatalf("expected no error but got %s", err)
	}
}

func newStore(t *testing.T) *SqliteDatastore {
	t.Helper()

	path := fmt.Sprintf("file://%s/%s", t.TempDir(), "test.db")
	store, err := NewSqliteDatastore(SqliteDatastoreConfig{path, true, slog.Default()})
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
