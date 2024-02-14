package cli

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/olekukonko/tablewriter"
)

func SendRequest(method string, baseurl string, endpoint string, r io.Reader) (*http.Response, error) {
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

	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func HandleResponse(resp *http.Response, respErr error) {
	if respErr != nil {
		fmt.Printf("api error: %s\n", respErr)
		os.Exit(0)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("api error: %s\n", resp.Status)
		os.Exit(0)
	}
}

func NewSimpleTable(c *CLIConfig) *tablewriter.Table {
	t := tablewriter.NewWriter(c.Out)
	t.SetBorder(false)
	t.SetHeaderLine(false)
	t.SetNoWhiteSpace(true)
	t.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	t.SetRowSeparator("")
	t.SetColumnSeparator("")
	t.SetCenterSeparator("")
	t.SetTablePadding("  ")
	return t
}
