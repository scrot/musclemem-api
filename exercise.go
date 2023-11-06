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
	ErrInvalidIdFormat  = errors.New("id is incorrectly formatted")
	ErrExerciseNotFound = errors.New("exercise not found")
)

type ExerciseRetreiver interface {
	ExerciseByID(uuid.UUID) (Exercise, error)
}

// HandleSingleExercise handles the request for a single exercise
// returning the details of exercise as json given an exerciseID
func (s *Server) HandleSingleExercise(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "exerciseID")

	s.logger.Debug("exercise %s requested", "id", id, "path", r.URL.Path)

	e, err := FetchSingleExerciseJSON(s.exercises, id)
	if err != nil {
		msg := fmt.Errorf("exercise with id %s: %w", id, err).Error()
		s.logger.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
	}

	payload, err := json.Marshal(&e)
	if err != nil {
		msg := fmt.Errorf("marshal exercise %s: %w", id, err).Error()
		s.logger.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(payload); err != nil {
		msg := fmt.Errorf("writing response: %w", err).Error()
		s.logger.Error(msg)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

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
