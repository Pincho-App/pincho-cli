# Pincho CLI Architecture

This document explains the architecture, design decisions, and implementation details of the Pincho CLI.

## Table of Contents

- [Overview](#overview)
- [Package Structure](#package-structure)
- [Configuration Priority](#configuration-priority)
- [Error Handling](#error-handling)
- [Retry Logic](#retry-logic)
- [Encryption](#encryption)
- [Design Decisions](#design-decisions)

## Overview

Pincho CLI is a command-line tool for sending push notifications via the Pincho API. It's designed to be:

- **Simple**: Easy to use with sensible defaults
- **Flexible**: Supports multiple configuration methods
- **Reliable**: Built-in retry logic and comprehensive error handling
- **Secure**: Encrypted messaging and secure config storage
- **CI/CD-friendly**: Exit codes for automation

## Package Structure

```
gitlab.com/pincho/cli/
├── cmd/                    # Command-line interface
│   ├── root.go            # Root command and global flags
│   ├── send.go            # Send command implementation
│   ├── notifai.go         # NotifAI command implementation
│   ├── config.go          # Config management commands
│   ├── version.go         # Version command
│   └── helpers.go         # Shared helper functions
│
├── pkg/
│   ├── client/            # Pincho API client
│   │   ├── client.go      # HTTP client with retry logic
│   │   └── client_test.go # Client tests
│   │
│   ├── config/            # Configuration management
│   │   ├── config.go      # Config file handling (Viper)
│   │   └── config_test.go # Config tests
│   │
│   ├── crypto/            # Message encryption
│   │   ├── crypto.go      # AES-128-CBC encryption
│   │   └── crypto_test.go # Crypto tests
│   │
│   ├── validation/        # Input validation
│   │   ├── tags.go        # Tag validation/normalization
│   │   └── tags_test.go   # Validation tests
│   │
│   ├── errors/            # Error handling
│   │   └── exit_codes.go  # CLI error types and exit codes
│   │
│   └── logging/           # Logging utilities
│       └── logger.go      # Verbose logging support
│
└── main.go                # Application entry point
```

### Package Responsibilities

**cmd**: Implements the command-line interface using Cobra. Each command (send, notifai, config) has its own file. The package handles argument parsing, flag management, and user interaction.

**pkg/client**: Provides the HTTP client for the Pincho API. Includes retry logic, timeout configuration, and response parsing. Handles both `/send` and `/notifai` endpoints.

**pkg/config**: Manages configuration file operations using Viper. Handles reading from and writing to `~/.pincho/config.yaml` with secure file permissions.

**pkg/crypto**: Implements AES-128-CBC encryption matching the Pincho mobile app. Enables end-to-end encrypted notifications.

**pkg/validation**: Validates and normalizes input parameters (currently tags). Provides early client-side validation before API calls.

**pkg/errors**: Defines CLI error types with standardized exit codes for CI/CD integration.

**pkg/logging**: Simple logging system with verbose output support for debugging.

## Configuration Priority

The CLI supports three configuration methods with the following precedence (highest to lowest):

1. **Command-line flags** (highest priority)
   ```bash
   pincho send "Title" --token abc123 --timeout 60
   ```

2. **Environment variables**
   ```bash
   export PINCHO_TOKEN=abc123
   export PINCHO_TIMEOUT=60
   pincho send "Title"
   ```

3. **Config file** (lowest priority)
   ```bash
   pincho config set token abc123
   pincho send "Title"
   ```

### Configuration Flow

```
Command Execution
    ↓
Check --flag value
    ↓ (if not set)
Check environment variable
    ↓ (if not set)
Check config file (~/.pincho/config.yaml)
    ↓ (if not set)
Use default value or return error
```

### Supported Configuration

| Parameter | Flag | Environment Variable | Config Key | Default |
|-----------|------|---------------------|------------|---------|
| API Token | `--token, -t` | `PINCHO_TOKEN` | `token` | *required* |
| API URL | (none) | `PINCHO_API_URL` | `api_url` | `https://api.pincho.dev/send` |
| Timeout | `--timeout` | `PINCHO_TIMEOUT` | `timeout` | `30` seconds |
| Max Retries | `--max-retries` | `PINCHO_MAX_RETRIES` | `max_retries` | `3` |
| Default Type | `--type` (send only) | (none) | `default_type` | (empty) |
| Default Tags | `--tag` (send only) | (none) | `default_tags` | (empty) |
| Verbose | `--verbose` | (none) | (none) | `false` |

**Note on defaults**: `default_type` and `default_tags` provide convenient defaults from config that are automatically applied when not specified via flags. Flags always override config values.

### Config File Examples

**Basic config** (`~/.pincho/config.yaml`):
```yaml
token: your-api-token-here
```

**Full config with all options**:
```yaml
# Authentication
token: your-api-token-here

# Custom API endpoint (for self-hosted instances)
api_url: https://custom-api.example.com/send

# Operational settings
timeout: 60        # 60 second timeout (slow network)
max_retries: 5     # Retry up to 5 times (flaky CI environment)

# Notification defaults
default_type: deploy                  # All notifications tagged as "deploy"
default_tags:                        # Always include these tags
  - production
  - automated
```

**Usage with config**:
```bash
# Uses all configured defaults
pincho send "Deploy Complete" "v1.2.3 deployed"
# → Uses: timeout=60s, retries=5, type=deploy, tags=[production, automated]

# Override specific settings
pincho send "Test Notification" --type test --tag staging
# → Uses: timeout=60s (from config), retries=5 (from config)
#         type=test (from flag), tags=[staging, production, automated] (merged)

# Completely override
pincho send "Emergency" --timeout 10 --max-retries 0
# → Uses: timeout=10s (from flag), retries=0 (from flag)
```

**Tag merging behavior**:
- Config has `default_tags: [production, automated]`
- Flag provides `--tag deploy --tag release`
- Result: `[deploy, release, production, automated]` (flags first, then defaults, no duplicates)

## Error Handling

### Exit Codes

The CLI uses standardized exit codes for CI/CD integration:

- **0**: Success - command completed successfully
- **1**: Usage Error - invalid input, missing parameters, authentication failure
- **2**: API Error - API returned an error (rate limits, validation, server errors)
- **3**: System Error - network failure, timeout, system-level error

### Error Flow

```
Error occurs
    ↓
Categorize error type (validation, auth, API, network)
    ↓
Wrap in appropriate CLIError (UsageError, APIError, SystemError)
    ↓
Add actionable suggestion if applicable
    ↓
HandleError() → print to stderr → exit with appropriate code
```

### Error Categorization

**UsageError (exit 1)**:
- Missing required parameters (title, token)
- Invalid input (text too short/long, invalid tags)
- Authentication failures (invalid token, unauthorized)

**APIError (exit 2)**:
- Rate limit exceeded (429)
- API validation errors (400)
- Server errors (500-599)

**SystemError (exit 3)**:
- Network connection failures
- Timeouts (request deadline exceeded)
- File system errors
- Configuration errors

### Actionable Error Messages

Errors include suggestions for resolution:

```bash
# Authentication error
Error: Authentication failed: invalid_api_token

Get your token: Open Pincho app → Settings → Help → Copy token
Or set it: pincho config set token YOUR_TOKEN

# Rate limit error
Error: Rate limit exceeded

The send endpoint allows 30 requests per hour. Please wait before trying again.

# Network error
Error: Network error: connection refused

Please check your internet connection and try again.
```

## Retry Logic

### Retry Strategy

The CLI implements automatic retries with exponential backoff for transient errors:

**Retryable Errors**:
- Network errors (connection refused, connection reset, timeout, EOF)
- Server errors (500, 502, 503, 504)
- Rate limit errors (429) with longer backoff

**Non-Retryable Errors**:
- Validation errors (400)
- Authentication errors (401, 403)
- Client errors (404, etc.)

### Exponential Backoff

Retry delays follow exponential backoff with a cap:

| Attempt | Delay (normal) | Delay (rate limit) |
|---------|---------------|-------------------|
| 1st retry | 1 second | 5 seconds |
| 2nd retry | 2 seconds | 10 seconds |
| 3rd retry | 4 seconds | 20 seconds |
| Maximum | 30 seconds | 30 seconds |

### Retry Flow

```
Send Request
    ↓
Error occurs?
    ↓ yes
Is error retryable?
    ↓ yes
Attempt < MaxRetries?
    ↓ yes
Calculate backoff (exponential)
    ↓
Wait (sleep)
    ↓
Clone request & retry
    ↓
(repeat until success or max retries exhausted)
```

### Configuration

```bash
# Disable retries
pincho send "Title" --max-retries 0

# Increase retries
pincho send "Title" --max-retries 5

# Via environment variable
export PINCHO_MAX_RETRIES=5
```

## Encryption

### Why AES-128-CBC with SHA1?

The encryption implementation uses AES-128-CBC with SHA1 key derivation to maintain **compatibility with the Pincho mobile app**. This enables end-to-end encrypted notifications that can be decrypted on the device.

**Design Constraints**:
- Must match iOS/Android app implementation
- Encrypted messages stored on server, decrypted on device
- Same password must work across CLI and mobile apps

**SHA1 Usage**: While SHA1 is deprecated for cryptographic hashing, it's used here only for key derivation (not signature verification). The risk is acceptable given:
- Passwords are user-chosen (not system-generated)
- Keys are not stored or transmitted
- Compatibility with existing mobile apps is required

**Future Consideration**: For new implementations, we recommend PBKDF2 or Argon2 for key derivation.

### Encryption Process

```
User provides password
    ↓
Derive 128-bit key from password (SHA1 hash, first 16 bytes)
    ↓
Generate random 16-byte IV
    ↓
Encrypt message with AES-128-CBC + PKCS7 padding
    ↓
Encode with custom Base64 (URL-safe with custom chars)
    ↓
Send encrypted message + IV (hex) to API
```

### Custom Base64 Encoding

The encryption uses custom Base64 encoding matching the mobile app:

| Standard | Custom |
|----------|--------|
| `+` | `-` |
| `/` | `_` |
| `=` | `.` |

This encoding is URL-safe and compatible with the mobile app's decryption.

### Example

```bash
# Send encrypted notification
pincho send "Secure Alert" "Sensitive data" \
  --encryption-password "secret123" \
  --type secure

# Mobile app decrypts using same password
```

## Design Decisions

### 1. Why Cobra + Viper?

**Cobra**: Industry-standard CLI framework used by kubectl, hugo, and others. Provides:
- Automatic help generation
- Subcommand structure
- Flag parsing and validation
- Shell completion support

**Viper**: Configuration management library that integrates with Cobra. Provides:
- Multiple config sources (files, env vars, flags)
- Automatic precedence handling
- YAML/JSON/TOML support

### 2. Why separate pkg/ packages?

Following Go best practices for CLI tools:
- Clear separation of concerns
- Easier testing (can test packages independently)
- Potential library reuse (pkg/client can be used by other Go programs)
- Better code organization as project grows

### 3. Why exit codes?

Exit codes enable CI/CD integration:

```bash
# Retry on network errors, fail on validation errors
if ! pincho send "Deploy" "$message"; then
  exit_code=$?
  if [ $exit_code -eq 3 ]; then
    echo "Network error, retrying..."
    sleep 5
    pincho send "Deploy" "$message"
  else
    echo "Command failed with exit code $exit_code"
    exit $exit_code
  fi
fi
```

### 4. Why automatic retries?

Transient network failures are common in CI/CD environments:
- Docker container networking
- Cloud service restarts
- Network blips

Automatic retries with backoff make the CLI more reliable without requiring users to implement retry logic in scripts.

### 5. Why verbose logging to stderr?

Keeping stdout clean allows:
```bash
# Parse JSON output
notification=$(pincho send "Title" --json | jq '.response.notificationID')

# Use in pipes
pincho send "Title" --json | jq '.response' | store-notification
```

Verbose debug output goes to stderr and doesn't interfere with structured output.

### 6. Why secure config file permissions?

API tokens are sensitive credentials. Config directory (0700) and files (0600) are readable only by the owner:

```bash
$ ls -la ~/.pincho/
drwx------  2 user  user   64 Nov 14 10:00 .pincho/  # 0700
-rw-------  1 user  user  123 Nov 14 10:00 config.yaml   # 0600
```

This prevents token exposure on shared systems.

### 7. Why NotifAI command?

The `/notifai` endpoint uses AI (Gemini) to convert free-form text into structured notifications. This enables:

```bash
# Instead of:
pincho send "Deploy Complete" \
  "Version 1.2.3 deployed to production" \
  --type deploy \
  --tag production --tag release

# You can just:
pincho notifai "v1.2.3 deployed to production"

# AI generates title, message, tags, and type
```

Perfect for quick notifications from scripts without formatting.

## Testing Strategy

### Test Coverage

- **pkg/crypto**: 87.9% - Encryption/decryption tested with known vectors
- **pkg/validation**: 100% - All validation paths covered (42 test cases)
- **pkg/config**: 71.2% - File operations and permissions tested
- **pkg/client**: 44.8% - HTTP client, retries, error handling

### Test Approach

1. **Unit tests**: Test individual functions in isolation
2. **Table-driven tests**: Test multiple scenarios with subtests
3. **Mock HTTP**: Use test servers for client testing
4. **Race detection**: All tests run with `-race` flag

### Future Testing

- Integration tests for full command execution
- End-to-end tests with mock API server
- Benchmarks for performance regression testing

## Contributing

When adding new features, consider:

1. **Configuration**: Add flag, env var, and config file support where appropriate
2. **Error handling**: Use appropriate CLI error type with helpful messages
3. **Logging**: Add verbose logging for debugging
4. **Testing**: Add tests with good coverage
5. **Documentation**: Update this file and package docs

## References

- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Viper Configuration](https://github.com/spf13/viper)
- [Pincho API Documentation](https://pincho.com/docs)
- [Go CLI Best Practices](https://github.com/cli/cli/blob/trunk/docs/project-layout.md)
