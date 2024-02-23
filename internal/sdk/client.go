package sdk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var (
	name    string
	version string
)

type userAgent struct {
	name    string
	version string
}

func (ua userAgent) String() string {
	return fmt.Sprintf("%s-sdk/%s", ua.name, ua.version)
}

type Client struct {
	baseURL    *url.URL
	userAgent  userAgent
	apiKey     string
	apiVersion string
}

func NewClient(baseURL string, apiKey string, apiVersion string) (*Client, error) {
	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	if apiKey == "" {
		return nil, errors.New("no api key provided")
	}

	userAgent := userAgent{name: name, version: version}

	return &Client{
		baseURL:    url,
		apiKey:     apiKey,
		userAgent:  userAgent,
		apiVersion: apiVersion,
	}, nil
}

var ErrStatusNotOK = errors.New("not http status ok")

func (c *Client) Send(ctx context.Context, method string, path string, body io.Reader) (*http.Response, error) {
	if path == "" {
		return nil, errors.New("no endpoint provided")
	}

	url, err := url.JoinPath(c.baseURL.String(), path)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("User-Agent", c.userAgent.String())
	req.Header.Set("X-Version", c.apiVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return resp, ErrStatusNotOK
	}

	return resp, nil
}
