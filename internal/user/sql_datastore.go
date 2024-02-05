package user

import (
	"fmt"
	"net/mail"

	"github.com/scrot/musclemem-api/internal/storage"
	"github.com/scrot/musclemem-api/internal/workout"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type SQLUsers struct {
	*storage.SqliteDatastore
}

func NewSQLUsers(ds *storage.SqliteDatastore) Users {
	return &SQLUsers{ds}
}

func (us *SQLUsers) Register(username, email, password string) (int, error) {
	const stmt = `
  INSERT INTO users (username, email, password)
  VALUES ({{ .Username}}, {{ .Email }}, {{ .Password }})
  `

	if username == "" || email == "" || password == "" {
		return 0, fmt.Errorf("Register: %w", ErrMissingFields)
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return 0, fmt.Errorf("Register: validate email: %w", ErrInvalidValue)
	}

	q, args, err := us.CompileStatement(stmt, User{Username: username, Email: email, Password: password})
	if err != nil {
		return 0, fmt.Errorf("Register: compile insert statement: %w", err)
	}

	res, err := us.Exec(q, args...)
	if err != nil {
		if liteErr, ok := err.(*sqlite.Error); ok {
			code := liteErr.Code()
			if code == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY {
				return 0, ErrUserExists
			}
		}
		return 0, fmt.Errorf("Register: execute insert: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Register: new id: %w", err)
	}

	if id <= 0 {
		return 0, fmt.Errorf("Register: invalid new id")
	}

	return int(id), nil
}

func (us *SQLUsers) UserWorkouts(username string) ([]workout.Workout, error) {
	const stmt = `
  SELECT *
  FROM workouts
  WHERE owner IN (SELECT user_id
    FROM users
    WHERE username = {{.}})
  `

	q, args, err := us.CompileStatement(stmt, username)
	if err != nil {
		return []workout.Workout{}, fmt.Errorf("UserWorkouts: %w", err)
	}

	rows, err := us.Query(q, args...)
	if err != nil {
		return []workout.Workout{}, fmt.Errorf("UserWorkouts: %w", err)
	}

	ws := []workout.Workout{}
	for rows.Next() {
		var w workout.Workout
		err := rows.Scan(&w.ID, &w.Owner, &w.Name)
		if err != nil {
			return []workout.Workout{}, fmt.Errorf("UserWorkouts: %w", err)
		}
		ws = append(ws, w)
	}

	return ws, nil
}
