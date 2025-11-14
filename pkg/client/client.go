// Package client provides a Go client for the WirePusher API.
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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.com/wirepusher/cli/pkg/crypto"
	"gitlab.com/wirepusher/cli/pkg/validation"
)

const (
	// DefaultAPIURL is the default WirePusher API endpoint (for send)
	DefaultAPIURL = "https://api.wirepusher.dev/send"

	// DefaultNotifAIURL is the default WirePusher NotifAI API endpoint
	DefaultNotifAIURL = "https://api.wirepusher.dev/notifai"

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default maximum number of retry attempts
	DefaultMaxRetries = 3

	// DefaultInitialBackoff is the default initial backoff duration for retries
	DefaultInitialBackoff = 1 * time.Second
)

// Client represents a WirePusher API client
type Client struct {
	APIURL         string
	HTTPClient     *http.Client
	Timeout        time.Duration // Custom timeout duration (uses DefaultTimeout if zero)
	MaxRetries     int           // Maximum number of retry attempts (uses DefaultMaxRetries if zero)
	InitialBackoff time.Duration // Initial backoff duration for retries (uses DefaultInitialBackoff if zero)
}

// SendOptions contains parameters for sending a notification
type SendOptions struct {
	Title              string   `json:"title"`
	Message            string   `json:"message"`
	Token              string   `json:"token,omitempty"`
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

// NotificationDetails represents a single notification object from the API
type NotificationDetails struct {
	NotificationID string   `json:"notificationID"`
	UserID         string   `json:"userID"`
	Title          string   `json:"title"`
	Body           string   `json:"body"`
	Type           string   `json:"type"`
	ImageURL       string   `json:"imageURL,omitempty"`
	ActionURL      string   `json:"actionURL,omitempty"`
	Timestamp      string   `json:"timestamp"`
	Tags           []string `json:"tags,omitempty"`
	TeamID         string   `json:"teamId,omitempty"`
	TeamName       string   `json:"teamName,omitempty"`
	Endpoint       string   `json:"endpoint"`
	IV             string   `json:"iv,omitempty"`
	ExpiresAt      string   `json:"expiresAt"`
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
	Text  string `json:"text"`
	Token string `json:"token,omitempty"`
	Type  string `json:"type,omitempty"`
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

// New creates a new WirePusher client with default settings
func New() *Client {
	return &Client{
		APIURL:         DefaultAPIURL,
		Timeout:        DefaultTimeout,
		MaxRetries:     DefaultMaxRetries,
		InitialBackoff: DefaultInitialBackoff,
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

	errStr := err.Error()

	// Retry on network errors
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
func (c *Client) calculateBackoff(attempt int, statusCode int) time.Duration {
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
func (c *Client) doRequestWithRetry(req *http.Request) (*http.Response, error) {
	maxRetries := c.MaxRetries
	if maxRetries == 0 {
		maxRetries = DefaultMaxRetries
	}

	var lastErr error
	var lastStatusCode int

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Clone the request for retry (body needs to be reset)
		reqClone := req.Clone(req.Context())
		if req.Body != nil {
			// For POST requests, we need to reset the body
			if req.GetBody != nil {
				body, err := req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("failed to reset request body: %w", err)
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
			backoff := c.calculateBackoff(attempt, lastStatusCode)
			time.Sleep(backoff)
		}
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
	}

	// This shouldn't happen, but just in case
	return nil, fmt.Errorf("request failed after %d retries with status %d", maxRetries, lastStatusCode)
}

// Send sends a notification via the WirePusher v1 API
// Returns SendResult with response details and rate limit info, or error if failed
func (c *Client) Send(opts *SendOptions) (*SendResult, error) {
	// Validate required fields
	if opts.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	if opts.Token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Normalize and validate tags
	if len(opts.Tags) > 0 {
		normalizedTags, err := validation.NormalizeAndValidateTags(opts.Tags)
		if err != nil {
			return nil, fmt.Errorf("tag validation failed: %w", err)
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
				return nil, fmt.Errorf("failed to generate IV: %w", err)
			}

			encrypted, err := crypto.EncryptMessage(opts.Message, opts.EncryptionPassword, ivBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt message: %w", err)
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
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with GetBody for retries
	req, err := http.NewRequest("POST", c.APIURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set GetBody for retry support
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(jsonData)), nil
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request with retry logic
	resp, err := c.doRequestWithRetry(req)
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
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error status codes
	if resp.StatusCode >= 400 {
		// Try to parse nested error response
		var errorResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Error.Message != "" {
			// Format error message with details
			errorMsg := errorResp.Error.Message
			if errorResp.Error.Param != "" {
				errorMsg = fmt.Sprintf("%s (parameter: %s)", errorMsg, errorResp.Error.Param)
			}

			// Add error code if available
			if errorResp.Error.Code != "" {
				errorMsg = fmt.Sprintf("%s [%s]", errorMsg, errorResp.Error.Code)
			}

			return nil, fmt.Errorf("%s", errorMsg)
		}

		// Fallback to generic error message if parsing fails
		errorMsg := string(bodyBytes)
		switch resp.StatusCode {
		case 400:
			return nil, fmt.Errorf("validation error: %s", errorMsg)
		case 401, 403:
			return nil, fmt.Errorf("authentication error: %s (check your token)", errorMsg)
		case 429:
			return nil, fmt.Errorf("rate limit exceeded: %s", errorMsg)
		default:
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errorMsg)
		}
	}

	// Parse success response
	var apiResp SendResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &SendResult{
		Response:  &apiResp,
		RateLimit: rateLimit,
	}, nil
}

// NotifAI sends a text-to-notification request via the WirePusher NotifAI API
// Returns NotifAIResult with response details and rate limit info, or error if failed
func (c *Client) NotifAI(opts *NotifAIOptions) (*NotifAIResult, error) {
	// Validate required fields
	if opts.Text == "" {
		return nil, fmt.Errorf("text is required")
	}

	if opts.Token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Validate text length
	if len(opts.Text) < 5 {
		return nil, fmt.Errorf("text must be at least 5 characters long")
	}
	if len(opts.Text) > 2500 {
		return nil, fmt.Errorf("text must be at most 2500 characters long")
	}

	// Build request body
	jsonData, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Determine NotifAI URL based on configured API URL
	notifaiURL := DefaultNotifAIURL
	if c.APIURL != "" && c.APIURL != DefaultAPIURL {
		// If custom API URL is set, derive NotifAI URL from it
		// Replace /send with /notifai
		notifaiURL = strings.Replace(c.APIURL, "/send", "/notifai", 1)
	}

	// Create HTTP request with GetBody for retries
	req, err := http.NewRequest("POST", notifaiURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set GetBody for retry support
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(jsonData)), nil
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request with retry logic
	resp, err := c.doRequestWithRetry(req)
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
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error status codes
	if resp.StatusCode >= 400 {
		// Try to parse nested error response
		var errorResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Error.Message != "" {
			// Format error message with details
			errorMsg := errorResp.Error.Message
			if errorResp.Error.Param != "" {
				errorMsg = fmt.Sprintf("%s (parameter: %s)", errorMsg, errorResp.Error.Param)
			}

			// Add error code if available
			if errorResp.Error.Code != "" {
				errorMsg = fmt.Sprintf("%s [%s]", errorMsg, errorResp.Error.Code)
			}

			return nil, fmt.Errorf("%s", errorMsg)
		}

		// Fallback to generic error message if parsing fails
		errorMsg := string(bodyBytes)
		switch resp.StatusCode {
		case 400:
			return nil, fmt.Errorf("validation error: %s", errorMsg)
		case 401, 403:
			return nil, fmt.Errorf("authentication error: %s (check your token)", errorMsg)
		case 429:
			return nil, fmt.Errorf("rate limit exceeded: %s", errorMsg)
		default:
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errorMsg)
		}
	}

	// Parse success response
	var apiResp NotifAIResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &NotifAIResult{
		Response:  &apiResp,
		RateLimit: rateLimit,
	}, nil
}
