package apiutils

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type TypeJSON int

const (
	TypeJSONObject TypeJSON = iota
	TypeJSONArray
	TypeJSONInvalid
)

// requestJSON reads r.Body into a slice of bytes
// it also detects if it is a json array or object
func RequestJSON(r *http.Request) ([]byte, TypeJSON, error) {
	var buf bytes.Buffer
	defer r.Body.Close()
	if _, err := buf.ReadFrom(r.Body); err != nil {
		return []byte{}, TypeJSONInvalid, err
	}
	body := buf.Bytes()

	if len(body) <= 0 {
		return []byte{}, TypeJSONInvalid, errors.New("empty request body")
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	t, err := dec.Token()
	if err != nil {
		return []byte{}, TypeJSONInvalid, err
	}

	switch t.(json.Delim) {
	case '{':
		return body, TypeJSONObject, nil
	case '[':
		return body, TypeJSONArray, nil
	default:
		return []byte{}, TypeJSONInvalid, nil
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
