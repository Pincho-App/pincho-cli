# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

The Pincho team takes security bugs seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report

**Please do NOT report security vulnerabilities through public GitLab issues.**

Instead, please report security vulnerabilities via email to:

**security@pincho.dev**

### What to Include

To help us triage and fix the issue quickly, please include:

1. **Type of vulnerability** (e.g., authentication bypass, injection, etc.)
2. **Full paths** of source files related to the vulnerability
3. **Location** of the affected source code (tag/branch/commit or direct URL)
4. **Step-by-step instructions** to reproduce the issue
5. **Proof-of-concept or exploit code** (if possible)
6. **Impact** of the vulnerability (what an attacker could achieve)
7. **Any mitigating factors** or workarounds you've identified

### What to Expect

After you submit a report:

1. **Acknowledgment** - We'll acknowledge receipt within 48 hours
2. **Assessment** - We'll assess the vulnerability and determine severity
3. **Updates** - We'll provide regular updates (at least every 7 days)
4. **Fix Timeline** - We aim to release fixes for:
   - **Critical** vulnerabilities: Within 7 days
   - **High** vulnerabilities: Within 14 days
   - **Medium** vulnerabilities: Within 30 days
   - **Low** vulnerabilities: Next regular release

5. **Disclosure** - We'll coordinate with you on public disclosure timing
6. **Credit** - We'll credit you in the security advisory (unless you prefer to remain anonymous)

## Security Best Practices

### For Users

When using the Pincho CLI:

1. **Keep the CLI updated** to the latest version
2. **Never commit credentials** to version control
3. **Use environment variables** for sensitive configuration
4. **Validate input** in scripts before passing to the CLI
5. **Handle errors gracefully** without exposing sensitive information
6. **Use HTTPS** for all network communication
7. **Protect config files** with secure permissions

### Credential Management

```bash
# ❌ Bad - Hardcoded token in script
pincho send "Alert" "Message" --token wpt_abc123

# ❌ Bad - Token in shell history
export PINCHO_TOKEN=wpt_abc123

# ✅ Good - Token from secure environment variable
PINCHO_TOKEN=$(cat ~/.secrets/pincho) pincho send "Alert" "Message"

# ✅ Good - Token stored in config file with secure permissions
pincho config set token wpt_abc123
# Config stored in ~/.pincho/config.yaml (permissions: 0600)

# ✅ Good - Token from password manager
pincho send "Alert" "Message" --token "$(pass show pincho/token)"
```

### Config File Security

```bash
# Check config file permissions
ls -la ~/.pincho/config.yaml
# Should show: -rw------- (0600)

# Fix permissions if needed
chmod 600 ~/.pincho/config.yaml
chmod 700 ~/.pincho/
```

### Error Handling in Scripts

```bash
# ❌ Bad - No error handling
pincho send "Alert" "Message"

# ✅ Good - Handle errors without exposing sensitive info
if ! pincho send "Alert" "Message" 2>/dev/null; then
  echo "Failed to send notification" >&2
  exit 1
fi

# ✅ Good - Check exit codes for specific errors
pincho send "Alert" "Message"
case $? in
  0) echo "Success" ;;
  1) echo "Usage error - check arguments" >&2 ;;
  2) echo "API error - rate limit or validation" >&2 ;;
  3) echo "System error - network issue" >&2 ;;
esac
```

### Input Validation in Scripts

```bash
# ❌ Bad - No validation
pincho send "$USER_INPUT_TITLE" "$USER_INPUT_MESSAGE"

# ✅ Good - Validate input before sending
title="$1"
message="$2"

# Check for empty values
if [[ -z "$title" ]]; then
  echo "Error: Title is required" >&2
  exit 1
fi

# Check for reasonable length
if [[ ${#title} -gt 256 ]]; then
  echo "Error: Title too long (max 256 chars)" >&2
  exit 1
fi

if [[ ${#message} -gt 4096 ]]; then
  echo "Error: Message too long (max 4096 chars)" >&2
  exit 1
fi

pincho send "$title" "$message"
```

### CI/CD Pipeline Security

```yaml
# ✅ Good - Use CI/CD secrets, not hardcoded values
# GitLab CI example
send_notification:
  script:
    - pincho send "Deploy Complete" "Version $CI_COMMIT_TAG deployed"
  variables:
    PINCHO_TOKEN: $PINCHO_TOKEN  # Set in CI/CD settings
```

## Known Security Considerations

### API Token Security

- Tokens are transmitted in `Authorization: Bearer` header over HTTPS
- Tokens are stored in plaintext in `~/.pincho/config.yaml`
- Config file permissions default to 0600 (owner read/write only)
- Compromised tokens can be used to send notifications as your user
- Rotate tokens regularly as a security best practice

### Network Communication

- All communication with Pincho API is over HTTPS
- The CLI uses Go's standard `net/http` package
- Certificate validation is handled by the Go runtime
- No custom TLS configuration - uses system defaults

### Binary Security

- Single static binary with no external dependencies
- No runtime code execution or dynamic loading
- No network connections except to Pincho API
- No local file access except for config file

### Dependencies

This CLI has **minimal dependencies** to reduce supply chain risks:
- `github.com/spf13/cobra` - CLI framework (widely used, well-audited)
- `github.com/spf13/viper` - Configuration management
- All dependencies are checked with `go mod verify`

## Vulnerability Disclosure Process

When we receive a security bug report:

1. **Confirm the vulnerability** and determine affected versions
2. **Develop and test a fix** for all supported versions
3. **Prepare security advisory** with:
   - Description of the vulnerability
   - Affected versions
   - Fixed versions
   - Workarounds (if any)
   - Credit to reporter
4. **Release patched versions**
5. **Publish security advisory** on GitLab
6. **Notify users** via:
   - GitLab security advisory
   - Project README update
   - Release notes

## Security Audit History

| Date | Type | Findings | Status |
|------|------|----------|--------|
| TBD  | TBD  | TBD      | TBD    |

## Security Hall of Fame

We thank the following individuals for responsibly disclosing security vulnerabilities:

- (None yet)

## Resources

- [Go Security Best Practices](https://go.dev/doc/security/best-practices)
- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
- [Shell Script Security](https://wiki.bash-hackers.org/howto/conffile)

## Questions?

For security-related questions that aren't reporting vulnerabilities:

- Email: security@pincho.dev
- General questions: support@pincho.dev

Thank you for helping keep Pincho and its users safe!
