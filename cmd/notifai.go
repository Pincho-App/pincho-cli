package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/wirepusher/cli/pkg/client"
	clierrors "gitlab.com/wirepusher/cli/pkg/errors"
	"gitlab.com/wirepusher/cli/pkg/logging"
)

// notifaiCmd represents the notifai command
var notifaiCmd = &cobra.Command{
	Use:   "notifai <text>",
	Short: "Send an AI-generated notification from free-form text",
	Long: `Send a notification using AI to convert free-form text into a structured notification.

The AI (Gemini) analyzes your text and automatically generates:
- A concise title (max 75 chars)
- A descriptive message (max 200 chars)
- Relevant tags (max 3 tags)
- An action URL if applicable

This is perfect for:
- Quick status updates: "deployment finished, v2.1.3 is live"
- Error reports: "database backup failed, disk full on prod-db-1"
- Event notifications: "new user signup from premium plan"

Text Requirements:
- Minimum: 5 characters
- Maximum: 2500 characters
- Can be provided as argument or via stdin

Rate Limit: 50 requests per hour per token

Examples:
  # Basic usage
  wirepusher notifai "deployment finished successfully, v2.1.3 is live on prod"

  # With notification type
  wirepusher notifai "cpu usage at 95% on web-server-3" --type alert

  # Read from stdin
  echo "backup completed, 2.3GB uploaded to S3" | wirepusher notifai --stdin

  # Read from file
  cat deploy-log.txt | wirepusher notifai --stdin --type deploy

  # JSON output
  wirepusher notifai "server restarted after update" --json
`,
	RunE: runNotifAI,
}

var (
	notifaiType  string
	notifaiStdin bool
	notifaiJSON  bool
)

func init() {
	rootCmd.AddCommand(notifaiCmd)

	// Flags specific to notifai command
	notifaiCmd.Flags().StringVar(&notifaiType, "type", "", "Notification type (optional)")
	notifaiCmd.Flags().BoolVar(&notifaiStdin, "stdin", false, "Read text from stdin")
	notifaiCmd.Flags().BoolVar(&notifaiJSON, "json", false, "Output response as JSON")
}

func runNotifAI(cmd *cobra.Command, args []string) error {
	// Get token from flags, env vars, or config
	token := getTokenOptional(cmd)

	if token == "" {
		return clierrors.NewUsageError(
			"API token is required",
			fmt.Errorf("no token provided via --token flag, WIREPUSHER_TOKEN environment variable, or config file"),
		)
	}

	logging.Verbose("Using token: %s...", token[:min(8, len(token))])

	// Parse text input
	text, err := parseText(args)
	if err != nil {
		return clierrors.NewUsageError("Invalid arguments", err)
	}

	// Validate text length
	if len(text) < 5 {
		return clierrors.NewUsageError("Text too short", fmt.Errorf("text must be at least 5 characters long (got %d)", len(text)))
	}
	if len(text) > 2500 {
		return clierrors.NewUsageError("Text too long", fmt.Errorf("text must be at most 2500 characters long (got %d)", len(text)))
	}

	logging.Verbose("Text length: %d characters", len(text))

	// Create client and send notifai request
	c := client.New()

	// Set API URL if configured (via env, config file, or default)
	if apiURL := getAPIURL(cmd); apiURL != "" {
		c.APIURL = apiURL
		logging.Verbose("Using API URL: %s", apiURL)
	}

	// Set timeout if configured (via flag, env var, or default)
	timeout := getTimeout(cmd)
	c.SetTimeout(timeout)
	logging.Verbose("Using timeout: %v", timeout)

	// Set retry configuration
	maxRetries := getMaxRetries(cmd)
	c.SetRetryConfig(maxRetries, client.DefaultInitialBackoff)
	logging.Verbose("Using max retries: %d", maxRetries)

	// Merge type with default from config
	finalType := mergeTypeWithDefault(notifaiType)
	if finalType != "" && finalType != notifaiType {
		logging.Verbose("Using default type from config: %s", finalType)
	}

	opts := &client.NotifAIOptions{
		Text:  text,
		Token: token,
		Type:  finalType,
	}

	logging.Verbose("Sending AI request to API...")
	result, err := c.NotifAI(opts)
	if err != nil {
		return categorizeNotifAIError(err)
	}

	logging.Verbose("AI-generated notification sent successfully")

	// Output response
	if notifaiJSON {
		// JSON output
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON response: %w", err)
		}
		fmt.Println(string(jsonBytes))
	} else {
		// Human-readable output
		displayNotifAIResult(result)
	}

	return nil
}

// parseText extracts text from args or stdin
func parseText(args []string) (string, error) {
	if notifaiStdin {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		text := strings.Join(lines, "\n")

		if text == "" {
			return "", fmt.Errorf("text cannot be empty when using --stdin")
		}

		return text, nil
	}

	// Get from positional argument
	if len(args) < 1 {
		return "", fmt.Errorf("text is required (or use --stdin)")
	}

	return args[0], nil
}

// categorizeNotifAIError converts a generic error into a CLI error with appropriate exit code
func categorizeNotifAIError(err error) error {
	errStr := err.Error()

	// Check for specific error patterns and categorize
	if strings.Contains(errStr, "validation error") || strings.Contains(errStr, "parameter_too_short") || strings.Contains(errStr, "parameter_too_long") {
		return clierrors.NewUsageError("Invalid input", err)
	}

	if strings.Contains(errStr, "authentication error") || strings.Contains(errStr, "invalid_api_token") {
		return clierrors.NewUsageError("Authentication failed", fmt.Errorf("%v\n\nGet your token: Open WirePusher app → Settings → Help → Copy token\nOr set it: wirepusher config set token YOUR_TOKEN", err))
	}

	if strings.Contains(errStr, "rate limit exceeded") {
		return clierrors.NewAPIError("Rate limit exceeded", fmt.Errorf("%v\n\nThe notifai endpoint allows 50 requests per hour. Please wait before trying again.", err))
	}

	if strings.Contains(errStr, "API error") {
		return clierrors.NewAPIError("API request failed", err)
	}

	if strings.Contains(errStr, "request failed") || strings.Contains(errStr, "connection") {
		return clierrors.NewSystemError("Network error", fmt.Errorf("%v\n\nPlease check your internet connection and try again.", err))
	}

	// Default to system error for unknown errors
	return clierrors.NewSystemError("Unexpected error", err)
}

// displayNotifAIResult formats and displays the notifai result in human-readable format
func displayNotifAIResult(result *client.NotifAIResult) {
	fmt.Println("✓ AI-generated notification sent successfully")
	fmt.Println()

	// Display AI-generated summary
	if result.Response.Summary != nil {
		fmt.Println("AI Summary:")
		fmt.Printf("  Title: %s\n", result.Response.Summary.Title)
		if result.Response.Summary.Message != "" {
			fmt.Printf("  Message: %s\n", result.Response.Summary.Message)
		}
		if len(result.Response.Summary.Tags) > 0 {
			fmt.Printf("  Tags: %s\n", strings.Join(result.Response.Summary.Tags, ", "))
		}
		if result.Response.Summary.ActionURL != "" {
			fmt.Printf("  Action URL: %s\n", result.Response.Summary.ActionURL)
		}
		fmt.Println()
	}

	// Display team or personal token result
	if result.Response.TeamID != "" {
		// Team token result
		fmt.Printf("Team: %s\n", result.Response.TeamID)
		fmt.Printf("Members notified: %d\n", result.Response.MemberCount)
	} else if result.Response.ReceivedNotification != nil {
		// Personal token result
		notif := result.Response.ReceivedNotification
		fmt.Printf("Notification ID: %s\n", notif.NotificationID)
		if notif.ExpiresAt != "" {
			fmt.Printf("Expires: %s\n", notif.ExpiresAt)
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
