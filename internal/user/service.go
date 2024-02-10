package user

import (
	"encoding/json"
	"fmt"
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
	l := s.logger

	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)

	var u User
	if err := dec.Decode(&u); err != nil {
		writeInternalError(l, w, err)
		return
	}

	l.Debug("decoded request body", "payload", u)

	newUser, err := s.users.New(u.Username, u.Email, u.Password)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	userJSON, err := json.Marshal(newUser)
	if err != nil {
		writeInternalError(l, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(userJSON); err != nil {
		writeInternalError(l, w, err)
		return
	}
}

func writeInternalError(l *slog.Logger, w http.ResponseWriter, err error) {
	msg := fmt.Errorf("exercise handler error: %w", err).Error()
	l.Error(msg)
	http.Error(w, msg, http.StatusInternalServerError)
}
