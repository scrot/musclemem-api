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

// HandleNewUser creates a new user if the e-mail is not already registered
func (s Service) HandleNewUser(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("new create new user request")

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

	id, err := s.users.Register(u.Email, u.Password)
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
