#!/bin/bash
# Simple notification example

# Send a basic notification
wirepusher send "Build Complete" "The build finished successfully"

# Send with notification type
wirepusher send "Alert" "CPU usage is high" --type alert

# Send with tags
wirepusher send "Deploy" "v1.2.3 deployed to production" \
  --tag production \
  --tag release
