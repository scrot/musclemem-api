package workout

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
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

func (s *Service) HandleExercises(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.logger.Debug("new exercises request", "id", id)

	wid, err := strconv.Atoi(id)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	ws, err := s.workouts.ByID(wid)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	xs := ws.Exercises(s.workouts)
	s.logger.Debug("fetching exercises", "count", len(xs))

	xsJSON, err := json.Marshal(xs)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(xsJSON)
}

// HandleNewWorkout creates a new workout given a user
func (s *Service) HandleNewWorkout(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("new create new workout request")

	var buf bytes.Buffer
	defer r.Body.Close()
	if _, err := buf.ReadFrom(r.Body); err != nil {
		s.writeInternalError(w, err)
		return
	}

	var wo Workout
	if err := json.Unmarshal(buf.Bytes(), &wo); err != nil {
		s.writeInternalError(w, err)
		return
	}

	s.logger.Debug("decoded request body", "payload", wo)

	id, err := s.workouts.New(wo.Owner, wo.Name)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	idJSON, err := json.Marshal(id)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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
