# WirePusher CLI

Official command-line tool for [WirePusher](https://wirepusher.com) push notifications.

Send push notifications from your terminal, CI/CD pipelines, monitoring scripts, and automation workflows with a single command.

## Features

- **Simple**: Send notifications with one command
- **Fast**: Single static binary, no runtime dependencies
- **Cross-platform**: Works on Linux, macOS, and Windows
- **CI/CD Ready**: Perfect for GitLab CI, GitHub Actions, Jenkins
- **Flexible Auth**: Command-line flags, environment variables, or config file
- **Stdin Support**: Pipe command output directly to notifications

## Installation

### Manual Download

Download the binary for your platform from [Releases](https://gitlab.com/wirepusher/cli/-/releases):

```bash
# Linux (amd64)
curl -LO https://gitlab.com/wirepusher/cli/-/releases/latest/downloads/wirepusher-linux-amd64
chmod +x wirepusher-linux-amd64
sudo mv wirepusher-linux-amd64 /usr/local/bin/wirepusher

# macOS (Apple Silicon)
curl -LO https://gitlab.com/wirepusher/cli/-/releases/latest/downloads/wirepusher-darwin-arm64
chmod +x wirepusher-darwin-arm64
sudo mv wirepusher-darwin-arm64 /usr/local/bin/wirepusher
```

### Build from Source

```bash
git clone https://gitlab.com/wirepusher/cli.git
cd cli
go build -o wirepusher main.go
sudo mv wirepusher /usr/local/bin/
```

##  Quick Start

1. **Get your credentials** from [WirePusher](https://wirepusher.com)
   - You need EITHER a token OR a user ID (not both)

2. **Configure once** (stores in `~/.wirepusher/config.yaml`):
   ```bash
   # Option 1: Using token (recommended)
   wirepusher config set token YOUR_TOKEN

   # Option 2: Using user ID
   wirepusher config set id YOUR_USER_ID
   ```

3. **Send a notification**:
   ```bash
   wirepusher send "Hello" "Your first notification!"
   ```

## Usage

### Basic Send

```bash
wirepusher send "Build Complete" "Deployment finished successfully"
```

### With Type and Tags

```bash
wirepusher send "Alert" "CPU usage high" \
  --type alert \
  --tag production \
  --tag monitoring
```

### With Image and Action URL

```bash
wirepusher send "Deploy Complete" "v1.2.3 deployed" \
  --type success \
  --image https://example.com/success.png \
  --action https://example.com/deploy/123
```

### Read from Stdin

```bash
# Pipe command output
echo "Server error detected" | wirepusher send "Error" --stdin

# Monitor logs
tail -f /var/log/app.log | grep ERROR | wirepusher send "Error Detected" --stdin
```

## Authentication

WirePusher supports two authentication methods for sending notifications. **Important:** These are mutually exclusive - use EITHER token OR ID, not both.

### Team Token (Recommended for Teams)

Team tokens (starting with `wpt_`) send notifications to **ALL members** of a team.

**Use cases:**
- Team-wide alerts and announcements
- Shared project notifications
- CI/CD pipelines broadcasting to teams
- Collaborative workflows

**Methods:**

```bash
# 1. Command-line flag
wirepusher send "Team Alert" "Server maintenance in 1 hour" --token wpt_abc123xyz

# 2. Environment variable
export WIREPUSHER_TOKEN="wpt_abc123xyz"
wirepusher send "Team Alert" "Server maintenance in 1 hour"

# 3. Config file (stores in ~/.wirepusher/config.yaml)
wirepusher config set token wpt_abc123xyz
wirepusher send "Team Alert" "Server maintenance in 1 hour"
```

### User ID (Personal Notifications)

User IDs send notifications to a **specific user's devices only**.

**Use cases:**
- Personal notifications
- User-specific alerts
- Individual reminders
- Single-user automation

**Methods:**

```bash
# 1. Command-line flag
wirepusher send "Personal Reminder" "Your task is due tomorrow" --id user_abc123

# 2. Environment variable
export WIREPUSHER_ID="user_abc123"
wirepusher send "Personal Reminder" "Your task is due tomorrow"

# 3. Config file (stores in ~/.wirepusher/config.yaml)
wirepusher config set id user_abc123
wirepusher send "Personal Reminder" "Your task is due tomorrow"
```

### Priority Order

If credentials are set in multiple places, the CLI uses this priority order:
1. Command-line flags (`--token` or `--id`)
2. Environment variables (`WIREPUSHER_TOKEN` or `WIREPUSHER_ID`)
3. Config file (`~/.wirepusher/config.yaml`)

**Important:** If both token and ID are configured, the CLI will return an error.

## Configuration Management

```bash
# Set values
wirepusher config set token wpt_abc123xyz
wirepusher config set id user-id-here

# Get specific value
wirepusher config get token

# List all configuration
wirepusher config list
```

Config is stored in `~/.wirepusher/config.yaml`.

## Encryption

The CLI supports AES-128-CBC encryption to secure notification messages before sending them to the API.

### How It Works

1. Message is encrypted client-side using AES-128-CBC
2. Encryption key is derived from your password using SHA1
3. A random initialization vector (IV) is generated for each message
4. Only the **message** field is encrypted (title, type, tags, imageURL, actionURL remain unencrypted)
5. WirePusher app decrypts the message using the matching password configured in your notification type

### Basic Usage

```bash
wirepusher send "Secure Alert" "This message is encrypted" \
  --encryption-password "your-secure-password" \
  --type secure
```

### Configuration

**In the WirePusher app:**
1. Create or edit a notification type
2. Enable encryption and set the same password
3. Save the type configuration

**In the CLI:**
- Use `--encryption-password` flag when sending notifications
- Password must match the type configuration in your app
- Password is never sent to the API (only used for local encryption)

### Important Notes

- **Password matching required**: The encryption password MUST match the password configured for the notification type in your app
- **Encryption scope**: Only the message body is encrypted; title, type, tags, images, and action URLs remain unencrypted for proper filtering and display
- **Security**: Passwords are used only for client-side encryption and are never transmitted to the API
- **Interoperability**: Encryption is compatible with all WirePusher SDKs (Python, JavaScript, Go, Java, C#, PHP, Rust)

### Example Workflow

```bash
# 1. Send encrypted notification with type "secure"
wirepusher send "Password Reset" "New password: xyz123" \
  --type secure \
  --encryption-password "my-secret-password"

# 2. Send encrypted alert with tags
wirepusher send "Database Alert" "Connection failed: timeout" \
  --type alert \
  --tag production \
  --tag database \
  --encryption-password "db-alert-password"

# 3. Pipe encrypted output from command
echo "Sensitive diagnostic info" | wirepusher send "Diagnostics" --stdin \
  --type diagnostic \
  --encryption-password "diagnostic-password"
```

See [examples/encryption.sh](examples/encryption.sh) for complete examples.

## CI/CD Integration

### GitLab CI

```yaml
deploy:
  script:
    - echo "Deploying..."
    - wirepusher send "Deploy Complete" "v${CI_COMMIT_TAG} deployed"
  variables:
    WIREPUSHER_TOKEN: $WIREPUSHER_TOKEN
    WIREPUSHER_ID: $WIREPUSHER_ID
```

### GitHub Actions

```yaml
- name: Notify on success
  run: |
    wirepusher send "Build Passed" "Commit ${{ github.sha }} succeeded" \
      --type success
  env:
    WIREPUSHER_TOKEN: ${{ secrets.WIREPUSHER_TOKEN }}
    WIREPUSHER_ID: ${{ secrets.WIREPUSHER_ID }}
```

## Commands

### send

Send a push notification.

```bash
wirepusher send [title] [message] [flags]
```

**Flags:**
- `--type string` - Notification type (alert, info, success, etc.)
- `--tag strings` - Tags for categorization (repeatable)
- `--image string` - Image URL to display
- `--action string` - Action URL to open on tap
- `--encryption-password string` - Password for AES-128-CBC encryption
- `--stdin` - Read message from stdin

**Examples:**
```bash
wirepusher send "Title" "Message"
wirepusher send "Error" "Details" --type alert --tag production
wirepusher send "Secure" "Encrypted message" --encryption-password "secret"
echo "Log output" | wirepusher send "Logs" --stdin
```

### config

Manage configuration.

```bash
wirepusher config set <key> <value>
wirepusher config get <key>
wirepusher config list
```

**Supported keys:**
- `token` - Your WirePusher API token
- `id` - Your WirePusher user ID

### version

Print version information.

```bash
wirepusher version
```

## Examples

See [examples/](examples/) directory for:
- [Simple notifications](examples/simple.sh)
- [CI/CD integration](examples/ci-cd.sh)
- [Log monitoring](examples/log-monitoring.sh)
- [Configuration management](examples/config-example.sh)

## Building

```bash
# Build for current platform
go build -o wirepusher main.go

# Build for all platforms
GOOS=linux GOARCH=amd64 go build -o wirepusher-linux-amd64 main.go
GOOS=darwin GOARCH=arm64 go build -o wirepusher-darwin-arm64 main.go
GOOS=windows GOARCH=amd64 go build -o wirepusher-windows-amd64.exe main.go
```

### With Version Info

```bash
go build -ldflags="-X 'gitlab.com/wirepusher/cli/cmd.version=1.0.0' \
  -X 'gitlab.com/wirepusher/cli/cmd.commit=$(git rev-parse --short HEAD)' \
  -X 'gitlab.com/wirepusher/cli/cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
  -o wirepusher main.go
```

## Testing

```bash
# Run all tests
go test ./...

# With coverage
go test ./... -cover

# Verbose
go test ./... -v
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## Security

For security vulnerabilities, email security@wirepusher.com. See [SECURITY.md](SECURITY.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- **Documentation:** https://wirepusher.com/docs
- **Issues:** https://gitlab.com/wirepusher/cli/-/issues
- **Email:** support@wirepusher.com

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.

---

Made with ❤️ by the WirePusher team
