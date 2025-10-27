package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	// ConfigDirName is the name of the config directory
	ConfigDirName = ".wirepusher"

	// ConfigFileName is the name of the config file (without extension)
	ConfigFileName = "config"
)

// Config represents the WirePusher CLI configuration
type Config struct {
	Token string `mapstructure:"token"`
	ID    string `mapstructure:"id"`
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
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
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
	viper.SetEnvPrefix("WIREPUSHER")
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
