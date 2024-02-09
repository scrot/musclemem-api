package user

import (
	"fmt"

	"github.com/scrot/musclemem-api/internal/storage"
)

type SQLUsers struct {
	*storage.SqliteDatastore
}

func NewSQLUsers(ds *storage.SqliteDatastore) Users {
	return &SQLUsers{ds}
}

func (us *SQLUsers) New(username, email, password string) (User, error) {
	const stmt = `
  INSERT INTO users (username, email, password)
  VALUES ({{ .Username}}, {{ .Email }}, {{ .Password }})
  `

	if username == "" || email == "" || password == "" {
		return User{}, fmt.Errorf("New: %w", ErrInvalidFields)
	}

	data := User{Username: username, Email: email, Password: password}
	q, args, err := us.CompileStatement(stmt, data)
	if err != nil {
		return User{}, fmt.Errorf("New: compile: %w", err)
	}

	if _, err := us.Exec(q, args...); err != nil {
		return User{}, fmt.Errorf("New: execute: %w", err)
	}

	u, err := us.ByUsername(username)
	if err != nil {
		return User{}, fmt.Errorf("New: fetch %s: %w", username, err)
	}

	return u, nil
}

func (us *SQLUsers) ByUsername(username string) (User, error) {
	const stmt = `
  SELECT username, email, password
  FROM users
  WHERE username = {{ . }}
  `

	q, args, err := us.CompileStatement(stmt, username)
	if err != nil {
		return User{}, fmt.Errorf("ByUsername: compile: %w", err)
	}

	var u User
	if err := us.QueryRow(q, args...).Scan(&u.Username, &u.Email, &u.Password); err != nil {
		return User{}, fmt.Errorf("ByUsername: query: %w", err)
	}

	return u, nil
}
