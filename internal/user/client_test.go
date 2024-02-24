package user

import (
	"context"
	"net/http"
	"testing"
)

const (
	validUserInput = `
  {
    "username": "test",
    "email": "test@gmail.com",
    "password": "secret"
  }
  `
	validUserOutput = `
  {
    "index": 1,
    "username": "test",
    "email": "test@gmail.com",
    "password": "secret"
  }
  `
)

func RunTestServer(t *testing.T, ctx context.Context) {
	t.Helper()
}

func TestRegisterUser(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	go RunTestServer(t, ctx)

	cs := []struct {
		name       string
		input      []byte
		want       []byte
		wantErr    error
		wantStatus int
	}{
		{"ValidUser", []byte(validUserInput), []byte(validUserOutput), nil, http.StatusOK},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			//
		})
	}
}
