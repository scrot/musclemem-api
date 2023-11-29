package workout

import (
	"errors"

	"github.com/scrot/musclemem-api/internal/exercise"
)

var (
	ErrInvalidID     = errors.New("invalid id")
	ErrNotFound      = errors.New("not found")
	ErrMissingFields = errors.New("missing fields")
)

// Workouts implementations allow for interaction with the
// workouts datastore
type Workouts interface {
	Retreiver
	Storer
}

// Retreiver implementations allow for exercises to be retreived
// that belong to that workout
type Retreiver interface {
	ByID(workoutID int) (Workout, error)
	WorkoutExercises(workoutID int) ([]exercise.Exercise, error)
}

// Storer implementations allow for new workouts to be created
type Storer interface {
	New(owner int, name string) (int, error)
	ChangeName(workoutID int, name string) error
}
