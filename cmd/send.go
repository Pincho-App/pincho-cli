package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/wirepusher/cli/pkg/client"
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send [title] [message]",
	Short: "Send a push notification",
	Long: `Send a push notification via WirePusher.

The title and message can be provided as positional arguments or via flags.
Authentication requires EITHER a token OR a user ID (not both) from flags,
environment variables, or the config file.

Examples:
  # Simple notification (using token)
  wirepusher send "Build Complete" "Deploy finished successfully"

  # With notification type
  wirepusher send "Alert" "CPU usage high" --type alert

  # With tags
  wirepusher send "Deploy" "v1.2.3 deployed" --tag production --tag release

  # With image and action URL
  wirepusher send "Success" "All tests passed" \
    --image https://example.com/success.png \
    --action https://example.com/build/123

  # With encryption (message encrypted with AES-128-CBC)
  wirepusher send "Secure Alert" "Sensitive data here" \
    --encryption-password "secret123" \
    --type secure

  # Read message from stdin with encryption
  echo "Confidential report" | wirepusher send "Report" --stdin \
    --encryption-password "secret123"

  # Using user ID instead of token
  wirepusher send "Test" "Message" --id user123

  # Override config with flags (token OR id, not both)
  wirepusher send "Test" "Message" --token abc123
`,
	RunE: runSend,
}

var (
	sendType               string
	sendTags               []string
	sendImageURL           string
	sendActionURL          string
	sendStdin              bool
	sendEncryptionPassword string
)

func init() {
	rootCmd.AddCommand(sendCmd)

	// Flags specific to send command
	sendCmd.Flags().StringVar(&sendType, "type", "", "Notification type (e.g., alert, info, success)")
	sendCmd.Flags().StringSliceVar(&sendTags, "tag", []string{}, "Tags for categorization (can be used multiple times)")
	sendCmd.Flags().StringVar(&sendImageURL, "image", "", "Image URL to display with notification")
	sendCmd.Flags().StringVar(&sendActionURL, "action", "", "Action URL to open when notification is tapped")
	sendCmd.Flags().BoolVar(&sendStdin, "stdin", false, "Read message from stdin")
	sendCmd.Flags().StringVar(&sendEncryptionPassword, "encryption-password", "", "Password for AES-128-CBC encryption (must match type configuration in app)")
}

func runSend(cmd *cobra.Command, args []string) error {
	// Get token and ID from flags, env vars, or config
	// Note: token and ID are mutually exclusive - only one should be provided
	token := getTokenOptional(cmd)
	id := getIDOptional(cmd)

	// Validate that we have either token or ID, but not both
	if token == "" && id == "" {
		return fmt.Errorf("either token or id is required (use --token/WIREPUSHER_TOKEN or --id/WIREPUSHER_ID)")
	}
	if token != "" && id != "" {
		return fmt.Errorf("token and id are mutually exclusive - use one or the other, not both")
	}

	// Parse title and message
	title, message, err := parseTitleAndMessage(cmd, args)
	if err != nil {
		return err
	}

	// Create client and send notification
	c := client.New()

	opts := &client.SendOptions{
		Title:              title,
		Message:            message,
		Token:              token,
		ID:                 id,
		Type:               sendType,
		Tags:               sendTags,
		ImageURL:           sendImageURL,
		ActionURL:          sendActionURL,
		EncryptionPassword: sendEncryptionPassword,
	}

	if err := c.Send(opts); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	// Success message
	fmt.Println("✓ Notification sent successfully")
	return nil
}

// getTokenOptional retrieves the token from flags, env vars, or config (in that order)
// Returns empty string if not found (caller should validate)
func getTokenOptional(cmd *cobra.Command) string {
	// Try flag first
	token, _ := cmd.Flags().GetString("token")
	if token != "" {
		return token
	}

	// Try environment variable
	token = os.Getenv("WIREPUSHER_TOKEN")
	if token != "" {
		return token
	}

	// Try config file
	token = viper.GetString("token")
	return token
}

// getIDOptional retrieves the user ID from flags, env vars, or config (in that order)
// Returns empty string if not found (caller should validate)
func getIDOptional(cmd *cobra.Command) string {
	// Try flag first
	id, _ := cmd.Flags().GetString("id")
	if id != "" {
		return id
	}

	// Try environment variable
	id = os.Getenv("WIREPUSHER_ID")
	if id != "" {
		return id
	}

	// Try config file
	id = viper.GetString("id")
	return id
}

// parseTitleAndMessage extracts title and message from args or stdin
func parseTitleAndMessage(cmd *cobra.Command, args []string) (string, string, error) {
	var title, message string

	if sendStdin {
		// Read message from stdin
		if len(args) < 1 {
			return "", "", fmt.Errorf("title is required when using --stdin")
		}
		title = args[0]

		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return "", "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		message = strings.Join(lines, "\n")

		if message == "" {
			return "", "", fmt.Errorf("message cannot be empty when using --stdin")
		}
	} else {
		// Get from positional arguments
		if len(args) < 2 {
			return "", "", fmt.Errorf("title and message are required (or use --stdin for message)")
		}
		title = args[0]
		message = args[1]
	}

	return title, message, nil
}
