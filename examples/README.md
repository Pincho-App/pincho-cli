# WirePusher CLI Examples

This directory contains example scripts demonstrating various use cases for the WirePusher CLI.

## Prerequisites

Before running these examples, you need:

1. WirePusher CLI installed
2. API credentials from your WirePusher account (EITHER token OR user ID - not both)

Set up your credentials:
```bash
# Option 1: Using token (recommended)
wirepusher config set token YOUR_TOKEN

# Option 2: Using user ID (alternative)
wirepusher config set id YOUR_USER_ID
```

Or use environment variables:
```bash
# Using token
export WIREPUSHER_TOKEN="your-token"

# Using user ID
export WIREPUSHER_ID="your-user-id"
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

Use in GitLab CI:
```yaml
deploy:
  script:
    - ./examples/ci-cd.sh
  variables:
    WIREPUSHER_TOKEN: $WIREPUSHER_TOKEN
    WIREPUSHER_ID: $WIREPUSHER_ID
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
1. Create notification types in WirePusher app with encryption enabled
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
  wirepusher send "Success" "Script completed" --type success || \
  wirepusher send "Failed" "Script failed" --type alert
```

### Notify with command output
```bash
df -h | wirepusher send "Disk Usage" --stdin
```

### Scheduled notifications (cron)
```bash
# Add to crontab
0 9 * * * wirepusher send "Daily Reminder" "Good morning! Time to check the dashboard"
```

## Authentication Methods

**Important:** Use EITHER token OR user ID - not both. These credentials are mutually exclusive.

The CLI supports three authentication methods (in priority order):

1. **Command-line flags** (highest priority)
   ```bash
   # Using token (recommended)
   wirepusher send "Title" "Message" --token abc123

   # Using user ID (alternative)
   wirepusher send "Title" "Message" --id user123
   ```

2. **Environment variables**
   ```bash
   # Using token
   export WIREPUSHER_TOKEN="your-token"
   wirepusher send "Title" "Message"

   # Using user ID
   export WIREPUSHER_ID="your-user-id"
   wirepusher send "Title" "Message"
   ```

3. **Config file** (lowest priority)
   ```bash
   # Using token
   wirepusher config set token your-token
   wirepusher send "Title" "Message"

   # Using user ID
   wirepusher config set id your-user-id
   wirepusher send "Title" "Message"
   ```

## Troubleshooting

### Permission denied
```bash
chmod +x examples/*.sh
```

### Token or ID not found
```bash
# Check your configuration
wirepusher config list

# Or set it
wirepusher config set token YOUR_TOKEN
wirepusher config set id YOUR_ID
```

### Test your setup
```bash
wirepusher send "Test" "Testing WirePusher CLI"
```

## More Information

- Full CLI documentation: [../README.md](../README.md)
- WirePusher website: https://wirepusher.com
- GitLab repository: https://gitlab.com/wirepusher/cli
