package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/viper"
)

// setupTestEnv creates a temporary config directory for testing
func setupTestEnv(t *testing.T) (string, func()) {
	t.Helper()

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "pincho-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Save original home
	originalHome := os.Getenv("HOME")

	// Set HOME to temp directory
	os.Setenv("HOME", tmpDir)

	// Reset viper between tests
	viper.Reset()

	// Cleanup function
	cleanup := func() {
		os.Setenv("HOME", originalHome)
		os.RemoveAll(tmpDir)
		viper.Reset()
	}

	return tmpDir, cleanup
}

func TestGetConfigDir(t *testing.T) {
	tmpHome, cleanup := setupTestEnv(t)
	defer cleanup()

	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() failed: %v", err)
	}

	expectedPath := filepath.Join(tmpHome, ConfigDirName)
	if configDir != expectedPath {
		t.Errorf("GetConfigDir() = %q, want %q", configDir, expectedPath)
	}
}

func TestGetConfigPath(t *testing.T) {
	tmpHome, cleanup := setupTestEnv(t)
	defer cleanup()

	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() failed: %v", err)
	}

	expectedPath := filepath.Join(tmpHome, ConfigDirName, ConfigFileName+".yaml")
	if configPath != expectedPath {
		t.Errorf("GetConfigPath() = %q, want %q", configPath, expectedPath)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	tmpHome, cleanup := setupTestEnv(t)
	defer cleanup()

	// Ensure directory is created
	err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() failed: %v", err)
	}

	// Check directory exists
	configDir := filepath.Join(tmpHome, ConfigDirName)
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Config directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Config path is not a directory")
	}

	// Check permissions (0700 = owner-only)
	if runtime.GOOS != "windows" {
		mode := info.Mode().Perm()
		expectedMode := os.FileMode(0700)
		if mode != expectedMode {
			t.Errorf("Config directory permissions = %o, want %o", mode, expectedMode)
		}
	}

	// Test idempotency - calling again should not error
	err = EnsureConfigDir()
	if err != nil {
		t.Errorf("EnsureConfigDir() on existing directory failed: %v", err)
	}
}

func TestInitConfig(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Should not error even if config file doesn't exist
	err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() failed: %v", err)
	}

	// Viper should be configured
	if viper.ConfigFileUsed() != "" {
		// Config file doesn't exist yet, so this should be empty
		t.Logf("Config file used: %s", viper.ConfigFileUsed())
	}
}

func TestSetAndGet(t *testing.T) {
	tmpHome, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "Set token",
			key:   "token",
			value: "test-token-123",
		},
		{
			name:  "Set API URL",
			key:   "api_url",
			value: "https://api.test.com",
		},
		{
			name:  "Set ID",
			key:   "id",
			value: "test-id-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set value
			err := Set(tt.key, tt.value)
			if err != nil {
				t.Fatalf("Set(%q, %q) failed: %v", tt.key, tt.value, err)
			}

			// Check config file exists
			configPath := filepath.Join(tmpHome, ConfigDirName, ConfigFileName+".yaml")
			info, err := os.Stat(configPath)
			if err != nil {
				t.Fatalf("Config file not created: %v", err)
			}

			// Check file permissions (0600 = owner read/write only)
			if runtime.GOOS != "windows" {
				mode := info.Mode().Perm()
				expectedMode := os.FileMode(0600)
				if mode != expectedMode {
					t.Errorf("Config file permissions = %o, want %o", mode, expectedMode)
				}
			}

			// Reset viper to force re-read from file
			viper.Reset()

			// Get value
			value, err := Get(tt.key)
			if err != nil {
				t.Fatalf("Get(%q) failed: %v", tt.key, err)
			}

			if value != tt.value {
				t.Errorf("Get(%q) = %q, want %q", tt.key, value, tt.value)
			}
		})
	}
}

func TestSetMultipleValues(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Set multiple values
	values := map[string]string{
		"token":   "token-123",
		"id":      "id-456",
		"api_url": "https://api.test.com",
	}

	for key, value := range values {
		err := Set(key, value)
		if err != nil {
			t.Fatalf("Set(%q, %q) failed: %v", key, value, err)
		}
	}

	// Reset viper to force re-read
	viper.Reset()

	// Get all values and verify
	for key, expectedValue := range values {
		value, err := Get(key)
		if err != nil {
			t.Fatalf("Get(%q) failed: %v", key, err)
		}
		if value != expectedValue {
			t.Errorf("Get(%q) = %q, want %q", key, value, expectedValue)
		}
	}
}

func TestGetNonExistentKey(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Get non-existent key should return empty string, not error
	value, err := Get("non_existent_key")
	if err != nil {
		t.Fatalf("Get(non_existent_key) failed: %v", err)
	}

	if value != "" {
		t.Errorf("Get(non_existent_key) = %q, want empty string", value)
	}
}

func TestGetAll(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Set multiple values
	expectedValues := map[string]string{
		"token":   "token-abc",
		"api_url": "https://api.example.com",
	}

	for key, value := range expectedValues {
		err := Set(key, value)
		if err != nil {
			t.Fatalf("Set(%q, %q) failed: %v", key, value, err)
		}
	}

	// Reset viper to force re-read
	viper.Reset()

	// Get all settings
	settings, err := GetAll()
	if err != nil {
		t.Fatalf("GetAll() failed: %v", err)
	}

	// Verify all expected values are present
	for key, expectedValue := range expectedValues {
		value, ok := settings[key]
		if !ok {
			t.Errorf("GetAll() missing key %q", key)
			continue
		}

		if value != expectedValue {
			t.Errorf("GetAll()[%q] = %v, want %q", key, value, expectedValue)
		}
	}
}

func TestLoad(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Set values
	err := Set("token", "test-token-789")
	if err != nil {
		t.Fatalf("Set(token) failed: %v", err)
	}

	err = Set("api_url", "https://api.load-test.com")
	if err != nil {
		t.Fatalf("Set(api_url) failed: %v", err)
	}

	// Reset viper to force re-read
	viper.Reset()

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify loaded values
	if cfg.Token != "test-token-789" {
		t.Errorf("Load() token = %q, want %q", cfg.Token, "test-token-789")
	}

	if cfg.APIURL != "https://api.load-test.com" {
		t.Errorf("Load() api_url = %q, want %q", cfg.APIURL, "https://api.load-test.com")
	}
}

func TestLoadWithoutConfigFile(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Load should not error even without config file
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() without config file failed: %v", err)
	}

	// Values should be empty
	if cfg.Token != "" {
		t.Errorf("Load() token = %q, want empty string", cfg.Token)
	}

	if cfg.APIURL != "" {
		t.Errorf("Load() api_url = %q, want empty string", cfg.APIURL)
	}
}

func TestSetOverwritesExistingValue(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Set initial value
	err := Set("token", "initial-token")
	if err != nil {
		t.Fatalf("Set(token, initial) failed: %v", err)
	}

	// Reset and verify initial value
	viper.Reset()
	value, err := Get("token")
	if err != nil {
		t.Fatalf("Get(token) after initial set failed: %v", err)
	}
	if value != "initial-token" {
		t.Errorf("Initial Get(token) = %q, want %q", value, "initial-token")
	}

	// Overwrite with new value
	err = Set("token", "updated-token")
	if err != nil {
		t.Fatalf("Set(token, updated) failed: %v", err)
	}

	// Reset and verify updated value
	viper.Reset()
	value, err = Get("token")
	if err != nil {
		t.Fatalf("Get(token) after update failed: %v", err)
	}
	if value != "updated-token" {
		t.Errorf("Updated Get(token) = %q, want %q", value, "updated-token")
	}
}
