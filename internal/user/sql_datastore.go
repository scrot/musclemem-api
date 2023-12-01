package user

import (
	"fmt"
	"net/mail"

	"github.com/scrot/musclemem-api/internal/storage"
	"github.com/scrot/musclemem-api/internal/workout"
)

type SQLUsers struct {
	*storage.SqliteDatastore
}

func NewSQLUsers(ds *storage.SqliteDatastore) Users {
	return &SQLUsers{ds}
}

func (us *SQLUsers) Register(email, password string) (int, error) {
	const stmt = `
  INSERT INTO users (email, password)
  VALUES ({{ .Email }}, {{ .Password }})
  `

	if email == "" || password == "" {
		return 0, fmt.Errorf("Register: %w", ErrMissingFields)
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return 0, fmt.Errorf("Register: validate email: %w", ErrInvalidValue)
	}

	q, args, err := us.CompileStatement(stmt, User{Email: email, Password: password})
	if err != nil {
		return 0, fmt.Errorf("Register: compile insert statement: %w", err)
	}

	res, err := us.Exec(q, args...)
	if err != nil {
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

func (us *SQLUsers) UserWorkouts(userID int) ([]workout.Workout, error) {
	return []workout.Workout{}, nil
}
