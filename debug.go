package musclememapi

import (
	"fmt"
	"net/http"
)

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "status: available\n")
	fmt.Fprintf(w, "environment: %s\n", s.Config.Environment)
	fmt.Fprintf(w, "version: %s\n", s.Config.Version)
	fmt.Fprintf(w, "commit: %s\n", s.Config.Commit)
	fmt.Fprintf(w, "date: %s\n", s.Config.Date)
	fmt.Fprintf(w, "maintainer: %s\n", s.Config.Maintainer)
}
