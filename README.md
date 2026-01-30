# Pincho CLI

Official command-line tool for [Pincho](https://pincho.app) push notifications.

Send push notifications from your terminal, scripts, and CI/CD pipelines.

## Installation

```bash
curl -sSL https://gitlab.com/pincho-app/pincho-cli/-/raw/main/install.sh | sh
```

Or download from [Releases](https://gitlab.com/pincho-app/pincho-cli/-/releases).

## Quick Start

```bash
# Get token: Open app → Settings → Help → Copy token

# Send notification
pincho send "Deploy Complete" "v1.2.3 is live" --token YOUR_TOKEN

# Or use environment variable
export PINCHO_TOKEN=YOUR_TOKEN
pincho send "Deploy Complete" "v1.2.3 is live"

# Or save in config (recommended)
pincho config set token YOUR_TOKEN
pincho send "Deploy Complete" "v1.2.3 is live"
```

## Features

### Send notifications

```bash
pincho send "Build Complete"
pincho send "Build Complete" "All tests passed"
pincho send "Deploy" "v1.2.3" --type deploy --tag production
```

### AI-generated notifications

```bash
pincho notifai "deployment finished successfully, v2.1.3 is live on prod"
# AI generates title, message, tags automatically
```

### Pipe command output

```bash
echo "Backup complete" | pincho send "Backup Status" --stdin
```

### CI/CD integration

```yaml
# GitLab CI
deploy:
  script:
    - ./deploy.sh
    - pincho send "Deploy Complete" "Version ${CI_COMMIT_TAG}"
  variables:
    PINCHO_TOKEN: $PINCHO_TOKEN
```

Exit codes for automation:
- `0` = Success
- `1` = Invalid input or auth error
- `2` = API error or rate limit
- `3` = Network error

### Encrypted messages

```bash
pincho send "Security Alert" "Sensitive data" \
  --type secure \
  --encryption-password "secret123"
```

## Configuration

```bash
pincho config set token YOUR_TOKEN
pincho config set timeout 60
pincho config set max_retries 5
pincho config list
```

Config file: `~/.pincho/config.yaml`

## Smart Retries

Automatic exponential backoff for network errors:
- Default: 3 retries (1s, 2s, 4s)
- Rate limits: Respects `Retry-After` header
- Disable: `--max-retries 0`

## Environment Variables

| Variable | Description |
|----------|-------------|
| `PINCHO_TOKEN` | API token |
| `PINCHO_TIMEOUT` | Request timeout (seconds) |
| `PINCHO_MAX_RETRIES` | Max retry attempts |

## Requirements

- Go 1.21+ (for building from source)
- Bash/sh (for install script)

## Links

- [Advanced Usage](docs/ADVANCED.md) - Detailed configuration, commands reference, validation
- [Architecture](docs/ARCHITECTURE.md) - Design and internals
- [Contributing](CONTRIBUTING.md) - Development guide
- [Changelog](CHANGELOG.md) - Release history
- [App](https://pincho.app)
- [API Docs](https://pincho.app/help)
- [Issues](https://gitlab.com/pincho-app/pincho-cli/-/issues)

## License

MIT License - see [LICENSE](LICENSE)
