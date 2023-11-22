package exercise

import (
	"encoding/json"
	"errors"
	"log/slog"

	_ "github.com/go-playground/validator/v10"
)

type API struct {
	Logger *slog.Logger
	Store  *Datastore
}

type Datastore interface {
	Retreiver
	Storer
	Orderer
}

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

	User struct {
		ID       int
		Email    string
		Password string
	}

	Workout struct {
		ID   int
		Name string
	}

	ExerciseRef struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)

func (e Exercise) Ref() ExerciseRef {
	return ExerciseRef{e.ID, e.Name}
}

// Implementation of the Retreiver interface expose exercises
type Retreiver interface {
	// ExerciseByID takes an excercise id and returns the exercise
	// from the exercises repository if it exists
	ExerciseByID(int) (Exercise, error)

	// ExercisesByWorkoutID takes a owner and workout id and returns all
	// exercises from the repository that belong to it
	ExercisesByWorkoutID(int, int) ([]Exercise, error)
}

var EmptyJSON = []byte("{}")

var (
	ErrNoID            = errors.New("no exercise id provided")
	ErrInvalidIdFormat = errors.New("id incorrectly formatted")
	ErrNotFound        = errors.New("exercise not found")
)

// FetchSingleExerciseJSON takes a exercise repository and an id and returns
// an exercise in JSON format. the id must be a valid UUID or ErrInvalidIdFormat is returned
// if no exercise with that id is found an ErrExerciseNotFound is returned
func FetchSingleExerciseJSON(exercises Retreiver, id int) ([]byte, error) {
	if id == 0 {
		return EmptyJSON, ErrNoID
	}

	exercise, err := exercises.ExerciseByID(id)
	if err != nil {
		return EmptyJSON, err
	}

	payload, err := json.Marshal(&exercise)
	if err != nil {
		return EmptyJSON, err
	}

	return payload, nil
}

var (
	ErrMissingFields = errors.New("fields missing")
	ErrInvalidJSON   = errors.New("invalid json")
)

// Implementation of the Storer interface manipulates exercises
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

// StoreExerciseJSON stores an Exercise formatted as JSON
// in the exercises repository provided by the ExerciseStorer.
// If an ID is provided, it will overwrite it if it exists or throws an error
func StoreExerciseJSON(exercises Storer, exerciseJSON []byte) error {
	var e Exercise

	// also throws the error if the id is of an invalid uuid length
	if err := json.Unmarshal(exerciseJSON, &e); err != nil {
		return ErrInvalidJSON
	}

	exercises.Store(e)

	return nil
}

// Implementation of the Orderer interface allows to order
// exercises in the repository
type Orderer interface {
	// SwapExercises swaps the position of two exercises
	Swap(int, int) error
}
