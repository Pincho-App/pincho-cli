// Package cmd implements the command-line interface for Pincho CLI.
//
// This package provides the main commands for interacting with the Pincho
// API, including:
//   - send: Send push notifications with title, message, and optional parameters
//   - notifai: Use AI to generate notifications from free-form text
//   - config: Manage CLI configuration settings
//   - version: Display version information
//
// Commands support configuration via flags, environment variables, or config
// files (in order of precedence: flags > env vars > config file).
//
// Global flags:
//
//	--token, -t: API token for authentication
//	--verbose: Enable detailed logging output
//	--timeout: HTTP request timeout in seconds
//	--max-retries: Maximum number of retry attempts
//
// Environment variables:
//
//	PINCHO_TOKEN: API token
//	PINCHO_API_URL: Custom API endpoint
//	PINCHO_TIMEOUT: Request timeout in seconds
//	PINCHO_MAX_RETRIES: Maximum retry attempts
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/pincho-app/pincho-cli/pkg/config"
	clierrors "gitlab.com/pincho-app/pincho-cli/pkg/errors"
	"gitlab.com/pincho-app/pincho-cli/pkg/logging"
)

var (
	// Version information (set by build flags)
	version = "dev"
	commit  = "none"
	date    = "unknown"

	// verbose enables detailed logging
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pincho",
	Short: "Official CLI tool for Pincho push notifications",
	Long: `Pincho CLI is a command-line tool for sending push notifications
via the Pincho API. Perfect for CI/CD pipelines, monitoring scripts,
and automation workflows.

Documentation: https://gitlab.com/pincho-app/pincho-cli
API Reference: https://pincho.app/help`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Enable verbose logging if flag is set
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			logging.SetVerbose(true)
			logging.Debug("Verbose logging enabled")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Handle errors with proper exit codes
		clierrors.HandleError(err)
	}
}

func init() {
	// Initialize configuration
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringP("token", "t", "", "Pincho API token (env: PINCHO_TOKEN)")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().Int("timeout", 30, "HTTP request timeout in seconds (env: PINCHO_TIMEOUT)")
	rootCmd.PersistentFlags().Int("max-retries", 3, "Maximum number of retry attempts (env: PINCHO_MAX_RETRIES)")
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if err := config.InitConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize config: %v\n", err)
	}
}

// GetVersionInfo returns version information as a formatted string
func GetVersionInfo() string {
	return fmt.Sprintf("pincho version %s\ncommit: %s\nbuilt: %s", version, commit, date)
}
