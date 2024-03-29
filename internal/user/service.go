package user

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/scrot/musclemem-api/internal/api"
)

func NewLoginHandler(l *slog.Logger, users Retreiver) http.Handler {
	l = l.With("handler", "LoginHandler")

	type input struct {
		Username string
		Password string
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i, err := api.ReadJSON[input](r)
		if err != nil {
			WriteUnauthorizedError(l, w, err)
			return
		}

		authenticated, err := users.Authenticate(i.Username, i.Password)
		if err != nil {
			if errors.Is(err, ErrWrongPassword) {
				WriteUnauthorizedError(l, w, err)
				return
			}
			api.WriteInternalError(l, w, err, "")
			return
		}

		if err := api.WriteJSON(w, http.StatusOK, authenticated); err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}
	})
}

func NewCreateHandler(l *slog.Logger, users Storer) http.Handler {
	l = l.With("handler", "CreateHandler")

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			user, err := api.ReadJSON[User](r)
			if err != nil {
				api.WriteInternalError(l, w, err, "")
			}

			l := l.With("username", user.Username, "email", user.Email, "password", user.Password)

			l.Debug("create new user")

			u, err := users.New(user.Username, user.Email, user.Password)
			if err != nil {
				if strings.Contains(err.Error(), "UNIQUE constraint failed") {
					msg := fmt.Sprintf("user %q already exists", u.Username)
					http.Error(w, msg, http.StatusConflict)
				}
				api.WriteInternalError(l, w, err, "")
				return
			}

			if err := api.WriteJSON(w, http.StatusOK, u); err != nil {
				api.WriteInternalError(l, w, err, "")
				return
			}
		},
	)
}

func WriteUnauthorizedError(l *slog.Logger, w http.ResponseWriter, err error) {
	l.Error(err.Error())
	http.Error(w, "invalid credentials", http.StatusUnauthorized)
}
