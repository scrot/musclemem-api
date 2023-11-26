package exercise

import (
	"errors"

	_ "github.com/go-playground/validator/v10"
)

// Exercise represents a specific exercise that belongs to a
// workout of a user, an exercise is a node in a linked list
// of workout exercises containing also its implicit position
// in the list of exercises
type (
	Exercise struct {
		ID          int         `json:"id" validate:"gt=0"`
		Owner       int         `json:"owner" validate:"required,gt=0"`
		Workout     int         `json:"workout" validate:"required"`
		Name        string      `json:"name" validate:"required"`
		Weight      float64     `json:"weight" validate:"required"`
		Repetitions int         `json:"repetitions" validate:"required"`
		Next        ExerciseRef `json:"next,omitempty"`
		Previous    ExerciseRef `json:"previous,omitempty"`
	}

	ExerciseRef struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)

// Ref takes an Exercise and returns a Exercise reference
// containing only basic information of the node
func (e Exercise) ToRef() ExerciseRef {
	return ExerciseRef{e.ID, e.Name}
}

// ToExercise returns an exercise belonging to that reference
// if it exists in the Exercises repository
func (r ExerciseRef) ToExercise(xs Exercises) (Exercise, error) {
	x, err := xs.WithID(r.ID)
	if err != nil {
		return Exercise{}, err
	}
	return x, nil
}

var (
	ErrInvalidID     = errors.New("invalid exercise id")
	ErrNotFound      = errors.New("exercise not found")
	ErrMissingFields = errors.New("missing required fields")
)

// Implementation of Exercises provide the means to persist
// exercises in a repository
type Exercises interface {
	Retreiver
	Storer
	Updater
	Deleter
	Orderer
}

// Implementation of the Retreiver interface enables querying exercises
type Retreiver interface {
	// ExerciseByID takes an excercise id and returns the exercise
	// from the exercises repository if it exists
	WithID(id int) (Exercise, error)
}

// Implementation of the Storer interface enables creating new exercises
type Storer interface {
	// StoreExercise stores an exercise at the tail,
	// updating the references and returns its id.
	New(owner, workout int, name string, weight float64, repetitions int) (int, error)
}

// Implemention of Updater interface enables updating exercises
type Updater interface {
	// ChangeName updates the name of an existing exercise
	ChangeName(id int, newName string) error

	// UpdateWeight updates the weight of an existing exercise
	UpdateWeight(id int, newWeight float64) error

	// UpdateRepetitions updates the repetitions of an existing exercise
	UpdateRepetitions(id int, newRepetitions int) error
}

type Deleter interface {
	// DeleteExercise deletes an exercise if exists
	// and updates the references in the linked list
	Delete(id int) error
}

// Implementation of the Orderer interface allows to reorder exercise positions
type Orderer interface {
	// SwapExercises swaps the position of two exercises
	// that belong to a user workout
	Swap(id1 int, id2 int) error
}
