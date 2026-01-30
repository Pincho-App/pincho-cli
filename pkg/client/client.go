// Package client provides a Go client for the Pincho API.
//
// The client supports two main endpoints:
//   - /send: Send push notifications with full control over title, message, and parameters
//   - /notifai: Use AI (Gemini) to generate structured notifications from free-form text
//
// Features:
//   - Configurable timeout and retry logic with exponential backoff
//   - Automatic tag validation and normalization
//   - AES-128-CBC message encryption support
//   - Rate limit information extraction from response headers
//   - Structured error responses with detailed error information
//
// Basic usage:
//
//	client := client.New()
//	client.SetTimeout(30 * time.Second)
//	client.SetRetryConfig(3, 1 * time.Second)
//
//	result, err := client.Send(&client.SendOptions{
//	    Title: "Build Complete",
//	    Message: "v1.2.3 deployed successfully",
//	    Token: "your-token",
//	    Tags: []string{"deploy", "production"},
//	})
//
// The client automatically retries on:
//   - Network errors (connection refused, timeout, etc.)
//   - Server errors (5xx status codes)
//   - Rate limit errors (429) with longer backoff
//
// Retries use exponential backoff (1s, 2s, 4s, 8s) capped at 30 seconds.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitlab.com/pincho-app/pincho-cli/pkg/crypto"
	"gitlab.com/pincho-app/pincho-cli/pkg/errors"
	"gitlab.com/pincho-app/pincho-cli/pkg/validation"
)

const (
	// DefaultAPIURL is the default Pincho API endpoint (for send)
	DefaultAPIURL = "https://api.pincho.app/send"

	// DefaultNotifAIURL is the default Pincho NotifAI API endpoint
	DefaultNotifAIURL = "https://api.pincho.app/notifai"

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default maximum number of retry attempts
	DefaultMaxRetries = 3

	// DefaultInitialBackoff is the default initial backoff duration for retries
	DefaultInitialBackoff = 1 * time.Second

	// Version is the client library version (can be overridden)
	Version = "1.0.0"
)

// Client represents a Pincho API client
type Client struct {
	APIURL         string
	HTTPClient     *http.Client
	Timeout        time.Duration // Custom timeout duration (uses DefaultTimeout if zero)
	MaxRetries     int           // Maximum number of retry attempts (uses DefaultMaxRetries if zero)
	InitialBackoff time.Duration // Initial backoff duration for retries (uses DefaultInitialBackoff if zero)
	Token          string        // API token for authentication (sent as Bearer token in Authorization header)
	UserAgent      string        // User-Agent header value (defaults to pincho-cli/{version})
}

// SendOptions contains parameters for sending a notification
type SendOptions struct {
	Title              string   `json:"title"`
	Message            string   `json:"message"`
	Type               string   `json:"type,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	ImageURL           string   `json:"imageURL,omitempty"`
	ActionURL          string   `json:"actionURL,omitempty"`
	IV                 string   `json:"iv,omitempty"`
	EncryptionPassword string   `json:"-"` // Not sent to API, used for local encryption
}

// SendResponse represents the API success response
type SendResponse struct {
	Status               string                `json:"status"`
	Message              string                `json:"message"`
	ReceivedNotification *NotificationDetails  `json:"receivedNotification,omitempty"` // For personal tokens
	TeamID               string                `json:"teamId,omitempty"`               // For team tokens
	MemberCount          int                   `json:"memberCount,omitempty"`          // For team tokens
	Notifications        []NotificationDetails `json:"notifications,omitempty"`        // For team tokens
}

// FirestoreTimestamp represents a Firestore timestamp object returned by the API
type FirestoreTimestamp struct {
	Seconds     int64 `json:"_seconds"`
	Nanoseconds int64 `json:"_nanoseconds"`
}

// NotificationDetails represents a single notification object from the API
type NotificationDetails struct {
	NotificationID string             `json:"notificationID"`
	UserID         string             `json:"userID"`
	Title          string             `json:"title"`
	Body           string             `json:"body"`
	Type           string             `json:"type"`
	ImageURL       string             `json:"imageURL,omitempty"`
	ActionURL      string             `json:"actionURL,omitempty"`
	Timestamp      string             `json:"timestamp"`
	Tags           []string           `json:"tags,omitempty"`
	TeamID         string             `json:"teamId,omitempty"`
	TeamName       string             `json:"teamName,omitempty"`
	Endpoint       string             `json:"endpoint"`
	IV             string             `json:"iv,omitempty"`
	ExpiresAt      FirestoreTimestamp `json:"expiresAt"`
}

// RateLimitInfo contains rate limiting information from response headers
type RateLimitInfo struct {
	Limit     string
	Remaining string
	Reset     string
}

// SendResult combines the response with additional metadata
type SendResult struct {
	Response  *SendResponse
	RateLimit *RateLimitInfo
}

// NotifAIOptions contains parameters for sending a NotifAI request
type NotifAIOptions struct {
	Text string `json:"text"`
	Type string `json:"type,omitempty"`
}

// NotifAIResponse represents the NotifAI API response
type NotifAIResponse struct {
	Status               string                `json:"status"`
	Message              string                `json:"message"`
	ReceivedNotification *NotificationDetails  `json:"receivedNotification,omitempty"` // For personal tokens
	TeamID               string                `json:"teamId,omitempty"`               // For team tokens
	MemberCount          int                   `json:"memberCount,omitempty"`          // For team tokens
	Notifications        []NotificationDetails `json:"notifications,omitempty"`        // For team tokens
	Summary              *NotifAISummary       `json:"summary,omitempty"`              // AI-generated summary
}

// NotifAISummary represents the AI-generated notification summary
type NotifAISummary struct {
	Title     string   `json:"title"`
	Message   string   `json:"message"`
	ActionURL string   `json:"actionURL,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

// NotifAIResult combines the NotifAI response with additional metadata
type NotifAIResult struct {
	Response  *NotifAIResponse
	RateLimit *RateLimitInfo
}

// ErrorResponse represents the API error response with nested structure
type ErrorResponse struct {
	Status string       `json:"status"`
	Error  ErrorDetails `json:"error"`
}

// ErrorDetails contains the nested error information
type ErrorDetails struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
}

// New creates a new Pincho client with default settings
func New() *Client {
	return &Client{
		APIURL:         DefaultAPIURL,
		Timeout:        DefaultTimeout,
		MaxRetries:     DefaultMaxRetries,
		InitialBackoff: DefaultInitialBackoff,
		UserAgent:      fmt.Sprintf("pincho-cli/%s", Version),
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// SetTimeout updates the client's HTTP timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		c.Timeout = timeout
		c.HTTPClient.Timeout = timeout
	}
}

// SetToken sets the API token for authentication
func (c *Client) SetToken(token string) {
	c.Token = token
}

// SetRetryConfig updates the client's retry configuration
func (c *Client) SetRetryConfig(maxRetries int, initialBackoff time.Duration) {
	if maxRetries >= 0 {
		c.MaxRetries = maxRetries
	}
	if initialBackoff > 0 {
		c.InitialBackoff = initialBackoff
	}
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	// Check if error implements APIError with IsRetryable method
	if errors.IsRetryableError(err) {
		return true
	}

	// Check error string for network errors (wrapped errors may not implement APIError)
	errStr := err.Error()
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "temporary failure") ||
		strings.Contains(errStr, "EOF") {
		return true
	}

	// Retry on 429 (rate limit) - but with longer backoff
	if statusCode == 429 {
		return true
	}

	// Retry on 500, 502, 503, 504 (server errors)
	if statusCode >= 500 && statusCode < 600 {
		return true
	}

	return false
}

// calculateBackoff calculates the backoff duration for a given attempt using exponential backoff
// If retryAfter is provided (from Retry-After header), it takes precedence for rate limit errors
func (c *Client) calculateBackoff(attempt int, statusCode int, retryAfter string) time.Duration {
	// If Retry-After header is provided for rate limit errors, use it
	if statusCode == 429 && retryAfter != "" {
		// Try to parse as seconds
		if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
			backoff := time.Duration(seconds) * time.Second
			// Cap at 30 seconds
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			return backoff
		}
	}

	baseBackoff := c.InitialBackoff
	if baseBackoff == 0 {
		baseBackoff = DefaultInitialBackoff
	}

	// For rate limit errors, use longer backoff
	if statusCode == 429 {
		baseBackoff = 5 * time.Second
	}

	// Exponential backoff: 1s, 2s, 4s, 8s, etc.
	backoff := baseBackoff * time.Duration(1<<uint(attempt))

	// Cap at 30 seconds
	if backoff > 30*time.Second {
		backoff = 30 * time.Second
	}

	return backoff
}

// doRequestWithRetry performs an HTTP request with retry logic
func (c *Client) doRequestWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	maxRetries := c.MaxRetries
	if maxRetries == 0 {
		maxRetries = DefaultMaxRetries
	}

	var lastErr error
	var lastStatusCode int
	var lastRetryAfter string

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check for context cancellation before each attempt
		select {
		case <-ctx.Done():
			return nil, errors.NewNetworkError("request cancelled", ctx.Err())
		default:
		}

		// Clone the request for retry (body needs to be reset)
		reqClone := req.Clone(ctx)
		if req.Body != nil {
			// For POST requests, we need to reset the body
			if req.GetBody != nil {
				body, err := req.GetBody()
				if err != nil {
					return nil, errors.NewNetworkError("failed to reset request body", err)
				}
				reqClone.Body = body
			}
		}

		// Perform the request
		resp, err := c.HTTPClient.Do(reqClone)

		// If successful, return immediately
		if err == nil && resp.StatusCode < 400 {
			return resp, nil
		}

		// Store error and status code for retry decision
		lastErr = err
		if resp != nil {
			lastStatusCode = resp.StatusCode
			// Extract Retry-After header for rate limit responses
			lastRetryAfter = resp.Header.Get("Retry-After")
		}

		// Check if error is retryable
		if !isRetryableError(lastErr, lastStatusCode) {
			// Not retryable, return the response/error immediately
			return resp, lastErr
		}

		// Will retry - close response body before retrying
		if resp != nil {
			resp.Body.Close()
		}

		// Don't sleep after the last attempt
		if attempt < maxRetries {
			backoff := c.calculateBackoff(attempt, lastStatusCode, lastRetryAfter)
			// Use select to respect context cancellation during sleep
			select {
			case <-ctx.Done():
				return nil, errors.NewNetworkError("request cancelled during retry backoff", ctx.Err())
			case <-time.After(backoff):
				// Continue with next retry attempt
			}
		}
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, errors.NewNetworkError(fmt.Sprintf("request failed after %d retries", maxRetries), lastErr)
	}

	// This shouldn't happen, but just in case
	return nil, errors.NewServerErrorWithStatus(fmt.Sprintf("request failed after %d retries", maxRetries), lastStatusCode)
}

// Send sends a notification via the Pincho API
// Returns SendResult with response details and rate limit info, or error if failed
func (c *Client) Send(ctx context.Context, opts *SendOptions) (*SendResult, error) {
	// Validate required fields
	if opts.Title == "" {
		return nil, errors.NewValidationError("title is required")
	}

	if c.Token == "" {
		return nil, errors.NewAuthenticationError("token is required")
	}

	// Normalize and validate tags
	if len(opts.Tags) > 0 {
		normalizedTags, err := validation.NormalizeAndValidateTags(opts.Tags)
		if err != nil {
			return nil, errors.NewValidationErrorWithDetails(fmt.Sprintf("tag validation failed: %v", err), "tags", "invalid_tags")
		}
		opts.Tags = normalizedTags
	}

	// Handle encryption if password provided
	finalMessage := opts.Message
	var ivHex string

	if opts.EncryptionPassword != "" {
		// Only encrypt if message is not empty
		if opts.Message != "" {
			ivBytes, ivHexStr, err := crypto.GenerateIV()
			if err != nil {
				return nil, errors.NewNetworkError("failed to generate IV", err)
			}

			encrypted, err := crypto.EncryptMessage(opts.Message, opts.EncryptionPassword, ivBytes)
			if err != nil {
				return nil, errors.NewNetworkError("failed to encrypt message", err)
			}

			finalMessage = encrypted
			ivHex = ivHexStr
		}
	}

	// Build request with encrypted message if applicable
	requestOpts := *opts
	requestOpts.Message = finalMessage
	requestOpts.IV = ivHex
	requestOpts.EncryptionPassword = "" // Don't send password to API

	// Build request body
	jsonData, err := json.Marshal(requestOpts)
	if err != nil {
		return nil, errors.NewNetworkError("failed to marshal request", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", c.APIURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, errors.NewNetworkError("failed to create request", err)
	}

	// Set GetBody for retry support
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(jsonData)), nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("User-Agent", c.UserAgent)

	// Send request with retry logic
	resp, err := c.doRequestWithRetry(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Extract rate limit headers
	rateLimit := &RateLimitInfo{
		Limit:     resp.Header.Get("RateLimit-Limit"),
		Remaining: resp.Header.Get("RateLimit-Remaining"),
		Reset:     resp.Header.Get("RateLimit-Reset"),
	}

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewNetworkError("failed to read response", err)
	}

	// Handle error status codes
	if resp.StatusCode >= 400 {
		// Try to parse nested error response
		var errorResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Error.Message != "" {
			// Return typed error based on status code
			switch resp.StatusCode {
			case 400, 404:
				return nil, errors.NewValidationErrorWithDetails(errorResp.Error.Message, errorResp.Error.Param, errorResp.Error.Code)
			case 401, 403:
				return nil, errors.NewAuthenticationErrorWithStatus(errorResp.Error.Message, resp.StatusCode)
			case 429:
				retryAfter := 0
				if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
					retryAfter, _ = strconv.Atoi(retryAfterStr)
				}
				return nil, errors.NewRateLimitErrorWithRetryAfter(errorResp.Error.Message, retryAfter)
			default:
				if resp.StatusCode >= 500 {
					return nil, errors.NewServerErrorWithStatus(errorResp.Error.Message, resp.StatusCode)
				}
				return nil, errors.NewValidationErrorWithDetails(errorResp.Error.Message, errorResp.Error.Param, errorResp.Error.Code)
			}
		}

		// Fallback to generic error message if parsing fails
		errorMsg := string(bodyBytes)
		switch resp.StatusCode {
		case 400, 404:
			return nil, errors.NewValidationError(fmt.Sprintf("validation error: %s", errorMsg))
		case 401, 403:
			return nil, errors.NewAuthenticationErrorWithStatus(fmt.Sprintf("authentication error: %s", errorMsg), resp.StatusCode)
		case 429:
			retryAfter := 0
			if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
				retryAfter, _ = strconv.Atoi(retryAfterStr)
			}
			return nil, errors.NewRateLimitErrorWithRetryAfter(fmt.Sprintf("rate limit exceeded: %s", errorMsg), retryAfter)
		default:
			if resp.StatusCode >= 500 {
				return nil, errors.NewServerErrorWithStatus(fmt.Sprintf("server error: %s", errorMsg), resp.StatusCode)
			}
			return nil, errors.NewValidationError(fmt.Sprintf("API error (%d): %s", resp.StatusCode, errorMsg))
		}
	}

	// Parse success response
	var apiResp SendResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, errors.NewNetworkError("failed to parse response", err)
	}

	return &SendResult{
		Response:  &apiResp,
		RateLimit: rateLimit,
	}, nil
}

// NotifAI sends a text-to-notification request via the Pincho NotifAI API
// Returns NotifAIResult with response details and rate limit info, or error if failed
func (c *Client) NotifAI(ctx context.Context, opts *NotifAIOptions) (*NotifAIResult, error) {
	// Validate required fields
	if opts.Text == "" {
		return nil, errors.NewValidationError("text is required")
	}

	if c.Token == "" {
		return nil, errors.NewAuthenticationError("token is required")
	}

	// Validate text length
	if len(opts.Text) < 5 {
		return nil, errors.NewValidationError("text must be at least 5 characters long")
	}
	if len(opts.Text) > 2500 {
		return nil, errors.NewValidationError("text must be at most 2500 characters long")
	}

	// Build request body
	jsonData, err := json.Marshal(opts)
	if err != nil {
		return nil, errors.NewNetworkError("failed to marshal request", err)
	}

	// Determine NotifAI URL based on configured API URL
	notifaiURL := DefaultNotifAIURL
	if c.APIURL != "" && c.APIURL != DefaultAPIURL {
		// If custom API URL is set, derive NotifAI URL from it
		// Replace /send with /notifai
		notifaiURL = strings.Replace(c.APIURL, "/send", "/notifai", 1)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", notifaiURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, errors.NewNetworkError("failed to create request", err)
	}

	// Set GetBody for retry support
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(jsonData)), nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("User-Agent", c.UserAgent)

	// Send request with retry logic
	resp, err := c.doRequestWithRetry(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Extract rate limit headers
	rateLimit := &RateLimitInfo{
		Limit:     resp.Header.Get("RateLimit-Limit"),
		Remaining: resp.Header.Get("RateLimit-Remaining"),
		Reset:     resp.Header.Get("RateLimit-Reset"),
	}

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewNetworkError("failed to read response", err)
	}

	// Handle error status codes
	if resp.StatusCode >= 400 {
		// Try to parse nested error response
		var errorResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Error.Message != "" {
			// Return typed error based on status code
			switch resp.StatusCode {
			case 400, 404:
				return nil, errors.NewValidationErrorWithDetails(errorResp.Error.Message, errorResp.Error.Param, errorResp.Error.Code)
			case 401, 403:
				return nil, errors.NewAuthenticationErrorWithStatus(errorResp.Error.Message, resp.StatusCode)
			case 429:
				retryAfter := 0
				if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
					retryAfter, _ = strconv.Atoi(retryAfterStr)
				}
				return nil, errors.NewRateLimitErrorWithRetryAfter(errorResp.Error.Message, retryAfter)
			default:
				if resp.StatusCode >= 500 {
					return nil, errors.NewServerErrorWithStatus(errorResp.Error.Message, resp.StatusCode)
				}
				return nil, errors.NewValidationErrorWithDetails(errorResp.Error.Message, errorResp.Error.Param, errorResp.Error.Code)
			}
		}

		// Fallback to generic error message if parsing fails
		errorMsg := string(bodyBytes)
		switch resp.StatusCode {
		case 400, 404:
			return nil, errors.NewValidationError(fmt.Sprintf("validation error: %s", errorMsg))
		case 401, 403:
			return nil, errors.NewAuthenticationErrorWithStatus(fmt.Sprintf("authentication error: %s", errorMsg), resp.StatusCode)
		case 429:
			retryAfter := 0
			if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
				retryAfter, _ = strconv.Atoi(retryAfterStr)
			}
			return nil, errors.NewRateLimitErrorWithRetryAfter(fmt.Sprintf("rate limit exceeded: %s", errorMsg), retryAfter)
		default:
			if resp.StatusCode >= 500 {
				return nil, errors.NewServerErrorWithStatus(fmt.Sprintf("server error: %s", errorMsg), resp.StatusCode)
			}
			return nil, errors.NewValidationError(fmt.Sprintf("API error (%d): %s", resp.StatusCode, errorMsg))
		}
	}

	// Parse success response
	var apiResp NotifAIResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, errors.NewNetworkError("failed to parse response", err)
	}

	return &NotifAIResult{
		Response:  &apiResp,
		RateLimit: rateLimit,
	}, nil
}
