package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/wirepusher/cli/pkg/config"
)

var (
	// Version information (set by build flags)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wirepusher",
	Short: "Official CLI tool for WirePusher push notifications",
	Long: `WirePusher CLI is a command-line tool for sending push notifications
via the WirePusher API. Perfect for CI/CD pipelines, monitoring scripts,
and automation workflows.

Documentation: https://gitlab.com/wirepusher/cli
API Reference: https://wirepusher.com/docs`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Initialize configuration
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringP("token", "t", "", "WirePusher API token (env: WIREPUSHER_TOKEN)")
	// Deprecated: Legacy authentication. Use --token flag instead.
	rootCmd.PersistentFlags().StringP("id", "i", "", "[DEPRECATED] WirePusher user ID - use --token instead (env: WIREPUSHER_ID)")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if err := config.InitConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize config: %v\n", err)
	}
}

// GetVersionInfo returns version information as a formatted string
func GetVersionInfo() string {
	return fmt.Sprintf("wirepusher version %s\ncommit: %s\nbuilt: %s", version, commit, date)
}
