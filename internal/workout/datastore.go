package workout

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidFields = errors.New("contains invalid fields")
)

// Workouts implementations allow for interaction with the
// workouts datastore
type Workouts interface {
	Retreiver
	Storer
	Deleter
}

// Retreiver implementations allow for exercises to be retreived
// that belong to that workout
type Retreiver interface {
	// ByID returns a workout belonging to an owner, given an workout index
	// includeExercises also include all exercises belonging to the workout
	ByID(owner string, workout int) (Workout, error)

	// ByOwner returns all workouts belonging to an owner
	ByOwner(owner string) ([]Workout, error)
}

// Storer implementations allow for new workouts to be created
type Storer interface {
	// New creates a new workout for the given owner
	New(owner string, name string) (Workout, error)

	// ChangeName updates the name of an existing workout
	ChangeName(owner string, workout int, name string) (Workout, error)
}

// Deleter implementations allow for existing workouts to be deleted
type Deleter interface {
	// Delete deletes an existing user workout,
	// It also updates the indices of adjecent exercises
	// The deleted workout is returned as a result
	Delete(owner string, workout int) (Workout, error)
}
