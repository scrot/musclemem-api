package main

import (
	"database/sql"
	"errors"
	"os"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	migrationsURL = "file://migrations/"
	exercisesURL  = "./repos/exercises.db"
)

type Exercises struct {
	*sql.DB
}

func NewExercises() (*Exercises, error) {
	if err := os.Remove(exercisesURL); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	if _, err := os.Create(exercisesURL); err != nil {
		return nil, err
	}

	m, err := migrate.New(migrationsURL, "sqlite3://"+exercisesURL)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil {
		return nil, err
	}

	xs, err := NewExercisesConn()
	if err != nil {
		return nil, err
	}

	return xs, nil
}

func NewExercisesConn() (*Exercises, error) {
	db, err := sql.Open("sqlite3", exercisesURL)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Exercises{db}, err
}

// GetExerciseByID implements the ExerciseRetreiver
// interface for sqlite3 databases
func (xs *Exercises) ExerciseByID(id uuid.UUID) (Exercise, error) {
	xs.Ping()
	return Exercise{}, nil
}
