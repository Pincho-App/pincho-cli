# WirePusher CLI Examples

This directory contains example scripts demonstrating various use cases for the WirePusher CLI.

## Prerequisites

Before running these examples, you need:

1. WirePusher CLI installed
2. API credentials from your WirePusher account (token)

Set up your credentials:
```bash
# Option 1: Using token (recommended)
wirepusher config set token YOUR_TOKEN

```

Or use environment variables:
```bash
# Using token
export WIREPUSHER_TOKEN="your-token"

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

**Important:** Use token for authentication.

The CLI supports three authentication methods (in priority order):

1. **Command-line flags** (highest priority)
   ```bash
   # Using token (recommended)
   wirepusher send "Title" "Message" --token abc123

   
   ```

2. **Environment variables**
   ```bash
   # Using token
   export WIREPUSHER_TOKEN="your-token"
   wirepusher send "Title" "Message"

   
   wirepusher send "Title" "Message"
   ```

3. **Config file** (lowest priority)
   ```bash
   # Using token
   wirepusher config set token your-token
   wirepusher send "Title" "Message"

   
   wirepusher send "Title" "Message"
   ```

## Troubleshooting

### Permission denied
```bash
chmod +x examples/*.sh
```

### Token not found
```bash
# Check your configuration
wirepusher config list

# Or set it
wirepusher config set token YOUR_TOKEN
```

### Test your setup
```bash
wirepusher send "Test" "Testing WirePusher CLI"
```

## More Information

- Full CLI documentation: [../README.md](../README.md)
- WirePusher website: https://wirepusher.com
- GitLab repository: https://gitlab.com/wirepusher/cli
