package internal

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/VauntDev/tqla"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	_ "github.com/mattn/go-sqlite3"

	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/scrot/musclemem-api/internal/user"
	"github.com/scrot/musclemem-api/internal/workout"
)

// Configuration for SqliteDatastore
type SqliteDatastoreConfig struct {
	DatabaseURL        string
	MigrationURL       string
	Overwrite          bool
	ForeignKeyEnforced bool
}

var DefaultSqliteConfig = SqliteDatastoreConfig{
	DatabaseURL:        "file://./musclemem.db",
	MigrationURL:       "file://./migrations/",
	Overwrite:          false,
	ForeignKeyEnforced: true,
}

// NewSqliteDatastore creates a new database at dbURL
// and runs the migrations in the defaultMigrations folder
// if overwrite is false, it returns the existing db
func NewSqliteDatastore(config SqliteDatastoreConfig) (*sql.DB, error) {
	path, hasFileScheme := strings.CutPrefix(config.DatabaseURL, "file://")
	if !hasFileScheme {
		err := errors.New("invalid scheme expected file://")
		return nil, err
	}

	if config.Overwrite {
		os.Remove(path)
	}

	dbDNS := fmt.Sprintf("file:%s?_foreign_keys=%t", path, config.ForeignKeyEnforced)
	db, err := sql.Open("sqlite3", dbDNS)
	if err != nil {
		return nil, err
	}

	drv, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(config.MigrationURL, config.DatabaseURL, drv)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return db, nil
		}
		return nil, err
	}

	return db, nil
}

///
/// Mock repository
///

type MockSqliteDatastore struct {
	Workouts  workout.Workouts
	Exercises exercise.Exercises
	Users     user.Users
	DB        *sql.DB
}

func NewMockSqliteDatastore(t *testing.T) *MockSqliteDatastore {
	t.Helper()

	dbURL := fmt.Sprintf("file://%s/%s", t.TempDir(), "test.db")
	mURL := "file://../../migrations"
	config := SqliteDatastoreConfig{
		DatabaseURL:        dbURL,
		MigrationURL:       mURL,
		Overwrite:          true,
		ForeignKeyEnforced: true,
	}

	db, err := NewSqliteDatastore(config)
	if err != nil {
		t.Errorf("expected no error but got %q", err)
	}

	xs := exercise.NewSQLExercises(db)
	ws := workout.NewSQLWorkouts(db)
	us := user.NewSQLUsers(db)

	return &MockSqliteDatastore{ws, xs, us, db}
}

func (ds *MockSqliteDatastore) WithUser(t *testing.T, u user.User) {
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

	if _, err := ds.DB.Exec(q, args...); err != nil {
		t.Fatalf("expected no error but got %s", err)
	}
}

func (ds *MockSqliteDatastore) WithWorkout(t *testing.T, w workout.Workout) {
	t.Helper()

	const stmt = `
    INSERT INTO workouts (workout_id, owner, name)
    VALUES({{.ID}}, {{.Owner}}, {{.Name}})
    `

	tmpl, _ := tqla.New()
	q, args, err := tmpl.Compile(stmt, w)
	if err != nil {
		t.Fatalf("expected no error but got %q", err)
	}

	if _, err := ds.DB.Exec(q, args...); err != nil {
		t.Fatalf("expected no error but got %q", err)
	}
}

func (ds *MockSqliteDatastore) WithExercise(t *testing.T, e exercise.Exercise) {
	t.Helper()

	const stmt = `
    INSERT INTO exercises (exercise_id, workout, name, weight, repetitions, previous, next)
    VALUES (
    {{.ID}},
    {{.Workout}},
    {{.Name}},
    {{.Weight}},
    {{.Repetitions}},
    {{.PreviousID}},
    {{.NextID}}    
    )
  `

	tmpl, _ := tqla.New()
	q, args, err := tmpl.Compile(stmt, e)
	if err != nil {
		t.Fatalf("expected no error but got %q", err)
	}

	if _, err := ds.DB.Exec(q, args...); err != nil {
		t.Fatalf("expected no error but got %q", err)
	}
}

func (ds *MockSqliteDatastore) TablesEqual(t *testing.T, wantTables []string) ([]string, bool) {
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

	rows, err := ds.DB.Query(stmt)
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
