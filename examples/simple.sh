#!/bin/bash
# Simple notification example

# Send a basic notification
pincho send "Build Complete" "The build finished successfully"

# Send with notification type
pincho send "Alert" "CPU usage is high" --type alert

# Send with tags
pincho send "Deploy" "v1.2.3 deployed to production" \
  --tag production \
  --tag release
