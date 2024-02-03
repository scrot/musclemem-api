package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func postJSON(baseurl string, endpoint string, r io.Reader) (*http.Response, error) {
	if baseurl == "" {
		return nil, errors.New("no base url provided")
	}

	if endpoint == "" {
		return nil, errors.New("no endpoint provided")
	}

	url, err := url.JoinPath(baseurl, endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	resp, err := http.Post(url, "application/json", r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getJSON(baseurl string, endpoint string) ([]byte, error) {
	if baseurl == "" {
		return []byte{}, errors.New("no base url provided")
	}

	if endpoint == "" {
		return []byte{}, errors.New("no endpoint provided")
	}

	url, err := url.JoinPath(baseurl, endpoint)
	fmt.Println(url)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}
