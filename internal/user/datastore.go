package user

import (
	"errors"
)

// UserStore represents the user repository
type UserStore interface {
	Storer
	Retreiver
}

var (
	ErrInvalidFields = errors.New("contains invalid fields")
	ErrWrongPassword = errors.New("wrong password")
)

// Storer allow for new users to be created
type Storer interface {
	// New is responsible for validating the email and
	// encrypting the password before storing. A new userID will
	// be returned on success otherwise an error is thrown
	New(username string, email string, password string) (User, error)
}

// Retreiver implementations can query data
type Retreiver interface {
	// ByUsername returns the User given a username
	// it returns an ErrNotFound if user does not exists
	// includeWorkouts and includeExercises include
	// the user's workouts and workout exercises as well
	ByUsername(username string) (User, error)

	// Authenticate returns the user after checking the
	// credentials. It returns an ErrWrongPassword if
	// username/password pair doesn't match
	Authenticate(username string, password []byte) (User, error)
}
