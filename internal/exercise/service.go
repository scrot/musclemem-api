package exercise

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	apiutils "github.com/scrot/musclemem-api/internal/server"
)

func NewFetchHandler(l *slog.Logger, exercises Retreiver) http.Handler {
	l = l.With("handler", "FetchHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			username = r.PathValue("username")
			workout  = r.PathValue("workout")
			exercise = r.PathValue("exercise")
		)

		l := l.With("user", username, "workout", workout)

		wi, err := strconv.Atoi(r.PathValue("workout"))
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		ei, err := strconv.Atoi(exercise)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		e, err := exercises.ByID(username, wi, ei)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		l.Debug(fmt.Sprintf("fetched exercise %s", e.Key()))

		payload, err := json.Marshal(&e)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if _, err := w.Write(payload); err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}
	})
}

func NewFetchAllHandler(l *slog.Logger, exercises Retreiver) http.Handler {
	l = l.With("handler", "FetchAllHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			username = r.PathValue("username")
			workout  = r.PathValue("workout")
		)

		l := l.With("user", username, "workout", workout)

		wi, err := strconv.Atoi(workout)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		xs, err := exercises.ByWorkout(username, wi)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		l.Debug("fetched exercises", "count", len(xs))

		xsJSON, err := json.Marshal(xs)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(xsJSON)
	})
}

func NewCreateHandler(l *slog.Logger, exercises Storer) http.Handler {
	l = l.With("handler", "CreateHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			username = r.PathValue("username")
			workout  = r.PathValue("workout")
		)

		l := l.With("user", username, "workout", workout)

		wid, err := strconv.Atoi(workout)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		js, typ, err := apiutils.RequestJSON(r)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		var responseBody []byte
		switch typ {
		case apiutils.TypeJSONObject:
			var ex Exercise
			if err := json.Unmarshal(js, &ex); err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}

			nex, err := exercises.New(username, wid, ex.Name, ex.Weight, ex.Repetitions)
			if err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}

			responseBody, err = json.Marshal(nex)
			if err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}

			l.Debug(fmt.Sprintf("exercise %s created", nex.Key()))
		case apiutils.TypeJSONArray:
			var xs []Exercise
			if err := json.Unmarshal(js, &xs); err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}

			var nxs []Exercise
			for _, ex := range xs {
				nex, err := exercises.New(username, wid, ex.Name, ex.Weight, ex.Repetitions)
				if err != nil {
					apiutils.WriteInternalError(l, w, err, "")
					return
				}
				nxs = append(nxs, nex)
			}

			var err error
			responseBody, err = json.Marshal(nxs)
			if err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}

			l.Debug(fmt.Sprintf("%d exercises created", len(nxs)))
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

func NewDeleteHandler(l *slog.Logger, exercises Deleter) http.Handler {
	l = l.With("handler", "DeleteHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			username = r.PathValue("username")
			workout  = r.PathValue("workout")
			exercise = r.PathValue("exercise")
		)

		l := l.With("user", username, "workout", workout, "exercise", exercise)

		wi, err := strconv.Atoi(workout)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		ei, err := strconv.Atoi(exercise)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		e, err := exercises.Delete(username, wi, ei)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		w.Header().Add("Content-Type", "application/json")

		// TODO: write into buffer first to prevent double writing
		// while writing the error message on err
		enc := json.NewEncoder(w)
		if err := enc.Encode(e); err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}
	})
}

func NewUpdateHandler(l *slog.Logger, exercises Updater) http.Handler {
	l = l.With("handler", "UpdateHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			username = r.PathValue("username")
			workout  = r.PathValue("workout")
			exercise = r.PathValue("exercise")
		)

		l := l.With("user", username, "workout", workout, "exercise", exercise)

		wi, err := strconv.Atoi(workout)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		ei, err := strconv.Atoi(exercise)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		defer r.Body.Close()
		dec := json.NewDecoder(r.Body)

		var patch Exercise
		if err := dec.Decode(&patch); err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		var ex Exercise
		switch {
		case patch.Name != "":
			cx, err := exercises.ChangeName(username, wi, ei, patch.Name)
			if err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}
			ex.Name = cx.Name
			fallthrough
		case patch.Weight > 0:
			cx, err := exercises.UpdateWeight(username, wi, ei, patch.Weight)
			if err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}
			ex.Weight = cx.Weight
			fallthrough
		case patch.Repetitions > 0:
			cx, err := exercises.UpdateRepetitions(username, wi, ei, patch.Repetitions)
			if err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}
			ex.Repetitions = cx.Repetitions
		default:
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		w.Header().Add("Content-Type", "application/json")

		enc := json.NewEncoder(w)
		if err := enc.Encode(ex); err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}
	})
}

// HandleMoveUpExercise moves a workout exercise one position up
// reducing the index with 1 but never lower than 1
// requires {username}, {workout}, and {exercise} path variables
func NewUpHandler(l *slog.Logger, exercises Orderer) http.Handler {
	l = l.With("handler", "UpHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			username = r.PathValue("username")
			workout  = r.PathValue("workout")
			exercise = r.PathValue("exercise")
		)

		l := l.With("user", username, "workout", workout, "exercise", exercise)

		wi, err := strconv.Atoi(workout)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		ei1, err := strconv.Atoi(exercise)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		if ei1 > 1 {
			ei2 := ei1 - 1
			l = l.With("to-index", ei2)
			if err := exercises.Swap(username, wi, ei1, ei2); err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}
		} else {
			err := fmt.Errorf("index %d out of range", ei1)
			apiutils.WriteInternalError(l, w, err, "index already at the top")
			return
		}

		l.Debug("moved exercise up")
	})
}

// HandleMoveDownExercise moves a workout exercise one position down
// increasing the exercise index with 1 but never higher than the exercise count
// requires {username}, {workout}, and {exercise} path variables
func NewDownHandler(l *slog.Logger, exercises Orderer) http.Handler {
	l = l.With("handler", "DownHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			username = r.PathValue("username")
			workout  = r.PathValue("workout")
			exercise = r.PathValue("exercise")
		)

		l := l.With("user", username, "workout", workout, "exercise", exercise)

		wi, err := strconv.Atoi(workout)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		ei1, err := strconv.Atoi(exercise)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		count, err := exercises.Len(username, wi)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		if ei1 < count-1 {
			ei2 := ei1 + 1
			l = l.With("to-index", ei2)

			if err := exercises.Swap(username, wi, ei1, ei2); err != nil {
				apiutils.WriteInternalError(l, w, err, "")
				return
			}
		} else {
			err := fmt.Errorf("index %d out of range", ei1)
			apiutils.WriteInternalError(l, w, err, "index already at bottom")
			return
		}

		l.Debug("moved exercise down")
	})
}

// HandleSwapExercises swaps the index of the exercise with one provided
// requires {username}, {workout}, and {exercise} path variables
// requires json payload {"with": INDEX}
func NewSwapHandler(l *slog.Logger, exercises Orderer) http.Handler {
	l = l.With("handler", "SwapHandler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			username = r.PathValue("username")
			workout  = r.PathValue("workout")
			exercise = r.PathValue("exercise")
		)

		l := l.With("user", username, "workout", workout, "exercise", exercise)

		wi, err := strconv.Atoi(workout)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		ei1, err := strconv.Atoi(exercise)
		if err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		defer r.Body.Close()
		dec := json.NewDecoder(r.Body)

		data := struct {
			With int `json:"with"`
		}{}
		if err := dec.Decode(&data); err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		ei2 := data.With
		l = l.With("to-index", ei2)

		if err := exercises.Swap(username, wi, ei1, ei2); err != nil {
			apiutils.WriteInternalError(l, w, err, "")
			return
		}

		l.Debug("swapped exercises")
	})
}
