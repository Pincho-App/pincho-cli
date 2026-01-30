// Package config provides configuration management for Pincho CLI.
//
// The package handles loading, saving, and retrieving configuration values
// from the user's config file (~/.pincho/config.yaml). Configuration
// files store persistent settings like API tokens, custom API URLs, and
// other preferences.
//
// Configuration Priority (highest to lowest):
//  1. Command-line flags (--token, --timeout, etc.)
//  2. Environment variables (PINCHO_TOKEN, PINCHO_API_URL, etc.)
//  3. Config file (~/.pincho/config.yaml)
//
// Security:
//   - Config directory created with 0700 permissions (owner-only access)
//   - Config files created with 0600 permissions (owner read/write only)
//   - Prevents token exposure to other users on shared systems
//
// Example usage:
//
//	// Set a value
//	err := config.Set("token", "your-api-token")
//
//	// Get a value
//	token, err := config.Get("token")
//
//	// Load all config
//	cfg, err := config.Load()
//
// Supported configuration keys:
//   - token: Pincho API token
//   - api_url: Custom API endpoint URL
//   - timeout: HTTP request timeout in seconds (overrides default 30s)
//   - max_retries: Maximum number of retry attempts (overrides default 3)
//   - default_type: Default notification type (e.g., "alert", "deploy", "info")
//   - default_tags: Default tags to include with all notifications (array of strings)
//
// Example config file (~/.pincho/config.yaml):
//
//	token: abc123
//	api_url: https://custom-api.example.com/send
//	timeout: 60
//	max_retries: 5
//	default_type: deploy
//	default_tags:
//	  - production
//	  - automated
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	// ConfigDirName is the name of the config directory
	ConfigDirName = ".pincho"

	// ConfigFileName is the name of the config file (without extension)
	ConfigFileName = "config"
)

// Config represents the Pincho CLI configuration
type Config struct {
	Token       string   `mapstructure:"token"`
	ID          string   `mapstructure:"id"`
	APIURL      string   `mapstructure:"api_url"`
	Timeout     int      `mapstructure:"timeout"`      // HTTP request timeout in seconds
	MaxRetries  int      `mapstructure:"max_retries"`  // Maximum number of retry attempts
	DefaultType string   `mapstructure:"default_type"` // Default notification type
	DefaultTags []string `mapstructure:"default_tags"` // Default tags to include with all notifications
}

// GetConfigDir returns the path to the config directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ConfigDirName), nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigFileName+".yaml"), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
// Uses 0700 permissions to ensure only the owner can read the config (which contains tokens)
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Use 0700 (owner-only) to protect tokens stored in config files
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

// InitConfig initializes the Viper configuration
func InitConfig() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Set config file location
	viper.AddConfigPath(configDir)
	viper.SetConfigName(ConfigFileName)
	viper.SetConfigType("yaml")

	// Set environment variable prefix
	viper.SetEnvPrefix("PINCHO")
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		// Ignore if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return nil
}

// Load loads the configuration from file and environment
func Load() (*Config, error) {
	if err := InitConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Set sets a configuration value and saves it to the config file
// Config file is created with 0600 permissions to protect sensitive data (tokens)
func Set(key, value string) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	if err := InitConfig(); err != nil {
		return err
	}

	viper.Set(key, value)

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Ensure config file has secure permissions (owner-only read/write)
	if err := os.Chmod(configPath, 0600); err != nil {
		return fmt.Errorf("failed to set config file permissions: %w", err)
	}

	return nil
}

// Get retrieves a configuration value
func Get(key string) (string, error) {
	if err := InitConfig(); err != nil {
		return "", err
	}

	value := viper.GetString(key)
	return value, nil
}

// GetAll returns all configuration values
func GetAll() (map[string]interface{}, error) {
	if err := InitConfig(); err != nil {
		return nil, err
	}

	return viper.AllSettings(), nil
}
