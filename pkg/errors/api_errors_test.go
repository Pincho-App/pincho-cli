package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestValidationError(t *testing.T) {
	tests := []struct {
		name           string
		err            *ValidationError
		expectedMsg    string
		expectedStatus int
		isRetryable    bool
	}{
		{
			name:           "simple validation error",
			err:            NewValidationError("title is required"),
			expectedMsg:    "title is required",
			expectedStatus: 400,
			isRetryable:    false,
		},
		{
			name:           "validation error with details",
			err:            NewValidationErrorWithDetails("Title is required", "title", "missing_parameter"),
			expectedMsg:    "Title is required (parameter: title) [missing_parameter]",
			expectedStatus: 400,
			isRetryable:    false,
		},
		{
			name:           "validation error with only param",
			err:            &ValidationError{Message: "Invalid value", Parameter: "timeout"},
			expectedMsg:    "Invalid value (parameter: timeout)",
			expectedStatus: 400,
			isRetryable:    false,
		},
		{
			name:           "validation error with only code",
			err:            &ValidationError{Message: "Invalid format", Code: "invalid_format"},
			expectedMsg:    "Invalid format [invalid_format]",
			expectedStatus: 400,
			isRetryable:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expectedMsg {
				t.Errorf("expected error message '%s', got '%s'", tt.expectedMsg, tt.err.Error())
			}
			if tt.err.StatusCode() != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, tt.err.StatusCode())
			}
			if tt.err.IsRetryable() != tt.isRetryable {
				t.Errorf("expected IsRetryable() to be %v, got %v", tt.isRetryable, tt.err.IsRetryable())
			}
		})
	}
}

func TestAuthenticationError(t *testing.T) {
	tests := []struct {
		name           string
		err            *AuthenticationError
		expectedMsg    string
		expectedStatus int
		isRetryable    bool
	}{
		{
			name:           "auth error default status",
			err:            NewAuthenticationError("invalid token"),
			expectedMsg:    "invalid token",
			expectedStatus: 401,
			isRetryable:    false,
		},
		{
			name:           "auth error with 403",
			err:            NewAuthenticationErrorWithStatus("forbidden", 403),
			expectedMsg:    "forbidden",
			expectedStatus: 403,
			isRetryable:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expectedMsg {
				t.Errorf("expected error message '%s', got '%s'", tt.expectedMsg, tt.err.Error())
			}
			if tt.err.StatusCode() != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, tt.err.StatusCode())
			}
			if tt.err.IsRetryable() != tt.isRetryable {
				t.Errorf("expected IsRetryable() to be %v, got %v", tt.isRetryable, tt.err.IsRetryable())
			}
		})
	}
}

func TestRateLimitError(t *testing.T) {
	tests := []struct {
		name           string
		err            *RateLimitError
		expectedMsg    string
		expectedStatus int
		isRetryable    bool
	}{
		{
			name:           "rate limit without retry-after",
			err:            NewRateLimitError("too many requests"),
			expectedMsg:    "too many requests",
			expectedStatus: 429,
			isRetryable:    true,
		},
		{
			name:           "rate limit with retry-after",
			err:            NewRateLimitErrorWithRetryAfter("rate limit exceeded", 60),
			expectedMsg:    "rate limit exceeded (retry after 60 seconds)",
			expectedStatus: 429,
			isRetryable:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expectedMsg {
				t.Errorf("expected error message '%s', got '%s'", tt.expectedMsg, tt.err.Error())
			}
			if tt.err.StatusCode() != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, tt.err.StatusCode())
			}
			if tt.err.IsRetryable() != tt.isRetryable {
				t.Errorf("expected IsRetryable() to be %v, got %v", tt.isRetryable, tt.err.IsRetryable())
			}
		})
	}
}

func TestServerError(t *testing.T) {
	tests := []struct {
		name           string
		err            *ServerError
		expectedMsg    string
		expectedStatus int
		isRetryable    bool
	}{
		{
			name:           "server error default status",
			err:            NewServerError("internal server error"),
			expectedMsg:    "internal server error",
			expectedStatus: 500,
			isRetryable:    true,
		},
		{
			name:           "server error with 502",
			err:            NewServerErrorWithStatus("bad gateway", 502),
			expectedMsg:    "bad gateway",
			expectedStatus: 502,
			isRetryable:    true,
		},
		{
			name:           "server error with 503",
			err:            NewServerErrorWithStatus("service unavailable", 503),
			expectedMsg:    "service unavailable",
			expectedStatus: 503,
			isRetryable:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expectedMsg {
				t.Errorf("expected error message '%s', got '%s'", tt.expectedMsg, tt.err.Error())
			}
			if tt.err.StatusCode() != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, tt.err.StatusCode())
			}
			if tt.err.IsRetryable() != tt.isRetryable {
				t.Errorf("expected IsRetryable() to be %v, got %v", tt.isRetryable, tt.err.IsRetryable())
			}
		})
	}
}

func TestNetworkError(t *testing.T) {
	tests := []struct {
		name           string
		err            *NetworkError
		expectedMsg    string
		expectedStatus int
		isRetryable    bool
		hasCause       bool
	}{
		{
			name:           "network error without cause",
			err:            NewNetworkError("connection refused", nil),
			expectedMsg:    "connection refused",
			expectedStatus: 0,
			isRetryable:    true,
			hasCause:       false,
		},
		{
			name:           "network error with cause",
			err:            NewNetworkError("request failed", fmt.Errorf("timeout after 30s")),
			expectedMsg:    "request failed: timeout after 30s",
			expectedStatus: 0,
			isRetryable:    true,
			hasCause:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expectedMsg {
				t.Errorf("expected error message '%s', got '%s'", tt.expectedMsg, tt.err.Error())
			}
			if tt.err.StatusCode() != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, tt.err.StatusCode())
			}
			if tt.err.IsRetryable() != tt.isRetryable {
				t.Errorf("expected IsRetryable() to be %v, got %v", tt.isRetryable, tt.err.IsRetryable())
			}
			if tt.hasCause && tt.err.Unwrap() == nil {
				t.Error("expected error to have cause, but Unwrap() returned nil")
			}
			if !tt.hasCause && tt.err.Unwrap() != nil {
				t.Error("expected error to have no cause, but Unwrap() returned non-nil")
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		isRetryable bool
	}{
		{
			name:        "ValidationError is not retryable",
			err:         NewValidationError("invalid"),
			isRetryable: false,
		},
		{
			name:        "AuthenticationError is not retryable",
			err:         NewAuthenticationError("unauthorized"),
			isRetryable: false,
		},
		{
			name:        "RateLimitError is retryable",
			err:         NewRateLimitError("too many requests"),
			isRetryable: true,
		},
		{
			name:        "ServerError is retryable",
			err:         NewServerError("internal error"),
			isRetryable: true,
		},
		{
			name:        "NetworkError is retryable",
			err:         NewNetworkError("connection refused", nil),
			isRetryable: true,
		},
		{
			name:        "generic error is not retryable",
			err:         errors.New("some error"),
			isRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			if result != tt.isRetryable {
				t.Errorf("expected IsRetryableError() to return %v, got %v", tt.isRetryable, result)
			}
		})
	}
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{
			name:       "ValidationError returns 400",
			err:        NewValidationError("invalid"),
			statusCode: 400,
		},
		{
			name:       "AuthenticationError returns 401",
			err:        NewAuthenticationError("unauthorized"),
			statusCode: 401,
		},
		{
			name:       "AuthenticationError returns 403",
			err:        NewAuthenticationErrorWithStatus("forbidden", 403),
			statusCode: 403,
		},
		{
			name:       "RateLimitError returns 429",
			err:        NewRateLimitError("too many requests"),
			statusCode: 429,
		},
		{
			name:       "ServerError returns 500",
			err:        NewServerError("internal error"),
			statusCode: 500,
		},
		{
			name:       "ServerError returns 502",
			err:        NewServerErrorWithStatus("bad gateway", 502),
			statusCode: 502,
		},
		{
			name:       "NetworkError returns 0",
			err:        NewNetworkError("connection refused", nil),
			statusCode: 0,
		},
		{
			name:       "generic error returns 0",
			err:        errors.New("some error"),
			statusCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStatusCode(tt.err)
			if result != tt.statusCode {
				t.Errorf("expected GetStatusCode() to return %d, got %d", tt.statusCode, result)
			}
		})
	}
}

func TestAPIErrorInterface(t *testing.T) {
	// Verify that all error types implement the APIError interface
	var _ APIError = &ValidationError{}
	var _ APIError = &AuthenticationError{}
	var _ APIError = &RateLimitError{}
	var _ APIError = &ServerError{}
	var _ APIError = &NetworkError{}
}

func TestCLIErrorStillWorks(t *testing.T) {
	// Test that the existing CLIError still works (backward compatibility)
	usageErr := NewUsageError("invalid input", fmt.Errorf("title is required"))
	if usageErr.ExitCode != ExitUsageError {
		t.Errorf("expected exit code %d, got %d", ExitUsageError, usageErr.ExitCode)
	}
	if usageErr.Error() != "invalid input: title is required" {
		t.Errorf("unexpected error message: %s", usageErr.Error())
	}

	apiErr := NewAPIError("rate limit", fmt.Errorf("too many requests"))
	if apiErr.ExitCode != ExitAPIError {
		t.Errorf("expected exit code %d, got %d", ExitAPIError, apiErr.ExitCode)
	}

	sysErr := NewSystemError("network", fmt.Errorf("connection refused"))
	if sysErr.ExitCode != ExitSystemError {
		t.Errorf("expected exit code %d, got %d", ExitSystemError, sysErr.ExitCode)
	}

	// Test Unwrap
	if usageErr.Unwrap() == nil {
		t.Error("expected Unwrap() to return cause")
	}
}
