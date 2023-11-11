package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/scrot/jsonapi"
)

type Exercise struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Weight      float64   `json:"weight"`
	Repetitions int       `json:"repetitions"`
	Next        uuid.UUID `json:"next,omitempty"`
	Previous    uuid.UUID `json:"previous,omitempty"`
}

var (
	emptyJSON = []byte("{}")
)

var (
	ErrNoIDProvided     = errors.New("no exercise id provided")
	ErrInvalidIdFormat  = errors.New("id incorrectly formatted")
	ErrExerciseNotFound = errors.New("exercise not found")
)

// ExerciseRetreiver is an interface for retreiving an Exercise
// from an exercises repository
type ExerciseRetreiver interface {
	ExerciseByID(uuid.UUID) (Exercise, error)
}

// HandleSingleExercise handles the request for a single exercise
// returning the details of exercise as json given an exerciseID
func (s *Server) HandleSingleExercise(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "exerciseID")

	s.logger.Debug("new single exercise request", "id", id, "path", r.URL.Path)

	e, err := FetchSingleExerciseJSON(s.exercises, id)
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
func FetchSingleExerciseJSON(exercises ExerciseRetreiver, id string) ([]byte, error) {
	if id == "" {
		return emptyJSON, ErrNoIDProvided

	}

	exerciseId, err := uuid.Parse(id)
	if err != nil {
		return emptyJSON, ErrInvalidIdFormat
	}

	exercise, err := exercises.ExerciseByID(exerciseId)
	if err != nil {
		return emptyJSON, err
	}

	payload, err := json.Marshal(&exercise)
	if err != nil {
		return emptyJSON, err
	}

	return payload, nil
}

var (
	ErrEmptyExercise = errors.New("empty exercise")
	ErrInvalidJSON   = errors.New("invalid json")
)

// ExerciseStorer is an interface for storing Exercises in
// an exercises repository
type ExerciseStorer interface {
	Exists(uuid.UUID) bool
	StoreExercise(Exercise) (uuid.UUID, error)
}

// StoreExerciseJSON stores an Exercise formatted as JSON in the exercises
// repository provided by the ExerciseStorer. If the JSON contains an ID
// it will overwrite, or throws an error if the ID doesn't exist
func StoreExerciseJSON(exercises ExerciseStorer, exerciseJSON []byte) (uuid.UUID, error) {
	var e Exercise

	// also throws the error if the id is of an invalid uuid lenght
	if err := json.Unmarshal(exerciseJSON, &e); err != nil {
		return uuid.Nil, ErrInvalidJSON
	}

	if (e == Exercise{}) {
		return uuid.Nil, ErrEmptyExercise
	}

	if e.ID != uuid.Nil && !exercises.Exists(e.ID) {
		return uuid.Nil, ErrExerciseNotFound
	}

	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}

	return exercises.StoreExercise(e)
}
