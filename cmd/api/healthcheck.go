package main

import (
	"fmt"
	"net/http"
)

func (api *api) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "status: available\n")
	fmt.Fprintf(w, "environment: %s\n", api.config.Env)
	fmt.Fprintf(w, "version: %s", api.config.Version)
}
