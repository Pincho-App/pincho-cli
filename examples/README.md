# Pincho CLI Examples

This directory contains example scripts demonstrating various use cases for the Pincho CLI.

## Prerequisites

Before running these examples, you need:

1. Pincho CLI installed
2. API credentials from your Pincho account (token)

Set up your credentials:
```bash
# Option 1: Using token (recommended)
pincho config set token YOUR_TOKEN

```

Or use environment variables:
```bash
# Using token
export PINCHO_TOKEN="your-token"

```

## Examples

### [simple.sh](simple.sh)
Basic notification examples showing:
- Simple title + message
- Adding notification types
- Using tags for categorization

```bash
chmod +x examples/simple.sh
./examples/simple.sh
```

### [ci-cd.sh](ci-cd.sh)
CI/CD pipeline integration showing:
- Build start/success/failure notifications
- Using CI environment variables
- Adding action URLs to link back to pipeline

Use in GitHub Actions:
```yaml
deploy:
  steps:
    - run: ./examples/ci-cd.sh
  env:
    PINCHO_TOKEN: ${{ secrets.PINCHO_TOKEN }}
```

### [log-monitoring.sh](log-monitoring.sh)
Log file monitoring showing:
- Reading from stdin (`--stdin` flag)
- Continuous monitoring with tail -f
- Filtering logs for specific patterns

```bash
chmod +x examples/log-monitoring.sh
./examples/log-monitoring.sh
```

### [config-example.sh](config-example.sh)
Configuration management showing:
- Setting up persistent configuration
- Listing and retrieving config values
- Overriding config with command-line flags

```bash
chmod +x examples/config-example.sh
./examples/config-example.sh
```

### [encryption.sh](encryption.sh)
End-to-end encryption examples showing:
- Basic encrypted notifications with AES-128-CBC
- Encrypted notifications with tags, images, and action URLs
- Piping sensitive data with encryption
- Using environment variables for passwords
- Batch sending with encryption

**Prerequisites for encryption:**
1. Create notification types in Pincho app with encryption enabled
2. Set the same password in the app as used in the examples

```bash
chmod +x examples/encryption.sh
./examples/encryption.sh
```

**Important:** Only the message body is encrypted; title, type, tags, images, and action URLs remain unencrypted for proper filtering and display.

## Common Patterns

### Notify on script completion
```bash
#!/bin/bash
./your-script.sh && \
  pincho send "Success" "Script completed" --type success || \
  pincho send "Failed" "Script failed" --type alert
```

### Notify with command output
```bash
df -h | pincho send "Disk Usage" --stdin
```

### Scheduled notifications (cron)
```bash
# Add to crontab
0 9 * * * pincho send "Daily Reminder" "Good morning! Time to check the dashboard"
```

## Authentication Methods

**Important:** Use token for authentication.

The CLI supports three authentication methods (in priority order):

1. **Command-line flags** (highest priority)
   ```bash
   # Using token (recommended)
   pincho send "Title" "Message" --token abc123

   
   ```

2. **Environment variables**
   ```bash
   # Using token
   export PINCHO_TOKEN="your-token"
   pincho send "Title" "Message"

   
   pincho send "Title" "Message"
   ```

3. **Config file** (lowest priority)
   ```bash
   # Using token
   pincho config set token your-token
   pincho send "Title" "Message"

   
   pincho send "Title" "Message"
   ```

## Troubleshooting

### Permission denied
```bash
chmod +x examples/*.sh
```

### Token not found
```bash
# Check your configuration
pincho config list

# Or set it
pincho config set token YOUR_TOKEN
```

### Test your setup
```bash
pincho send "Test" "Testing Pincho CLI"
```

## More Information

- Full CLI documentation: [../README.md](../README.md)
- Pincho website: https://pincho.com
- GitHub repository: https://github.com/Pincho-App/pincho-cli
