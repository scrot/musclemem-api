package cli

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ExitCode to be used with os.Exit() for proper
// error handling of cli tools
type ExitCode int

const (
	ExitOK      ExitCode = 0
	ExitError   ExitCode = 1
	ExitCancel  ExitCode = 2
	ExitAuth    ExitCode = 4
	ExitPending ExitCode = 8
)

// ErrExists indicates the resource to be added already exists
// the ExitCode should be ExitOK to not terminate batch execution
var (
	ErrExists = errors.New("resource already exists")
	ErrAuth   = errors.New("requires logged-in user")
)

// NewCLIError standardises the error text, representing a cli error
func NewCLIError(err error) error {
	return fmt.Errorf("cli error: %w", err)
}

// NewJSONError standardises the error text, representing a json error
func NewJSONError(err error) error {
	return fmt.Errorf("json error: %w", err)
}

// NewAPIError standardises the error text, representing an api error
func NewAPIError(err error) error {
	return fmt.Errorf("api error: %w", err)
}

// NewAPIStatusError standardises the error text, representing an api error
// the error messages in the response body is parsed and wrapped in the error
func NewAPIStatusError(resp *http.Response) error {
	defer resp.Body.Close()
	msg, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewAPIError(err)
	}
	return fmt.Errorf("api error: %w", errors.New(string(msg)))
}
