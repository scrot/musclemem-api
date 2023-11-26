package exercise

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/VauntDev/tqla"
	"github.com/google/go-cmp/cmp"
)

type SqliteExercises struct {
	db *sql.DB
}

func NewSqliteExercises(db *sql.DB) *SqliteExercises {
	return &SqliteExercises{db}
}

// ExerciseByID returns an exercise from the database if exists
// otherwise returns NotFound error
func (ds *SqliteExercises) WithID(id int) (Exercise, error) {
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
		return Exercise{}, ErrNoID
	}

	tmpl, err := tqla.New()
	if err != nil {
		return Exercise{}, err
	}

	q, args, err := tmpl.Compile(stmt, id)
	if err != nil {
		return Exercise{}, fmt.Errorf("WithID: %w", err)
	}

	var (
		e    Exercise
		prev int
		next int
	)

	if err := ds.db.QueryRow(q, args...).Scan(
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
			return Exercise{}, fmt.Errorf("WithID: %w", ErrNotFound)
		}
		return Exercise{}, fmt.Errorf("WithID: %w", err)
	}

	if prev != 0 {
		var pRef ExerciseRef

		q, args, err = tmpl.Compile(refStmt, prev)
		if err != nil {
			return Exercise{}, fmt.Errorf("WithID: %w", err)
		}

		if err := ds.db.QueryRow(q, args...).Scan(&pRef.ID, &pRef.Name); err != nil {
			return Exercise{}, fmt.Errorf("WithID: %w", err)
		}

		e.Previous = pRef
	}

	if next != 0 {
		var nRef ExerciseRef

		q, args, err = tmpl.Compile(refStmt, next)
		if err != nil {
			return Exercise{}, fmt.Errorf("WithID: %w", err)
		}

		if err := ds.db.QueryRow(q, args...).Scan(&nRef.ID, &nRef.Name); err != nil {
			return Exercise{}, fmt.Errorf("WithID: %w", err)
		}

		e.Next = nRef
	}

	return e, nil
}

// ExercisesByWorkoutID returns all exercises part of a workout
// given a owner and workout id pair
func (ds *SqliteExercises) FromWorkout(uid int, wid int) ([]Exercise, error) {
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

	rs, err := ds.db.Query(q, args...)
	if err != nil {
		return []Exercise{}, err
	}

	var xs []Exercise
	for rs.Next() {
		var (
			x          Exercise
			prev, next int
		)

		err := rs.Scan(&x.ID, &x.Owner, &x.Workout, &x.Name, &x.Weight,
			&x.Repetitions, &prev, &next)
		if err != nil {
			return []Exercise{}, err
		}

		if prev != 0 {
			q, args, err := tmpl.Compile(selectRefsStmt, prev)
			if err != nil {
				return []Exercise{}, err
			}

			if err := ds.db.QueryRow(q, args...).Scan(&x.Previous.ID, &x.Previous.Name); err != nil {
				return []Exercise{}, err
			}
		}

		if next != 0 {
			q, args, err := tmpl.Compile(selectRefsStmt, next)
			if err != nil {
				return []Exercise{}, err
			}

			if err := ds.db.QueryRow(q, args...).Scan(&x.Next.ID, &x.Next.Name); err != nil {
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
func (ds *SqliteExercises) Store(x Exercise) (int, error) {
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

	// required fields
	if x.Owner <= 0 || x.Workout <= 0 || x.Name == "" ||
		x.Weight <= 0 || x.Repetitions <= 0 {
		return 0, fmt.Errorf("Store: %w", ErrMissingFields)
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
			return 0, fmt.Errorf("Store: %w", err)
		}

		// insert new exercise
		tx, err := ds.db.Begin()
		if err != nil {
			return 0, nil
		}

		insert, args, err := tmpl.Compile(insertStmt, x)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("Store: insert new exercise: %w", err)
		}

		res, err := tx.Exec(insert, args...)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("Store: insert new exercise: %w", err)
		}

		newID, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("Store: insert new exercise: %w", err)
		}

		// update exercise references
		if !cmp.Equal(tail, Exercise{}) {
			newExPatch := struct {
				ID         int
				PreviousID int
			}{
				ID:         int(newID),
				PreviousID: tail.ID,
			}

			patchNew, args, err := tmpl.Compile(updatePreviousStmt, newExPatch)
			if err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("Store: patch new exercise: %w", err)
			}

			_, err = tx.Exec(patchNew, args...)
			if err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("Store: patch new exercise: %w", err)
			}

			oldExPatch := struct {
				ID     int
				NextID int
			}{
				ID:     tail.ID,
				NextID: int(newID),
			}

			patchOld, args, err := tmpl.Compile(updateNextStmt, oldExPatch)
			if err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("Store: patch old exercise: %w", err)
			}

			_, err = tx.Exec(patchOld, args...)
			if err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("Store: patch old exercise: %w", err)
			}
		}

		if err := tx.Commit(); err != nil {
			return 0, fmt.Errorf("Store: commit transaction: %w", err)
		}

		return int(newID), nil

	case ds.exists(x.ID):
		q, args, err := tmpl.Compile(updateStmt, x)
		if err != nil {
			return 0, fmt.Errorf("Store: update existing exercise: %w", err)
		}

		_, err = ds.db.Exec(q, args...)
		if err != nil {
			return 0, fmt.Errorf("Store: update existing exercise: %w", err)
		}

		return x.ID, nil

	default:
		return 0, fmt.Errorf("Store: id provided but exercise not found")
	}
}

func (ds *SqliteExercises) Delete(id int) error {
	deleteStmt := `
  DELETE FROM exercises
  WHERE exercise_id = {{.}}
  `

	if id <= 0 {
		return fmt.Errorf("Delete: %w", ErrNoID)
	}

	tmpl, err := tqla.New()
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}

	q, args, err := tmpl.Compile(deleteStmt, id)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	res, err := ds.db.Exec(q, args...)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	c, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}

	if c == 0 {
		return fmt.Errorf("Delete: %w", ErrNotFound)
	}

	return nil
}

// TODO: Implement
func (ds *SqliteExercises) Swap(i int, j int) error {
	return nil
}

// tail returns last exercise in linked list or empty exercise if
// no exercises exists for that workout
func tail(ds *SqliteExercises, owner, workout int) (Exercise, error) {
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

	// find last exercise if at least one in user workout
	q, args, err := tmpl.Compile(tailStmt, data)
	if err != nil {
		return Exercise{}, fmt.Errorf("tail: %w", err)
	}

	var tailID int

	err = ds.db.QueryRow(q, args...).Scan(&tailID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Exercise{}, nil
		}
		return Exercise{}, fmt.Errorf("tail: %w", err)
	}

	tailExercise, err := ds.WithID(tailID)
	if err != nil {
		return Exercise{}, fmt.Errorf("tail: %w", err)
	}
	return tailExercise, nil
}

func (ds *SqliteExercises) exists(id int) bool {
	stmt := `
  SELECT 1
  FROM exercises
  WHERE exercise_id = {{ . }}
  `

	tmpl, err := tqla.New()
	if err != nil {
		return false
	}

	q, args, err := tmpl.Compile(stmt, id)
	if err != nil {
		return false
	}

	var exists bool
	if err := ds.db.QueryRow(q, args...).Scan(&exists); err != nil {
		return false
	}

	return exists
}
