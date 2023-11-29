package user

import (
	"database/sql"
	"fmt"
	"net/mail"

	"github.com/VauntDev/tqla"
)

type SQLUsers struct {
	db *sql.DB
}

func NewSQLUsers(db *sql.DB) Users {
	return &SQLUsers{db}
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

	tmpl, err := tqla.New()
	if err != nil {
		return 0, fmt.Errorf("Register: %w", err)
	}

	q, args, err := tmpl.Compile(stmt, User{Email: email, Password: password})
	if err != nil {
		return 0, fmt.Errorf("Register: compile insert statement: %w", err)
	}

	res, err := us.db.Exec(q, args...)
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
