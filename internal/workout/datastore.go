package workout

import (
	"errors"

	"github.com/scrot/musclemem-api/internal/exercise"
)

var (
	ErrInvalidID     = errors.New("invalid workout id")
	ErrNotFound      = errors.New("workout doesn't exist")
	ErrUserNotFound  = errors.New("user doesn't exist")
	ErrMissingFields = errors.New("missing required fields")
)

type Workouts interface {
	Retreiver
	Storer
}

type Retreiver interface {
	WorkoutExercises(workoutID int) ([]exercise.Exercise, error)
}

type Storer interface {
	New(owner int, name string) (int, error)
}
