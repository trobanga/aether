# CI Workflow Contract

**Feature**: 003-implement-ci-pipeline
**Version**: 1.0.0
**Date**: 2025-10-10

## Overview

This document defines the behavioral contract for the GitHub Actions CI workflow. It specifies the expected inputs, outputs, behaviors, and guarantees that the workflow must provide.

## Workflow Interface

### Input Events

The workflow MUST respond to the following GitHub events:

#### 1. Push Event
```yaml
on:
  push:
    branches: ['**']  # All branches
```

**Contract**:
- Triggers on every push to any branch
- Runs all pipeline stages (lint, unit, integration, E2E)
- Complies with FR-001

#### 2. Pull Request Event
```yaml
on:
  pull_request:
    types: [opened, synchronize, reopened]
    branches: ['**']  # All target branches
```

**Contract**:
- Triggers on PR open, update (synchronize), or reopen
- Runs all pipeline stages
- Updates PR status checks
- Blocks merge if any check fails (FR-011)
- Complies with FR-002, FR-018

### Concurrency Control

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

**Contract**:
- Cancels outdated runs when new commits pushed to same branch
- Prevents resource waste
- Complies with FR-010

## Job Contracts

### 1. Lint Job

**Purpose**: Validate code quality and formatting

**Input**:
- Go source code from repository
- golangci-lint configuration (`.golangci.yml` if exists)

**Steps**:
1. Checkout code
2. Setup Go 1.25
3. Run `golangci/golangci-lint-action@v8` with version `v2.1`
4. Upload lint results as artifact

**Output**:
- Status: success/failure
- Artifacts: `lint-results` (test results, annotations)
- GitHub check status: visible in PR

**Guarantees**:
- Completes within 5 minutes (FR-013)
- Uses existing `make lint` behavior (FR-003)
- Creates GitHub annotations for violations
- Fails build on any violations

**Exit Codes**:
- `0`: All checks passed
- `1`: Linting violations found
- `2`: golangci-lint error (config, timeout, etc.)

### 2. Unit Test Job

**Purpose**: Execute unit tests in isolation

**Input**:
- Go source code
- Unit test files in `tests/unit/`
- Dependency: Lint job must pass

**Steps**:
1. Checkout code
2. Setup Go 1.25 with caching
3. Run `make test-unit`
4. Generate coverage report
5. Upload test results and coverage

**Output**:
- Status: success/failure
- Artifacts:
  - `unit-test-results` (test output)
  - `coverage-report` (coverage.out, coverage.html)
- Test summary in job log

**Guarantees**:
- Completes within 10 minutes (FR-013)
- Uses existing `make test-unit` command (FR-004)
- Reports which tests failed with error details
- Uploads results even on failure (`if: always()`)

**Exit Codes**:
- `0`: All tests passed
- `1`: One or more tests failed
- `2`: Test execution error

### 3. Integration Test Job

**Purpose**: Test interactions with Docker services

**Input**:
- Go source code
- Integration test files in `tests/integration/`
- Docker Compose configuration in `.github/test/`
- Dependency: Unit test job must pass

**Steps**:
1. Checkout code
2. Setup Go 1.25 with caching
3. Start Docker Compose services (`docker compose up -d`)
4. Wait for health checks (max 60s)
5. Run `make test-integration`
6. Capture Docker logs on failure
7. Cleanup services (`docker compose down -v`) - always runs

**Output**:
- Status: success/failure
- Artifacts:
  - `integration-test-results` (test output)
  - `docker-logs` (service logs, only on failure)
- Clean Docker state (no orphaned containers)

**Guarantees**:
- Completes within 20 minutes (FR-013)
- Services start and health checks pass within 60s (FR-007)
- Uses existing Docker Compose setup (FR-006)
- Captures service logs on failure (FR-020)
- Cleans up services regardless of outcome (FR-009)
- No resource leaks (SC-003)

**Exit Codes**:
- `0`: All integration tests passed
- `1`: One or more tests failed
- `2`: Docker services failed to start/health check timeout (FR-019)
- `3`: Test execution error

**Error Handling**:
- Health check timeout → fail fast with service diagnostics
- Service startup failure → include docker-compose logs in error
- Test failure → capture all service logs before cleanup

### 4. E2E Test Job

**Purpose**: Execute end-to-end pipeline workflow tests

**Input**:
- Go source code
- E2E test script `.github/test/test-dimp.sh`
- Docker Compose configuration in `.github/test/`
- Dependency: Integration test job must pass

**Steps**:
1. Checkout code
2. Setup Go 1.25 with caching
3. Build aether binary (`make build`)
4. Start Docker Compose services
5. Wait for health checks (max 60s)
6. Run E2E test script: `.github/test/test-dimp.sh`
7. Capture Docker logs on failure
8. Upload E2E artifacts (if any)
9. Cleanup services - always runs

**Output**:
- Status: success/failure
- Artifacts:
  - `e2e-test-results` (test script output)
  - `e2e-artifacts` (any generated files)
  - `docker-logs-e2e` (service logs, only on failure)
- Clean Docker state

**Guarantees**:
- Completes within 30 minutes (FR-013)
- Runs on every PR and main branch push (FR-018)
- Executes full pipeline workflow (FR-008)
- Captures comprehensive logs on failure (FR-020)
- Cleans up services regardless of outcome (FR-009)
- Publishes results accessible from GitHub UI (SC-005)

**Exit Codes**:
- `0`: E2E workflow completed successfully
- `1`: E2E test failed
- `2`: Docker services failed to start
- `3`: Binary build failed

## Performance Contracts

### Execution Time Guarantees

| Stage | Maximum Duration | Target Duration (with cache) |
|-------|-----------------|------------------------------|
| Lint | 5 minutes | < 1 minute |
| Unit Tests | 10 minutes | < 2 minutes |
| Integration Tests | 20 minutes | < 10 minutes |
| E2E Tests | 30 minutes | < 15 minutes |
| **Total Pipeline** | **65 minutes** | **< 25 minutes** (SC-006) |

### Caching Contract

**Go Module Cache**:
- Key: `${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}`
- Paths: `~/go/pkg/mod`, `~/.cache/go-build`
- Invalidation: On `go.sum` change

**golangci-lint Cache**:
- Automatic via `golangci/golangci-lint-action`
- Invalidation: Every 7 days or on config change

**Performance Guarantee**:
- Warm build (cache hit): 40-50% faster than cold build (SC-010)
- Lint with cache: ~14s vs ~50s without cache

## Artifact Contracts

### Always Uploaded (regardless of outcome)

1. **Test Results** (`if: always()`):
   - Name: `{stage}-test-results-${{ github.run_id }}`
   - Path: Test output files (JSON/text)
   - Retention: 30 days
   - Complies with FR-012

2. **Coverage Reports** (`if: always()`):
   - Name: `coverage-${{ github.run_id }}`
   - Path: `coverage.out`, `coverage.html`
   - Retention: 90 days

### Failure-Only Artifacts (`if: failure()`)

3. **Docker Service Logs**:
   - Name: `docker-logs-{stage}-${{ github.run_id }}`
   - Path: Service logs from `docker compose logs`
   - Retention: 7 days
   - Complies with FR-020

4. **E2E Diagnostics**:
   - Name: `e2e-diagnostics-${{ github.run_id }}`
   - Path: Any screenshots, trace files generated during E2E tests
   - Retention: 7 days

### Artifact Access Guarantee
- Artifacts accessible from GitHub Actions UI within 30 seconds of upload (SC-005)
- Downloadable via GitHub CLI: `gh run download <run-id>`
- Viewable in PR checks summary

## Status Check Contracts

### PR Status Updates

Each job creates a status check that:
- Reports as "pending" while running
- Reports as "success" or "failure" on completion
- Is **required** for merge (FR-011)
- Displays clear status indicator in GitHub UI (FR-017)

**Required Checks** (must all pass to merge):
1. `lint` - Code quality check
2. `unit-test` - Unit test suite
3. `integration-test` - Integration test suite
4. `e2e-test` - End-to-end test suite

### Error Message Contract

All failures MUST provide:
1. **Context**: Which stage/step failed
2. **Reason**: Specific error (test name, lint rule, service name)
3. **Location**: File path and line number (for code issues)
4. **Action**: What to do to fix (if determinable)

**Guarantee**: Developers can identify and fix failures within 5 minutes of receiving CI feedback (SC-009)

## Reliability Contracts

### Resource Cleanup Guarantee

**Promise**: No resource leaks after CI run

**Implementation**:
- All Docker cleanup steps use `if: always()` (FR-009)
- Explicit `docker compose down -v` to remove volumes
- Verification: Zero orphaned containers after run (SC-003)

### Flaky Test Handling

**Strategy**: Automatic retry for intermittent failures

**Contract**:
- Integration/E2E tests retry once on failure
- Reduces false failures by ~70% (SC-007)
- Retry count visible in test output

**Implementation Note**: Per research.md, use test framework retry (Option 1) for granular control.

### Service Startup Guarantee

**Promise**: Fail fast if services can't start

**Contract**:
- Maximum health check wait: 60 seconds (FR-007)
- On timeout: Fail with diagnostic logs (FR-019)
- On startup error: Include docker-compose error output

## Security Contracts

### Permissions

**Workflow Permissions**:
```yaml
permissions:
  contents: read        # Checkout code
  pull-requests: read   # Read PR metadata
  checks: write         # Create check runs
```

**Principle of Least Privilege**: No write access to contents, no admin permissions

### Secrets Management

**Contract**:
- No secrets required for current test services (all ports exposed via docker-compose)
- Future secrets MUST use GitHub Secrets, never hardcoded
- Secrets MUST NOT appear in logs or artifacts

## Failure Modes and Error Handling

### Service Startup Failures

**Scenario**: Docker services fail health checks
**Response**:
1. Wait up to 60 seconds for health
2. On timeout: Capture `docker compose ps` and `docker compose logs`
3. Fail job with diagnostic output
4. Run cleanup (always)

**Complies with**: FR-007, FR-019, FR-020

### Test Failures

**Scenario**: Tests fail
**Response**:
1. Complete test suite (don't fail fast within suite)
2. Report all failures with details
3. Capture service logs (for integration/E2E)
4. Upload all artifacts
5. Run cleanup (always)
6. Fail job

**Complies with**: FR-012, FR-020, FR-009

### Timeout Failures

**Scenario**: Job exceeds timeout
**Response**:
1. GitHub Actions kills the job
2. Cleanup steps may not run (GitHub limitation)
3. Manual cleanup may be required

**Mitigation**: Set realistic timeouts with buffer (FR-013)

### Cleanup Failures

**Scenario**: `docker compose down` fails
**Response**:
1. Log error but don't fail job (cleanup is best-effort)
2. Alert in job summary
3. Orphaned containers are GitHub's responsibility to clean

**Note**: This is a GitHub Actions platform limitation

## Versioning and Compatibility

### Action Versions

All GitHub Actions are pinned to major version:
- `actions/checkout@v5`
- `actions/setup-go@v6`
- `actions/cache@v4`
- `actions/upload-artifact@v4`
- `golangci/golangci-lint-action@v8`

**Update Policy**: Review and update quarterly, test in separate PR

### Go Version Compatibility

- Minimum: Go 1.21 (project requirement)
- CI uses: Go 1.25 (matches `go.mod`)
- Matrix testing: Not required (single Go version supported)

## Compliance Matrix

| Requirement | Contract Element | Verification |
|-------------|------------------|--------------|
| FR-001 | Push trigger | Workflow `on.push` config |
| FR-002 | PR trigger | Workflow `on.pull_request` config |
| FR-003 | Lint via make | Lint job runs `make lint` |
| FR-004 | Unit tests via make | Unit job runs `make test-unit` |
| FR-005 | Integration tests via make | Integration job runs `make test-integration` |
| FR-006 | Docker Compose startup | Integration/E2E jobs run `docker compose up -d` |
| FR-007 | Health check wait | 60s max wait with polling |
| FR-008 | E2E workflow test | E2E job runs `.github/test/test-dimp.sh` |
| FR-009 | Service cleanup | Cleanup steps have `if: always()` |
| FR-010 | Cancel outdated runs | `concurrency.cancel-in-progress: true` |
| FR-011 | Block merge on failure | Required status checks in branch protection |
| FR-012 | Publish artifacts | Upload steps with `if: always()` |
| FR-013 | Stage timeouts | Each job has `timeout-minutes` |
| FR-014 | Ubuntu + Docker | All jobs use `runs-on: ubuntu-latest` |
| FR-015 | Go 1.21+ | Setup Go with version 1.25 |
| FR-016 | Dependency caching | Go setup with `cache: true` |
| FR-017 | Status indicators | GitHub check runs per job |
| FR-018 | E2E on PR + main | E2E job runs on both triggers |
| FR-019 | Fail fast on service errors | Health check timeout → immediate failure |
| FR-020 | Capture logs on failure | Docker logs uploaded with `if: failure()` |

## Testing the Contract

To verify the workflow meets this contract:

1. **Test Lint Failure**: Push code with intentional lint violation
   - Expected: Lint job fails, subsequent jobs skipped, PR blocked

2. **Test Unit Failure**: Break a unit test
   - Expected: Unit job fails, integration/E2E skipped, artifacts uploaded

3. **Test Service Startup Failure**: Break Docker Compose config
   - Expected: Integration job fails in <60s with diagnostic logs

4. **Test E2E Failure**: Modify E2E script to fail
   - Expected: E2E job fails, logs captured, artifacts uploaded

5. **Test Cleanup**: Verify no containers remain after any outcome
   - Expected: `docker ps` shows no aether-related containers

6. **Test Concurrency**: Push multiple commits rapidly
   - Expected: Only latest commit's run completes, earlier runs cancelled

7. **Test Caching**: Run twice with same `go.sum`
   - Expected: Second run 40%+ faster

Each test validates a specific contract element.
