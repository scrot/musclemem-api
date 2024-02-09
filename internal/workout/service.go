package workout

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/scrot/musclemem-api/internal/xhttp"
)

// Service represents a workout service
type Service struct {
	logger   *slog.Logger
	workouts Workouts
}

// NewService returns a new workout service that
// can interact with the workouts datastore
func NewService(l *slog.Logger, ws Workouts) *Service {
	return &Service{
		logger:   l.With("svc", "workout"),
		workouts: ws,
	}
}

// HandleWorkouts returns a list of all workouts of the user in the parameters
func (s Service) HandleWorkouts(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	l := s.logger.With("user", username)

	ws, err := s.workouts.ByOwner(username)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}
	l.Debug("fetched user workouts", "count", len(ws))

	wsJson, err := json.Marshal(ws)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(wsJson); err != nil {
		writeInternalError(l, w, err)
		return
	}
}

// HandleNewWorkout creates a new workout given a user
// multiple workouts can be added using an json array
func (s *Service) HandleNewWorkout(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	l := s.logger.With("user", username)

	js, typ, err := xhttp.RequestJSON(r)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	var responseBody []byte
	switch typ {
	case xhttp.TypeJSONObject:
		var wo Workout
		if err := json.Unmarshal(js, &wo); err != nil {
			writeInternalError(l, w, err)
			return
		}

		nwo, err := s.workouts.New(username, wo.Name)
		if err != nil {
			writeInternalError(l, w, err)
			return
		}

		responseBody, err = json.Marshal(nwo)
		if err != nil {
			writeInternalError(l, w, err)
			return
		}

		l.Debug(fmt.Sprintf("workout %s created", nwo.Key()))
	case xhttp.TypeJSONArray:
		var ws []Workout
		if err := json.Unmarshal(js, &ws); err != nil {
			writeInternalError(l, w, err)
			return
		}

		var nws []Workout
		for _, wo := range ws {
			nwo, err := s.workouts.New(username, wo.Name)
			if err != nil {
				writeInternalError(l, w, err)
				return
			}
			nws = append(nws, nwo)
		}

		var err error
		responseBody, err = json.Marshal(nws)
		if err != nil {
			writeInternalError(l, w, err)
			return
		}

		l.Debug(fmt.Sprintf("%d workouts created", len(nws)))
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

func writeInternalError(l *slog.Logger, w http.ResponseWriter, err error) {
	msg := fmt.Errorf("%w", err).Error()
	l.Error(msg)
	http.Error(w, msg, http.StatusInternalServerError)
}
