package user

import (
	"errors"

	"github.com/scrot/musclemem-api/internal/workout"
)

// Users represents the user repository
type Users interface {
	Registerer
	Retreiver
}

var (
	ErrInvalidValue  = errors.New("invalid value")
	ErrMissingFields = errors.New("missing fields")
	ErrUserExists    = errors.New("user already exists")
)

// Registerer allow for new users to be created
type Registerer interface {
	// Register is responsible for validating the email and
	// encrypting the password before storing. A new userID will
	// be returned on success otherwise an error is thrown
	Register(username, email, password string) (int, error)
}

// Retreiver implementations can query data
type Retreiver interface {
	// UserWorkouts returns all workouts that belong to the user
	UserWorkouts(username string) ([]workout.Workout, error)
}
