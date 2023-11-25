package internal

import (
	"fmt"
	"net/http"
)

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "status: available\n")
	fmt.Fprintf(w, "environment: %s\n", s.config.Environment)
	fmt.Fprintf(w, "version: %s\n", s.config.Version)
	fmt.Fprintf(w, "commit: %s\n", s.config.Commit)
	fmt.Fprintf(w, "date: %s\n", s.config.Date)
	fmt.Fprintf(w, "maintainer: %s\n", s.config.Maintainer)
}
