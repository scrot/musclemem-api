package internal

import (
	"log/slog"
	"net/http"

	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/scrot/musclemem-api/internal/user"
	"github.com/scrot/musclemem-api/internal/workout"
)

type ServerConfig struct {
	listenAddr string
}

func NewServer(
	logger *slog.Logger,
	users user.UserStore,
	workouts workout.WorkoutStore,
	exercises exercise.ExerciseStore,
) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, logger, users, workouts, exercises)

	// TODO: middleware
	var handler http.Handler = mux
	// handler = someMiddleware(handler)
	// handler = someMiddleware2(handler)
	// handler = someMiddleware3(handler)
	return handler
}
