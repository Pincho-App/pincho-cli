// Package errors provides error handling and exit code management for WirePusher CLI.
//
// The package implements a structured error system with standardized exit codes
// to enable better integration with CI/CD pipelines and automation workflows.
//
// Exit Codes:
//   - 0: Success - command completed successfully
//   - 1: Usage Error - invalid user input, missing required parameters, or authentication failures
//   - 2: API Error - API returned an error (rate limits, validation, server errors)
//   - 3: System Error - network failures, timeouts, or other system-level errors
//
// Error Types:
//   - UsageError: Input validation, missing parameters, authentication issues
//   - APIError: API-level errors including rate limits and server errors
//   - SystemError: Network connectivity, timeouts, file system errors
//
// Example usage:
//
//	if token == "" {
//	    return errors.NewUsageError("API token required", fmt.Errorf("no token provided"))
//	}
//
//	if resp.StatusCode == 429 {
//	    return errors.NewAPIError("Rate limit exceeded", err)
//	}
//
//	if err := network.Call(); err != nil {
//	    return errors.NewSystemError("Network error", err)
//	}
//
// Exit code handling enables shell scripts to distinguish between different
// failure modes and take appropriate action (e.g., retry on system errors,
// fail fast on usage errors).
package errors

import (
	"fmt"
	"os"
)

// Exit codes following standard conventions
const (
	// ExitSuccess indicates successful execution
	ExitSuccess = 0

	// ExitUsageError indicates invalid command usage (missing args, invalid flags, etc.)
	ExitUsageError = 1

	// ExitAPIError indicates an API error (4xx, 5xx responses)
	ExitAPIError = 2

	// ExitSystemError indicates a system error (network issues, file I/O, etc.)
	ExitSystemError = 3
)

// CLIError represents an error with an associated exit code
type CLIError struct {
	Message  string
	ExitCode int
	Cause    error
}

// Error implements the error interface
func (e *CLIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap implements error unwrapping for Go 1.13+
func (e *CLIError) Unwrap() error {
	return e.Cause
}

// NewUsageError creates a user error (exit code 1)
func NewUsageError(message string, cause error) *CLIError {
	return &CLIError{
		Message:  message,
		ExitCode: ExitUsageError,
		Cause:    cause,
	}
}

// NewAPIError creates an API error (exit code 2)
func NewAPIError(message string, cause error) *CLIError {
	return &CLIError{
		Message:  message,
		ExitCode: ExitAPIError,
		Cause:    cause,
	}
}

// NewSystemError creates a system error (exit code 3)
func NewSystemError(message string, cause error) *CLIError {
	return &CLIError{
		Message:  message,
		ExitCode: ExitSystemError,
		Cause:    cause,
	}
}

// HandleError prints the error and exits with the appropriate code
// If the error is a CLIError, uses its exit code
// Otherwise, uses ExitSystemError (3)
func HandleError(err error) {
	if err == nil {
		return
	}

	// Check if it's a CLIError with a specific exit code
	if cliErr, ok := err.(*CLIError); ok {
		fmt.Fprintf(os.Stderr, "Error: %s\n", cliErr.Message)
		if cliErr.Cause != nil {
			fmt.Fprintf(os.Stderr, "Cause: %v\n", cliErr.Cause)
		}
		os.Exit(cliErr.ExitCode)
	}

	// Generic error - treat as system error
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(ExitSystemError)
}

// HandleErrorWithSuggestion prints the error with an actionable suggestion and exits
func HandleErrorWithSuggestion(err error, suggestion string) {
	if err == nil {
		return
	}

	// Check if it's a CLIError with a specific exit code
	if cliErr, ok := err.(*CLIError); ok {
		fmt.Fprintf(os.Stderr, "Error: %s\n", cliErr.Message)
		if cliErr.Cause != nil {
			fmt.Fprintf(os.Stderr, "Cause: %v\n", cliErr.Cause)
		}
		if suggestion != "" {
			fmt.Fprintf(os.Stderr, "\nSuggestion: %s\n", suggestion)
		}
		os.Exit(cliErr.ExitCode)
	}

	// Generic error
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	if suggestion != "" {
		fmt.Fprintf(os.Stderr, "\nSuggestion: %s\n", suggestion)
	}
	os.Exit(ExitSystemError)
}
