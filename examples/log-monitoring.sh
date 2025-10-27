#!/bin/bash
# Log monitoring example - send notification when errors are detected

# Monitor log file for errors
tail -f /var/log/app.log | grep --line-buffered "ERROR" | while read line
do
  echo "$line" | wirepusher send "Error Detected" --stdin --type alert --tag monitoring
done

# Alternative: Check for critical errors once
if grep -q "CRITICAL" /var/log/app.log; then
  grep "CRITICAL" /var/log/app.log | tail -1 | \
    wirepusher send "Critical Error" --stdin --type alert --tag critical
fi
