package commands

import (
	"errors"
	"io"
	"net/http"
)

func postJSON(baseurl string, endpoint string, r io.Reader) (*http.Response, error) {
	if baseurl == "" {
		return nil, errors.New("no base url provided")
	}

	if endpoint == "" {
		return nil, errors.New("no endpoint provided")
	}

	resp, err := http.Post(baseurl+endpoint, "application/json", r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
