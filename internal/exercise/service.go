package exercise

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/scrot/musclemem-api/internal/api"
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
			api.WriteInternalError(l, w, err, "")
			return
		}

		ei, err := strconv.Atoi(exercise)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		fetched, err := exercises.ByID(username, wi, ei)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		l.Debug(fmt.Sprintf("fetched exercise %s", fetched.Ref()))

		if err := api.WriteJSON(w, http.StatusOK, fetched); err != nil {
			api.WriteInternalError(l, w, err, "")
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
			api.WriteInternalError(l, w, err, "")
			return
		}

		xs, err := exercises.ByWorkout(username, wi)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		l.Debug("fetched exercises", "count", len(xs))

		if err := api.WriteJSON(w, http.StatusOK, xs); err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}
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
			api.WriteInternalError(l, w, err, "")
			return
		}

		add, err := api.ReadJSON[Exercise](r)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		created, err := exercises.New(username, wid, add.Name, add.Weight, add.Repetitions)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		l.Debug(fmt.Sprintf("exercise %s created", created.Ref()))

		if err := api.WriteJSON(w, http.StatusOK, created); err != nil {
			api.WriteInternalError(l, w, err, "")
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
			api.WriteInternalError(l, w, err, "")
			return
		}

		ei, err := strconv.Atoi(exercise)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		deleted, err := exercises.Delete(username, wi, ei)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		l.Debug("exercise deleted", "key", deleted.Ref())

		if err := api.WriteJSON(w, http.StatusOK, deleted); err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}
	})
}

// TODO merge all update functions into a single one
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
			api.WriteInternalError(l, w, err, "")
			return
		}

		ei, err := strconv.Atoi(exercise)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		patch, err := api.ReadJSON[Exercise](r)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		var updated Exercise
		if patch.Name != "" {
			l = l.With("name", patch.Name)
			cx, err := exercises.ChangeName(username, wi, ei, patch.Name)
			if err != nil {
				api.WriteInternalError(l, w, err, "")
				return
			}
			updated.Name = cx.Name
		}

		if patch.Weight > 0 {
			l = l.With("weight", patch.Weight)
			cx, err := exercises.UpdateWeight(username, wi, ei, patch.Weight)
			if err != nil {
				api.WriteInternalError(l, w, err, "")
				return
			}
			updated.Weight = cx.Weight

		}

		if patch.Repetitions > 0 {
			l = l.With("repetitions", patch.Repetitions)
			cx, err := exercises.UpdateRepetitions(username, wi, ei, patch.Repetitions)
			if err != nil {
				api.WriteInternalError(l, w, err, "")
				return
			}
			updated.Repetitions = cx.Repetitions
		}

		if err := api.WriteJSON(w, http.StatusOK, updated); err != nil {
			api.WriteInternalError(l, w, err, "")
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
			api.WriteInternalError(l, w, err, "")
			return
		}

		ei1, err := strconv.Atoi(exercise)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		if ei1 > 1 {
			ei2 := ei1 - 1
			l = l.With("to-index", ei2)
			if err := exercises.Swap(username, wi, ei1, ei2); err != nil {
				api.WriteInternalError(l, w, err, "")
				return
			}
		} else {
			err := fmt.Errorf("index %d out of range", ei1)
			api.WriteInternalError(l, w, err, "index already at the top")
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
			api.WriteInternalError(l, w, err, "")
			return
		}

		ei1, err := strconv.Atoi(exercise)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		count, err := exercises.Len(username, wi)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		if ei1 < count-1 {
			ei2 := ei1 + 1
			l = l.With("to-index", ei2)

			if err := exercises.Swap(username, wi, ei1, ei2); err != nil {
				api.WriteInternalError(l, w, err, "")
				return
			}
		} else {
			err := fmt.Errorf("index %d out of range", ei1)
			api.WriteInternalError(l, w, err, "index already at bottom")
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

	type request struct {
		Index int `json:"with"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			username = r.PathValue("username")
			workout  = r.PathValue("workout")
			exercise = r.PathValue("exercise")
		)

		l := l.With("user", username, "workout", workout, "exercise", exercise)

		wi, err := strconv.Atoi(workout)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		ei1, err := strconv.Atoi(exercise)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}
		with, err := api.ReadJSON[request](r)
		if err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}
		l = l.With("with", with.Index)

		if err := exercises.Swap(username, wi, ei1, with.Index); err != nil {
			api.WriteInternalError(l, w, err, "")
			return
		}

		l.Debug("swapped exercises")
	})
}
