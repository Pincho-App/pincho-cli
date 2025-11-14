package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	client := New()

	if client.APIURL != DefaultAPIURL {
		t.Errorf("expected APIURL to be %s, got %s", DefaultAPIURL, client.APIURL)
	}

	if client.HTTPClient == nil {
		t.Error("expected HTTPClient to be initialized")
	}

	if client.HTTPClient.Timeout != DefaultTimeout {
		t.Errorf("expected timeout to be %v, got %v", DefaultTimeout, client.HTTPClient.Timeout)
	}
}

func TestClient_Send_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		// Verify Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Send success response
		w.WriteHeader(200)
		w.Write([]byte(`{"status": "success", "message": "Notification sent"}`))
	}))
	defer server.Close()

	// Create client and send notification
	client := New()
	client.APIURL = server.URL

	opts := &SendOptions{
		Title:   "Test Title",
		Message: "Test Message",
		Token:   "test-token",
	}

	result, err := client.Send(opts)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Error("expected result, got nil")
	}
	if result.Response == nil {
		t.Error("expected response, got nil")
	}
}

func TestClient_Send_WithAllOptions(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	// Create client and send notification with all options
	client := New()
	client.APIURL = server.URL

	opts := &SendOptions{
		Title:   "Test Title",
		Message: "Test Message",
		Token:   "test-token",

		Type:      "alert",
		Tags:      []string{"tag1", "tag2"},
		ImageURL:  "https://example.com/image.png",
		ActionURL: "https://example.com/action",
	}

	result, err := client.Send(opts)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Error("expected result, got nil")
	}
}

func TestClient_Send_ValidationErrors(t *testing.T) {
	client := New()

	tests := []struct {
		name    string
		opts    *SendOptions
		wantErr string
	}{
		{
			name: "missing title",
			opts: &SendOptions{
				Message: "Test",
				Token:   "token",
			},
			wantErr: "title is required",
		},
		{
			name: "missing token",
			opts: &SendOptions{
				Title:   "Test",
				Message: "Test",
			},
			wantErr: "token is required",
		},
		{
			name: "invalid tags - too many",
			opts: &SendOptions{
				Title: "Test",
				Token: "token",
				Tags:  []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
			},
			wantErr: "maximum of 10 tags allowed",
		},
		{
			name: "invalid tags - too long",
			opts: &SendOptions{
				Title: "Test",
				Token: "token",
				Tags:  []string{"this-is-a-very-long-tag-name-that-exceeds-the-fifty-character-limit-by-far"},
			},
			wantErr: "exceeds maximum length of 50 characters",
		},
		{
			name: "invalid tags - invalid characters",
			opts: &SendOptions{
				Title: "Test",
				Token: "token",
				Tags:  []string{"invalid tag with spaces"},
			},
			wantErr: "contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.Send(tt.opts)
			if err == nil {
				t.Error("expected error, got nil")
			}
			if result != nil {
				t.Error("expected nil result on error, got result")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing '%s', got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestClient_Send_HTTPErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    string
	}{
		{
			name:       "400 validation error",
			statusCode: 400,
			response:   `{"status": "error", "message": "Invalid parameters"}`,
			wantErr:    "validation error",
		},
		{
			name:       "401 auth error",
			statusCode: 401,
			response:   `{"status": "error", "message": "Unauthorized"}`,
			wantErr:    "authentication error",
		},
		{
			name:       "403 auth error",
			statusCode: 403,
			response:   `{"status": "error", "message": "Forbidden"}`,
			wantErr:    "authentication error",
		},
		{
			name:       "429 rate limit",
			statusCode: 429,
			response:   `{"status": "error", "message": "Rate limit exceeded"}`,
			wantErr:    "rate limit exceeded",
		},
		{
			name:       "500 server error",
			statusCode: 500,
			response:   `{"status": "error", "message": "Internal server error"}`,
			wantErr:    "API error (500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := New()
			client.APIURL = server.URL

			opts := &SendOptions{
				Title:   "Test",
				Message: "Test",
				Token:   "token",
			}

			result, err := client.Send(opts)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if result != nil {
				t.Error("expected nil result on error, got result")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing '%s', got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestClient_Send_WithEncryption(t *testing.T) {
	// Track the request body to verify encryption occurred
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and store the request body
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])

		w.WriteHeader(200)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	client := New()
	client.APIURL = server.URL

	opts := &SendOptions{
		Title:              "Test Title",
		Message:            "Secret message",
		Token:              "test-token",
		EncryptionPassword: "test-password",
	}

	result, err := client.Send(opts)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// Verify that the message was encrypted (not the original plaintext)
	if strings.Contains(receivedBody, "Secret message") {
		t.Error("expected message to be encrypted, but found plaintext in request body")
	}

	// Verify IV was sent
	if !strings.Contains(receivedBody, "\"iv\":") {
		t.Error("expected IV field in request body when encryption is enabled")
	}
}
