package workout

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/scrot/musclemem-api/internal/api"
)

func NewFetchAllHandler(l *slog.Logger, workouts Retreiver) http.Handler {
	l = l.With("handler", "FetchAllHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.PathValue("username")

		l := l.With("user", username)

		ws, err := workouts.ByOwner(username)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}
		l.Debug("fetched user workouts", "count", len(ws))

		if err := api.WriteJSON(w, http.StatusOK, ws); err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}
	})
}

func NewCreateHandler(l *slog.Logger, workouts Storer) http.Handler {
	l = l.With("handler", "CreateHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.PathValue("username")
		l := l.With("user", username)

		wo, err := api.ReadJSON[Workout](r)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		nwo, err := workouts.New(username, wo.Name)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		l.Debug(fmt.Sprintf("workout %s created", nwo.Ref()))

		if err := api.WriteJSON(w, http.StatusOK, nwo); err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}
	})
}

func NewDeleteHandler(l *slog.Logger, workouts Deleter) http.Handler {
	l = l.With("handler", "DeleteHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			username = r.PathValue("username")
			workout  = r.PathValue("workout")
		)

		l := l.With("username", username, "workout", workout)

		windex, err := strconv.Atoi(workout)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		deleted, err := workouts.Delete(username, windex)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		l.Debug("workout deleted", "key", deleted.Ref())

		if err := api.WriteJSON(w, http.StatusOK, deleted); err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}
	})
}

func NewUpdateHandler(l *slog.Logger, workouts Updater) http.Handler {
	l = l.With("handler", "UpdateHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			username = r.PathValue("username")
			workout  = r.PathValue("workout")
		)

		l := l.With("username", username, "workout", workout)

		wid, err := strconv.Atoi(workout)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		defer r.Body.Close()
		dec := json.NewDecoder(r.Body)

		var patch Workout
		if err := dec.Decode(&patch); err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		if patch.Name != "" {
			l = l.With("name", patch.Name)
			updated, err := workouts.ChangeName(username, wid, patch.Name)
			if err != nil {
				api.WriteInternalError(l, w, err, "nothing to update")
				return
			}
			if err := api.WriteJSON(w, http.StatusOK, updated); err != nil {
				api.WriteInternalError(l, w, err, "")
				return
			}
		} else {
			api.WriteInternalError(l, w, err, "")
			return
		}
	})
}
