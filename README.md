# WirePusher CLI

Official command-line tool for [WirePusher](https://wirepusher.dev) push notifications.

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

**1. Get your token:** Open app → Settings → Help → copy token

**2. Configure once** (stores in `~/.wirepusher/config.yaml`):
```bash
wirepusher config set token YOUR_TOKEN
```

**3. Send a notification**:
```bash
wirepusher send "Deploy Complete" "Version 1.2.3 deployed to production"
```

## Usage

### Basic Send

```bash
wirepusher send "Deploy Complete" "Version 1.2.3 deployed to production"
```

### With Type and Tags

```bash
wirepusher send "Deploy Complete" "Version 1.2.3 deployed to production" \
  --type deployment \
  --tag production \
  --tag backend
```

### With Image and Action URL

```bash
wirepusher send "Deploy Complete" "Version 1.2.3 deployed to production" \
  --type deployment \
  --image https://cdn.example.com/success.png \
  --action https://dash.example.com/deploy/123
```

### Read from Stdin

```bash
# Pipe command output
echo "Build successful" | wirepusher send "Build Status" --stdin

# Monitor logs
tail -f /var/log/app.log | grep ERROR | wirepusher send "Error Detected" --stdin
```

### Authentication Methods

The CLI supports three ways to provide your token (priority order):

```bash
# 1. Command-line flag
wirepusher send "Deploy Complete" "Version 1.2.3 deployed" --token wpu_abc123xyz

# 2. Environment variable
export WIREPUSHER_TOKEN="wpu_abc123xyz"
wirepusher send "Deploy Complete" "Version 1.2.3 deployed"

# 3. Config file (set once, use everywhere)
wirepusher config set token wpu_abc123xyz
wirepusher send "Deploy Complete" "Version 1.2.3 deployed"
```

Config is stored in `~/.wirepusher/config.yaml`.

## Encryption

The CLI supports AES-128-CBC encryption to secure notification messages. Only the `message` field is encrypted—`title`, `type`, `tags`, `image`, and `action` remain unencrypted for filtering and display.

### Setup

**In the WirePusher app:**
1. Create or edit a notification type
2. Enable encryption and set a password
3. Save the type configuration

**In the CLI:**
```bash
wirepusher send "Security Alert" "Unauthorized access detected" \
  --type security \
  --encryption-password "your-secure-password"
```

### Important Notes

- **Password matching required**: Encryption password MUST match the type configuration in your app
- **Encryption scope**: Only message body is encrypted
- **Security**: Passwords used for client-side encryption only, never transmitted
- **Interoperability**: Compatible with all WirePusher SDKs

### Example

```bash
wirepusher send "Database Alert" "Connection failed: timeout" \
  --type alert \
  --tag production \
  --tag database \
  --encryption-password "db-alert-password"
```

## CI/CD Integration

### GitLab CI

```yaml
deploy:
  script:
    - echo "Deploying..."
    - wirepusher send "Deploy Complete" "Version ${CI_COMMIT_TAG} deployed to production"
  variables:
    WIREPUSHER_TOKEN: $WIREPUSHER_TOKEN
```

### GitHub Actions

```yaml
- name: Notify on success
  run: |
    wirepusher send "Deploy Complete" "Commit ${{ github.sha }} deployed to production" \
      --type deployment
  env:
    WIREPUSHER_TOKEN: ${{ secrets.WIREPUSHER_TOKEN }}
```

## Commands

### send

Send a push notification.

```bash
wirepusher send [title] [message] [flags]
```

**Flags:**
- `--token string` - WirePusher token (or use env var/config)
- `--type string` - Notification type (deployment, alert, info, etc.)
- `--tag strings` - Tags for categorization (repeatable)
- `--image string` - Image URL to display
- `--action string` - Action URL to open on tap
- `--encryption-password string` - Password for AES-128-CBC encryption
- `--stdin` - Read message from stdin

**Examples:**
```bash
wirepusher send "Deploy Complete" "Version 1.2.3 deployed"
wirepusher send "Alert" "CPU usage high" --type alert --tag production
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
- `token` - Your WirePusher token

### version

Print version information.

```bash
wirepusher version
```

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

## Links

- **Documentation**: https://wirepusher.dev/help
- **Repository**: https://gitlab.com/wirepusher/cli
- **Issues**: https://gitlab.com/wirepusher/cli/-/issues
- **Releases**: https://gitlab.com/wirepusher/cli/-/releases

## License

MIT License - see [LICENSE](LICENSE) for details.
