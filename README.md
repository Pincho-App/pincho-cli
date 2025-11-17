# WirePusher CLI

Official command-line tool for [WirePusher](https://wirepusher.dev) push notifications.

Send push notifications from your terminal, scripts, and CI/CD pipelines.

## Installation

```bash
curl -sSL https://gitlab.com/wirepusher/wirepusher-cli/-/raw/main/install.sh | sh
```

Or download from [Releases](https://gitlab.com/wirepusher/wirepusher-cli/-/releases).

## Quick Start

```bash
# Get token: Open app → Settings → Help → Copy token

# Send notification
wirepusher send "Deploy Complete" "v1.2.3 is live" --token YOUR_TOKEN

# Or use environment variable
export WIREPUSHER_TOKEN=YOUR_TOKEN
wirepusher send "Deploy Complete" "v1.2.3 is live"

# Or save in config (recommended)
wirepusher config set token YOUR_TOKEN
wirepusher send "Deploy Complete" "v1.2.3 is live"
```

## Features

### Send notifications

```bash
wirepusher send "Build Complete"
wirepusher send "Build Complete" "All tests passed"
wirepusher send "Deploy" "v1.2.3" --type deploy --tag production
```

### AI-generated notifications

```bash
wirepusher notifai "deployment finished successfully, v2.1.3 is live on prod"
# AI generates title, message, tags automatically
```

### Pipe command output

```bash
echo "Backup complete" | wirepusher send "Backup Status" --stdin
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

Exit codes for automation:
- `0` = Success
- `1` = Invalid input or auth error
- `2` = API error or rate limit
- `3` = Network error

### Encrypted messages

```bash
wirepusher send "Security Alert" "Sensitive data" \
  --type secure \
  --encryption-password "secret123"
```

## Configuration

```bash
wirepusher config set token YOUR_TOKEN
wirepusher config set timeout 60
wirepusher config set max_retries 5
wirepusher config list
```

Config file: `~/.wirepusher/config.yaml`

## Smart Retries

Automatic exponential backoff for network errors:
- Default: 3 retries (1s, 2s, 4s)
- Rate limits: Respects `Retry-After` header
- Disable: `--max-retries 0`

## Environment Variables

| Variable | Description |
|----------|-------------|
| `WIREPUSHER_TOKEN` | API token |
| `WIREPUSHER_TIMEOUT` | Request timeout (seconds) |
| `WIREPUSHER_MAX_RETRIES` | Max retry attempts |

## Requirements

- Go 1.21+ (for building from source)
- Bash/sh (for install script)

## Links

- [Advanced Usage](docs/ADVANCED.md) - Detailed configuration, commands reference, validation
- [Architecture](docs/ARCHITECTURE.md) - Design and internals
- [Contributing](CONTRIBUTING.md) - Development guide
- [Changelog](CHANGELOG.md) - Release history
- [App](https://wirepusher.dev)
- [API Docs](https://wirepusher.com/docs)
- [Issues](https://gitlab.com/wirepusher/wirepusher-cli/-/issues)

## License

MIT License - see [LICENSE](LICENSE)
