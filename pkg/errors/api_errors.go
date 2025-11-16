// Package errors provides structured error types for the WirePusher CLI.
package errors

import (
	"fmt"
)

// APIError is the base interface for all API-related errors
type APIError interface {
	error
	IsRetryable() bool
	StatusCode() int
}

// ValidationError represents a client-side validation error (4xx - typically 400)
// These are NOT retryable as the request is malformed
type ValidationError struct {
	Message    string
	Parameter  string
	Code       string
	statusCode int
}

func (e *ValidationError) Error() string {
	msg := e.Message
	if e.Parameter != "" {
		msg = fmt.Sprintf("%s (parameter: %s)", msg, e.Parameter)
	}
	if e.Code != "" {
		msg = fmt.Sprintf("%s [%s]", msg, e.Code)
	}
	return msg
}

func (e *ValidationError) IsRetryable() bool {
	return false // Validation errors are never retryable
}

func (e *ValidationError) StatusCode() int {
	if e.statusCode == 0 {
		return 400
	}
	return e.statusCode
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message, statusCode: 400}
}

// NewValidationErrorWithDetails creates a new validation error with additional details
func NewValidationErrorWithDetails(message, param, code string) *ValidationError {
	return &ValidationError{
		Message:    message,
		Parameter:  param,
		Code:       code,
		statusCode: 400,
	}
}

// AuthenticationError represents an authentication or authorization failure (401, 403)
// These are NOT retryable as the token is invalid
type AuthenticationError struct {
	Message    string
	statusCode int
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

func (e *AuthenticationError) IsRetryable() bool {
	return false // Auth errors are never retryable
}

func (e *AuthenticationError) StatusCode() int {
	if e.statusCode == 0 {
		return 401
	}
	return e.statusCode
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{Message: message, statusCode: 401}
}

// NewAuthenticationErrorWithStatus creates a new authentication error with specific status code
func NewAuthenticationErrorWithStatus(message string, statusCode int) *AuthenticationError {
	return &AuthenticationError{Message: message, statusCode: statusCode}
}

// RateLimitError represents a rate limit exceeded error (429)
// These ARE retryable after waiting for the rate limit to reset
type RateLimitError struct {
	Message    string
	RetryAfter int // Seconds to wait before retry (from Retry-After header)
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%s (retry after %d seconds)", e.Message, e.RetryAfter)
	}
	return e.Message
}

func (e *RateLimitError) IsRetryable() bool {
	return true // Rate limit errors are retryable after waiting
}

func (e *RateLimitError) StatusCode() int {
	return 429
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(message string) *RateLimitError {
	return &RateLimitError{Message: message}
}

// NewRateLimitErrorWithRetryAfter creates a new rate limit error with retry information
func NewRateLimitErrorWithRetryAfter(message string, retryAfter int) *RateLimitError {
	return &RateLimitError{Message: message, RetryAfter: retryAfter}
}

// ServerError represents a server-side error (5xx)
// These ARE retryable as the server may recover
type ServerError struct {
	Message    string
	statusCode int
}

func (e *ServerError) Error() string {
	return e.Message
}

func (e *ServerError) IsRetryable() bool {
	return true // Server errors are retryable
}

func (e *ServerError) StatusCode() int {
	if e.statusCode == 0 {
		return 500
	}
	return e.statusCode
}

// NewServerError creates a new server error
func NewServerError(message string) *ServerError {
	return &ServerError{Message: message, statusCode: 500}
}

// NewServerErrorWithStatus creates a new server error with specific status code
func NewServerErrorWithStatus(message string, statusCode int) *ServerError {
	return &ServerError{Message: message, statusCode: statusCode}
}

// NetworkError represents a network connectivity error (connection refused, timeout, etc.)
// These ARE retryable as the network may become available
type NetworkError struct {
	Message string
	Cause   error
}

func (e *NetworkError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *NetworkError) IsRetryable() bool {
	return true // Network errors are retryable
}

func (e *NetworkError) StatusCode() int {
	return 0 // No HTTP status code for network errors
}

func (e *NetworkError) Unwrap() error {
	return e.Cause
}

// NewNetworkError creates a new network error
func NewNetworkError(message string, cause error) *NetworkError {
	return &NetworkError{Message: message, Cause: cause}
}

// IsRetryableError checks if any error implements the APIError interface and is retryable
func IsRetryableError(err error) bool {
	if apiErr, ok := err.(APIError); ok {
		return apiErr.IsRetryable()
	}
	return false
}

// GetStatusCode returns the HTTP status code for an API error, or 0 if not applicable
func GetStatusCode(err error) int {
	if apiErr, ok := err.(APIError); ok {
		return apiErr.StatusCode()
	}
	return 0
}
