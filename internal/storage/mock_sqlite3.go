package storage

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type MockSqliteDatastore struct {
	*SqliteDatastore
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

	return &MockSqliteDatastore{db}
}

func (ds *MockSqliteDatastore) WithUser(t *testing.T, id int, email string, pass string) {
	t.Helper()

	const stmt = `
    INSERT INTO users (user_id, email, password)
    VALUES ({{.ID}}, {{.Email}}, {{.Password}})
    `

	data := struct {
		ID       int
		Email    string
		Password string
	}{id, email, pass}

	q, args, err := ds.CompileStatement(stmt, data)
	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	if _, err := ds.Exec(q, args...); err != nil {
		t.Fatalf("expected no error but got %s", err)
	}
}

func (ds *MockSqliteDatastore) WithWorkout(t *testing.T, workoutID int, owner int, name string) {
	t.Helper()

	const stmt = `
    INSERT INTO workouts (workout_id, owner, name)
    VALUES({{.ID}}, {{.Owner}}, {{.Name}})
    `

	data := struct {
		ID    int
		Owner int
		Name  string
	}{workoutID, owner, name}

	q, args, err := ds.CompileStatement(stmt, data)
	if err != nil {
		t.Fatalf("expected no error but got %q", err)
	}

	if _, err := ds.Exec(q, args...); err != nil {
		t.Fatalf("expected no error but got %q", err)
	}
}

func (ds *MockSqliteDatastore) WithExercise(t *testing.T, id int, workout int, name string, weight float64, reps int, prev int, next int) {
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

	data := struct {
		ID          int
		Workout     int
		Name        string
		Weight      float64
		Repetitions int
		PreviousID  int
		NextID      int
	}{id, workout, name, weight, reps, prev, next}

	q, args, err := ds.CompileStatement(stmt, data)
	if err != nil {
		t.Fatalf("expected no error but got %q", err)
	}

	if _, err := ds.Exec(q, args...); err != nil {
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

	rows, err := ds.Query(stmt)
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
