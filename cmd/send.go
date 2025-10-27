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
Authentication requires a token and user ID from flags, environment variables,
or the config file.

Examples:
  # Simple notification
  wirepusher send "Build Complete" "Deploy finished successfully"

  # With notification type
  wirepusher send "Alert" "CPU usage high" --type alert

  # With tags
  wirepusher send "Deploy" "v1.2.3 deployed" --tag production --tag release

  # With image and action URL
  wirepusher send "Success" "All tests passed" \
    --image https://example.com/success.png \
    --action https://example.com/build/123

  # Read message from stdin
  echo "Server error detected" | wirepusher send "Error" --stdin

  # Override config with flags
  wirepusher send "Test" "Message" --token abc123 --id user123
`,
	RunE: runSend,
}

var (
	sendType      string
	sendTags      []string
	sendImageURL  string
	sendActionURL string
	sendStdin     bool
)

func init() {
	rootCmd.AddCommand(sendCmd)

	// Flags specific to send command
	sendCmd.Flags().StringVar(&sendType, "type", "", "Notification type (e.g., alert, info, success)")
	sendCmd.Flags().StringSliceVar(&sendTags, "tag", []string{}, "Tags for categorization (can be used multiple times)")
	sendCmd.Flags().StringVar(&sendImageURL, "image", "", "Image URL to display with notification")
	sendCmd.Flags().StringVar(&sendActionURL, "action", "", "Action URL to open when notification is tapped")
	sendCmd.Flags().BoolVar(&sendStdin, "stdin", false, "Read message from stdin")
}

func runSend(cmd *cobra.Command, args []string) error {
	// Get token and ID from flags, env vars, or config
	token, err := getToken(cmd)
	if err != nil {
		return err
	}

	id, err := getID(cmd)
	if err != nil {
		return err
	}

	// Parse title and message
	title, message, err := parseTitleAndMessage(cmd, args)
	if err != nil {
		return err
	}

	// Create client and send notification
	c := client.New()

	opts := &client.SendOptions{
		Title:     title,
		Message:   message,
		Token:     token,
		ID:        id,
		Type:      sendType,
		Tags:      sendTags,
		ImageURL:  sendImageURL,
		ActionURL: sendActionURL,
	}

	if err := c.Send(opts); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	// Success message
	fmt.Println("✓ Notification sent successfully")
	return nil
}

// getToken retrieves the token from flags, env vars, or config (in that order)
func getToken(cmd *cobra.Command) (string, error) {
	// Try flag first
	token, _ := cmd.Flags().GetString("token")
	if token != "" {
		return token, nil
	}

	// Try environment variable
	token = os.Getenv("WIREPUSHER_TOKEN")
	if token != "" {
		return token, nil
	}

	// Try config file
	token = viper.GetString("token")
	if token != "" {
		return token, nil
	}

	return "", fmt.Errorf("token is required (use --token, WIREPUSHER_TOKEN env var, or 'wirepusher config set token <value>')")
}

// getID retrieves the user ID from flags, env vars, or config (in that order)
func getID(cmd *cobra.Command) (string, error) {
	// Try flag first
	id, _ := cmd.Flags().GetString("id")
	if id != "" {
		return id, nil
	}

	// Try environment variable
	id = os.Getenv("WIREPUSHER_ID")
	if id != "" {
		return id, nil
	}

	// Try config file
	id = viper.GetString("id")
	if id != "" {
		return id, nil
	}

	return "", fmt.Errorf("id is required (use --id, WIREPUSHER_ID env var, or 'wirepusher config set id <value>')")
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
