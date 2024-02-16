package user

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	apiutils "github.com/scrot/musclemem-api/internal/server"
)

func NewCreateHandler(l *slog.Logger, users Storer) http.Handler {
	l = l.With("handler", "CreateHandler")

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			dec := json.NewDecoder(r.Body)

			var u User
			if err := dec.Decode(&u); err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}

			l.Debug("decoded request body", "payload", u)

			newUser, err := users.New(u.Username, u.Email, u.Password)
			if err != nil {
				if strings.Contains(err.Error(), "UNIQUE constraint failed") {
					msg := fmt.Sprintf("user %q already exists", u.Username)
					http.Error(w, msg, http.StatusConflict)
				}
				apiutils.WriteInternalError(l, w, err, "")
				return
			}

			userJSON, err := json.Marshal(newUser)
			if err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}

			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write(userJSON); err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}
		},
	)
}
