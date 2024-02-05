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

type Service struct {
	logger    *slog.Logger
	exercises Exercises
}

func NewService(l *slog.Logger, xs Exercises) *Service {
	return &Service{
		logger:    l.With("svc", "exercise"),
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

// HandleNewExercise creates a new exercise given a workout
// to batch add multiple exercises, use a json array in the request body
// the exercise(s) are added add the end of the workout by default
func (s Service) HandleNewExercise(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("new create new exercise request")

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		s.writeInternalError(w, err)
		return
	}

	payload := bytes.TrimLeft(buf.Bytes(), "\t\r\n")
	if len(payload) <= 0 {
		s.writeInternalError(w, errors.New("no data in the request body"))
		return
	}

	switch payload[0] {
	case '[':
		var es []Exercise
		if err := json.Unmarshal(payload, &es); err != nil {
			s.writeInternalError(w, err)
			return
		}

		s.logger.Debug("decoded exercises in body", "count", len(es))

		for _, e := range es {
			_, err := s.exercises.New(e.Workout, e.Name, e.Weight, e.Repetitions)
			if err != nil {
				s.logger.Error(fmt.Sprintf("error adding exercise: %s", err), "exercise", e.Name)
			}
		}
	case '{':
		var e Exercise
		if err := json.Unmarshal(payload, &e); err != nil {
			s.writeInternalError(w, err)
			return
		}

		s.logger.Debug("decoded exercise request body", "payload", e)

		_, err := s.exercises.New(e.Workout, e.Name, e.Weight, e.Repetitions)
		if err != nil {
			s.writeInternalError(w, err)
			return
		}
	default:
		s.writeInternalError(w, errors.New("invalid json, first token not { or ["))
		return
	}
}

func (s *Service) writeInternalError(w http.ResponseWriter, err error) {
	s.logger.Error(err.Error())
	http.Error(w, "whoops", http.StatusInternalServerError)
}
