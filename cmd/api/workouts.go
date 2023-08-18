package main

import (
	"fmt"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/julienschmidt/httprouter"
)

func (api *api) createWorkoutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create new workout...")

}

func (api *api) showWorkoutHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id := params.ByName("id")
	if _, err := uuid.FromString(id); err != nil {
		api.logger.Error(err.Error())
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "this is workout %s", id)
}
