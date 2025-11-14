#!/bin/bash
# Configuration management example

# Set up configuration (do this once)
wirepusher config set token "wpt_your_token_here"

# View configuration
wirepusher config list

# Get specific values
wirepusher config get token

# Now you can send without specifying token each time
wirepusher send "Test" "Configuration is working"

# You can still override config with flags if needed
wirepusher send "Test" "Using different credentials" \
  --token "temporary-token" \
  
