package workout

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	apiutils "github.com/scrot/musclemem-api/internal/server"
)

func NewFetchAllHandler(l *slog.Logger, workouts Retreiver) http.Handler {
	l = l.With("handler", "FetchAllHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.PathValue("username")

		l := l.With("user", username)

		ws, err := workouts.ByOwner(username)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}
		l.Debug("fetched user workouts", "count", len(ws))

		wsJson, err := json.Marshal(ws)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		w.Header().Add("Content-Type", "application/json")
		if _, err := w.Write(wsJson); err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}
	})
}

func NewCreateHandler(l *slog.Logger, workouts Storer) http.Handler {
	l = l.With("handler", "CreateHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.PathValue("username")
		l := l.With("user", username)

		js, typ, err := apiutils.RequestJSON(r)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		var responseBody []byte
		switch typ {
		case apiutils.TypeJSONObject:
			var wo Workout
			if err := json.Unmarshal(js, &wo); err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}

			nwo, err := workouts.New(username, wo.Name)
			if err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}

			responseBody, err = json.Marshal(nwo)
			if err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}

			l.Debug(fmt.Sprintf("workout %s created", nwo.Key()))
		case apiutils.TypeJSONArray:
			var ws []Workout
			if err := json.Unmarshal(js, &ws); err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}

			var nws []Workout
			for _, wo := range ws {
				nwo, err := workouts.New(username, wo.Name)
				if err != nil {
					apiutils.WriteInternalError(l, w, err, "")
					return
				}
				nws = append(nws, nwo)
			}

			var err error
			responseBody, err = json.Marshal(nws)
			if err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}
			l.Debug(fmt.Sprintf("%d workouts created", len(nws)))
		default:
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(responseBody); err != nil {
			apiutils.WriteInternalError(l, w, err, "")
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
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		delWo, err := workouts.Delete(username, windex)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		w.Header().Add("Content-Type", "application/json")
		respBody, err := json.Marshal(delWo)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		if _, err := w.Write(respBody); err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		l.Debug("deleted workout")
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
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		defer r.Body.Close()
		dec := json.NewDecoder(r.Body)

		var patch Workout
		if err := dec.Decode(&patch); err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		var wo Workout
		switch {
		case patch.Name != "":
			l = l.With("name", patch.Name)
			wo, err = workouts.ChangeName(username, wid, patch.Name)
			if err != nil {
				apiutils.WriteInternalError(l, w, err, "nothing to update")
				return
			}
		default: // remove due to fallthrough
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		w.Header().Add("Content-Type", "application/json")

		enc := json.NewEncoder(w)
		if err := enc.Encode(wo); err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}
	})
}
