package main

import (
	"fmt"
	"net/http"
)

func (api *api) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "status: available\n")
	fmt.Fprintf(w, "environment: %s\n", api.config.Env)
	fmt.Fprintf(w, "version: %s\n", api.config.Version)
	fmt.Fprintf(w, "commit: %s\n", api.config.Commit)
	fmt.Fprintf(w, "date: %s\n", api.config.Date)
	fmt.Fprintf(w, "maintainer: %s\n", api.config.Maintainer)
}
