package user

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/scrot/musclemem-api/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type SQLUserStore struct {
	*storage.SqlDatastore
}

func NewSQLUserStore(ds *storage.SqlDatastore) UserStore {
	return &SQLUserStore{ds}
}

func (us *SQLUserStore) Authenticate(username string, password string) (User, error) {
	const stmt = `
  SELECT password
  FROM users
  WHERE username = {{ . }}
  `

	if username == "" || len(password) <= 0 {
		return User{}, fmt.Errorf("Authenticate: %w", ErrEmptyField)
	}

	q, args, err := us.CompileStatement(stmt, username)
	if err != nil {
		return User{}, fmt.Errorf("Authenticate: compile: %w", err)
	}

	var hash []byte
	if err := us.QueryRow(q, args...).Scan(&hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, fmt.Errorf("Authenticate: query: %w", ErrUnknownUser)
		}
		return User{}, fmt.Errorf("Authenticate: query: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		return User{}, ErrWrongPassword
	}

	u, err := us.ByUsername(username)
	if err != nil {
		return User{}, fmt.Errorf("Authenticate: fetch user: %w", err)
	}

	return u, nil
}

func (us *SQLUserStore) New(username string, email string, password string) (User, error) {
	const stmt = `
  INSERT INTO users (username, email, password)
  VALUES ({{ .Username}}, {{ .Email }}, {{ .Password }})
  `

	if username == "" || email == "" || password == "" {
		return User{}, fmt.Errorf("New: %w", ErrEmptyField)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("New: generate hash: %w", err)
	}

	data := struct {
		Username string
		Email    string
		Password []byte
	}{Username: username, Email: email, Password: hash}

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

	if username == "" {
		return User{}, fmt.Errorf("ByUsername: %w", ErrEmptyField)
	}

	q, args, err := us.CompileStatement(stmt, username)
	if err != nil {
		return User{}, fmt.Errorf("ByUsername: compile: %w", err)
	}

	var u User
	if err := us.QueryRow(q, args...).Scan(&u.Username, &u.Email, &u.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUnknownUser
		}
		return User{}, fmt.Errorf("ByUsername: query: %w", err)
	}

	return u, nil
}
