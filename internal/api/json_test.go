package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadJSON(t *testing.T) {
	type input struct {
		Value string
	}

	want := input{"ok"}

	inputJson, _ := json.Marshal(want)
	r := bytes.NewReader(inputJson)

	req := httptest.NewRequest(http.MethodGet, "/", r)

	got, err := ReadJSON[input](req)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(got, want) {
		t.Fatalf("want %v but got %v", want, got)
	}
}

func TestWriteJSON(t *testing.T) {
	type input struct {
		Value string `json:"value"`
	}

	i := input{"ok"}
	want := "{\"value\":\"ok\"}"

	rec := httptest.NewRecorder()
	err := WriteJSON(rec, http.StatusOK, &i)
	if err != nil {
		t.Fatal(err)
	}
	got := rec.Body.String()

	if want == got {
		t.Fatalf("want %v but got %v", want, got)
	}
}
