package exercise

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/scrot/jsonapi"
)

// Exercise contains details of a single workout exercise
// an exercise is a node in a linked list, determining the order
type Exercise struct {
	ID          int
	Workout     int     `json:"workout-id" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Weight      float64 `json:"weight" validate:"required"`
	Repetitions int     `json:"repetitions" validate:"required"`
	NextID      int
	PreviousID  int
}

// Next returns next Exercise node in linked list
// returns empty exercise if already at the tail
func (e Exercise) Next(r Retreiver) Exercise {
	x, err := r.WithID(e.NextID)
	if err != nil {
		return Exercise{}
	}
	return x
}

// Previous returns previous Exercise node in linked list
// returns empty exercise if already at the head
func (e Exercise) Previous(r Retreiver) Exercise {
	x, err := r.WithID(e.PreviousID)
	if err != nil {
		return Exercise{}
	}
	return x
}

type Service struct {
	logger    *slog.Logger
	exercises Exercises
}

func NewService(l *slog.Logger, xs Exercises) *Service {
	return &Service{
		logger:    l,
		exercises: xs,
	}
}

var ErrInvalidJSON = errors.New("invalid json")

// HandleSingleExercise handles the request for a single exercise
// returning the details of exercise as json given an exerciseID
func (s *Service) HandleSingleExercise(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "exerciseID")

	s.logger.Debug("new single exercise request", "id", idParam, "path", r.URL.Path)

	id, err := strconv.Atoi(idParam)
	if err != nil {
		msg := fmt.Sprintf("%d not a valid id: %s", id, err)
		s.logger.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
	}

	if id == 0 {
		s.writeInternalError(w, err)
		return
	}

	exercise, err := s.exercises.WithID(id)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	payload, err := json.Marshal(&exercise)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(payload); err != nil {
		msg := fmt.Errorf("exercise response error: %w", err).Error()
		s.logger.Error(msg)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s Service) HandleNewExercise(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("new create new exercise request")

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		s.writeInternalError(w, err)
		return
	}

	s.logger.Debug("encoded body", "payload", buf.String())

	var e Exercise
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		s.writeInternalError(w, err)
		return
	}

	id, err := s.exercises.New(e.Workout, e.Name, e.Weight, e.Repetitions)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	idJSON, err := json.Marshal(id)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)
	if _, err := w.Write(idJSON); err != nil {
		s.writeInternalError(w, err)
		return
	}
}

func (s *Service) writeInternalError(w http.ResponseWriter, err error) {
	msg := fmt.Errorf("exercise handler error: %w", err).Error()
	s.logger.Error(msg)
	http.Error(w, msg, http.StatusInternalServerError)
}
