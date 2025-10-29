# Changelog

All notable changes to the WirePusher CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of WirePusher CLI
- `send` command for sending notifications
- `config` command for managing configuration
- `version` command for version information
- Support for v1 WirePusher API
- Configuration via flags, environment variables, or config file
- Stdin support for piping command output
- Support for notification types, tags, images, and action URLs
- Message encryption with AES-128-CBC (--encryption-password flag)
- Comprehensive examples for CI/CD, log monitoring, and more
- Cross-platform binaries (Linux, macOS, Windows)
- GoReleaser configuration for automated GitHub Releases
- GitHub Actions workflows for CI/CD

### Deprecated
- `--id` flag for user ID authentication (use `--token` instead, will be removed in v2.0.0)

## [1.0.0] - TBD

Initial stable release.

### Features
- Send push notifications from command line
- Configuration persistence (`~/.wirepusher/config.yaml`)
- Multiple authentication methods
- Stdin support for log monitoring
- Full notification features (types, tags, images, actions)
- CI/CD ready

---

## Version History

- **1.0.0** - Initial stable release (TBD)

## Upgrading

### To 1.0.0

Initial release - no migration needed.

## Support

For questions about specific versions:
- Check the [documentation](https://gitlab.com/wirepusher/cli)
- Open an [issue](https://gitlab.com/wirepusher/cli/-/issues)
- Email support@wirepusher.com
