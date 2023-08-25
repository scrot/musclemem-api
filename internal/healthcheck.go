package internal

import (
	"fmt"
	"net/http"
)

func (s *Server) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "status: available\n")
	fmt.Fprintf(w, "environment: %s\n", s.Environment)
	fmt.Fprintf(w, "version: %s\n", s.Version)
	fmt.Fprintf(w, "commit: %s\n", s.Commit)
	fmt.Fprintf(w, "date: %s\n", s.Date)
	fmt.Fprintf(w, "maintainer: %s\n", s.Maintainer)
}
