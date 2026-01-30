// Package logging provides structured logging utilities for Pincho CLI.
//
// The package uses Go's standard log/slog for structured, leveled logging.
// Output is controlled by the --verbose flag, with all output going to stderr
// to keep stdout clean for structured output (like JSON responses).
//
// Log Levels:
//   - Debug: Detailed debugging information (only shown with --verbose flag)
//   - Info: Informational messages shown to all users
//   - Error: Error messages shown to all users
//
// All log output goes to stderr to keep stdout clean for JSON output.
//
// Example usage:
//
//	logging.SetVerbose(true)  // Set by --verbose flag
//	logging.Debug("Using token", "token_prefix", token[:8])
//	logging.Info("Notification sent")
//	logging.Error("Failed to connect", "error", err)
//
// The structured logging supports key-value pairs for better observability:
//
//	logging.Debug("Request configured",
//	    "api_url", "https://api.pincho.app/send",
//	    "timeout", "30s",
//	    "max_retries", 3)
package logging

import (
	"log/slog"
	"os"
)

var (
	// logger is the global logger instance
	logger *slog.Logger

	// verboseEnabled tracks if verbose mode is active
	verboseEnabled = false
)

func init() {
	// Initialize with default logger (Info level, text handler to stderr)
	updateLogger()
}

// updateLogger recreates the logger with current verbosity settings
func updateLogger() {
	level := slog.LevelInfo
	if verboseEnabled {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
		// Custom replacement to format output for CLI use
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove time for cleaner output
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			// Simplify level names
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				switch level {
				case slog.LevelDebug:
					a.Value = slog.StringValue("VERBOSE")
				case slog.LevelInfo:
					a.Value = slog.StringValue("INFO")
				case slog.LevelError:
					a.Value = slog.StringValue("ERROR")
				}
			}
			return a
		},
	}

	logger = slog.New(slog.NewTextHandler(os.Stderr, opts))
}

// SetVerbose enables or disables verbose logging
func SetVerbose(enabled bool) {
	verboseEnabled = enabled
	updateLogger()
}

// IsVerbose returns whether verbose logging is enabled
func IsVerbose() bool {
	return verboseEnabled
}

// Debug logs a debug-level message (only shown when verbose mode is enabled)
// Accepts key-value pairs for structured logging
func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

// Info logs an info-level message
// Accepts key-value pairs for structured logging
func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

// Error logs an error-level message
// Accepts key-value pairs for structured logging
func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

// With returns a new logger with additional attributes
// Useful for adding context that applies to multiple log messages
func With(args ...any) *slog.Logger {
	return logger.With(args...)
}

// GetLogger returns the underlying slog.Logger for advanced use cases
func GetLogger() *slog.Logger {
	return logger
}
