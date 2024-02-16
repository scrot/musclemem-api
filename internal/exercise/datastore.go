package exercise

import (
	"errors"

	_ "github.com/go-playground/validator/v10"
)

var (
	ErrNotFound = errors.New("not found")

	// TODO: custom error containing which fields are invalid
	ErrInvalidFields = errors.New("contains invalid fields")
)

type ExerciseStore interface {
	Retreiver
	Storer
	Updater
	Deleter
	Orderer
}

// Implementation of the Retreiver interface enables querying exercises
type Retreiver interface {
	// ByID takes an excercise id and returns the exercise
	// from the exercises repository if it exists
	ByID(owner string, workout int, exercise int) (Exercise, error)

	// ByWorkout returns all exercises belongign to an user's workout
	ByWorkout(owner string, workout int) ([]Exercise, error)
}

// Implementation of the Storer interface enables creating new exercises
type Storer interface {
	// New stores an exercise at the tail,
	// updating the references and returns the exercise.
	// user and workout must exist before adding exercise
	New(owner string, workout int, name string, weight float64, repetitions int) (Exercise, error)
}

// Implemention of Updater interface enables updating exercises
type Updater interface {
	// ChangeName updates the name of an existing exercise
	ChangeName(owner string, workout int, exercise int, newName string) (Exercise, error)

	// UpdateWeight updates the weight of an existing exercise
	UpdateWeight(owner string, workout int, exercise int, newWeight float64) (Exercise, error)

	// UpdateRepetitions updates the repetitions of an existing exercise
	UpdateRepetitions(owner string, workout int, exercise int, newRepetitions int) (Exercise, error)
}

type Deleter interface {
	// DeleteExercise deletes an exercise if exists
	// and updates the references in the linked list
	Delete(owner string, workout int, exercise int) (Exercise, error)
}

type Orderer interface {
	// Swap swaps the indices from the given exercises
	// if the workout or index doesn't exist it returns an error
	Swap(owner string, workout int, e1 int, e2 int) error

	// Len returns the length of all exercises of a workout
	Len(owner string, workout int) (int, error)
}
