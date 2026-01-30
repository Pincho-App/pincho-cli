#!/bin/bash
# Configuration management example

# Set up configuration (do this once)
pincho config set token "wpt_your_token_here"

# View configuration
pincho config list

# Get specific values
pincho config get token

# Now you can send without specifying token each time
pincho send "Test" "Configuration is working"

# You can still override config with flags if needed
pincho send "Test" "Using different credentials" \
  --token "temporary-token" \
  
