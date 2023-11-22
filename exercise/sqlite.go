package exercise

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
	Logger       *slog.Logger
	DatabaseURL  string
	MigrationURL string
	Overwrite    bool
}

// NewSqliteDatastore creates a new database at dbURL
// and runs the migrations in the defaultMigrations folder
// if overwrite is false, it returns the existing db
func NewSqliteDatastore(config SqliteDatastoreConfig) (Datastore, error) {
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

		if err := migrateSchema(config.MigrationURL, db); err != nil {
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

func (s *SqliteDatastore) Exists(id int) bool {
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
func (ds *SqliteDatastore) ExerciseByID(id int) (Exercise, error) {
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

	if err := ds.QueryRow(q, args...).Scan(
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
			return Exercise{}, ErrNotFound
		}
		return Exercise{}, err
	}

	if prev != 0 {
		var pRef ExerciseRef

		q, args, err = tmpl.Compile(refStmt, prev)
		if err != nil {
			return Exercise{}, err
		}

		if err := ds.QueryRow(q, args...).Scan(&pRef.ID, &pRef.Name); err != nil {
			return Exercise{}, err
		}

		e.Previous = pRef
	}

	if next != 0 {
		var nRef ExerciseRef

		q, args, err = tmpl.Compile(refStmt, next)
		if err != nil {
			return Exercise{}, err
		}

		if err := ds.QueryRow(q, args...).Scan(&nRef.ID, &nRef.Name); err != nil {
			return Exercise{}, err
		}

		e.Next = nRef
	}

	return e, nil
}

// ExercisesByWorkoutID returns all exercises part of a workout
// given a owner and workout id pair
func (ds *SqliteDatastore) ExercisesByWorkoutID(uid int, wid int) ([]Exercise, error) {
	const (
		selectStmt = `
    SELECT exercise_id, owner, workout, name, weight, repetitions, previous, next
    FROM exercises
    WHERE owner = {{ .OwnerID }} AND workout = {{ .WorkoutID }}
    `
		selectRefsStmt = `
    SELECT exercise_id, name
    FROM exercises
    WHERE exercise_id = {{.}}
    `
	)

	if uid == 0 || wid == 0 {
		return []Exercise{}, ErrNoID
	}

	tmpl, err := tqla.New()
	if err != nil {
		return []Exercise{}, err
	}

	data := struct {
		OwnerID   int
		WorkoutID int
	}{uid, wid}

	q, args, err := tmpl.Compile(selectStmt, data)
	if err != nil {
		return []Exercise{}, err
	}

	rs, err := ds.Query(q, args...)
	if err != nil {
		return []Exercise{}, err
	}

	var xs []Exercise
	for rs.Next() {
		var (
			x          Exercise
			prev, next int
		)

		rs.Scan(&x.ID, &x.Owner, &x.Workout, &x.Weight,
			&x.Repetitions, &prev, &next)

		if x.Previous.ID != 0 {
			q, args, err := tmpl.Compile(selectRefsStmt, x.Previous.ID)
			if err != nil {
				return []Exercise{}, err
			}

			if err := ds.QueryRow(q, args...).Scan(&x.Previous.ID, &x.Previous.Name); err != nil {
				return []Exercise{}, err
			}
		}

		if x.Next.ID != 0 {
			q, args, err := tmpl.Compile(selectRefsStmt, x.Next.ID)
			if err != nil {
				return []Exercise{}, err
			}

			if err := ds.QueryRow(q, args...).Scan(&x.Next.ID, &x.Next.Name); err != nil {
				return []Exercise{}, err
			}
		}

		xs = append(xs, x)
	}

	if len(xs) == 0 {
		return []Exercise{}, ErrNotFound
	}

	return xs, nil
}

// StoreExercise stores the given exrcise in the sqlite database
// If the id already exists it updates the existing record
// ErrMissingFields is returned when fields are missing
func (ds *SqliteDatastore) Store(x Exercise) (int, error) {
	const (
		insertStmt = `
    INSERT INTO exercises (owner, workout, name, weight, repetitions, previous, next)
    VALUES ({{.Owner}}, {{.Workout}}, {{.Name}}, {{.Weight}}, {{.Repetitions}}, {{.Previous.ID}}, {{.Next.ID}});
    `
		updateStmt = `
    UPDATE exercises
    SET name = {{ .Name }}, 
        weight = {{ .Weight }},
        repetitions = {{ .Repetitions }}
    WHERE exercise_id = {{ .ID }}
    `
		updateNextStmt = `
    UPDATE exercises
    SET next = {{ .NextID }}
    WHERE exercise_id = {{ .ID }}
    `
		updatePreviousStmt = `
    UPDATE exercises
    SET previous = {{.PreviousID }}   
    WHERE exercise_id = {{ .ID }}
    `
	)

	if x.Owner <= 0 || x.Workout <= 0 || x.Name == "" ||
		x.Weight <= 0 || x.Repetitions <= 0 {
		return 0, ErrMissingFields
	}

	tmpl, err := tqla.New()
	if err != nil {
		return 0, err
	}

	switch {
	case x.ID == 0:
		// get last exercise in workout
		tail, err := tail(ds, x.Owner, x.Workout)
		if err != nil {
			return 0, fmt.Errorf("get last exercise: %w", err)
		}

		tx, err := ds.Begin()
		if err != nil {
			return 0, nil
		}

		// insert new exercise
		insert, args, err := tmpl.Compile(insertStmt, x)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("insert new exercise (compile): %w", err)
		}

		res, err := tx.Exec(insert, args...)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("insert new exercise (exec): %w", err)
		}

		newID, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("insert new exercise (id from res): %w", err)
		}

		// update previous for tail
		updatePreviousData := struct {
			ID         int
			PreviousID int
		}{
			ID:         int(newID),
			PreviousID: tail.ID,
		}

		updateNew, args, err := tmpl.Compile(updatePreviousStmt, updatePreviousData)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("update ref for n-1 (compile): %w", err)
		}

		_, err = tx.Exec(updateNew, args...)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("update ref for n-1 (exec): %w", err)
		}

		// update next for tail-1
		updateNextData := struct {
			ID     int
			NextID int
		}{
			ID:     tail.ID,
			NextID: int(newID),
		}

		updateLast, args, err := tmpl.Compile(updateNextStmt, updateNextData)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("update ref for n (compile): %w", err)
		}

		_, err = tx.Exec(updateLast, args...)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("update ref for n (exec): %w", err)
		}

		if err := tx.Commit(); err != nil {
			return 0, fmt.Errorf("commit new exercise transaction: %w", err)
		}

		return int(newID), nil

	case ds.Exists(x.ID):
		q, args, err := tmpl.Compile(updateStmt, x)
		if err != nil {
			return 0, err
		}

		_, err = ds.Exec(q, args...)
		if err != nil {
			return 0, err
		}

		return x.ID, nil

	default:
		return 0, fmt.Errorf("id provided but exercise not found")
	}
}

func tail(s *SqliteDatastore, owner, workout int) (Exercise, error) {
	const (
		tailStmt = `
    SELECT exercise_id   
    FROM exercises
    WHERE owner={{.Owner}} AND workout={{.Workout}} AND next=0
    `
	)

	tmpl, err := tqla.New()
	if err != nil {
		return Exercise{}, err
	}

	data := struct {
		Owner   int
		Workout int
	}{owner, workout}

	tail, args, err := tmpl.Compile(tailStmt, data)
	if err != nil {
		return Exercise{}, fmt.Errorf("get tail id (compile): %w", err)
	}

	var tailID int

	err = s.QueryRow(tail, args...).Scan(&tailID)
	if err != nil {
		return Exercise{}, fmt.Errorf("get tail id (query): %w", err)
	}

	tailExercise, err := s.ExerciseByID(tailID)
	if err != nil {
		return Exercise{}, fmt.Errorf("get tail exercise: %w", err)
	}
	return tailExercise, nil
}

func updateExercise(s *SqliteDatastore, id int, name string, weight float64, repetitions int) error {
	return nil
}

func updateReferences(s *SqliteDatastore, id, next, previous int) error {
	return nil
}

// TODO: Implement
func (ds *SqliteDatastore) Delete(id int) error {
	return nil
}

// TODO: Implement
func (ds *SqliteDatastore) Swap(i int, j int) error {
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
