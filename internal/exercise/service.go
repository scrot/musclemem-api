package exercise

import (
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

	l.Debug(fmt.Sprintf("fetched exercise %s", e.Key()))

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
// requires {username} and {workout} path variables
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

	l.Debug("fetched exercises", "count", len(xs))

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

// HandleChangeExercise
func (s *Service) HandleChangeExercise(w http.ResponseWriter, r *http.Request) {
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

	ei, err := strconv.Atoi(exercise)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)

	var patch Exercise
	if err := dec.Decode(&patch); err != nil {
		writeInternalError(l, w, err)
		return
	}

	var ex Exercise
	switch {
	case patch.Name != "":
		ex, err = s.exercises.ChangeName(username, wi, ei, patch.Name)
		if err != nil {
			writeInternalError(l, w, err)
			return
		}
		fallthrough
	case patch.Weight > 0:
		ex, err = s.exercises.UpdateWeight(username, wi, ei, patch.Weight)
		if err != nil {
			writeInternalError(l, w, err)
			return
		}
		fallthrough
	case patch.Repetitions > 0:
		ex, err = s.exercises.UpdateRepetitions(username, wi, ei, patch.Repetitions)
		if err != nil {
			writeInternalError(l, w, err)
			return
		}
	default:
		writeInternalError(l, w, errors.New("nothing to update"))
		return
	}

	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(ex); err != nil {
		writeInternalError(l, w, err)
		return
	}
}

// HandleMoveUpExercise moves a workout exercise one position up
// reducing the index with 1 but never lower than 1
// requires {username}, {workout}, and {exercise} path variables
func (s *Service) HandleMoveUpExercise(w http.ResponseWriter, r *http.Request) {
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

	if ei1 > 1 {
		ei2 := ei1 - 1
		l = l.With("to-index", ei2)
		if err := s.exercises.Swap(username, wi, ei1, ei2); err != nil {
			writeInternalError(l, w, err)
			return
		}
	} else {
		writeInternalError(l, w, errors.New("already first exercise"))
		return
	}

	l.Debug("moved exercise up")
}

// HandleMoveDownExercise moves a workout exercise one position down
// increasing the exercise index with 1 but never higher than the exercise count
// requires {username}, {workout}, and {exercise} path variables
func (s *Service) HandleMoveDownExercise(w http.ResponseWriter, r *http.Request) {
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

	xs, err := s.exercises.ByWorkout(username, wi)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	if ei1 < len(xs)-1 {
		ei2 := ei1 + 1
		l = l.With("to-index", ei2)

		if err := s.exercises.Swap(username, wi, ei1, ei2); err != nil {
			writeInternalError(l, w, err)
			return
		}
	} else {
		writeInternalError(l, w, errors.New("already last exercise"))
		return
	}

	l.Debug("moved exercise down")
}

// HandleSwapExercises swaps the index of the exercise with one provided
// requires {username}, {workout}, and {exercise} path variables
// requires json payload {"with": INDEX}
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

	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)

	data := struct {
		With int `json:"with"`
	}{}
	if err := dec.Decode(&data); err != nil {
		writeInternalError(l, w, err)
		return
	}

	ei2 := data.With
	l = l.With("to-index", ei2)

	if err := s.exercises.Swap(username, wi, ei1, ei2); err != nil {
		writeInternalError(l, w, err)
		return
	}

	l.Debug("swapped exercises")
}

func writeInternalError(l *slog.Logger, w http.ResponseWriter, err error) {
	msg := fmt.Errorf("%w", err).Error()
	l.Error(msg)
	http.Error(w, msg, http.StatusInternalServerError)
}
