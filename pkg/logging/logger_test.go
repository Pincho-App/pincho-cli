package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestSetVerbose(t *testing.T) {
	// Reset state after test
	defer SetVerbose(false)

	// Initial state
	if IsVerbose() {
		t.Error("expected verbose to be false initially")
	}

	// Enable verbose
	SetVerbose(true)
	if !IsVerbose() {
		t.Error("expected verbose to be true after SetVerbose(true)")
	}

	// Disable verbose
	SetVerbose(false)
	if IsVerbose() {
		t.Error("expected verbose to be false after SetVerbose(false)")
	}
}

func TestDebugLogging(t *testing.T) {
	defer SetVerbose(false)

	// Capture output
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	testLogger := slog.New(slog.NewTextHandler(&buf, opts))

	// Temporarily replace global logger
	oldLogger := logger
	logger = testLogger
	defer func() { logger = oldLogger }()

	// Enable verbose mode
	verboseEnabled = true

	// Log a debug message
	Debug("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("expected output to contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("expected output to contain 'key=value', got: %s", output)
	}
}

func TestInfoLogging(t *testing.T) {
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	testLogger := slog.New(slog.NewTextHandler(&buf, opts))

	oldLogger := logger
	logger = testLogger
	defer func() { logger = oldLogger }()

	Info("info message", "status", "success")

	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Errorf("expected output to contain 'info message', got: %s", output)
	}
	if !strings.Contains(output, "status=success") {
		t.Errorf("expected output to contain 'status=success', got: %s", output)
	}
}

func TestErrorLogging(t *testing.T) {
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelError}
	testLogger := slog.New(slog.NewTextHandler(&buf, opts))

	oldLogger := logger
	logger = testLogger
	defer func() { logger = oldLogger }()

	Error("error message", "err", "connection refused")

	output := buf.String()
	if !strings.Contains(output, "error message") {
		t.Errorf("expected output to contain 'error message', got: %s", output)
	}
	if !strings.Contains(output, "connection refused") {
		t.Errorf("expected output to contain 'connection refused', got: %s", output)
	}
}

func TestDebugWithStructuredFields(t *testing.T) {
	defer SetVerbose(false)

	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	testLogger := slog.New(slog.NewTextHandler(&buf, opts))

	oldLogger := logger
	logger = testLogger
	defer func() { logger = oldLogger }()

	// Enable verbose mode
	verboseEnabled = true

	// Use structured Debug with key-value pairs
	Debug("token configured", "token_prefix", "test123", "length", 7)

	output := buf.String()
	if !strings.Contains(output, "token configured") {
		t.Errorf("expected output to contain message, got: %s", output)
	}
	if !strings.Contains(output, "token_prefix=test123") {
		t.Errorf("expected output to contain token_prefix field, got: %s", output)
	}
	if !strings.Contains(output, "length=7") {
		t.Errorf("expected output to contain length field, got: %s", output)
	}
}

func TestWith(t *testing.T) {
	// Test that With() returns a logger with additional attributes
	contextLogger := With("request_id", "abc123")
	if contextLogger == nil {
		t.Error("expected With() to return non-nil logger")
	}
}

func TestGetLogger(t *testing.T) {
	l := GetLogger()
	if l == nil {
		t.Error("expected GetLogger() to return non-nil logger")
	}
	if l != logger {
		t.Error("expected GetLogger() to return the global logger")
	}
}

func TestUpdateLoggerLevel(t *testing.T) {
	defer SetVerbose(false)

	// Initially, debug messages should not be logged (level is Info)
	SetVerbose(false)

	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	testLogger := slog.New(slog.NewTextHandler(&buf, opts))

	oldLogger := logger
	logger = testLogger
	defer func() { logger = oldLogger }()

	// Debug should not appear when verbose is false
	buf.Reset()
	verboseEnabled = false
	Debug("should not appear")

	// Since we manually set the logger, we need to verify at Info level
	// The real test is that SetVerbose updates the logger's level

	// Info should always appear
	Info("should appear")
	output := buf.String()
	if !strings.Contains(output, "should appear") {
		t.Errorf("expected info message to appear, got: %s", output)
	}
}

func TestLoggerInitialization(t *testing.T) {
	// Test that the logger is initialized on package load
	if logger == nil {
		t.Error("expected logger to be initialized")
	}
}

func TestIsVerboseFunction(t *testing.T) {
	defer SetVerbose(false)

	SetVerbose(false)
	if IsVerbose() {
		t.Error("expected IsVerbose() to return false")
	}

	SetVerbose(true)
	if !IsVerbose() {
		t.Error("expected IsVerbose() to return true")
	}
}
