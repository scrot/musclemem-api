package musclememapi

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/VauntDev/tqla"
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
	*slog.Logger
}

// Configuration for SqliteDatastore
type SqliteDatastoreConfig struct {
	Logger      *slog.Logger
	DatabaseURL string
	Overwrite   bool
}

// NewSqliteDatastore creates a new database at dbURL
// and runs the migrations in the defaultMigrations folder
// if overwrite is false, it returns the existing db
func NewSqliteDatastore(config SqliteDatastoreConfig) (*SqliteDatastore, error) {
	log := config.Logger

	var db *sql.DB

	path, hasFileScheme := strings.CutPrefix(config.DatabaseURL, "file://")
	if !hasFileScheme {
		err := errors.New("invalid scheme expected file://")
		log.Error(fmt.Sprintf("%s: %s", config.DatabaseURL, err))
		return nil, err
	}

	// create new if not exist or overwrite is true
	if _, err := os.Stat(path); err != nil || config.Overwrite {
		db, err = newLocalDatabase(path)
		if err != nil {
			log.Error(fmt.Sprintf("creating new database %s: %s", path, err))
			return nil, err
		}
		log.Debug("new database created", "db-url", config.DatabaseURL)

		if err := migrateSchema(defaultMigrationsURL, db); err != nil {
			log.Error(fmt.Sprintf("running schema migrations %s against %s: %s", defaultMigrationsURL, path, err))
			return nil, err
		}
		log.Debug("schemas migrated", "migrations", defaultMigrationsURL, "db-url", config.DatabaseURL)

	} else {
		db, err = sql.Open("sqlite3", path)
		if err != nil {
			log.Error(fmt.Sprintf("opening existing database %s: %s", path, err))
			return nil, err
		}
		log.Debug("existing database opened", "db-url", config.DatabaseURL)
	}

	return &SqliteDatastore{db, config.Logger}, nil
}

func (s *SqliteDatastore) ExerciseExists(id int) bool {
	log := s.Logger

	stmt := `
  SELECT 1
  FROM exercises
  WHERE exercise_id = {{ . }}
  `

	tmpl, err := tqla.New()
	if err != nil {
		log.Error(fmt.Sprintf("new sql template: %s", err))
		return false
	}

	q, args, err := tmpl.Compile(stmt, id)
	if err != nil {
		log.Error(fmt.Sprintf("compiling query: %s", err))
		return false
	}

	var exists bool
	if err := s.QueryRow(q, args...).Scan(&exists); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Error(fmt.Sprintf("querying database: %s", err))
		}
		return false
	}

	return exists
}

// ExerciseByID returns an exercise from the database if exists
// otherwise returns NotFound error
func (s *SqliteDatastore) ExerciseByID(id int) (Exercise, error) {
	const (
		stmt = `
    SELECT exercise_id, owner, workout, name, weight, repetitions, previous, next
    FROM exercises
    WHERE exercise_id = {{ . }}
    `
		refStmt = `
    SELECT exercise_id, name
    FROM exercises
    WHERE {{ . }} = exercise_id
    `
	)

	if id <= 0 {
		return Exercise{}, ErrInvalidIdFormat
	}

	tmpl, err := tqla.New()
	if err != nil {
		return Exercise{}, err
	}

	q, args, err := tmpl.Compile(stmt, id)
	if err != nil {
		return Exercise{}, err
	}

	var (
		e    Exercise
		prev int
		next int
	)

	if err := s.QueryRow(q, args...).Scan(
		&e.ID,
		&e.Owner,
		&e.Workout,
		&e.Name,
		&e.Weight,
		&e.Repetitions,
		&prev,
		&next,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Exercise{}, ErrExerciseNotFound
		}
		return Exercise{}, err
	}

	if prev != 0 {
		var pRef ExerciseRef

		q, args, err = tmpl.Compile(refStmt, prev)
		if err != nil {
			return Exercise{}, err
		}

		if err := s.QueryRow(q, args...).Scan(&pRef.ID, &pRef.Name); err != nil {
			return Exercise{}, err
		}

		e.Previous = &pRef
	}

	if next != 0 {
		var nRef ExerciseRef

		q, args, err = tmpl.Compile(refStmt, next)
		if err != nil {
			return Exercise{}, err
		}

		if err := s.QueryRow(q, args...).Scan(&nRef.ID, &nRef.Name); err != nil {
			return Exercise{}, err
		}

		e.Next = &nRef
	}

	return e, nil
}

// StoreExercise stores the given exrcise in the sqlite database
// If the id already exists it updates the existing record
// ErrMissingFields is returned when fields are missing
func (s *SqliteDatastore) StoreExercise(x Exercise) error {
	const (
		insertStmt = `
    INSERT INTO exercises (owner, workout, name, weight, repetitions)
    VALUES ({{.Owner}}, {{.Workout}}, {{.Name}}, {{.Weight}}, {{.Repetitions}});
    `
		updateStmt = `
    UPDATE exercises
    SET name = {{ .Name }}, 
        weight = {{ .Weight }},
        repetitions = {{ .Repetitions }}
    WHERE exercise_id = {{ .ID }}
    `
		tailStmt = `
    SELECT id, name
    FROM exercises
    WHERE workout={.Workout} AND previous IS NULL
    `
	)

	if (x == Exercise{}) {
		return ErrMissingFields
	}

	tmpl, err := tqla.New()
	if err != nil {
		return err
	}

	if !s.ExerciseExists(x.ID) {
		if x.Owner <= 0 || x.Workout <= 0 || x.Name == "" || x.Weight <= 0 || x.Repetitions <= 0 {
			return ErrMissingFields
		}
		q, args, err := tmpl.Compile(insertStmt, x)
		if err != nil {
			return err
		}

		_, err = s.Exec(q, args...)
		if err != nil {
			return err
		}
	}

	return nil
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
