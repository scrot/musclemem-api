package workout

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/scrot/musclemem-api/internal/storage"
)

type SQLWorkouts struct {
	*storage.SqliteDatastore
}

func NewSQLWorkouts(db *storage.SqliteDatastore) *SQLWorkouts {
	return &SQLWorkouts{db}
}

func (ws *SQLWorkouts) New(owner string, name string) (Workout, error) {
	const stmt = `
    INSERT INTO workouts (owner, workout_index, name)
    VALUES ({{ .Owner }}, {{ .Index }}, {{ .Name }})
    `

	if owner == "" || name == "" {
		return Workout{}, fmt.Errorf("New: %w", ErrInvalidFields)
	}

	if !ws.userExists(owner) {
		return Workout{}, fmt.Errorf("New: validate user: %w", ErrNotFound)
	}

	last, err := ws.lastIndex(owner)
	if err != nil {
		return Workout{}, fmt.Errorf("New: last index: %w", err)
	}

	data := Workout{Owner: owner, Index: last + 1, Name: name}
	q, args, err := ws.CompileStatement(stmt, data)
	if err != nil {
		return Workout{}, fmt.Errorf("New: compile: %w", err)
	}

	if _, err := ws.Exec(q, args...); err != nil {
		return Workout{}, fmt.Errorf("New: execute: %w", err)
	}

	workout, err := ws.ByID(owner, last+1)
	if err != nil {
		return Workout{}, fmt.Errorf("New: fetch workout %d: %w", last+1, err)
	}

	return workout, nil
}

func (ws *SQLWorkouts) Delete(owner string, workout int) (Workout, error) {
	const (
		stmt = `
    DELETE FROM workouts
    WHERE owner = {{ .Owner }} AND workout_index = {{ .Workout }}
    `

		indStmt = `
    UPDATE workouts
    SET workout_index = {{ .NewIndex }}
    WHERE owner = {{ .Owner }} AND workout_index = {{ .Index }}
    `
	)
	if owner == "" || workout <= 0 {
		return Workout{}, fmt.Errorf("Delete: %w", ErrInvalidFields)
	}

	wo, err := ws.ByID(owner, workout)
	if err != nil {
		return Workout{}, fmt.Errorf("Delete: fetch workout: %w", err)
	}

	tx, err := ws.Begin()
	if err != nil {
		return Workout{}, fmt.Errorf("Delete: begin transaction: %w", err)
	}

	data := struct {
		Owner   string
		Workout int
	}{owner, workout}

	q, args, err := ws.CompileStatement(stmt, data)
	if err != nil {
		return Workout{}, fmt.Errorf("Delete: compile delete: %w", err)
	}

	if _, err := tx.Exec(q, args...); err != nil {
		tx.Rollback()
		return Workout{}, fmt.Errorf("Delete: execute delete: %w", err)
	}

	wos, err := ws.ByOwner(owner)
	if err != nil {
		return Workout{}, fmt.Errorf("Delete: fetch workouts: %w", err)
	}

	// update all subsequent workout indices
	for _, w := range wos {
		if w.Index > workout {
			data := struct {
				Owner           string
				Index, NewIndex int
			}{owner, w.Index, w.Index - 1}

			q, args, err := ws.CompileStatement(indStmt, data)
			if err != nil {
				return Workout{}, fmt.Errorf("Delete: compile update %d: %w", w.Index, err)
			}

			if _, err := tx.Exec(q, args...); err != nil {
				tx.Rollback()
				return Workout{}, fmt.Errorf("Delete: execute update %d: %w", w.Index, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return Workout{}, fmt.Errorf("Delete: commit transaction %w", err)
	}

	return wo, nil
}

func (ws *SQLWorkouts) ByID(owner string, workout int) (Workout, error) {
	const stmt = `
  SELECT owner, workout_index, name
  FROM workouts
  WHERE owner = {{ .Owner }} AND workout_index = {{ .Workout }}
  `

	if owner == "" || workout <= 0 {
		return Workout{}, ErrInvalidFields
	}

	data := struct {
		Owner   string
		Workout int
	}{owner, workout}

	q, args, err := ws.CompileStatement(stmt, data)
	if err != nil {
		return Workout{}, fmt.Errorf("ByID: compile: %w", err)
	}

	var w Workout
	if err := ws.QueryRow(q, args...).Scan(&w.Owner, &w.Index, &w.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Workout{}, fmt.Errorf("ByID: %w", ErrNotFound)
		}
		return Workout{}, fmt.Errorf("ByID: query: %w", err)
	}

	return w, nil
}

func (ws *SQLWorkouts) ByOwner(owner string) ([]Workout, error) {
	const stmt = `
  SELECT owner, workout_index, name
  FROM workouts
  WHERE owner = {{ . }}
  `

	if owner == "" {
		return []Workout{}, ErrInvalidFields
	}

	q, args, err := ws.CompileStatement(stmt, owner)
	if err != nil {
		return []Workout{}, fmt.Errorf("ByOwner: compile: %w", err)
	}

	rows, err := ws.Query(q, args...)
	if err != nil {
		return []Workout{}, fmt.Errorf("ByOwner: query: %w", err)
	}

	var wos []Workout
	for rows.Next() {
		var w Workout
		if err := rows.Scan(&w.Owner, &w.Index, &w.Name); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return []Workout{}, fmt.Errorf("ByOwner: query: %w", ErrNotFound)
			}
			return []Workout{}, fmt.Errorf("ByOwner: query: %w", err)
		}
		wos = append(wos, w)
	}

	return wos, nil
}

func (ws *SQLWorkouts) ChangeName(owner string, workout int, name string) (Workout, error) {
	const stmt = `
  UPDATE workouts
  SET name = {{ .Name }}
  WHERE owner = {{ .Owner }} AND workout_index = {{ .Workout }}
  `

	if owner == "" || workout <= 0 || name == "" {
		return Workout{}, fmt.Errorf("ChangeName: %w", ErrInvalidFields)
	}

	if _, err := ws.ByID(owner, workout); err != nil {
		return Workout{}, fmt.Errorf("ChangeName: workout %s/%d: %w", owner, workout, ErrNotFound)
	}

	data := struct {
		Owner   string
		Workout int
		Name    string
	}{owner, workout, name}

	q, args, err := ws.CompileStatement(stmt, data)
	if err != nil {
		return Workout{}, fmt.Errorf("ChangeName: compile: %w", err)
	}

	if _, err := ws.Exec(q, args...); err != nil {
		return Workout{}, fmt.Errorf("ChangeName: execute: %w", err)
	}

	w, err := ws.ByID(owner, workout)
	if err != nil {
		return Workout{}, fmt.Errorf("ChangeName: fetch %s/%d: %w", owner, workout, err)
	}

	return w, nil
}

// lastIndex returns the last workout index of a user
// if the index is 0 and no error, then there are no workouts
func (xs *SQLWorkouts) lastIndex(owner string) (int, error) {
	const (
		lastIndexStmt = `
    SELECT MAX(workout_index)
    FROM workouts
    WHERE owner = {{.}}
    `
	)

	// find last exercise if at least one in user workout
	q, args, err := xs.CompileStatement(lastIndexStmt, owner)
	if err != nil {
		return 0, fmt.Errorf("lastIndex: %w", err)
	}

	var index sql.NullInt32
	err = xs.QueryRow(q, args...).Scan(&index)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("lastIndex: %w", err)
	}

	if !index.Valid {
		return 0, nil
	}

	return int(index.Int32), nil
}

func (ws *SQLWorkouts) userExists(username string) bool {
	const stmt = `
  SELECT 1
  FROM users
  WHERE username = {{ . }}
  `

	q, args, err := ws.CompileStatement(stmt, username)
	if err != nil {
		return false
	}

	var exists bool
	if err := ws.QueryRow(q, args...).Scan(&exists); err != nil {
		return false
	}

	return exists
}
