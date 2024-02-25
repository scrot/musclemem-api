package internal

import (
	"log/slog"
	"net/http"

	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/scrot/musclemem-api/internal/user"
	"github.com/scrot/musclemem-api/internal/workout"
)

func RegisterEndpoints(
	mux *http.ServeMux,
	logger *slog.Logger,
	users user.UserStore,
	workouts workout.WorkoutStore,
	exercises exercise.ExerciseStore,
) {
	mux.Handle("GET /ready", NewReadyHandler(logger))
	mux.Handle("POST /users", user.NewCreateHandler(logger, users))
	mux.Handle("GET /users/{username}/workouts", workout.NewFetchAllHandler(logger, workouts))
	mux.Handle("POST /users/{username}/workouts", workout.NewCreateHandler(logger, workouts))
	mux.Handle("DELETE /users/{username}/workouts/{workout}", workout.NewDeleteHandler(logger, workouts))
	mux.Handle("PATCH /users/{username}/workouts/{workout}", workout.NewUpdateHandler(logger, workouts))
	mux.Handle("GET /users/{username}/workouts/{workout}/exercises", exercise.NewFetchAllHandler(logger, exercises))
	mux.Handle("POST /users/{username}/workouts/{workout}/exercises", exercise.NewCreateHandler(logger, exercises))
	mux.Handle("GET /users/{username}/workouts/{workout}/exercises/{exercise}", exercise.NewFetchHandler(logger, exercises))
	mux.Handle("PATCH /users/{username}/workouts/{workout}/exercises/{exercise}", exercise.NewUpdateHandler(logger, exercises))
	mux.Handle("DELETE /users/{username}/workouts/{workout}/exercises/{exercise}", exercise.NewDeleteHandler(logger, exercises))
	mux.Handle("PUT /users/{username}/workouts/{workout}/exercises/{exercise}/up", exercise.NewUpHandler(logger, exercises))
	mux.Handle("PUT /users/{username}/workouts/{workout}/exercises/{exercise}/down", exercise.NewDownHandler(logger, exercises))
	mux.Handle("POST /users/{username}/workouts/{workout}/exercises/{exercise}/swap", exercise.NewSwapHandler(logger, exercises))
}

func NewReadyHandler(l *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.Info("answered readiness probe")
		w.Write([]byte("ready"))
	})
}
