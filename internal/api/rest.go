package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type TypeJSON int

const (
	TypeJSONObject TypeJSON = iota
	TypeJSONArray
	TypeJSONUnknown
	TypeJSONInvalid
)

// WriteJSON encodes v as json and writes to the ResponseWriter
// setting the proper headers. It protects writing invalid json
// and returns an error instead.
func WriteJSON[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

// ReadJSON reads the Request and returns the decoded payload
// as T. The request body needs to be closed by the caller
func ReadJSON[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

// JSONType detects if the JSON payload is an array or an object
func JSONType(r io.Reader) TypeJSON {
	t, err := json.NewDecoder(r).Token()
	if err != nil {
		return TypeJSONInvalid
	}

	switch t.(json.Delim) {
	case '{':
		return TypeJSONObject
	case '[':
		return TypeJSONArray
	default:
		return TypeJSONUnknown
	}
}

// WriteInternalError handles generic server errors logging the error and writing
// the error response
func WriteInternalError(l *slog.Logger, w http.ResponseWriter, err error, msg string) {
	l.Error(err.Error())
	if msg == "" {
		msg = "Whoeps! something went wrong"
	}
	http.Error(w, msg, http.StatusInternalServerError)
}
