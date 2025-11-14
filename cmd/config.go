package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/wirepusher/cli/pkg/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage WirePusher CLI configuration",
	Long: `Manage configuration settings for the WirePusher CLI.

Configuration is stored in ~/.wirepusher/config.yaml and can be set, retrieved,
or listed using the subcommands.

Priority order for configuration values:
  1. Command-line flags (--token)
  2. Environment variables (WIREPUSHER_TOKEN)
  3. Config file (~/.wirepusher/config.yaml)

Examples:
  # Set configuration values
  wirepusher config set token wpt_abc123xyz

  # Get a specific value
  wirepusher config get token

  # List all configuration
  wirepusher config list
`,
}

// configSetCmd represents the 'config set' command
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value and save it to the config file.

Supported keys:
  - token: Your WirePusher API token

Example:
  wirepusher config set token wpt_abc123xyz
`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

// configGetCmd represents the 'config get' command
var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get a specific configuration value from the config file or environment.

Example:
  wirepusher config get token
`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

// configListCmd represents the 'config list' command
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long: `List all configuration values from the config file.

Example:
  wirepusher config list
`,
	RunE: runConfigList,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Validate key
	if key != "token" {
		return fmt.Errorf("invalid key '%s' (supported: token)", key)
	}

	if err := config.Set(key, value); err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("✓ Set %s in %s\n", key, configPath)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	value, err := config.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	if value == "" {
		fmt.Printf("%s: (not set)\n", key)
	} else {
		// Mask sensitive values
		if key == "token" && len(value) > 8 {
			fmt.Printf("%s: %s...%s\n", key, value[:4], value[len(value)-4:])
		} else {
			fmt.Printf("%s: %s\n", key, value)
		}
	}

	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	all, err := config.GetAll()
	if err != nil {
		return fmt.Errorf("failed to list config: %w", err)
	}

	if len(all) == 0 {
		fmt.Println("No configuration set")
		fmt.Println("\nTo get started:")
		fmt.Println("  wirepusher config set token YOUR_TOKEN")
		return nil
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("Configuration from %s:\n\n", configPath)

	for key, value := range all {
		valueStr := fmt.Sprintf("%v", value)

		// Mask sensitive values
		if key == "token" && len(valueStr) > 8 {
			fmt.Printf("  %s: %s...%s\n", key, valueStr[:4], valueStr[len(valueStr)-4:])
		} else {
			fmt.Printf("  %s: %s\n", key, valueStr)
		}
	}

	return nil
}
