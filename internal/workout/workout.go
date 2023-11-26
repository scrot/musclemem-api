package workout

import (
	"errors"

	"github.com/scrot/musclemem-api/internal/exercise"
)

type Workout struct {
	ID   int
	Name string
}

var (
	ErrInvalidID     = errors.New("invalid user or workout id")
	ErrNotFound      = errors.New("workout not found")
	ErrMissingFields = errors.New("missing required fields")
)

type Workouts interface {
	Retreiver
	Storer
}

type Retreiver interface {
	Exercises(userID int, workoutID int) ([]exercise.Exercise, error)
}

type Storer interface {
	New(name string) (int, error)
}
