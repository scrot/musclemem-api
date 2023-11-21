package musclememapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/scrot/jsonapi"

	_ "github.com/go-playground/validator/v10"
)

type Exercise struct {
	ID          int         `json:"id" validate:"gt=0"`
	Owner       int         `json:"owner" validate:"required,gt=0"`
	Workout     int         `json:"workout" validate:"required"`
	Name        string      `json:"name" validate:"required"`
	Weight      float64     `json:"weight" validate:"required"`
	Repetitions int         `json:"repetitions" validate:"required"`
	Next        ExerciseRef `json:"next,omitempty"`
	Previous    ExerciseRef `json:"previous,omitempty"`
}

func (e Exercise) Ref() ExerciseRef {
	return ExerciseRef{e.ID, e.Name}
}

type ExerciseRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var EmptyJSON = []byte("{}")

var (
	ErrNoIDProvided     = errors.New("no exercise id provided")
	ErrInvalidIdFormat  = errors.New("id incorrectly formatted")
	ErrExerciseNotFound = errors.New("exercise not found")
)

// ExerciseRetreiver is an interface for retreiving an Exercise
// from an exercises repository
type ExerciseRetreiver interface {
	ExerciseByID(int) (Exercise, error)
}

// HandleSingleExercise handles the request for a single exercise
// returning the details of exercise as json given an exerciseID
func (s *Server) HandleSingleExercise(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "exerciseID")

	s.logger.Debug("new single exercise request", "id", idParam, "path", r.URL.Path)

	id, err := strconv.Atoi(idParam)
	if err != nil {
		msg := fmt.Sprintf("%d not a valid id: %s", id, err)
		s.logger.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)

	}

	e, err := FetchSingleExerciseJSON(s.store, id)
	if err != nil {
		msg := fmt.Errorf("exercise retrieval error: %w", err).Error()
		s.logger.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(e); err != nil {
		msg := fmt.Errorf("exercise response error: %w", err).Error()
		s.logger.Error(msg)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// FetchSingleExerciseJSON takes a exercise repository and an id and returns
// an exercise in JSON format. the id must be a valid UUID or ErrInvalidIdFormat is returned
// if no exercise with that id is found an ErrExerciseNotFound is returned
func FetchSingleExerciseJSON(exercises ExerciseRetreiver, id int) ([]byte, error) {
	if id == 0 {
		return EmptyJSON, ErrNoIDProvided
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

// ExerciseStorer is an interface for storing Exercises in
// an exercises repository
type ExerciseStorer interface {
	ExerciseExists(int) bool
	StoreExercise(Exercise) (int, error)
}

// StoreExerciseJSON stores an Exercise formatted as JSON
// in the exercises repository provided by the ExerciseStorer.
// If an ID is provided, it will overwrite it if it exists or throws an error
func StoreExerciseJSON(exercises ExerciseStorer, exerciseJSON []byte) error {
	var exercise Exercise

	// also throws the error if the id is of an invalid uuid length
	if err := json.Unmarshal(exerciseJSON, &exercise); err != nil {
		return ErrInvalidJSON
	}

	return nil
}

// BatchStoreExerciseJSON stores a list of exercises
// in the exercise repository provided by the ExerciseStorer
// If an ID is provided, it will overwrite it if it exists or throws an error
func BatchStoreExerciseJSON(exercises ExerciseStorer, exerciseJSON []byte) error {
	var xs []Exercise

	// also throws the error if the id is of an invalid uuid length
	if err := json.Unmarshal(exerciseJSON, &xs); err != nil {
		return ErrInvalidJSON
	}

	return nil
}
