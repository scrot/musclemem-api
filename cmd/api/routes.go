package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (api *api) routes() *chi.Mux {
	router := chi.NewRouter()

	router.MethodFunc(http.MethodGet, "/v1/health", api.healthcheckHandler)
	router.MethodFunc(http.MethodPost, "/v1/workouts", api.createWorkoutHandler)
	router.MethodFunc(http.MethodGet, "/v1/workouts/:id", api.showWorkoutHandler)

	return router
}
