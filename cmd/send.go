package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/Pincho-App/pincho-cli/pkg/client"
	clierrors "github.com/Pincho-App/pincho-cli/pkg/errors"
	"github.com/Pincho-App/pincho-cli/pkg/logging"
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send <title> [message]",
	Short: "Send a push notification",
	Long: `Send a push notification via Pincho.

The title is required, and message is optional.
Authentication requires a token from flags,
environment variables, or the config file.

Examples:
  # Simple notification (using token)
  pincho send "Build Complete" "Deploy finished successfully"

  # Title-only notification (message is optional)
  pincho send "Deploy Complete"

  # With notification type
  pincho send "Alert" "CPU usage high" --type alert

  # With tags (normalized to lowercase, max 10 tags, 50 chars each)
  pincho send "Deploy" "v1.2.3 deployed" --tag production --tag release

  # With image and action URL
  pincho send "Success" "All tests passed" \
    --image-url https://example.com/success.png \
    --action-url https://example.com/build/123

  # With encryption (message encrypted with AES-128-CBC)
  pincho send "Secure Alert" "Sensitive data here" \
    --encryption-password "secret123" \
    --type secure

  # Read message from stdin with encryption
  echo "Confidential report" | pincho send "Report" --stdin \
    --encryption-password "secret123"


  # Override config with flags
  pincho send "Test" "Message" --token abc123
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
	sendJSON               bool
)

func init() {
	rootCmd.AddCommand(sendCmd)

	// Flags specific to send command
	sendCmd.Flags().StringVar(&sendType, "type", "", "Notification type (e.g., alert, info, success)")
	sendCmd.Flags().StringSliceVar(&sendTags, "tag", []string{}, "Tags for categorization (can be used multiple times)")
	sendCmd.Flags().StringVar(&sendImageURL, "image-url", "", "Image URL to display with notification")
	sendCmd.Flags().StringVar(&sendActionURL, "action-url", "", "Action URL to open when notification is tapped")
	sendCmd.Flags().BoolVar(&sendStdin, "stdin", false, "Read message from stdin")
	sendCmd.Flags().StringVar(&sendEncryptionPassword, "encryption-password", "", "Password for AES-128-CBC encryption (must match type configuration in app)")
	sendCmd.Flags().BoolVar(&sendJSON, "json", false, "Output response as JSON")
}

func runSend(cmd *cobra.Command, args []string) error {
	// Get token and ID from flags, env vars, or config
	token := getTokenOptional(cmd)

	if token == "" {
		return clierrors.NewUsageError(
			"API token is required",
			fmt.Errorf("no token provided via --token flag, PINCHO_TOKEN environment variable, or config file"),
		)
	}

	logging.Debug("Token configured", "token_prefix", token[:min(8, len(token))])

	// Parse title and message
	title, message, err := parseTitleAndMessage(cmd, args)
	if err != nil {
		return clierrors.NewUsageError("Invalid arguments", err)
	}

	logging.Debug("Notification content parsed", "title", title, "message_length", len(message))

	// Create client and send notification
	c := client.New()

	// Set token for authentication
	c.SetToken(token)

	// Set API URL if configured (via env, config file, or default)
	if apiURL := getAPIURL(cmd); apiURL != "" {
		c.APIURL = apiURL
	}
	logging.Debug("API client configured", "api_url", c.APIURL)

	// Set timeout if configured (via flag, env var, or default)
	timeout := getTimeout(cmd)
	c.SetTimeout(timeout)

	// Set retry configuration
	maxRetries := getMaxRetries(cmd)
	c.SetRetryConfig(maxRetries, client.DefaultInitialBackoff)

	logging.Debug("Client settings", "timeout", timeout, "max_retries", maxRetries)

	// Merge type with default from config
	finalType := mergeTypeWithDefault(sendType)

	// Merge tags with defaults from config
	finalTags := mergeTagsWithDefaults(sendTags)

	logging.Debug("Notification options", "type", finalType, "tags", finalTags, "has_encryption", sendEncryptionPassword != "")

	opts := &client.SendOptions{
		Title:              title,
		Message:            message,
		Type:               finalType,
		Tags:               finalTags,
		ImageURL:           sendImageURL,
		ActionURL:          sendActionURL,
		EncryptionPassword: sendEncryptionPassword,
	}

	logging.Debug("Sending notification to API")
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := c.Send(ctx, opts)
	if err != nil {
		return categorizeError(err)
	}

	logging.Debug("Notification sent successfully")

	// Output response
	if sendJSON {
		// JSON output
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON response: %w", err)
		}
		fmt.Println(string(jsonBytes))
	} else {
		// Human-readable output
		displaySendResult(result)
	}

	return nil
}

// parseTitleAndMessage extracts title and message from args or stdin
// Message is optional - can be empty string
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

		// Message can be empty - backend allows null message
	} else {
		// Get from positional arguments
		if len(args) < 1 {
			return "", "", fmt.Errorf("title is required")
		}
		title = args[0]

		// Message is optional (can be omitted)
		if len(args) >= 2 {
			message = args[1]
		}
	}

	return title, message, nil
}

// categorizeError converts a generic error into a CLI error with appropriate exit code
func categorizeError(err error) error {
	// Check for typed API errors
	switch e := err.(type) {
	case *clierrors.ValidationError:
		return clierrors.NewUsageError("Invalid input", e)
	case *clierrors.AuthenticationError:
		return clierrors.NewUsageError("Authentication failed", fmt.Errorf("%v\n\nGet your token: Open Pincho app → Settings → Help → Copy token\nOr set it: pincho config set token YOUR_TOKEN", e))
	case *clierrors.RateLimitError:
		return clierrors.NewAPIError("Rate limit exceeded", fmt.Errorf("%v\n\nThe send endpoint allows 30 requests per hour. Please wait before trying again.", e))
	case *clierrors.ServerError:
		return clierrors.NewAPIError("Server error", e)
	case *clierrors.NetworkError:
		return clierrors.NewSystemError("Network error", fmt.Errorf("%v\n\nPlease check your internet connection and try again.", e))
	default:
		// Unknown error type - treat as system error
		return clierrors.NewSystemError("Unexpected error", err)
	}
}

// displaySendResult formats and displays the send result in human-readable format
func displaySendResult(result *client.SendResult) {
	fmt.Println("✓ Notification sent successfully")
	fmt.Println()

	// Display team or personal token result
	if result.Response.TeamID != "" {
		// Team token result
		fmt.Printf("Team: %s\n", result.Response.TeamID)
		fmt.Printf("Members notified: %d\n", result.Response.MemberCount)
	} else if result.Response.ReceivedNotification != nil {
		// Personal token result
		notif := result.Response.ReceivedNotification
		fmt.Printf("Notification ID: %s\n", notif.NotificationID)
		fmt.Printf("Title: %s\n", notif.Title)
		if notif.Body != "" {
			fmt.Printf("Message: %s\n", notif.Body)
		}
		if notif.Type != "" {
			fmt.Printf("Type: %s\n", notif.Type)
		}
		if len(notif.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(notif.Tags, ", "))
		}
		if notif.ExpiresAt.Seconds > 0 {
			expiresTime := time.Unix(notif.ExpiresAt.Seconds, notif.ExpiresAt.Nanoseconds)
			fmt.Printf("Expires: %s\n", expiresTime.Format(time.RFC3339))
		}
	}

	// Display rate limit info if available
	if result.RateLimit != nil && result.RateLimit.Limit != "" {
		fmt.Println()
		fmt.Printf("Rate Limit: %s/%s remaining", result.RateLimit.Remaining, result.RateLimit.Limit)
		if result.RateLimit.Reset != "" {
			fmt.Printf(" (resets at %s)", result.RateLimit.Reset)
		}
		fmt.Println()
	}
}
