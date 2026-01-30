#!/bin/bash
# CI/CD pipeline notification example

# Set credentials from environment variables
# PINCHO_TOKEN should be set in CI/CD secrets

# Notify on build start
pincho send "Build Started" "Pipeline #$CI_PIPELINE_ID started on $CI_COMMIT_BRANCH" \
  --type info \
  --tag ci

# Notify on build success
if [ $? -eq 0 ]; then
  pincho send "Build Passed" "Pipeline #$CI_PIPELINE_ID completed successfully" \
    --type success \
    --tag ci \
    --action-url "https://gitlab.com/your-project/-/pipelines/$CI_PIPELINE_ID"
else
  pincho send "Build Failed" "Pipeline #$CI_PIPELINE_ID failed" \
    --type alert \
    --tag ci \
    --action-url "https://gitlab.com/your-project/-/pipelines/$CI_PIPELINE_ID"
fi
