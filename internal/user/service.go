package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type Service struct {
	logger *slog.Logger
	users  Users
}

func NewService(l *slog.Logger, us Users) *Service {
	return &Service{
		logger: l.With("svc", "user"),
		users:  us,
	}
}

// HandleWorkouts returns a list of all workouts of the user in the parameters
func (s Service) HandleWorkouts(w http.ResponseWriter, r *http.Request) {
	u := r.PathValue("username")
	s.logger.Debug("user workouts request", "user", u)

	ws, err := s.users.UserWorkouts(u)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}
	s.logger.Debug("fetched user workouts", "count", len(ws))

	wsJson, err := json.Marshal(ws)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(wsJson); err != nil {
		s.writeInternalError(w, err)
		return
	}
}

// HandleNewUser creates a new user if the e-mail is not already registered
func (s Service) HandleNewUser(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("create new user request")

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		s.writeInternalError(w, err)
		return
	}

	var u User
	if err := json.Unmarshal(buf.Bytes(), &u); err != nil {
		s.writeInternalError(w, err)
		return
	}

	s.logger.Debug("decoded request body", "payload", u)

	id, err := s.users.Register(u.Username, u.Email, u.Password)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			s.writeInternalError(w, ErrUserExists)
		} else {
			s.writeInternalError(w, err)
		}
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

func (s Service) writeInternalError(w http.ResponseWriter, err error) {
	s.logger.Error(err.Error())
	http.Error(w, "whoops!", http.StatusInternalServerError)
}
