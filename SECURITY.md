# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

The WirePusher team takes security bugs seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report

**Please do NOT report security vulnerabilities through public GitLab issues.**

Instead, please report security vulnerabilities via email to:

**security@wirepusher.com**

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

When using the WirePusher JavaScript SDK:

1. **Keep the SDK updated** to the latest version
2. **Never commit credentials** to version control
3. **Use environment variables** for sensitive configuration
4. **Validate input** before sending to the SDK
5. **Handle errors gracefully** without exposing sensitive information
6. **Use HTTPS** for all network communication
7. **Limit token scope** to minimum required permissions

### Credential Management

```typescript
// ❌ Bad - Hardcoded credentials
const client = new WirePusher({
  token: 'wpt_abc123',
  userId: 'user123'
});

// ✅ Good - Environment variables
const client = new WirePusher({
  token: process.env.WIREPUSHER_TOKEN!,
  userId: process.env.WIREPUSHER_USER_ID!
});
```

### Error Handling

```typescript
// ❌ Bad - Exposes sensitive information
try {
  await client.send(title, message);
} catch (error) {
  console.error(error); // May log tokens or user IDs
  throw error;
}

// ✅ Good - Safe error handling
try {
  await client.send(title, message);
} catch (error) {
  if (error instanceof WirePusherError) {
    console.error('Notification failed:', error.message);
    // Handle without exposing credentials
  }
}
```

### Input Validation

```typescript
// ❌ Bad - No validation
app.post('/notify', async (req, res) => {
  await client.send(req.body.title, req.body.message);
});

// ✅ Good - Validate input
app.post('/notify', async (req, res) => {
  const { title, message } = req.body;

  if (!title || !message) {
    return res.status(400).json({ error: 'Missing required fields' });
  }

  if (title.length > 256 || message.length > 4096) {
    return res.status(400).json({ error: 'Content too long' });
  }

  await client.send(title, message);
});
```

## Known Security Considerations

### API Token Security

- Tokens are transmitted in API requests and should be kept confidential
- Tokens are stored in plaintext by the SDK (secure storage is the user's responsibility)
- Compromised tokens can be used to send notifications as your user
- Rotate tokens regularly as a security best practice

### Network Communication

- All communication with WirePusher API is over HTTPS
- The SDK uses native fetch which respects system-level TLS/SSL settings
- Certificate validation is handled by the Node.js runtime

### Dependencies

This SDK has **zero runtime dependencies** to minimize supply chain risks:
- Uses native fetch API (Node.js 18+)
- All development dependencies are thoroughly vetted
- Regular dependency audits via `npm audit`

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
4. **Release patched versions** to npm
5. **Publish security advisory** on GitLab and npm
6. **Notify users** via:
   - GitLab security advisory
   - npm security advisory
   - Project README update

## Security Audit History

| Date | Type | Findings | Status |
|------|------|----------|--------|
| TBD  | TBD  | TBD      | TBD    |

## Security Hall of Fame

We thank the following individuals for responsibly disclosing security vulnerabilities:

- (None yet)

## Resources

- [npm Security Best Practices](https://docs.npmjs.com/packages-and-modules/securing-your-code)
- [Node.js Security Best Practices](https://nodejs.org/en/docs/guides/security/)
- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)

## Questions?

For security-related questions that aren't reporting vulnerabilities:

- Email: security@wirepusher.com
- General questions: support@wirepusher.com

Thank you for helping keep WirePusher and its users safe!
