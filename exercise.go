package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/scrot/jsonapi"
)

type Exercise struct {
	ID          int       `jsonapi:"primary,exercises"`
	Name        string    `jsonapi:"attr,name"`
	Weight      float64   `jsonapi:"attr,weight"`
	Repetitions int       `jsonapi:"attr,repetitions"`
	Next        *Exercise `jsonapi:"relation,next,omitempty"`
	Previous    *Exercise `jsonapi:"relation,previous,omitempty"`
}

var (
	ErrExerciseNotFound = errors.New("exercise not found")
)

type ExerciseStorer interface {
	ExerciseByID(int) (Exercise, error)
}

func (e Exercise) JSONAPILinks() *jsonapi.Links {
	ls := jsonapi.Links{
		"self": fmt.Sprintf("/exercises/%d", e.ID),
	}
	return &ls
}

func (e Exercise) JSONAPIRelationshipLinks(relation string) *jsonapi.Links {
	switch relation {
	case "next_exercise":
		return &jsonapi.Links{
			"related": fmt.Sprintf("/exercises/%d", e.Next.ID),
		}
	case "previous_exercise":
		return &jsonapi.Links{
			"related": fmt.Sprintf("/exercises/%d", e.Previous.ID),
		}
	default:
		return nil
	}
}

func (s *Server) HandleExerciseDetails(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "exerciseID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	s.logger.Debug("exercise details request", "id", id, "path", r.URL.Path)

	e, err := s.exercises.ExerciseByID(id)
	if err != nil {
		msg := fmt.Errorf("exercise with id %d: %w", id, err).Error()
		s.logger.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
	}

	var buf bytes.Buffer
	if err := jsonapi.MarshalPayload(&buf, &e); err != nil {
		msg := fmt.Errorf("marshal exercise %d: %w", id, err).Error()
		s.logger.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(http.StatusOK)
	if _, err := buf.WriteTo(w); err != nil {
		msg := fmt.Errorf("writing response: %w", err).Error()
		s.logger.Error(msg)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
