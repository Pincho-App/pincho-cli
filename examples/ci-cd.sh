#!/bin/bash
# CI/CD pipeline notification example

# Set credentials from environment variables
# WIREPUSHER_TOKEN and WIREPUSHER_ID should be set in CI/CD secrets

# Notify on build start
wirepusher send "Build Started" "Pipeline #$CI_PIPELINE_ID started on $CI_COMMIT_BRANCH" \
  --type info \
  --tag ci

# Notify on build success
if [ $? -eq 0 ]; then
  wirepusher send "Build Passed" "Pipeline #$CI_PIPELINE_ID completed successfully" \
    --type success \
    --tag ci \
    --action "https://gitlab.com/your-project/-/pipelines/$CI_PIPELINE_ID"
else
  wirepusher send "Build Failed" "Pipeline #$CI_PIPELINE_ID failed" \
    --type alert \
    --tag ci \
    --action "https://gitlab.com/your-project/-/pipelines/$CI_PIPELINE_ID"
fi
