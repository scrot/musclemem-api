package workout

import (
	"errors"

	"github.com/scrot/musclemem-api/internal/exercise"
)

var (
	ErrUserNotExist    = errors.New("user doesn't exist")
	ErrWorkoutNotExist = errors.New("workout doesn't exist")
	ErrMissingFields   = errors.New("missing required fields")
)

type Workouts interface {
	Retreiver
	Storer
}

type Retreiver interface {
	Exercises(userID int, workoutID int) ([]exercise.Exercise, error)
}

type Storer interface {
	New(userID int, name string) (int, error)
}
