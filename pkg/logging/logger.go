// Package logging provides logging utilities for WirePusher CLI.
//
// The package implements a simple logging system with support for verbose
// output controlled by the --verbose flag. Logging is designed to help
// users debug issues without cluttering normal output.
//
// Log Levels:
//   - Verbose: Detailed debugging information (only shown with --verbose flag)
//   - Info: Informational messages shown to all users
//   - Error: Error messages shown to all users
//
// All output goes to stderr to keep stdout clean for structured output
// (like JSON responses with --json flag).
//
// Example usage:
//
//	logging.VerboseEnabled = true  // Set by --verbose flag
//	logging.Verbose("Using token: %s...", token[:8])
//	logging.Info("Notification sent successfully")
//	logging.Error("Failed to connect: %v", err)
//
// Verbose logging includes:
//   - Token usage (first 8 characters only)
//   - API URL configuration
//   - Timeout and retry configuration
//   - Request progress and timing
package logging

import (
	"fmt"
	"os"
)

var (
	// VerboseEnabled controls whether verbose logging is enabled
	VerboseEnabled = false
)

// Verbose prints a message only if verbose logging is enabled
func Verbose(format string, args ...interface{}) {
	if VerboseEnabled {
		fmt.Fprintf(os.Stderr, "[VERBOSE] "+format+"\n", args...)
	}
}

// Info prints an informational message
func Info(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// Error prints an error message
func Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}
