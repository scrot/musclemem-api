package exercise

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/scrot/musclemem-api/internal/xhttp"
)

type Service struct {
	logger    *slog.Logger
	exercises Exercises
}

// NewService creates a new http/json service for handling exercises
// It interacts with a storage controller implementing the Exercises
// interface
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
	var (
		username = r.PathValue("username")
		workout  = r.PathValue("workout")
		exercise = r.PathValue("exercise")
	)

	l := s.logger.With("user", username, "workout", workout)

	wi, err := strconv.Atoi(r.PathValue("workout"))
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	ei, err := strconv.Atoi(exercise)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	e, err := s.exercises.ByID(username, wi, ei)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	payload, err := json.Marshal(&e)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write(payload); err != nil {
		writeInternalError(l, w, err)
		return
	}
}

// HandleExercises retrieves all exercises of a user's workout
func (s *Service) HandleExercises(w http.ResponseWriter, r *http.Request) {
	var (
		username = r.PathValue("username")
		workout  = r.PathValue("workout")
	)

	l := s.logger.With("user", username, "workout", workout)

	wi, err := strconv.Atoi(workout)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	xs, err := s.exercises.ByWorkout(username, wi)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	l.Debug("fetching exercises", "count", len(xs))

	xsJSON, err := json.Marshal(xs)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(xsJSON)
}

// HandleNewExercise creates a new exercise given a workout
// to batch add multiple exercises, use a json array in the request body
// the exercise(s) are added add the end of the workout by default
func (s Service) HandleNewExercise(w http.ResponseWriter, r *http.Request) {
	var (
		username = r.PathValue("username")
		workout  = r.PathValue("workout")
	)

	l := s.logger.With("user", username, "workout", workout)

	wid, err := strconv.Atoi(workout)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	js, typ, err := xhttp.RequestJSON(r)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	var responseBody []byte
	switch typ {
	case xhttp.TypeJSONObject:
		var ex Exercise
		if err := json.Unmarshal(js, &ex); err != nil {
			writeInternalError(l, w, err)
			return
		}

		nex, err := s.exercises.New(username, wid, ex.Name, ex.Weight, ex.Repetitions)
		if err != nil {
			writeInternalError(l, w, err)
			return
		}

		responseBody, err = json.Marshal(nex)
		if err != nil {
			writeInternalError(l, w, err)
			return
		}

		l.Debug(fmt.Sprintf("exercise %s created", nex.Key()))
	case xhttp.TypeJSONArray:
		var xs []Exercise
		if err := json.Unmarshal(js, &xs); err != nil {
			writeInternalError(l, w, err)
			return
		}

		var nxs []Exercise
		for _, ex := range xs {
			nex, err := s.exercises.New(username, wid, ex.Name, ex.Weight, ex.Repetitions)
			if err != nil {
				writeInternalError(l, w, err)
				return
			}
			nxs = append(nxs, nex)
		}

		var err error
		responseBody, err = json.Marshal(nxs)
		if err != nil {
			writeInternalError(l, w, err)
			return
		}

		l.Debug(fmt.Sprintf("%d exercises created", len(nxs)))
	default:
		writeInternalError(l, w, errors.New("invalid json in request body"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(responseBody); err != nil {
		writeInternalError(l, w, err)
		return
	}
}

func (s *Service) createExercise(l *slog.Logger, w http.ResponseWriter, e Exercise) {
}

func (s *Service) HandleSwapExercises(w http.ResponseWriter, r *http.Request) {
	var (
		username = r.PathValue("username")
		workout  = r.PathValue("workout")
		exercise = r.PathValue("exercise")
	)

	l := s.logger.With("user", username, "workout", workout, "exercise", exercise)

	wi, err := strconv.Atoi(workout)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	ei1, err := strconv.Atoi(exercise)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	var buf bytes.Buffer
	defer r.Body.Close()
	if _, err := buf.ReadFrom(r.Body); err != nil {
		writeInternalError(l, w, err)
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		writeInternalError(l, w, err)
		return
	}

	val, ok := payload["with"]
	if !ok {
		writeInternalError(l, w, fmt.Errorf("require body parameters %q", "with"))
		return
	}

	ei2, ok := val.(int)
	if !ok {
		writeInternalError(l, w, fmt.Errorf("body parameter should be a digit"))
		return
	}

	l = l.With("with", ei2)

	if err := s.exercises.Swap(username, wi, ei1, ei2); err != nil {
		writeInternalError(l, w, err)
		return
	}
}

func writeInternalError(l *slog.Logger, w http.ResponseWriter, err error) {
	msg := fmt.Errorf("%w", err).Error()
	l.Error(msg)
	http.Error(w, msg, http.StatusInternalServerError)
}
