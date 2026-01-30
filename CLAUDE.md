# CLAUDE.md - Pincho CLI

Context file for AI-powered development assistance on the Pincho CLI project.

## Project Overview

**Pincho CLI** is a command-line tool for sending push notifications via [Pincho](https://pincho.app).

- **Language**: Go 1.23+
- **Framework**: Cobra (CLI) + Viper (config)
- **Purpose**: Send notifications from terminal, scripts, CI/CD pipelines
- **Philosophy**: Simple, fast, zero dependencies (single static binary)

## Architecture

```
pincho-cli/
├── main.go                 # Entry point
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command, global flags
│   ├── send.go            # Send notifications
│   ├── notifai.go         # AI-generated notifications
│   ├── config.go          # Manage config
│   ├── version.go         # Version info
│   └── helpers.go         # Shared utilities
├── pkg/
│   ├── client/            # Pincho API client
│   │   ├── client.go      # HTTP client with retry logic
│   │   └── client_test.go
│   ├── config/            # Configuration management
│   │   ├── config.go      # Viper-based config
│   │   └── config_test.go
│   ├── crypto/            # AES-128-CBC encryption
│   │   ├── crypto.go
│   │   └── crypto_test.go
│   ├── validation/        # Input validation
│   │   ├── tags.go        # Tag normalization
│   │   └── tags_test.go
│   ├── errors/            # Error handling & exit codes
│   │   └── exit_codes.go
│   └── logging/           # Logging utilities
│       └── logger.go
└── docs/                   # Documentation
    ├── ARCHITECTURE.md     # Detailed architecture guide
    ├── CONTRIBUTING.md     # Contributor guide
    └── CHANGELOG.md        # Release history
```

## Commands

### send
Send push notifications with full control:
```bash
pincho send "Title" "Message" [flags]
```

**Key features:**
- Optional message (title-only notifications supported)
- Type and tags
- Image and action URLs
- AES-128-CBC encryption
- Stdin support for piping

### notifai
AI-powered notifications (Gemini converts free-form text to structured notifications):
```bash
pincho notifai "deployment finished, v2.1.3 is live" [flags]
```

AI generates title, message, tags, and action URL automatically.

### config
Manage persistent configuration:
```bash
pincho config set <key> <value>
pincho config get <key>
pincho config list
```

Stores in `~/.pincho/config.yaml` with secure permissions (0700 dir, 0600 file).

## Configuration Priority

1. **Command-line flags** (highest)
2. **Environment variables** (PINCHO_TOKEN, etc.)
3. **Config file** (`~/.pincho/config.yaml`)
4. **Defaults** (lowest)

**Supported config keys:**
- `token` - API token
- `api_url` - Custom API endpoint
- `timeout` - Request timeout (seconds)
- `max_retries` - Max retry attempts
- `default_type` - Auto-apply notification type
- `default_tags` - Auto-merge tags (array)

## Key Features

### Exit Codes (CI/CD Integration)
- `0` - Success
- `1` - Usage error (invalid input, auth failure)
- `2` - API error (rate limit, validation, server error)
- `3` - System error (network, timeout)

### Automatic Retries
- **Retryable**: Network errors, 5xx errors, 429 (rate limit)
- **Non-retryable**: 400 (validation), 401/403 (auth)
- **Backoff**: Exponential (1s, 2s, 4s, 8s, capped at 30s)
- **Rate limits**: Special handling with 5s initial backoff
- **Configurable**: `--max-retries` flag or `max_retries` config

### Tag Validation & Normalization
- Automatic lowercase conversion
- Whitespace trimming
- Duplicate removal
- Character validation (alphanumeric, hyphens, underscores)

### Encryption
- **Algorithm**: AES-128-CBC
- **Key derivation**: SHA1 (for mobile app compatibility)
- **Scope**: Only message body encrypted
- **Password**: Must match type config in Pincho app

### Verbose Logging
- `--verbose` flag enables debug output
- Shows: token (first 8 chars), API URL, timeout, retries, request progress
- All logging goes to stderr (keeps stdout clean for JSON output)

## Development

### Building
```bash
go build -o pincho
```

With version info:
```bash
go build -ldflags="-X 'gitlab.com/pincho-app/pincho-cli/cmd.version=1.0.0'" -o pincho
```

### Testing
```bash
go test ./...              # All tests
go test ./... -v -race     # With race detection
go test -cover ./...       # With coverage
```

**Current coverage:**
- validation: 100%
- crypto: 87.9%
- config: 71.2%
- client: 44.8%

### Code Style
- Standard Go formatting (`go fmt`)
- Vet before commit (`go vet`)
- Table-driven tests with subtests
- Race detection enabled

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

No runtime dependencies. Compiles to single static binary.

## API Integration

### Endpoints
- `POST /send` - Send notifications
- `POST /notifai` - AI-generated notifications

### Authentication
Token via:
- `--token` flag
- `PINCHO_TOKEN` env var
- `token` config key

### Response Format
Structured JSON with nested error details:
```json
{
  "status": "error",
  "error": {
    "type": "validation_error",
    "code": "invalid_parameter",
    "message": "Title is required",
    "param": "title"
  }
}
```

## Testing Philosophy

- **Unit tests**: Test functions in isolation
- **Table-driven**: Multiple scenarios with subtests
- **Mock HTTP**: Use test servers for client testing
- **Race detection**: All tests run with `-race`
- **100% coverage goal**: For critical paths (validation, crypto)

## Common Development Tasks

### Adding a new command
1. Create `cmd/newcommand.go`
2. Define cobra.Command
3. Add to `rootCmd` in `cmd/root.go`
4. Add flags and logic
5. Update README.md
6. Add tests

### Adding a config key
1. Add field to `Config` struct in `pkg/config/config.go`
2. Add getter helper in `cmd/helpers.go`
3. Use in command
4. Update documentation

### Adding a new validation rule
1. Add logic to `pkg/validation/`
2. Add comprehensive tests (table-driven)
3. Integrate into client or command

## Troubleshooting

### Tests failing
```bash
go test ./... -v          # Verbose output
go test ./pkg/client -v   # Specific package
go clean -testcache       # Clear test cache
```

### Build issues
```bash
go mod tidy              # Clean dependencies
go mod verify            # Verify dependencies
go clean -cache          # Clear build cache
```

## Links

- **Repository**: https://gitlab.com/pincho-app/pincho-cli
- **Issues**: https://gitlab.com/pincho-app/pincho-cli/-/issues
- **API Docs**: https://pincho.app/help
- **App**: https://pincho.app

## Notes for AI Assistants

- **Simplicity is key**: CLI should be as simple as the Pincho app
- **No hardcoded limits**: Don't document API limits (they may change)
- **Config is optional**: CLI works with zero setup (just pass `--token`)
- **Flags override everything**: Maintain priority chain (flag > env > config > default)
- **Test before commit**: Always run `go test -v -race ./...`
- **Documentation tone**: Concise, clear, actionable, developer-focused
- **Exit codes matter**: CI/CD users rely on proper exit codes
- **Verbose logging**: Help users debug without cluttering normal output

## Project Status

**Current**: Production-ready, feature-complete CLI

**Completed improvements:**
- ✅ Fixed critical bugs and security issues
- ✅ Added comprehensive test coverage
- ✅ Implemented professional error handling
- ✅ Added retry logic and timeout configuration
- ✅ Created complete documentation
- ✅ Added config defaults for power users

**Not needed:**
- ❌ Integration tests (simple 2-command CLI, unit tests sufficient)
- ❌ Progress indicators (API calls are fast)
- ❌ Additional commands (send + notifai cover all use cases)
