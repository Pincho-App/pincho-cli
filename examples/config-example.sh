#!/bin/bash
# Configuration management example

# Set up configuration (do this once)
wirepusher config set token "wpt_your_token_here"
wirepusher config set id "your-user-id"

# View configuration
wirepusher config list

# Get specific values
wirepusher config get token
wirepusher config get id

# Now you can send without specifying token/id each time
wirepusher send "Test" "Configuration is working"

# You can still override config with flags if needed
wirepusher send "Test" "Using different credentials" \
  --token "temporary-token" \
  --id "temporary-id"
