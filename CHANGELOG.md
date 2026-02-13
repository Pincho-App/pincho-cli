# Changelog

All notable changes to Pincho CLI are documented here.

Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), following [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **NotifAI command**: AI-powered notifications using Gemini to convert free-form text into structured notifications
- **Exit codes**: Standardized exit codes for CI/CD integration (0=success, 1=usage error, 2=API error, 3=system error)
- **Automatic retries**: Exponential backoff for network errors, server errors, and rate limits
- **Timeout configuration**: Configurable via `--timeout` flag, `PINCHO_TIMEOUT` env var, or `timeout` config key
- **Retry configuration**: Configurable via `--max-retries` flag, `PINCHO_MAX_RETRIES` env var, or `max_retries` config key
- **Verbose logging**: `--verbose` flag for debugging (shows token, API URL, timeout, retries, progress)
- **Config defaults**: `default_type` and `default_tags` in config file for automatic notification defaults
- **Tag merging**: Intelligent merging of flag tags with config default_tags (no duplicates)
- **Comprehensive tests**: 100% coverage for validation, 87.9% for crypto, 71.2% for config
- **Package documentation**: Comprehensive package-level docs for all 7 packages
- **Architecture guide**: ARCHITECTURE.md with design decisions and implementation details
- **AI context file**: CLAUDE.md for AI-powered development assistance

### Changed
- **Config file permissions**: Changed from 0755 to 0700 (owner-only) for security
- **Config file permissions**: Files created with 0600 (owner read/write only)
- **Tag validation**: Tags automatically normalized (lowercase, trimmed, deduplicated)
- **Error messages**: Improved with actionable suggestions (how to get token, rate limit info, etc.)
- **Message parameter**: Made optional in send command (title-only notifications supported)
- **Documentation**: Simplified README with use-case driven examples, removed hardcoded limits

### Fixed
- **Broken client tests**: Fixed 5 test functions with incorrect signature
- **Security vulnerability**: Config directory permissions too open (world-readable tokens)
- **Code duplication**: Extracted 50+ lines of duplicate code into helpers.go
- **Unused verbose flag**: Implemented proper verbose logging system
- **Missing test coverage**: Added 42 tag validation tests, 11 config tests

### Technical Improvements
- **Retry logic**: Retries on network errors, 5xx errors, 429 (rate limit) with exponential backoff (1s, 2s, 4s, 8s)
- **Error categorization**: Automatic categorization of errors (validation, auth, API, network)
- **Tag normalization**: Automatic lowercase, whitespace trim, deduplication, character validation
- **Configuration priority**: Maintained consistent priority (flag > env > config > default)
- **Exit code handling**: Proper exit codes enable better automation and CI/CD integration
- **Verbose debug output**: Helps users troubleshoot without cluttering normal output

## [1.0.0] - TBD

Initial stable release.

### Features
- Send push notifications via `send` command
- AI-generated notifications via `notifai` command
- Configuration management via `config` command
- Multiple authentication methods (flag, env var, config file)
- Stdin support for piping command output
- Full notification features (types, tags, images, actions)
- Message encryption with AES-128-CBC
- Automatic retries with exponential backoff
- CI/CD ready with proper exit codes
- Cross-platform binaries (Linux, macOS, Windows)
- Zero runtime dependencies (single static binary)

---

## Upgrade Guide

### To Unreleased

If upgrading from an early version:

1. **Config permissions**: Config files will be recreated with secure permissions (0700/0600) on first use
2. **Optional message**: `send` command now supports title-only notifications
3. **New config keys**: Add `timeout`, `max_retries`, `default_type`, `default_tags` to config file if desired
4. **Exit codes**: Update CI/CD scripts to handle new exit codes (0, 1, 2, 3)
5. **Retry behavior**: Network errors now retry automatically (disable with `--max-retries 0`)

### Configuration Example

```yaml
# ~/.pincho/config.yaml
token: your-token-here
timeout: 60                 # Optional
max_retries: 5              # Optional
default_type: deploy        # Optional
default_tags:               # Optional
  - production
  - automated
```

---

## Support

- **Repository**: https://github.com/Pincho-App/pincho-cli
- **Issues**: https://github.com/Pincho-App/pincho-cli/issues
- **Documentation**: See README.md and docs/ARCHITECTURE.md
