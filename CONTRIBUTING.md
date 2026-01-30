# Contributing to Pincho CLI

Thanks for considering contributing! This is a small project with a small team, so every contribution makes a real difference.

## Code of Conduct

Be respectful and constructive. See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for details.

## Quick Start

```bash
# Get the code
git clone https://gitlab.com/pincho/pincho-cli.git
cd pincho-cli

# Build and test
go build -o pincho
go test ./...

# Make sure everything passes
go test -race ./...
go fmt ./...
go vet ./...
```

**Requirements:** Go 1.23+

## How to Contribute

### Report a Bug

[Check existing issues](https://gitlab.com/pincho/pincho-cli/-/issues) first, then create a new one with:

- What you did (exact command)
- What happened vs what you expected
- Your environment (CLI version, OS, Go version)
- Error output (use `--verbose` for more details)

**Example:**
```
## Retry logic doesn't work for 429 errors

**Command:** pincho send "Title" "Message" --verbose
**Expected:** Retries with backoff
**Actual:** Exits immediately
**Environment:** CLI v1.0.0, Go 1.23, Ubuntu 22.04

[Paste error output]
```

### Suggest a Feature

Create an issue explaining:
- What you want and why
- How it would work
- Any alternatives you considered

### Submit Code

1. **Fork** the repo
2. **Create a branch**: `git checkout -b fix-something`
3. **Make your changes**
4. **Add tests** for new functionality
5. **Run tests**: `go test -v -race ./...`
6. **Format**: `go fmt ./... && go vet ./...`
7. **Commit** with a clear message
8. **Push** to your fork
9. **Open a Merge Request**

## Code Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go) conventions
- Keep functions small and focused
- Write table-driven tests
- Document exported functions
- Update README for user-facing changes

**Good commit messages:**
```
Add retry logic for network errors
Fix tag validation edge case
Update docs with config examples
```

**Bad commit messages:**
```
fix bug
update
changes
```

## Project Structure

```
cmd/          # CLI commands (Cobra)
pkg/          # Packages (client, config, crypto, validation, errors, logging)
docs/         # Documentation
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## Common Tasks

**Add a command:** Create `cmd/newcommand.go`, add to `rootCmd`, write tests

**Add config key:** Add to `pkg/config/config.go`, add getter in `cmd/helpers.go`, update docs

**Add validation:** Add logic in `pkg/validation/`, write tests, integrate

## Testing

```bash
go test ./...              # All tests
go test -v ./...           # Verbose
go test -cover ./...       # With coverage
go test -race ./...        # Race detection (important!)
go test ./pkg/client -v    # Specific package
```

## Building

```bash
# Local build
go build -o pincho

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o pincho-linux-amd64
GOOS=darwin GOARCH=arm64 go build -o pincho-darwin-arm64
```

## CI/CD Pipeline

Every push to the repository triggers Cloud Build:

- Runs tests with race detector and coverage thresholds
- Checks code formatting (gofmt)
- Runs static analysis (go vet)
- Verifies builds for all 7 target platforms

Git tags (`v*.*.*`) trigger full release pipeline with GoReleaser.

For CI/CD setup details, see [docs/CI_CD_SETUP.md](docs/CI_CD_SETUP.md).

## Releasing (For Maintainers)

Releases are automated via Cloud Build and GoReleaser. Here's the process:

### Pre-Release Checklist

1. **Update CHANGELOG.md**
   - Move items from `[Unreleased]` to new version section
   - Add release date: `## [1.0.0] - 2024-01-15`

2. **Verify test coverage**
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -func=coverage.out | grep total
   # Must be >= 30% overall
   ```

3. **Run all quality checks**
   ```bash
   go fmt ./...
   go vet ./...
   go test -race ./...
   ```

4. **Commit and push**
   ```bash
   git add -A
   git commit -m "chore: Prepare for v1.0.0 release"
   git push origin main
   ```

### Create Release

```bash
# Create annotated tag
git tag -a v1.0.0 -m "Release v1.0.0"

# Push tag (triggers full release pipeline)
git push origin v1.0.0
```

### Post-Release Verification

1. **Monitor Cloud Build**
   - Check: https://console.cloud.google.com/cloud-build/builds?project=pincho-dev
   - All steps should pass (tests, GoReleaser)

2. **Verify GitLab Release**
   - Check: https://gitlab.com/pincho/pincho-cli/-/releases
   - Should have 7 platform binaries + checksums.txt

3. **Test Binary Download**
   ```bash
   # Download and verify
   curl -LO https://gitlab.com/pincho/pincho-cli/-/releases/v1.0.0/downloads/pincho_1.0.0_darwin_arm64.tar.gz
   tar -xzf pincho_1.0.0_darwin_arm64.tar.gz
   ./pincho --version
   ```

### Version Numbering

Follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking API changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

Examples:
- `v1.0.0` - First stable release
- `v1.1.0` - Added new command
- `v1.1.1` - Fixed bug in existing command
- `v2.0.0` - Changed flag names (breaking)

Pre-releases:
- `v1.0.0-alpha.1` - Early testing
- `v1.0.0-beta.1` - Feature complete, needs testing
- `v1.0.0-rc.1` - Release candidate

## Need Help?

- **Architecture questions?** See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- **AI development?** See [CLAUDE.md](CLAUDE.md)
- **Stuck?** Open an issue or ask in your MR

## License

By contributing, you agree your contributions will be licensed under the MIT License.

---

Thanks for contributing! 🙌
