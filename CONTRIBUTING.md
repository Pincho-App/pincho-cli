# Contributing to WirePusher CLI

Thanks for considering contributing! This is a small project with a small team, so every contribution makes a real difference.

## Code of Conduct

Be respectful and constructive. See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for details.

## Quick Start

```bash
# Get the code
git clone https://gitlab.com/wirepusher/wirepusher-cli.git
cd wirepusher-cli

# Build and test
go build -o wirepusher
go test ./...

# Make sure everything passes
go test -race ./...
go fmt ./...
go vet ./...
```

**Requirements:** Go 1.23+

## How to Contribute

### Report a Bug

[Check existing issues](https://gitlab.com/wirepusher/wirepusher-cli/-/issues) first, then create a new one with:

- What you did (exact command)
- What happened vs what you expected
- Your environment (CLI version, OS, Go version)
- Error output (use `--verbose` for more details)

**Example:**
```
## Retry logic doesn't work for 429 errors

**Command:** wirepusher send "Title" "Message" --verbose
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
go build -o wirepusher

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o wirepusher-linux-amd64
GOOS=darwin GOARCH=arm64 go build -o wirepusher-darwin-arm64
```

## Need Help?

- **Architecture questions?** See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- **AI development?** See [CLAUDE.md](CLAUDE.md)
- **Stuck?** Open an issue or ask in your MR

## License

By contributing, you agree your contributions will be licensed under the MIT License.

---

Thanks for contributing! 🙌
