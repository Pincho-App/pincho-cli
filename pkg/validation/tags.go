// Package validation provides input validation utilities for Pincho CLI.
//
// The package implements validation and normalization logic for notification
// parameters, ensuring they meet API requirements before submission.
//
// Tag Validation Rules:
//   - Maximum 10 tags per notification
//   - Maximum 50 characters per tag
//   - Allowed characters: lowercase letters, numbers, hyphens, underscores
//   - Tags are automatically normalized: lowercased, trimmed, deduplicated
//
// Example usage:
//
//	tags := []string{"Production", "DEPLOY ", "production", "release-v1"}
//	normalized, err := validation.NormalizeAndValidateTags(tags)
//	// Returns: ["production", "deploy", "release-v1"] (normalized and deduplicated)
//
// The validation logic matches the Pincho API requirements to provide
// early client-side validation and better error messages.
package validation

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// MaxTags is the maximum number of tags allowed per notification
	MaxTags = 10

	// MaxTagLength is the maximum length of a single tag
	MaxTagLength = 50

	// TagPattern defines the valid characters for tags (alphanumeric, hyphens, underscores)
	TagPattern = `^[a-z0-9_-]+$`
)

var tagRegex = regexp.MustCompile(TagPattern)

// NormalizeAndValidateTags normalizes and validates a list of tags according to backend rules.
// Normalization: lowercase, trim whitespace, remove duplicates, filter empty strings
// Validation: max 10 tags, max 50 chars each, pattern ^[a-z0-9_-]+$
//
// Returns normalized tags or error if validation fails.
func NormalizeAndValidateTags(tags []string) ([]string, error) {
	if len(tags) == 0 {
		return []string{}, nil
	}

	// Normalize tags
	seen := make(map[string]bool)
	normalized := make([]string, 0, len(tags))

	for _, tag := range tags {
		// Trim whitespace and convert to lowercase
		tag = strings.TrimSpace(tag)
		tag = strings.ToLower(tag)

		// Skip empty strings
		if tag == "" {
			continue
		}

		// Skip duplicates
		if seen[tag] {
			continue
		}
		seen[tag] = true

		// Validate tag length
		if len(tag) > MaxTagLength {
			return nil, fmt.Errorf("tag '%s' exceeds maximum length of %d characters", tag, MaxTagLength)
		}

		// Validate tag pattern
		if !tagRegex.MatchString(tag) {
			return nil, fmt.Errorf("tag '%s' contains invalid characters (only lowercase letters, numbers, hyphens, and underscores allowed)", tag)
		}

		normalized = append(normalized, tag)
	}

	// Validate tag count
	if len(normalized) > MaxTags {
		return nil, fmt.Errorf("maximum of %d tags allowed, got %d", MaxTags, len(normalized))
	}

	return normalized, nil
}
