package cli

import (
	"fmt"
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

// NewAPIStatusError creates an error out of a status code, representing an api error
func NewAPIStatusError(code int) error {
	return fmt.Errorf("api error: %s", http.StatusText(code))
}
