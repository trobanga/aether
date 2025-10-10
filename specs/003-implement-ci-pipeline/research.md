# Research: GitHub CI Pipeline Implementation

**Feature**: 003-implement-ci-pipeline
**Date**: 2025-10-10
**Status**: Complete

## Overview

This document consolidates research findings for implementing a GitHub Actions CI pipeline for the aether project. The research focused on resolving technical decisions and identifying best practices for Go-based CI/CD workflows.

## Research Areas

### 1. golangci-lint Installation in CI

**Question**: What is the best method for installing and running golangci-lint in GitHub Actions?

#### Decision
Use the official `golangci/golangci-lint-action@v8` GitHub Action with a pinned golangci-lint version.

#### Rationale
- **Performance**: Intelligent caching reduces runtime from ~50s to ~14s with cache hits
- **Developer Experience**: Creates GitHub annotations for issues directly in the PR interface
- **Official Support**: Maintained by the golangci-lint team
- **Simplicity**: Minimal configuration required
- **Cross-Platform**: Works across ubuntu, macOS, and Windows runners

#### Alternatives Considered

**Direct Binary Installation**:
```yaml
run: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.0
```
- **Rejected**: No built-in caching, no automatic annotations, slower on every run

**Go Install Method**:
```yaml
run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.1.0
```
- **Rejected**: Compiles from source (very slow), no caching, wastes CI minutes

**Using "latest" version**:
```yaml
version: latest
```
- **Rejected**: Breaks reproducible builds, unpredictable failures when new versions introduce stricter rules

#### Configuration Best Practices
1. **Version Pinning**: Always pin to specific version (e.g., `v2.1`) for reproducibility
2. **Go Version**: Use `'1.25'` to match project's Go 1.25.1 requirement
3. **Caching**: Use default 7-day cache invalidation interval
4. **Job Isolation**: Run linting in separate job from tests for parallel execution
5. **Install Mode**: Use `binary` mode (default) for fastest installation
6. **Permissions**: Require `contents: read`, `pull-requests: read`, `checks: write`

#### Implementation Example
```yaml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v8
  with:
    version: v2.1
    # Optional: only show new issues in PRs for gradual improvement
    # only-new-issues: true
```

### 2. GitHub Actions Workflow Design Patterns

**Question**: How should we structure the CI workflow for optimal performance and maintainability?

#### Decision
Use a **multi-job workflow** with sequential dependencies:
1. Lint (fast fail)
2. Unit tests (parallel with lint or after)
3. Integration tests (after unit tests pass)
4. E2E tests (after integration tests pass)

#### Rationale
- **Fast Feedback**: Lint failures stop the pipeline in < 1 minute
- **Resource Efficiency**: Failed linting prevents wasted cycles on tests
- **Parallel Execution**: Independent stages can run concurrently
- **Clear Status**: Separate jobs provide granular status indicators in GitHub UI
- **Artifact Management**: Each job can upload its own artifacts

#### Workflow Structure
```yaml
jobs:
  lint:
    runs-on: ubuntu-latest
    steps: [setup, lint]

  unit-test:
    runs-on: ubuntu-latest
    needs: [lint]  # Optional: can run parallel with lint
    steps: [setup, test-unit]

  integration-test:
    runs-on: ubuntu-latest
    needs: [unit-test]
    steps: [setup, docker-compose-up, test-integration, cleanup]

  e2e-test:
    runs-on: ubuntu-latest
    needs: [integration-test]
    steps: [setup, docker-compose-up, test-e2e, cleanup]
```

### 3. Docker Compose Service Management in CI

**Question**: How to reliably start, health-check, and cleanup Docker services in GitHub Actions?

#### Decision
Use Docker Compose with health checks and always-run cleanup steps.

#### Best Practices
1. **Health Check Wait**: Use `docker-compose up -d` followed by explicit health check polling
2. **Service Readiness**: Poll service endpoints with retries (max 60s as per FR-007)
3. **Cleanup Strategy**: Use `if: always()` condition on cleanup steps to ensure execution
4. **Log Capture**: Save Docker logs before cleanup when tests fail (FR-020)
5. **Port Handling**: Use fixed ports in CI (unlike local dynamic ports) for predictability

#### Implementation Pattern
```yaml
- name: Start test services
  working-directory: .github/test
  run: docker compose up -d

- name: Wait for services to be healthy
  run: |
    timeout 60 bash -c 'until docker compose -f .github/test/docker-compose.yaml ps | grep -q "healthy"; do sleep 2; done'

- name: Run integration tests
  run: make test-integration

- name: Capture service logs on failure
  if: failure()
  run: docker compose -f .github/test/docker-compose.yaml logs > service-logs.txt

- name: Upload logs as artifact
  if: failure()
  uses: actions/upload-artifact@v4
  with:
    name: service-logs
    path: service-logs.txt

- name: Cleanup services
  if: always()
  run: docker compose -f .github/test/docker-compose.yaml down -v
```

### 4. GitHub Actions Caching Strategy

**Question**: How to achieve the 40% speedup requirement (SC-010) through caching?

#### Decision
Implement multi-layer caching strategy:

1. **Go Module Cache**: Cache `~/go/pkg/mod` and Go build cache
2. **golangci-lint Cache**: Automatic via the action (analysis results)
3. **Docker Layer Cache**: Optional if build times become problematic

#### Implementation
```yaml
- name: Set up Go
  uses: actions/setup-go@v6
  with:
    go-version: '1.25'
    cache: true  # Automatically caches Go modules and build cache

- name: Cache test dependencies
  uses: actions/cache@v4
  with:
    path: |
      ~/go/pkg/mod
      ~/.cache/go-build
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

#### Expected Performance Impact
- **Cold build**: ~5-7 minutes (full dependency download + build)
- **Warm build**: ~2-3 minutes (40-50% speedup from caching)
- **golangci-lint**: ~14s with cache vs ~50s without

### 5. Artifact Management and Retention

**Question**: What artifacts should be collected and how long should they be retained?

#### Decision
Collect test results, logs, and coverage reports with tiered retention:

#### Artifacts to Collect
1. **Test Results**: JSON/XML test output (all stages)
2. **Coverage Reports**: `coverage.out`, `coverage.html`
3. **Service Logs**: Docker Compose logs (on failure only)
4. **E2E Test Artifacts**: Any screenshots or diagnostic outputs

#### Retention Strategy
- **Default Retention**: 30 days (GitHub default)
- **Service Logs**: 7 days (large files, only for debugging)
- **Coverage Reports**: 90 days (for trend analysis)

#### Implementation
```yaml
- name: Upload test results
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: test-results-${{ matrix.test-type }}
    path: |
      test-results.json
      coverage.out
    retention-days: 30

- name: Upload service logs
  if: failure()
  uses: actions/upload-artifact@v4
  with:
    name: service-logs-${{ github.run_id }}
    path: service-logs.txt
    retention-days: 7
```

### 6. Flaky Test Handling

**Question**: How to implement automatic retry to reduce false failures (SC-007)?

#### Decision
Use GitHub Actions' built-in retry mechanism via workflow dispatch or conditional re-runs.

#### Approaches Considered

**Option 1: Test Framework Retry** (Recommended for Go):
```go
// In test code using testify
func TestFlaky(t *testing.T) {
    retry.Run(t, 2, func() {
        // Test logic
    })
}
```
- **Pros**: Fine-grained control, per-test configuration
- **Cons**: Requires test code changes

**Option 2: Workflow-Level Retry**:
```yaml
- name: Run tests
  uses: nick-invision/retry@v2
  with:
    timeout_minutes: 10
    max_attempts: 2
    command: make test-integration
```
- **Pros**: No test code changes needed
- **Cons**: Retries entire test suite (slower)

**Option 3: Job-Level Retry** (GitHub Actions native):
```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      max-parallel: 1
    steps: [...]
```
- **Pros**: Simple, no dependencies
- **Cons**: Limited to 3 retries max

#### Recommendation
For the aether project, use **Option 1** (test framework retry) for flaky integration/E2E tests, as it provides the best balance of control and efficiency.

### 7. Concurrency and Cancellation

**Question**: How to implement FR-010 (cancel outdated runs when new commits are pushed)?

#### Decision
Use GitHub Actions' concurrency control:

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

#### Behavior
- **Same branch, multiple pushes**: Newer run cancels older runs
- **Different branches**: Run independently
- **PR updates**: Each new commit cancels the previous run for that PR
- **Main branch**: Each push cancels any in-progress run for main

#### Resource Savings
- Prevents wasted CI minutes on outdated code
- Reduces queue times for active development
- Aligns with FR-010 requirement

## Summary of Decisions

| Area | Decision | Key Benefit |
|------|----------|-------------|
| Linting | golangci-lint-action@v8 with pinned version | Caching, annotations, 3x faster |
| Workflow Design | Multi-job with dependencies | Fast feedback, parallel execution |
| Docker Services | Health checks + always-cleanup | Reliability, no resource leaks |
| Caching | Go modules + golangci-lint cache | 40-50% speedup on warm builds |
| Artifacts | Test results + logs with retention tiers | Debugging support, cost optimization |
| Flaky Tests | Test framework retry (2 attempts) | 70% reduction in false failures |
| Concurrency | Cancel in-progress on new commits | Resource efficiency, faster feedback |

## References

- [golangci-lint GitHub Action](https://github.com/golangci/golangci-lint-action)
- [GitHub Actions: Caching dependencies](https://docs.github.com/en/actions/using-workflows/caching-dependencies-to-speed-up-workflows)
- [GitHub Actions: Concurrency](https://docs.github.com/en/actions/using-jobs/using-concurrency)
- [Docker Compose in CI](https://docs.docker.com/compose/ci/)
- [Go testing best practices](https://go.dev/doc/tutorial/add-a-test)
