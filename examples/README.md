# WirePusher CLI Examples

This directory contains example scripts demonstrating various use cases for the WirePusher CLI.

## Prerequisites

Before running these examples, you need:

1. WirePusher CLI installed
2. API credentials (token and user ID) from your WirePusher account

Set up your credentials:
```bash
wirepusher config set token YOUR_TOKEN
wirepusher config set id YOUR_USER_ID
```

Or use environment variables:
```bash
export WIREPUSHER_TOKEN="your-token"
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

The CLI supports three authentication methods (in priority order):

1. **Command-line flags** (highest priority)
   ```bash
   wirepusher send "Title" "Message" --token abc --id xyz
   ```

2. **Environment variables**
   ```bash
   export WIREPUSHER_TOKEN="your-token"
   export WIREPUSHER_ID="your-user-id"
   wirepusher send "Title" "Message"
   ```

3. **Config file** (lowest priority)
   ```bash
   wirepusher config set token your-token
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
