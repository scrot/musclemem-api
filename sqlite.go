package musclememapi

import (
	"database/sql"
	"errors"
	"os"
	"strings"

	"github.com/VauntDev/tqla"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	defaultDatabaseURL   = "file://./musclemem.db"
	defaultMigrationsURL = "file://./migrations/"
)

// Implements the Datastore interface for sqlite3
type SqliteDatastore struct {
	*sql.DB
}

// NewSqliteDatastore creates a new database at dbURL
// and runs the migrations in the defaultMigrations folder
// if overwrite is false, it returns the existing db
func NewSqliteDatastore(dbURL string, overwrite bool) (*SqliteDatastore, error) {
	if path, hasFileScheme := strings.CutPrefix(dbURL, "file://"); hasFileScheme {
		if _, err := os.Stat(path); !overwrite && err == nil {

			db, err := sql.Open("sqlite3", path)
			if err != nil {
				return nil, err
			}

			return &SqliteDatastore{db}, nil
		}

		db, err := newLocalDatabase(path)
		if err != nil {
			return nil, err
		}

		if err := migrateSchema(defaultMigrationsURL, db); err != nil {
			return nil, err
		}

		return &SqliteDatastore{db}, nil
	}

	return nil, errors.New("scheme not supported, use file://")

}

// newLocalDatabase create a new database
// if a database already exists, it will be overwritten
func newLocalDatabase(path string) (*sql.DB, error) {
	if err := os.Remove(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	if _, err := os.Create(path); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// migrateSchema populates the musclemem database
// based on the migrations provided by the migrationsPath
func migrateSchema(migrationsURL string, db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsURL, "musclemem", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		return err
	}

	return nil
}

func (xs *SqliteDatastore) ExerciseByID(id uuid.UUID) (Exercise, error) {
	xs.Ping()
	return Exercise{}, nil
}

func (xs *SqliteDatastore) StoreExercise(x Exercise) error {
	if (x == Exercise{}) {
		return ErrEmptyExercise
	}

	switch {
	case x.ID == uuid.Nil:
		return xs.createExercise(x)
	default:
		return xs.overwriteExercise(x)

	}
}

func (xs *SqliteDatastore) getLastExercise() ExerciseRef {
	stmt := `
  SELECT id, name
  FROM exercises
  WHERE previous IS NULL
  `

	row := xs.QueryRow(stmt)

	var ref ExerciseRef
	row.Scan(&ref.ID, &ref.Name)

	return ref
}

func (xs *SqliteDatastore) createExercise(x Exercise) error {
	x.ID = uuid.New()

	last := xs.getLastExercise()
	x.Previous = &last

	template, err := tqla.New()
	if err != nil {
		return err
	}

	insertExerciseStmt := `
  INSERT INTO 'exercises' ('id', 'name', 'weight', 'repetitions', 'next', 'previous')
  VALUES({{ .ID }}, {{ .Name }}, {{ .Weight }}, {{ .Repetitions }}, {{ .Next }}, {{ .Previous.ID}});
  `

	insert, insertArgs, err := template.Compile(insertExerciseStmt, x)
	if err != nil {
		return err
	}

	if _, err := xs.Exec(insert, insertArgs...); err != nil {
		return err
	}

	return nil
}

func (xs *SqliteDatastore) overwriteExercise(x Exercise) error {
	template, err := tqla.New()
	if err != nil {
		return err
	}

	updateExerciseStmt := `
  UPDATE exercises
  SET name = {{ .Name }}, 
      weight = {{ .Weight }},
      repetitions = {{ .Repetitions }}
  WHERE id = {{ .ID }}
  `

	update, updateArgs, err := template.Compile(updateExerciseStmt, x)
	if err != nil {
		return err
	}

	if _, err := xs.Exec(update, updateArgs...); err != nil {
		return err
	}

	return nil
}

func (xs *SqliteDatastore) Exists(id uuid.UUID) bool {
	template, err := tqla.New()
	if err != nil {
		return false
	}

	stmt := `
  SELECT id
  FROM exercises
  WHERE id = {{ . }}
  `

	query, queryArg, err := template.Compile(stmt, id)
	if err != nil {
		return false
	}

	row := xs.QueryRow(query, queryArg...)

	var hit uuid.UUID
	if err := row.Scan(&hit); err != nil {
		return false
	}

	return true
}
