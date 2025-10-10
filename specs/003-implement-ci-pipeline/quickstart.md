# Quickstart: GitHub CI Pipeline

**Feature**: 003-implement-ci-pipeline
**Audience**: Developers implementing the CI pipeline
**Time to Complete**: ~30 minutes for basic setup, 1-2 hours for full implementation with testing

## Overview

This quickstart guide walks you through implementing the GitHub Actions CI pipeline for the aether project. You'll create a workflow that automatically runs linting, unit tests, integration tests, and E2E tests on every push and pull request.

## Prerequisites

✅ **Before you begin**:
- [ ] Feature branch `003-implement-ci-pipeline` checked out
- [ ] Existing Makefile with `lint`, `test-unit`, `test-integration` targets
- [ ] Docker Compose test infrastructure in `.github/test/`
- [ ] E2E test script at `.github/test/test-dimp.sh`
- [ ] golangci-lint installed locally (for testing): `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

## Step-by-Step Implementation

### Step 1: Create Workflow Directory

```bash
mkdir -p .github/workflows
```

### Step 2: Create Basic Workflow File

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: ['**']
  pull_request:
    types: [opened, synchronize, reopened]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

permissions:
  contents: read
  pull-requests: read
  checks: write

jobs:
  # Jobs will be added in steps 3-6
```

**What this does**:
- Triggers on all pushes and PR events (FR-001, FR-002)
- Cancels outdated runs for the same branch (FR-010)
- Sets minimal required permissions

### Step 3: Add Lint Job

Add this job to `ci.yml`:

```yaml
jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    timeout-minutes: 5
    steps:
      - name: Checkout code
        uses: actions/checkout@v5

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version: '1.25'
          cache: true

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v8
        with:
          version: v2.1
          # Optional: only show new issues in PRs
          # only-new-issues: true

      - name: Upload lint results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: lint-results-${{ github.run_id }}
          path: |
            golangci-lint-report.xml
          retention-days: 30
          if-no-files-found: ignore
```

**Test it**:
```bash
# Commit and push to trigger CI
git add .github/workflows/ci.yml
git commit -m "feat: add lint job to CI pipeline"
git push origin 003-implement-ci-pipeline

# Watch the workflow run
gh run watch
```

### Step 4: Add Unit Test Job

Add after the lint job:

```yaml
  unit-test:
    name: Unit Tests
    runs-on: ubuntu-latest
    needs: [lint]
    timeout-minutes: 10
    steps:
      - name: Checkout code
        uses: actions/checkout@v5

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version: '1.25'
          cache: true

      - name: Run unit tests
        run: make test-unit

      - name: Generate coverage report
        run: make coverage

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: unit-test-results-${{ github.run_id }}
          path: |
            coverage.out
            coverage.html
          retention-days: 90
```

**Test it**:
```bash
# Ensure unit tests pass locally first
make test-unit

# Commit and push
git add .github/workflows/ci.yml
git commit -m "feat: add unit test job to CI pipeline"
git push
```

### Step 5: Add Integration Test Job

Add after the unit test job:

```yaml
  integration-test:
    name: Integration Tests
    runs-on: ubuntu-latest
    needs: [unit-test]
    timeout-minutes: 20
    steps:
      - name: Checkout code
        uses: actions/checkout@v5

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version: '1.25'
          cache: true

      - name: Start Docker services
        working-directory: .github/test
        run: docker compose up -d

      - name: Wait for services to be healthy
        run: |
          timeout 60 bash -c 'until docker compose -f .github/test/docker-compose.yaml ps | grep -q "healthy"; do
            echo "Waiting for services to be healthy..."
            sleep 2
          done' || {
            echo "Services failed to become healthy within 60 seconds"
            docker compose -f .github/test/docker-compose.yaml ps
            docker compose -f .github/test/docker-compose.yaml logs
            exit 2
          }

      - name: Run integration tests
        run: make test-integration

      - name: Capture service logs on failure
        if: failure()
        run: |
          docker compose -f .github/test/docker-compose.yaml logs > docker-logs-integration.txt

      - name: Upload integration test results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: integration-test-results-${{ github.run_id }}
          path: |
            test-results-integration.json
          retention-days: 30
          if-no-files-found: ignore

      - name: Upload service logs
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: docker-logs-integration-${{ github.run_id }}
          path: docker-logs-integration.txt
          retention-days: 7

      - name: Cleanup Docker services
        if: always()
        run: docker compose -f .github/test/docker-compose.yaml down -v
```

**Test it**:
```bash
# Test Docker services locally
cd .github/test
docker compose up -d
docker compose ps  # Should show healthy services
docker compose down -v
cd ../..

# Commit and push
git add .github/workflows/ci.yml
git commit -m "feat: add integration test job with Docker services"
git push
```

### Step 6: Add E2E Test Job

Add after the integration test job:

```yaml
  e2e-test:
    name: E2E Tests
    runs-on: ubuntu-latest
    needs: [integration-test]
    timeout-minutes: 30
    steps:
      - name: Checkout code
        uses: actions/checkout@v5

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version: '1.25'
          cache: true

      - name: Build aether binary
        run: make build

      - name: Start Docker services
        working-directory: .github/test
        run: docker compose up -d

      - name: Wait for services to be healthy
        run: |
          timeout 60 bash -c 'until docker compose -f .github/test/docker-compose.yaml ps | grep -q "healthy"; do
            echo "Waiting for services to be healthy..."
            sleep 2
          done' || {
            echo "Services failed to become healthy within 60 seconds"
            docker compose -f .github/test/docker-compose.yaml ps
            docker compose -f .github/test/docker-compose.yaml logs
            exit 2
          }

      - name: Run E2E tests
        run: |
          chmod +x .github/test/test-dimp.sh
          .github/test/test-dimp.sh

      - name: Capture service logs on failure
        if: failure()
        run: |
          docker compose -f .github/test/docker-compose.yaml logs > docker-logs-e2e.txt

      - name: Upload E2E test results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: e2e-test-results-${{ github.run_id }}
          path: |
            e2e-results.txt
          retention-days: 30
          if-no-files-found: ignore

      - name: Upload service logs
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: docker-logs-e2e-${{ github.run_id }}
          path: docker-logs-e2e.txt
          retention-days: 7

      - name: Cleanup Docker services
        if: always()
        run: docker compose -f .github/test/docker-compose.yaml down -v
```

**Test it**:
```bash
# Test E2E script locally
make build
cd .github/test
docker compose up -d
./test-dimp.sh
docker compose down -v
cd ../..

# Commit and push
git add .github/workflows/ci.yml
git commit -m "feat: add E2E test job to CI pipeline"
git push
```

### Step 7: Enable Branch Protection

Configure branch protection rules for `main` branch:

**Via GitHub UI**:
1. Go to repository Settings → Branches
2. Add rule for `main` branch
3. Enable "Require status checks to pass before merging"
4. Select required checks:
   - `lint`
   - `unit-test`
   - `integration-test`
   - `e2e-test`
5. Enable "Require branches to be up to date before merging"
6. Save changes

**Via GitHub CLI**:
```bash
gh api repos/:owner/:repo/branches/main/protection \
  -X PUT \
  -f required_status_checks[strict]=true \
  -f required_status_checks[checks][][context]=lint \
  -f required_status_checks[checks][][context]=unit-test \
  -f required_status_checks[checks][][context]=integration-test \
  -f required_status_checks[checks][][context]=e2e-test
```

This enforces FR-011 (block merge on failure).

## Validation Checklist

After implementation, verify the pipeline works correctly:

### ✅ Lint Validation
- [ ] Push code with linting violation
- [ ] Verify lint job fails with specific violation details
- [ ] Verify PR shows "checks failed" status
- [ ] Fix violation and verify lint job passes

### ✅ Unit Test Validation
- [ ] Break a unit test
- [ ] Verify unit-test job fails with test details
- [ ] Verify coverage report is uploaded
- [ ] Fix test and verify job passes

### ✅ Integration Test Validation
- [ ] Verify Docker services start successfully
- [ ] Verify health checks pass within 60 seconds
- [ ] Break integration test
- [ ] Verify service logs are captured and uploaded
- [ ] Verify cleanup runs even on failure
- [ ] Check `docker ps` - no orphaned containers

### ✅ E2E Test Validation
- [ ] Verify E2E test script executes
- [ ] Verify pipeline workflow completes end-to-end
- [ ] Break E2E test
- [ ] Verify diagnostics and logs are captured
- [ ] Verify cleanup runs

### ✅ Performance Validation
- [ ] First run (cold cache): Record total time
- [ ] Second run (warm cache): Should be 40%+ faster
- [ ] Lint + unit should complete within 3 minutes

### ✅ Concurrency Validation
- [ ] Push multiple commits rapidly to same branch
- [ ] Verify only latest commit's run completes
- [ ] Verify earlier runs are cancelled

### ✅ Branch Protection Validation
- [ ] Create PR with failing tests
- [ ] Verify merge button is disabled
- [ ] Fix tests
- [ ] Verify merge button is enabled

## Troubleshooting

### Issue: golangci-lint fails with "no Go files to analyze"

**Cause**: golangci-lint can't find Go files in the working directory

**Solution**: Ensure checkout step runs before lint step, or specify paths:
```yaml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v8
  with:
    version: v2.1
    args: ./...  # Explicitly specify paths
```

### Issue: Docker services fail health checks

**Cause**: Services taking longer than expected to start, or health check misconfigured

**Solution**:
1. Check `.github/test/docker-compose.yaml` health check configuration
2. Increase timeout if needed (but keep under 60s per FR-007)
3. Verify services work locally with same compose file

### Issue: Integration tests pass locally but fail in CI

**Cause**: Port conflicts, network issues, or timing differences

**Solution**:
1. Check service logs: `docker compose logs`
2. Verify network connectivity between containers
3. Ensure tests don't depend on timing assumptions
4. Add retries for flaky network calls

### Issue: Cleanup doesn't run and containers are orphaned

**Cause**: Missing `if: always()` condition on cleanup step

**Solution**: Ensure cleanup step has:
```yaml
- name: Cleanup Docker services
  if: always()  # ← This is required
  run: docker compose -f .github/test/docker-compose.yaml down -v
```

### Issue: Artifacts not uploaded

**Cause**: File path doesn't exist or incorrect pattern

**Solution**:
1. Verify files are created before upload step
2. Use `if-no-files-found: ignore` to prevent failure
3. Check artifact path matches actual file location

### Issue: Cache not improving performance

**Cause**: Cache key changing on every run, or cache not being saved

**Solution**:
1. Verify cache key uses `hashFiles('**/go.sum')`
2. Ensure `go.sum` isn't changing
3. Check cache hit/miss in job logs
4. Verify `cache: true` is set in setup-go step

## Next Steps

After completing this quickstart:

1. **Monitor CI Performance**: Track pipeline execution times and optimize as needed
2. **Add Coverage Tracking**: Integrate coverage reports with codecov.io or similar
3. **Expand E2E Tests**: Add more comprehensive workflow scenarios
4. **Add Notifications**: Configure Slack/email notifications for main branch failures
5. **Document for Team**: Share this guide with team members

## Quick Reference

### Useful Commands

```bash
# Watch live CI run
gh run watch

# List recent runs
gh run list

# View run details
gh run view <run-id>

# Download artifacts
gh run download <run-id>

# Rerun failed jobs
gh run rerun <run-id> --failed

# Trigger workflow manually (if configured)
gh workflow run ci.yml
```

### File Locations

- Workflow: `.github/workflows/ci.yml`
- Docker services: `.github/test/docker-compose.yaml`
- E2E script: `.github/test/test-dimp.sh`
- Makefile: `./Makefile`
- Test directories: `tests/unit/`, `tests/integration/`, `tests/contract/`

### Key Configuration Values

| Setting | Value | Requirement |
|---------|-------|-------------|
| Lint timeout | 5 min | FR-013 |
| Unit test timeout | 10 min | FR-013 |
| Integration timeout | 20 min | FR-013 |
| E2E timeout | 30 min | FR-013 |
| Service health check max wait | 60s | FR-007 |
| Go version | 1.25 | FR-015 |
| golangci-lint version | v2.1 | Research decision |
| Runner OS | ubuntu-latest | FR-014 |

## Support

- Review [research.md](research.md) for technical decisions and rationale
- Check [workflow-contract.md](contracts/workflow-contract.md) for detailed behavior specifications
- See [data-model.md](data-model.md) for workflow entity relationships
- Consult [plan.md](plan.md) for overall implementation strategy
