#!/bin/bash
# Encryption example - AES-128-CBC encrypted notifications

# Prerequisites:
# 1. Configure authentication (token or user ID)
# 2. In WirePusher app, create notification types with encryption enabled
# 3. Set the same encryption password in the app as used in these examples

echo "=== WirePusher CLI Encryption Examples ==="
echo

# Example 1: Basic encrypted notification
echo "1. Basic encrypted notification"
wirepusher send "Secure Alert" "This message is encrypted with AES-128-CBC" \
  --type secure \
  --encryption-password "my-secret-password"
echo

# Example 2: Encrypted notification with tags
echo "2. Encrypted notification with tags"
wirepusher send "Database Alert" "Connection failed: password mismatch" \
  --type alert \
  --tag production \
  --tag database \
  --encryption-password "db-alert-password"
echo

# Example 3: Encrypted notification with image
echo "3. Encrypted notification with image"
wirepusher send "Security Report" "Unauthorized access attempt detected" \
  --type security \
  --image https://example.com/security-icon.png \
  --encryption-password "security-password"
echo

# Example 4: Encrypted notification with action URL
echo "4. Encrypted notification with action URL"
wirepusher send "Password Reset" "Your password reset link is ready" \
  --type account \
  --action https://example.com/reset-password \
  --encryption-password "account-password"
echo

# Example 5: Pipe sensitive data with encryption
echo "5. Piping sensitive data with encryption"
echo "API Key: abc123xyz456" | wirepusher send "API Credentials" --stdin \
  --type credentials \
  --encryption-password "credentials-password"
echo

# Example 6: Environment variable for password (more secure)
echo "6. Using environment variable for password"
export ENCRYPTION_PASSWORD="env-secret-password"
wirepusher send "Deployment Credentials" "SSH key generated: $(date)" \
  --type deployment \
  --encryption-password "$ENCRYPTION_PASSWORD"
unset ENCRYPTION_PASSWORD
echo

# Example 7: Multiple notifications with same password
echo "7. Sending multiple encrypted notifications"
PASSWORD="batch-password"
for i in {1..3}; do
  wirepusher send "Batch Update $i" "Processing batch item $i with encrypted data" \
    --type batch \
    --encryption-password "$PASSWORD"
  sleep 1
done
echo

echo "=== Examples Complete ==="
echo
echo "Important notes:"
echo "- Only the message body is encrypted"
echo "- Title, type, tags, images, and action URLs remain unencrypted"
echo "- Encryption password must match the type configuration in WirePusher app"
echo "- Passwords are never sent to the API (used only for local encryption)"
echo "- Compatible with all WirePusher SDKs (Python, JavaScript, Go, Java, C#, PHP, Rust)"
