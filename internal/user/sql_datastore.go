package user

import (
	"fmt"
	"slices"

	"github.com/scrot/musclemem-api/internal/storage"
)

type SQLUserStore struct {
	*storage.SqliteDatastore
}

func NewSQLUserStore(ds *storage.SqliteDatastore) UserStore {
	return &SQLUserStore{ds}
}

func (us *SQLUserStore) Authenticate(username string, password []byte) (User, error) {
	const stmt = `
  SELECT password
  FROM users
  WHERE username = {{ . }}
  `

	if username == "" || len(password) <= 0 {
		return User{}, fmt.Errorf("Authenticate: %w", ErrInvalidFields)
	}

	q, args, err := us.CompileStatement(stmt, username)
	if err != nil {
		return User{}, fmt.Errorf("Authenticate: compile: %w", err)
	}

	var actual []byte
	if err := us.QueryRow(q, args...).Scan(&actual); err != nil {
		return User{}, fmt.Errorf("Authenticate: query: %w", err)
	}

	if slices.Equal(password, actual) {
		return User{}, ErrWrongPassword
	}

	u, err := us.ByUsername(username)
	if err != nil {
		return User{}, fmt.Errorf("Authenticate: fetch user: %w", err)
	}

	return u, nil
}

func (us *SQLUserStore) New(username string, email string, password []byte) (User, error) {
	const stmt = `
  INSERT INTO users (username, email, password)
  VALUES ({{ .Username}}, {{ .Email }}, {{ .Password }})
  `

	if username == "" || email == "" || len(password) <= 0 {
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

func (us *SQLUserStore) ByUsername(username string) (User, error) {
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
