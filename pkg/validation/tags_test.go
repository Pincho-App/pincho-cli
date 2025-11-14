package validation

import (
	"strings"
	"testing"
)

func TestNormalizeAndValidateTags_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty tags",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single valid tag",
			input:    []string{"production"},
			expected: []string{"production"},
		},
		{
			name:     "multiple valid tags",
			input:    []string{"production", "backend", "critical"},
			expected: []string{"production", "backend", "critical"},
		},
		{
			name:     "uppercase to lowercase",
			input:    []string{"PRODUCTION", "Backend", "CriTicAl"},
			expected: []string{"production", "backend", "critical"},
		},
		{
			name:     "whitespace trimming",
			input:    []string{"  production  ", " backend ", "critical   "},
			expected: []string{"production", "backend", "critical"},
		},
		{
			name:     "duplicate removal",
			input:    []string{"production", "backend", "production", "critical", "backend"},
			expected: []string{"production", "backend", "critical"},
		},
		{
			name:     "empty strings filtered",
			input:    []string{"production", "", "backend", "  ", "critical"},
			expected: []string{"production", "backend", "critical"},
		},
		{
			name:     "hyphens allowed",
			input:    []string{"production-server", "api-v1", "feature-flag"},
			expected: []string{"production-server", "api-v1", "feature-flag"},
		},
		{
			name:     "underscores allowed",
			input:    []string{"production_server", "api_v1", "feature_flag"},
			expected: []string{"production_server", "api_v1", "feature_flag"},
		},
		{
			name:     "numbers allowed",
			input:    []string{"server1", "v2", "test123"},
			expected: []string{"server1", "v2", "test123"},
		},
		{
			name:     "combined normalization",
			input:    []string{"  PRODUCTION  ", "Backend", "production", "", "BACKEND", "critical"},
			expected: []string{"production", "backend", "critical"},
		},
		{
			name:     "exactly 10 tags",
			input:    []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			expected: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
		},
		{
			name:     "exactly 50 chars",
			input:    []string{strings.Repeat("a", 50)},
			expected: []string{strings.Repeat("a", 50)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeAndValidateTags(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d tags, got %d: %v", len(tt.expected), len(result), result)
			}

			for i, tag := range result {
				if i >= len(tt.expected) {
					t.Errorf("unexpected tag at index %d: %s", i, tag)
					continue
				}
				if tag != tt.expected[i] {
					t.Errorf("expected tag %d to be '%s', got '%s'", i, tt.expected[i], tag)
				}
			}
		})
	}
}

func TestNormalizeAndValidateTags_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		wantErr string
	}{
		{
			name:    "too many tags - 11",
			input:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
			wantErr: "maximum of 10 tags allowed, got 11",
		},
		{
			name:    "too many tags - 15",
			input:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"},
			wantErr: "maximum of 10 tags allowed, got 15",
		},
		{
			name:    "tag too long - 51 chars",
			input:   []string{strings.Repeat("a", 51)},
			wantErr: "exceeds maximum length of 50 characters",
		},
		{
			name:    "tag too long - 100 chars",
			input:   []string{strings.Repeat("a", 100)},
			wantErr: "exceeds maximum length of 50 characters",
		},
		{
			name:    "spaces not allowed",
			input:   []string{"production server"},
			wantErr: "contains invalid characters",
		},
		{
			name:    "special chars - @",
			input:   []string{"production@server"},
			wantErr: "contains invalid characters",
		},
		{
			name:    "special chars - #",
			input:   []string{"production#server"},
			wantErr: "contains invalid characters",
		},
		{
			name:    "special chars - $",
			input:   []string{"production$server"},
			wantErr: "contains invalid characters",
		},
		{
			name:    "special chars - .",
			input:   []string{"production.server"},
			wantErr: "contains invalid characters",
		},
		{
			name:    "special chars - /",
			input:   []string{"production/server"},
			wantErr: "contains invalid characters",
		},
		{
			name:    "uppercase letters before normalization fail",
			input:   []string{"PRODUCTION SERVER"},
			wantErr: "contains invalid characters",
		},
		{
			name:    "mixed valid and invalid - spaces",
			input:   []string{"production", "backend server", "critical"},
			wantErr: "contains invalid characters",
		},
		{
			name:    "mixed valid and invalid - special chars",
			input:   []string{"production", "backend@server", "critical"},
			wantErr: "contains invalid characters",
		},
		{
			name:    "mixed valid and invalid - too long",
			input:   []string{"production", strings.Repeat("a", 51), "critical"},
			wantErr: "exceeds maximum length of 50 characters",
		},
		{
			name:    "too many after normalization",
			input:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "", ""},
			wantErr: "maximum of 10 tags allowed, got 11",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeAndValidateTags(tt.input)
			if err == nil {
				t.Fatalf("expected error containing '%s', got nil (result: %v)", tt.wantErr, result)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing '%s', got: %v", tt.wantErr, err)
			}

			// Result should be nil on error
			if result != nil {
				t.Errorf("expected nil result on error, got: %v", result)
			}
		})
	}
}

func TestNormalizeAndValidateTags_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
		wantErr  bool
	}{
		{
			name:     "all empty strings",
			input:    []string{"", "", ""},
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "all whitespace",
			input:    []string{"   ", "\t", "\n"},
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "duplicates after normalization",
			input:    []string{"Production", "PRODUCTION", "production"},
			expected: []string{"production"},
			wantErr:  false,
		},
		{
			name:     "whitespace and duplicates",
			input:    []string{"  production  ", "Production", "  PRODUCTION"},
			expected: []string{"production"},
			wantErr:  false,
		},
		{
			name:     "single hyphen",
			input:    []string{"-"},
			expected: []string{"-"},
			wantErr:  false,
		},
		{
			name:     "single underscore",
			input:    []string{"_"},
			expected: []string{"_"},
			wantErr:  false,
		},
		{
			name:     "single number",
			input:    []string{"1"},
			expected: []string{"1"},
			wantErr:  false,
		},
		{
			name:     "hyphen at start",
			input:    []string{"-production"},
			expected: []string{"-production"},
			wantErr:  false,
		},
		{
			name:     "hyphen at end",
			input:    []string{"production-"},
			expected: []string{"production-"},
			wantErr:  false,
		},
		{
			name:     "underscore at start",
			input:    []string{"_production"},
			expected: []string{"_production"},
			wantErr:  false,
		},
		{
			name:     "underscore at end",
			input:    []string{"production_"},
			expected: []string{"production_"},
			wantErr:  false,
		},
		{
			name:     "number at start",
			input:    []string{"1production"},
			expected: []string{"1production"},
			wantErr:  false,
		},
		{
			name:     "all numbers",
			input:    []string{"123456789"},
			expected: []string{"123456789"},
			wantErr:  false,
		},
		{
			name:     "all hyphens and underscores",
			input:    []string{"-_-_-"},
			expected: []string{"-_-_-"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeAndValidateTags(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil (result: %v)", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d tags, got %d: %v", len(tt.expected), len(result), result)
			}

			for i, tag := range result {
				if i >= len(tt.expected) {
					t.Errorf("unexpected tag at index %d: %s", i, tag)
					continue
				}
				if tag != tt.expected[i] {
					t.Errorf("expected tag %d to be '%s', got '%s'", i, tt.expected[i], tag)
				}
			}
		})
	}
}
