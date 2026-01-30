# CI/CD Setup Guide

Setup guide for Pincho CLI's CI/CD pipeline using Google Cloud Build and GoReleaser with GitLab releases.

## Overview

The CI/CD pipeline is designed to:

1. **On every push**: Run tests, linting, and build verification
2. **On Git tags (v*.*.*)**: Build multi-platform binaries and create GitLab release

## Prerequisites

- Google Cloud Build triggers configured (via Terraform)
- GitLab repository: `https://gitlab.com/pincho/pincho-cli`
- GitLab Personal Access Token with `api` scope

## Pipeline Steps

### 1. Download Dependencies
Downloads Go modules using `go mod download`.

### 2. Verify Dependencies
Verifies module checksums with `go mod verify`.

### 3. Format Check
Ensures all Go code is properly formatted using `gofmt -l .`.

### 4. Static Analysis
Runs `go vet ./...` to catch common issues.

### 5. Tests with Coverage
- Runs all tests with race detector
- Enforces **30% minimum overall coverage**
- Enforces **55% minimum coverage for pkg/ layer**
- Generates coverage reports

### 6. Multi-Platform Build Verification
Verifies successful builds for:
- Linux (amd64, arm64, arm v7)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

### 7. GoReleaser (Tags Only)
On Git tags matching `v*.*.*`:
- Validates semantic version format
- Builds binaries for all 7 platforms
- Creates GitLab release with assets
- Generates checksums

### 8. Release Notification (Tags Only)
Sends push notification via Pincho API on successful release:
- Uses `/send` endpoint with structured message
- Includes release version, platform count, and link to GitLab release
- Dogfooding: We use our own product for CI/CD notifications

## Setup Steps

### 1. Verify GitLab Token Secret

The GitLab Personal Access Token is already configured in Secret Manager as `gitlab-api-token-cloudbuild`. This secret is managed by Terraform in the main frontend repository and Cloud Build service account already has access.

**Verify the secret exists**:
```bash
# Check secret exists (requires Secret Manager access)
gcloud secrets describe gitlab-api-token-cloudbuild --project=pincho-dev

# Cloud Build access is already granted via Terraform
# See: frontend/terraform/modules/cloudbuild/main.tf
```

The cloudbuild.yaml references this secret as:
```yaml
availableSecrets:
  secretManager:
    - versionName: projects/$PROJECT_ID/secrets/gitlab-api-token-cloudbuild/versions/latest
      env: 'GITLAB_TOKEN'
```

### 2. Verify Pincho Token Secret

The Pincho API token for release notifications is stored in Secret Manager as `pincho-token`. This enables dogfooding - we use our own product for CI/CD notifications.

**Verify the secret exists**:
```bash
gcloud secrets describe pincho-token --project=pincho-app-dev
```

Cloud Build access is granted via Terraform in `frontend/terraform/modules/cloudbuild/main.tf`:
```hcl
locals {
  secrets = [
    # ... other secrets ...
    "pincho-token"
  ]
}
```

### 3. Verify Cloud Build Triggers

The triggers should already exist from Terraform:

```bash
# List triggers
gcloud builds triggers list --project=pincho-dev --region=us-central1

# Expected triggers:
# - Main branch trigger: Runs on push to main (tests only)
# - Release trigger: Runs on tags matching v*.*.* (full release)
```

If triggers don't exist, they can be created via Terraform or manually in GCP Console.

## Testing the Pipeline

### Test 1: Main Branch (Tests Only)

```bash
# Make a small change
echo "# Test" >> README.md
git add README.md
git commit -m "test: Trigger Cloud Build on main branch"
git push origin main
```

**Expected**: Cloud Build runs tests, format check, vet, and build verification. No release created.

Monitor: https://console.cloud.google.com/cloud-build/builds?project=pincho-dev

### Test 2: Release Tag (Full Pipeline)

```bash
# Create a test release tag
git tag v0.1.0-alpha.1
git push origin v0.1.0-alpha.1
```

**Expected**:
1. Cloud Build runs all tests
2. GoReleaser builds 7 platform binaries
3. GitLab release created at: https://gitlab.com/pincho/pincho-cli/-/releases

### Cleanup Test Release

```bash
# Delete local tag
git tag -d v0.1.0-alpha.1

# Delete remote tag
git push origin :refs/tags/v0.1.0-alpha.1
```

Delete GitLab release manually if needed.

## Release Process (For Maintainers)

See [CONTRIBUTING.md](../CONTRIBUTING.md#releasing-for-maintainers) for the full release checklist.

Quick reference:

```bash
# 1. Update CHANGELOG.md
# 2. Commit changes
git add -A
git commit -m "chore: Prepare for v1.0.0 release"
git push origin main

# 3. Create and push tag
git tag v1.0.0
git push origin v1.0.0

# 4. Monitor Cloud Build
# 5. Verify GitLab release
```

## Troubleshooting

### Build Fails: Coverage Below Threshold

The pipeline enforces 30% overall and 55% pkg layer coverage:

```bash
# Check current coverage locally
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total

# Check pkg layer coverage
go test -coverprofile=pkg_coverage.out ./pkg/...
go tool cover -func=pkg_coverage.out | grep total

# Identify uncovered code
go tool cover -html=coverage.out
```

### GoReleaser Fails: Missing GITLAB_TOKEN

1. Verify secret exists:
   ```bash
   gcloud secrets describe gitlab-api-token-cloudbuild --project=pincho-dev
   ```

2. Verify Cloud Build has access (should be via Terraform):
   ```bash
   gcloud secrets get-iam-policy gitlab-api-token-cloudbuild --project=pincho-dev
   ```

3. Check secret reference in `cloudbuild.yaml`:
   ```yaml
   availableSecrets:
     secretManager:
       - versionName: projects/$PROJECT_ID/secrets/gitlab-api-token-cloudbuild/versions/latest
         env: 'GITLAB_TOKEN'
   ```

### GoReleaser Fails: Token Permissions

The GitLab token must have `api` scope for creating releases. The token is managed by Terraform - check the main frontend repository's terraform configuration if permissions need updating.

### Build Fails: Format Check

```bash
# Fix formatting locally
gofmt -w .
```

### Build Fails: Static Analysis (go vet)

```bash
# Run vet locally to see issues
go vet ./...
```

## Architecture Diagram

```
GitLab Push
     |
     v
Cloud Build Trigger
     |
     v
┌─────────────────────┐
│ Download & Verify   │
│ Dependencies        │
└─────────────────────┘
     |
     v
┌─────────────────────┐
│ Format Check        │
│ Static Analysis     │
└─────────────────────┘
     |
     v
┌─────────────────────┐
│ Tests with Coverage │
│ (30% overall,       │
│  55% pkg layer)     │
└─────────────────────┘
     |
     v
┌─────────────────────┐
│ Multi-Platform      │
│ Build Verification  │
└─────────────────────┘
     |
     v (if tag v*.*.*)
┌─────────────────────┐
│ GoReleaser          │
│ → Build binaries    │
│ → Create release    │
│ → Upload assets     │
└─────────────────────┘
     |
     v
GitLab Releases
```

## Security Considerations

- **GitLab token**: Stored in Secret Manager, not in code
- **Token rotation**: Rotate annually or as needed
- **Minimal permissions**: Token only has `api` scope
- **No secrets in logs**: Cloud Build masks secret values

## Related Files

- `cloudbuild.yaml` - Cloud Build configuration
- `.goreleaser.yml` - GoReleaser configuration for multi-platform builds and GitLab releases
- `CONTRIBUTING.md` - Release process documentation
- `CHANGELOG.md` - Version history

## Future Enhancements

- Homebrew tap distribution (requires GitHub mirror)
- Scoop bucket for Windows package manager
- Snap package for Linux

## References

- [Google Cloud Build Documentation](https://cloud.google.com/build/docs)
- [GoReleaser Documentation](https://goreleaser.com/)
- [GoReleaser GitLab Configuration](https://goreleaser.com/customization/release/#gitlab)
- [GitLab Personal Access Tokens](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html)
