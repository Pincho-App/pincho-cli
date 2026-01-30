# Advanced Usage

This document covers advanced features of the Pincho CLI.

## Table of Contents

- [Commands Reference](#commands-reference)
- [Validation Philosophy](#validation-philosophy)
- [Retry Logic](#retry-logic)
- [Rate Limits](#rate-limits)
- [Encryption](#encryption)
- [Configuration](#configuration)
- [Exit Codes](#exit-codes)
- [Verbose Mode](#verbose-mode)
- [Advanced Examples](#advanced-examples)
- [Building from Source](#building-from-source)
- [Testing](#testing)

## Commands Reference

### send

```bash
pincho send <title> [message] [flags]
```

**Flags:**
- `--type string` - Notification type (deploy, alert, etc.)
- `--tag strings` - Tags (repeatable, max 10)
- `--image-url string` - Image URL
- `--action-url string` - Action URL (opens on tap)
- `--encryption-password string` - Encrypt message with AES-128-CBC
- `--stdin` - Read message from stdin
- `--json` - JSON output format
- `--timeout int` - Request timeout in seconds (default: 30)
- `--max-retries int` - Max retries (default: 3)
- `--verbose` - Debug output
- `--token string` - API token (overrides config)

**Examples:**
```bash
pincho send "Deploy" "v1.2.3 deployed"
pincho send "Alert" "CPU high" --type alert --tag production
pincho send "Secure" "Encrypted" --encryption-password "secret"
echo "Output" | pincho send "Logs" --stdin
pincho send "Deploy" --json  # Machine-readable output
```

### notifai

AI-powered notifications using Gemini:

```bash
pincho notifai <text> [flags]
```

**Flags:**
- `--type string` - Override AI-generated type
- `--stdin` - Read text from stdin
- `--json` - JSON output format
- `--timeout int` - Request timeout (default: 30)
- `--max-retries int` - Max retries (default: 3)
- `--verbose` - Debug output
- `--token string` - API token

**Examples:**
```bash
pincho notifai "deployment finished, v2.1.3 is live on prod"
pincho notifai "cpu at 95% on web-3" --type alert
cat log.txt | pincho notifai --stdin
```

### config

Manage persistent configuration:

```bash
pincho config set <key> <value>
pincho config get <key>
pincho config list
```

**Supported keys:**

| Key | Description | Example |
|-----|-------------|---------|
| `token` | API token | `pincho config set token abc123` |
| `api_url` | Custom API endpoint | `pincho config set api_url https://custom.com/send` |
| `timeout` | Request timeout (seconds) | `pincho config set timeout 60` |
| `max_retries` | Max retry attempts | `pincho config set max_retries 5` |
| `default_type` | Default notification type | `pincho config set default_type deploy` |

**Note:** `default_tags` must be set directly in `~/.pincho/config.yaml` (YAML array):

```yaml
default_tags:
  - production
  - automated
```

### version

```bash
pincho version
```

Shows version, commit hash, and build date.

## Validation Philosophy

The CLI validates **more strictly** than library clients to provide immediate user feedback.

### What CLI Validates

- **Required parameters**: Title and token must be present
- **Tag limits**: Max 10 tags, 50 characters each
- **Tag format**: Alphanumeric, hyphens, underscores only
- **Immediate feedback**: Errors shown before API call

### What CLI Normalizes

- **Tags**: Lowercase conversion, whitespace trimming, deduplication
- **Automatic**: No extra flags needed

### Why Stricter Validation?

CLI serves interactive users who benefit from fast-fail:
- Saves API quota by catching errors locally
- Clear error messages with actionable guidance
- No surprise failures after network round-trip
- Matches documented API limits upfront

**Note:** Pincho libraries (Go, Python, JS) perform minimal validation and let the API be the source of truth. This is intentional - CLI serves interactive users, libraries serve programmatic integrations.

## Retry Logic

The CLI automatically retries failed requests with exponential backoff.

### Retryable Errors

- **Network errors**: Connection refused, timeout, EOF
- **Server errors**: HTTP 5xx status codes
- **Rate limits**: HTTP 429 with smart backoff

### Non-Retryable Errors

- **Validation errors**: HTTP 400 (invalid parameters)
- **Authentication errors**: HTTP 401, 403 (invalid token)
- **Not found**: HTTP 404

### Backoff Strategy

```
Attempt 1: Initial backoff (1 second)
Attempt 2: 2 seconds
Attempt 3: 4 seconds
Attempt 4: 8 seconds
...
Maximum: 30 seconds (capped)
```

For rate limit errors (429), the CLI uses:
- `Retry-After` header value if provided by the API
- Otherwise, starts with 5 seconds and doubles

### Configuration

```bash
# Set max retries (default: 3)
pincho send "Title" "Message" --max-retries 5

# Disable retries
pincho send "Title" "Message" --max-retries 0

# Configure via environment variable
PINCHO_MAX_RETRIES=5 pincho send "Title" "Message"

# Configure permanently
pincho config set max_retries 5
```

## Rate Limits

The Pincho API enforces rate limits:

- **Send endpoint**: 30 requests per hour per token
- **NotifAI endpoint**: 50 requests per hour per token

### Rate Limit Headers

The API returns rate limit information in response headers:

```
RateLimit-Limit: 30
RateLimit-Remaining: 25
RateLimit-Reset: 2024-01-15T10:00:00Z
```

When a rate limit is hit (HTTP 429), the API provides:

```
Retry-After: 120  (seconds until reset)
```

The CLI respects the `Retry-After` header and waits appropriately.

### Monitoring Rate Limits

Use verbose mode to see rate limit information:

```bash
pincho send "Title" "Message" --verbose
# Output includes:
# Rate Limit: 25/30 remaining (resets at 2024-01-15T10:00:00Z)
```

### Best Practices

1. **Batch notifications** when possible
2. **Cache tokens** to avoid unnecessary auth calls
3. **Monitor remaining requests** in CI/CD pipelines
4. **Implement application-level queuing** for high-volume scenarios

## Encryption

The CLI supports AES-128-CBC encryption for message content.

### How It Works

1. Password is hashed using SHA-1 to derive a 16-byte key
2. Random 16-byte IV is generated for each message
3. Message is encrypted using AES-128-CBC with PKCS7 padding
4. Result is encoded in custom Base64 (URL-safe: `-`, `_`, `.`)

### Usage

```bash
# Encrypt message with password
pincho send "Secure Alert" "Sensitive data" \
  --encryption-password "your-secret-password" \
  --type secure
```

### What Gets Encrypted

- **Encrypted**: Message body only
- **Not encrypted**: Title, type, tags, URLs

This allows the mobile app to filter and categorize notifications while keeping the actual message content private.

### Mobile App Configuration

The encryption password must match the type configuration in your Pincho mobile app:

1. Open Pincho app
2. Go to Settings
3. Select the notification type (e.g., "secure")
4. Enter the same encryption password

### Security Considerations

- Use strong, unique passwords for each notification type
- Store passwords securely (not in scripts or version control)
- The SHA-1 key derivation is for compatibility with mobile app
- Messages are encrypted end-to-end (API never sees plaintext)

## Configuration

### Priority Order

Configuration values are resolved in this order (highest priority first):

1. **Command-line flags** (`--token`, `--timeout`)
2. **Environment variables** (`PINCHO_TOKEN`, `PINCHO_TIMEOUT`)
3. **Config file** (`~/.pincho/config.yaml`)
4. **Defaults** (built into the CLI)

### Config File Location

```bash
# Default location
~/.pincho/config.yaml

# Check path
pincho config list
# Configuration from /Users/you/.pincho/config.yaml:
```

### Available Settings

```bash
# API token (required)
pincho config set token wpt_abc123xyz

# Request timeout in seconds (default: 30)
pincho config set timeout 60

# Max retry attempts (default: 3)
pincho config set max_retries 5

# Custom API URL (for testing)
pincho config set api_url https://staging.api.pincho.dev/send
```

### Environment Variables

```bash
PINCHO_TOKEN       # API token
PINCHO_TIMEOUT     # Request timeout (seconds)
PINCHO_MAX_RETRIES # Max retry attempts
PINCHO_API_URL     # Custom API endpoint
```

### Config File Format

```yaml
# ~/.pincho/config.yaml
token: wpt_abc123xyz
timeout: "60"
max_retries: "5"
api_url: https://api.pincho.dev/send
default_type: alert
default_tags:
  - production
  - automated
```

### Default Type and Tags

```yaml
# Auto-apply notification type
default_type: alert

# Auto-merge tags with explicit tags
default_tags:
  - automated
  - ci-cd
```

When you run:
```bash
pincho send "Build" "Complete" --tag production
```

The notification will have tags: `["production", "automated", "ci-cd"]`

## Exit Codes

The CLI uses specific exit codes for CI/CD integration:

| Code | Meaning | Example |
|------|---------|---------|
| 0 | Success | Notification sent successfully |
| 1 | Usage error | Invalid arguments, missing token, auth failure |
| 2 | API error | Rate limit exceeded, validation error |
| 3 | System error | Network timeout, connection refused |

### Usage in Scripts

```bash
#!/bin/bash
pincho send "Deploy" "Version $VERSION deployed"
exit_code=$?

case $exit_code in
  0)
    echo "Notification sent"
    ;;
  1)
    echo "Configuration error - check token" >&2
    exit 1
    ;;
  2)
    echo "API error - check rate limits" >&2
    exit 2
    ;;
  3)
    echo "Network error - retry later" >&2
    exit 3
    ;;
esac
```

### CI/CD Pipeline Example

```yaml
# GitLab CI
notify_success:
  script:
    - pincho send "Deploy Success" "Version $CI_COMMIT_TAG"
  allow_failure: false  # Pipeline fails if notification fails

notify_failure:
  script:
    - pincho send "Deploy Failed" "Pipeline $CI_PIPELINE_ID failed" || true
  when: on_failure
  allow_failure: true  # Don't double-fail
```

## Verbose Mode

Enable detailed logging to debug issues:

```bash
pincho send "Title" "Message" --verbose
```

### Output Includes

- Token (first 8 characters, masked)
- API URL being used
- Timeout configuration
- Max retries setting
- Request progress
- Rate limit information
- Response details

### Example Output

```
[VERBOSE] Verbose logging enabled
[VERBOSE] Using token: wpt_abc1...
[VERBOSE] Using API URL: https://api.pincho.dev/send
[VERBOSE] Using timeout: 30s
[VERBOSE] Using max retries: 3
[VERBOSE] Title: Deploy Complete
[VERBOSE] Message: Version 1.2.3 deployed
[VERBOSE] Sending notification to API...
[VERBOSE] Notification sent successfully

✓ Notification sent successfully

Notification ID: notif_123456
Title: Deploy Complete
Message: Version 1.2.3 deployed
Expires: 2024-01-16T10:00:00Z

Rate Limit: 25/30 remaining (resets at 2024-01-15T11:00:00Z)
```

### Redirecting Output

All verbose logging goes to stderr, keeping stdout clean for JSON output:

```bash
# Capture JSON, ignore verbose
pincho send "Title" "Message" --json --verbose 2>/dev/null

# Log verbose to file
pincho send "Title" "Message" --verbose 2>debug.log

# See both in terminal
pincho send "Title" "Message" --verbose 2>&1 | tee log.txt
```

## Advanced Examples

### Monitoring Script

```bash
#!/bin/bash
# Monitor disk usage and alert if high

THRESHOLD=90
USAGE=$(df -h / | awk 'NR==2 {print $5}' | tr -d '%')

if [ "$USAGE" -gt "$THRESHOLD" ]; then
  pincho send \
    "Disk Alert" \
    "Disk usage at ${USAGE}% on $(hostname)" \
    --type alert \
    --tag monitoring \
    --tag disk-space
fi
```

### Build Notification

```bash
#!/bin/bash
# Notify on build completion

if make build; then
  pincho send \
    "Build Success" \
    "$(git rev-parse --short HEAD) built successfully" \
    --tag build \
    --tag success
else
  pincho send \
    "Build Failed" \
    "Build failed for $(git rev-parse --short HEAD)" \
    --type alert \
    --tag build \
    --tag failure
  exit 1
fi
```

### Encrypted Status Report

```bash
#!/bin/bash
# Send encrypted system status

STATUS=$(top -l 1 | head -10)
pincho send \
  "System Status" \
  "$STATUS" \
  --encryption-password "$ENCRYPTION_KEY" \
  --type secure
```

## Building from Source

```bash
git clone https://gitlab.com/pincho/pincho-cli.git
cd pincho-cli
go build -o pincho
```

With version info:
```bash
go build -ldflags="-X 'gitlab.com/pincho/cli/cmd.version=1.0.0'" -o pincho
```

## Testing

```bash
go test ./...              # Run all tests
go test ./... -cover       # With coverage
go test ./... -v -race     # Verbose with race detection
```

Current coverage targets:
- Overall: 30% minimum
- pkg/ layer: 55% minimum (validation: 100%, crypto: 88%, config: 71%)
