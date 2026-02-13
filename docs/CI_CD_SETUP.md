# CI/CD Setup Guide

Setup guide for Pincho CLI's CI/CD pipeline using GitHub Actions and GoReleaser with GitHub releases.

## Overview

The CI/CD pipeline is designed to:

1. **On every push**: Run tests, linting, and build verification
2. **On Git tags (v*.*.*)**: Build multi-platform binaries and create GitHub release

## Prerequisites

- Google Cloud Build triggers configured (via Terraform)
- GitHub repository: `https://github.com/Pincho-App/pincho-cli`
- GitHub Personal Access Token or `GITHUB_TOKEN` (provided by GitHub Actions)

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
- Creates GitHub release with assets
- Generates checksums

### 8. Release Notification (Tags Only)
Sends push notification via Pincho API on successful release:
- Uses `/send` endpoint with structured message
- Includes release version, platform count, and link to GitHub release
- Dogfooding: We use our own product for CI/CD notifications

## Setup Steps

### 1. Verify GitHub Token

GitHub Actions provides `GITHUB_TOKEN` automatically. No additional token configuration is needed for creating releases. The `goreleaser-action` uses `GITHUB_TOKEN` by default.

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
3. GitHub release created at: https://github.com/Pincho-App/pincho-cli/releases

### Cleanup Test Release

```bash
# Delete local tag
git tag -d v0.1.0-alpha.1

# Delete remote tag
git push origin :refs/tags/v0.1.0-alpha.1
```

Delete GitHub release manually if needed.

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
# 5. Verify GitHub release
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

### GoReleaser Fails: Missing GITHUB_TOKEN

The `GITHUB_TOKEN` is automatically provided by GitHub Actions. Ensure the workflow has `permissions: contents: write` set for the release job.

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
GitHub Push
     |
     v
GitHub Actions Trigger
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
GitHub Releases
```

## Security Considerations

- **GitHub token**: Provided automatically by GitHub Actions
- **Token rotation**: Rotate annually or as needed
- **Minimal permissions**: Token only has `api` scope
- **No secrets in logs**: Cloud Build masks secret values

## Related Files

- `cloudbuild.yaml` - Cloud Build configuration
- `.goreleaser.yml` - GoReleaser configuration for multi-platform builds and GitHub releases
- `CONTRIBUTING.md` - Release process documentation
- `CHANGELOG.md` - Version history

## Future Enhancements

- Homebrew tap distribution (requires GitHub mirror)
- Scoop bucket for Windows package manager
- Snap package for Linux

## References

- [Google Cloud Build Documentation](https://cloud.google.com/build/docs)
- [GoReleaser Documentation](https://goreleaser.com/)
- [GoReleaser GitHub Configuration](https://goreleaser.com/customization/release/#github)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
