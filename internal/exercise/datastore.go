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
	ErrNoID          = errors.New("no (valid) id provided")
	ErrNotFound      = errors.New("exercise not found")
	ErrMissingFields = errors.New("missing required fields")
)

// Implementation of Exercises provide the means to persist
// exercises in a repository
type Exercises interface {
	Retreiver
	Storer
	Orderer
}

// Implementation of the Retreiver interface enables querying exercises
type Retreiver interface {
	// ExerciseByID takes an excercise id and returns the exercise
	// from the exercises repository if it exists
	WithID(int) (Exercise, error)

	// ExercisesByWorkoutID takes a owner and workout id and returns all
	// exercises from the repository that belong to it
	FromWorkout(int, int) ([]Exercise, error)
}

// Implementation of the Storer interface enables manipulating exercises
type Storer interface {
	// StoreExercise stores an exercise at the tail,
	// updating the references and returns its id.
	// it overwrites the exercise if it already exists
	Store(Exercise) (int, error)

	// DeleteExercise deletes an exercise if exists
	// updates the references of the previous and next
	// exercise
	Delete(int) error
}

// Implementation of the Orderer interface allows to reorder exercise positions
type Orderer interface {
	// SwapExercises swaps the position of two exercises
	Swap(int, int) error
}
