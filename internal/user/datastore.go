package user

import "errors"

// Collection of users
type Users interface {
	Registerer
}

var (
	ErrInvalidValue  = errors.New("invalid value")
	ErrMissingFields = errors.New("missing fields")
)

// Registerer allow for new users to be created
type Registerer interface {
	// Register is responsible for validating the email and
	// encrypting the password before storing. A new userID will
	// be returned on success otherwise an error is thrown
	Register(email, password string) (int, error)
}
