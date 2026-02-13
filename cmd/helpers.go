package cmd

import (
	"os"
	"strconv"
	"time"

	"github.com/Pincho-App/pincho-cli/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getTokenOptional retrieves the token from flags, env vars, or config (in that order)
// Returns empty string if not found (caller should validate)
func getTokenOptional(cmd *cobra.Command) string {
	// Try flag first
	token, _ := cmd.Flags().GetString("token")
	if token != "" {
		return token
	}

	// Try environment variable
	token = os.Getenv("PINCHO_TOKEN")
	if token != "" {
		return token
	}

	// Try config file
	token = viper.GetString("token")
	return token
}

// getAPIURL retrieves the API URL from env vars or config (in that order)
// Returns empty string if not found (client will use default)
func getAPIURL(cmd *cobra.Command) string {
	// Try environment variable first
	apiURL := os.Getenv("PINCHO_API_URL")
	if apiURL != "" {
		return apiURL
	}

	// Try config file
	apiURL = viper.GetString("api_url")
	return apiURL
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getTimeout retrieves the timeout from flags, env vars, config file, or returns default
// Priority: flag > env var > config file > default
// Returns timeout in seconds as time.Duration
func getTimeout(cmd *cobra.Command) time.Duration {
	// Try flag first
	if timeout, err := cmd.Flags().GetInt("timeout"); err == nil && timeout > 0 {
		return time.Duration(timeout) * time.Second
	}

	// Try environment variable
	if timeoutStr := os.Getenv("PINCHO_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil && timeout > 0 {
			return time.Duration(timeout) * time.Second
		}
	}

	// Try config file
	if timeout := viper.GetInt("timeout"); timeout > 0 {
		return time.Duration(timeout) * time.Second
	}

	// Return default
	return client.DefaultTimeout
}

// getMaxRetries retrieves the max retry count from flags, env vars, config file, or returns default
// Priority: flag > env var > config file > default
func getMaxRetries(cmd *cobra.Command) int {
	// Try flag first
	if retries, err := cmd.Flags().GetInt("max-retries"); err == nil && retries >= 0 {
		return retries
	}

	// Try environment variable
	if retriesStr := os.Getenv("PINCHO_MAX_RETRIES"); retriesStr != "" {
		if retries, err := strconv.Atoi(retriesStr); err == nil && retries >= 0 {
			return retries
		}
	}

	// Try config file (use viper.IsSet to distinguish 0 from unset)
	if viper.IsSet("max_retries") {
		return viper.GetInt("max_retries")
	}

	// Return default
	return client.DefaultMaxRetries
}

// getDefaultType retrieves the default notification type from config file
// Only checks config file (not flag or env var, as flags are command-specific)
// Returns empty string if not configured
func getDefaultType() string {
	return viper.GetString("default_type")
}

// getDefaultTags retrieves the default tags from config file
// Only checks config file (not flag or env var, as flags are command-specific)
// Returns empty slice if not configured
func getDefaultTags() []string {
	return viper.GetStringSlice("default_tags")
}

// mergeTypeWithDefault returns the provided type if non-empty, otherwise returns configured default
func mergeTypeWithDefault(providedType string) string {
	if providedType != "" {
		return providedType
	}
	return getDefaultType()
}

// mergeTagsWithDefaults merges provided tags with configured default tags
// Provided tags take precedence and are listed first
func mergeTagsWithDefaults(providedTags []string) []string {
	defaultTags := getDefaultTags()
	if len(defaultTags) == 0 {
		return providedTags
	}

	// If no provided tags, use defaults
	if len(providedTags) == 0 {
		return defaultTags
	}

	// Merge: provided tags first, then defaults (avoiding duplicates)
	seen := make(map[string]bool)
	merged := make([]string, 0, len(providedTags)+len(defaultTags))

	// Add provided tags first
	for _, tag := range providedTags {
		if !seen[tag] {
			merged = append(merged, tag)
			seen[tag] = true
		}
	}

	// Add default tags that aren't already present
	for _, tag := range defaultTags {
		if !seen[tag] {
			merged = append(merged, tag)
			seen[tag] = true
		}
	}

	return merged
}
