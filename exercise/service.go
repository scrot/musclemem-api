package exercise

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/scrot/jsonapi"
)

// HandleSingleExercise handles the request for a single exercise
// returning the details of exercise as json given an exerciseID
func (a *API) HandleSingleExercise(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "exerciseID")

	a.Logger.Debug("new single exercise request", "id", idParam, "path", r.URL.Path)

	id, err := strconv.Atoi(idParam)
	if err != nil {
		msg := fmt.Sprintf("%d not a valid id: %s", id, err)
		a.Logger.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)

	}

	e, err := FetchSingleExerciseJSON(*a.Store, id)
	if err != nil {
		msg := fmt.Errorf("exercise retrieval error: %w", err).Error()
		a.Logger.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(e); err != nil {
		msg := fmt.Errorf("exercise response error: %w", err).Error()
		a.Logger.Error(msg)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
