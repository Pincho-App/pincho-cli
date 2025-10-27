package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// DefaultAPIURL is the default WirePusher API endpoint
	DefaultAPIURL = "https://wirepusher.com/send"

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second
)

// Client represents a WirePusher API client
type Client struct {
	APIURL     string
	HTTPClient *http.Client
}

// SendOptions contains parameters for sending a notification
type SendOptions struct {
	Title     string   `json:"title"`
	Message   string   `json:"message"`
	Token     string   `json:"token"`
	ID        string   `json:"id"`
	Type      string   `json:"type,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	ImageURL  string   `json:"imageURL,omitempty"`
	ActionURL string   `json:"actionURL,omitempty"`
}

// SendResponse represents the API response
type SendResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// New creates a new WirePusher client
func New() *Client {
	return &Client{
		APIURL: DefaultAPIURL,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// Send sends a notification via the WirePusher v1 API
func (c *Client) Send(opts *SendOptions) error {
	// Validate required fields
	if opts.Title == "" {
		return fmt.Errorf("title is required")
	}
	if opts.Message == "" {
		return fmt.Errorf("message is required")
	}
	if opts.Token == "" {
		return fmt.Errorf("token is required")
	}
	if opts.ID == "" {
		return fmt.Errorf("id is required")
	}

	// Build request body
	jsonData, err := json.Marshal(opts)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", c.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error status codes
	if resp.StatusCode >= 400 {
		// Try to parse error message from response
		var apiResp SendResponse
		errorMsg := string(bodyBytes)
		if err := json.Unmarshal(bodyBytes, &apiResp); err == nil && apiResp.Message != "" {
			errorMsg = apiResp.Message
		}

		switch resp.StatusCode {
		case 400:
			return fmt.Errorf("validation error: %s", errorMsg)
		case 401, 403:
			return fmt.Errorf("authentication error: %s (check your token and id)", errorMsg)
		case 429:
			return fmt.Errorf("rate limit exceeded: %s", errorMsg)
		default:
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, errorMsg)
		}
	}

	return nil
}
