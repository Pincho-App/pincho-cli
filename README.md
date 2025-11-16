# WirePusher CLI

Official command-line tool for [WirePusher](https://wirepusher.dev) push notifications.

Send push notifications from your terminal, scripts, and CI/CD pipelines. One command. No complexity.

## Quick Start

**1. Get your token:** Open app → Settings → Help → Copy token

**2. Send a notification:**

```bash
# Option 1: Pass token directly (no setup)
wirepusher send "Deploy Complete" "v1.2.3 is live" --token YOUR_TOKEN

# Option 2: Environment variable (great for CI/CD)
export WIREPUSHER_TOKEN=YOUR_TOKEN
wirepusher send "Deploy Complete" "v1.2.3 is live"

# Option 3: Save in config (best for frequent use)
wirepusher config set token YOUR_TOKEN
wirepusher send "Deploy Complete" "v1.2.3 is live"
```

Pick whichever fits your workflow.

## Installation

### Quick Install (Recommended)

```bash
curl -sSL https://gitlab.com/wirepusher/wirepusher-cli/-/raw/main/install.sh | sh
```

This automatically detects your OS and architecture, downloads the latest release, and installs to `/usr/local/bin`.

### Manual Install

Download the binary for your platform from [Releases](https://gitlab.com/wirepusher/wirepusher-cli/-/releases):

```bash
# Linux (amd64)
curl -LO https://gitlab.com/wirepusher/wirepusher-cli/-/releases/v1.0.0/downloads/wirepusher_1.0.0_linux_amd64.tar.gz
tar -xzf wirepusher_1.0.0_linux_amd64.tar.gz
sudo mv wirepusher /usr/local/bin/

# macOS (Apple Silicon)
curl -LO https://gitlab.com/wirepusher/wirepusher-cli/-/releases/v1.0.0/downloads/wirepusher_1.0.0_darwin_arm64.tar.gz
tar -xzf wirepusher_1.0.0_darwin_arm64.tar.gz
sudo mv wirepusher /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://gitlab.com/wirepusher/wirepusher-cli/-/releases/v1.0.0/downloads/wirepusher_1.0.0_windows_amd64.zip" -OutFile wirepusher.zip
Expand-Archive wirepusher.zip
Move-Item wirepusher\wirepusher.exe C:\Windows\System32\
```

### Uninstall

```bash
curl -sSL https://gitlab.com/wirepusher/wirepusher-cli/-/raw/main/uninstall.sh | sh
```

## Common Use Cases

### Send notifications

```bash
# Basic
wirepusher send "Build Complete"

# With message
wirepusher send "Build Complete" "All tests passed"

# With type and tags
wirepusher send "Deploy" "v1.2.3 to production" --type deploy --tag production
```

### AI-generated notifications

Let AI structure your notifications automatically:

```bash
wirepusher notifai "deployment finished successfully, v2.1.3 is live on prod"

# AI generates title, message, tags, and action URL
```

### Pipe command output

```bash
echo "Backup complete" | wirepusher send "Backup Status" --stdin

# Monitor logs
tail -f app.log | grep ERROR | wirepusher send "Error Alert" --stdin
```

### CI/CD integration

```yaml
# GitLab CI
deploy:
  script:
    - ./deploy.sh
    - wirepusher send "Deploy Complete" "Version ${CI_COMMIT_TAG}"
  variables:
    WIREPUSHER_TOKEN: $WIREPUSHER_TOKEN
```

Exit codes make automation easy:
- `0` = Success
- `1` = Invalid input or auth error → **fix your command**
- `2` = API error or rate limit → **maybe retry**
- `3` = Network error → **definitely retry**

```bash
# Retry on network errors only
if ! wirepusher send "Deploy" "$message"; then
  [[ $? -eq 3 ]] && wirepusher send "Deploy" "$message"  # Retry network errors
fi
```

### Encrypted messages

```bash
# Setup: In WirePusher app, enable encryption for a type with password "secret123"

# Send encrypted notification
wirepusher send "Security Alert" "Sensitive data" \
  --type secure \
  --encryption-password "secret123"
```

Only the message is encrypted. Title, type, and tags remain visible for filtering.

## Configuration

Config file: `~/.wirepusher/config.yaml`

### Set defaults once, use everywhere

```bash
# Basic
wirepusher config set token YOUR_TOKEN

# Advanced: Set operational defaults
wirepusher config set timeout 60          # 60s timeout
wirepusher config set max_retries 5       # Retry 5 times
wirepusher config set default_type deploy # Auto-type all notifications
```

### Full config example

```yaml
# ~/.wirepusher/config.yaml
token: your-token-here
timeout: 60                    # 60s timeout (slow networks)
max_retries: 5                 # 5 retries (flaky CI)
default_type: deploy           # Auto-type all notifications
default_tags:                  # Auto-tag all notifications
  - production
  - automated
```

Now every notification automatically gets these settings:

```bash
wirepusher send "Deploy" "v1.2.3"
# → Uses: timeout=60s, retries=5, type=deploy, tags=[production, automated]
```

Flags always override config:

```bash
wirepusher send "Test" --type test --tag staging
# → type=test (flag wins), tags=[staging, production, automated] (merged)
```

### Supported config keys

| Key | Description | Example |
|-----|-------------|---------|
| `token` | API token | `wirepusher config set token wpu_abc123` |
| `api_url` | Custom API endpoint | `wirepusher config set api_url https://custom.com/send` |
| `timeout` | Request timeout (seconds) | `wirepusher config set timeout 60` |
| `max_retries` | Max retry attempts | `wirepusher config set max_retries 5` |
| `default_type` | Default notification type | `wirepusher config set default_type deploy` |

**Note**: `default_tags` must be set directly in the YAML file (it's an array).

## Authentication

Three ways to provide your token (priority order):

```bash
# 1. Flag (highest priority)
wirepusher send "Title" --token wpu_abc123

# 2. Environment variable
export WIREPUSHER_TOKEN=wpu_abc123
wirepusher send "Title"

# 3. Config file (set once, recommended)
wirepusher config set token wpu_abc123
wirepusher send "Title"
```

## Reliability Features

### Automatic retries

Network errors? The CLI automatically retries with exponential backoff:

- **Default**: 3 retries (1s, 2s, 4s delay)
- **Rate limits** (429): Longer backoff (5s, 10s, 20s)
- **Configurable**: `--max-retries 5` or `wirepusher config set max_retries 5`

Disable retries:
```bash
wirepusher send "Title" --max-retries 0
```

### Custom timeout

```bash
# Slow network? Increase timeout
wirepusher send "Title" --timeout 60

# Or set in config
wirepusher config set timeout 60
```

### Debug with verbose mode

```bash
wirepusher send "Title" --verbose

# Shows:
# - Token (first 8 chars)
# - API URL
# - Timeout and retry settings
# - Request progress
```

## Validation Philosophy

The WirePusher CLI validates **more strictly** than the library clients to provide immediate user feedback:

### ✅ CLI Validates

- **Required parameters**: Title and token
- **Tag limits**: Max 10 tags, 50 characters each (API limits)
- **Tag format**: Alphanumeric, hyphens, underscores only
- **Immediate feedback**: Errors shown before API call

### ✅ CLI Normalizes

- **Tags**: Lowercase conversion, whitespace trimming, deduplication
- **Automatic**: No extra flags needed

### Why Stricter Validation in CLI?

**User experience matters.** CLI users benefit from immediate feedback before making API calls:

- ✅ Fast-fail saves API quota
- ✅ Clear error messages with actionable guidance
- ✅ No surprise failures after network round-trip
- ✅ Matches documented API limits upfront

**Note:** WirePusher libraries (Go, Python) perform minimal validation and let the API be the source of truth. This difference is intentional - CLI serves interactive users, libraries serve programmatic integrations.

## Commands Reference

### send

```bash
wirepusher send <title> [message] [flags]
```

**Common flags:**
- `--type string` - Notification type (deploy, alert, etc.)
- `--tag strings` - Tags (repeatable)
- `--image-url string` - Image URL
- `--action-url string` - Action URL (opens on tap)
- `--encryption-password string` - Encrypt message
- `--stdin` - Read message from stdin
- `--json` - JSON output
- `--timeout int` - Request timeout in seconds (default: 30)
- `--max-retries int` - Max retries (default: 3)
- `--verbose` - Debug output

**Examples:**
```bash
wirepusher send "Deploy" "v1.2.3 deployed"
wirepusher send "Alert" "CPU high" --type alert --tag production
wirepusher send "Secure" "Encrypted" --encryption-password "secret"
echo "Output" | wirepusher send "Logs" --stdin
wirepusher send "Deploy" --json  # Machine-readable output
```

### notifai

AI-powered notifications (Gemini):

```bash
wirepusher notifai <text> [flags]
```

AI automatically generates a structured notification from your free-form text.

**Examples:**
```bash
wirepusher notifai "deployment finished, v2.1.3 is live on prod"
wirepusher notifai "cpu at 95% on web-3" --type alert
cat log.txt | wirepusher notifai --stdin
```

### config

```bash
wirepusher config set <key> <value>
wirepusher config get <key>
wirepusher config list
```

See [Supported config keys](#supported-config-keys) above.

### version

```bash
wirepusher version
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `WIREPUSHER_TOKEN` | API token |
| `WIREPUSHER_API_URL` | Custom API endpoint |
| `WIREPUSHER_TIMEOUT` | Request timeout (seconds) |
| `WIREPUSHER_MAX_RETRIES` | Max retry attempts |

## Building from Source

```bash
git clone https://gitlab.com/wirepusher/wirepusher-cli.git
cd wirepusher-cli
go build -o wirepusher
```

With version info:
```bash
go build -ldflags="-X 'gitlab.com/wirepusher/cli/cmd.version=1.0.0'" -o wirepusher
```

## Testing

```bash
go test ./...              # Run all tests
go test ./... -cover       # With coverage
go test ./... -v -race     # Verbose with race detection
```

## Documentation

- **Architecture**: See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for design details
- **Contributing**: See [CONTRIBUTING.md](CONTRIBUTING.md) for development guide
- **Changelog**: See [CHANGELOG.md](CHANGELOG.md) for release history

## Links

- **App**: https://wirepusher.dev
- **Help**: https://wirepusher.dev/help
- **API Docs**: https://wirepusher.com/docs
- **Repository**: https://gitlab.com/wirepusher/wirepusher-cli
- **Issues**: https://gitlab.com/wirepusher/wirepusher-cli/-/issues

## License

MIT License - see [LICENSE](LICENSE)
