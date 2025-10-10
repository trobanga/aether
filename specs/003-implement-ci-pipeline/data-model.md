# Data Model: GitHub CI Pipeline

**Feature**: 003-implement-ci-pipeline
**Date**: 2025-10-10

## Overview

This feature is infrastructure configuration (GitHub Actions YAML), not application code, so traditional data modeling doesn't apply. However, we can model the workflow entities and their relationships to understand the pipeline structure.

## Workflow Entities

### 1. CI Workflow
**Definition**: The top-level GitHub Actions workflow that orchestrates all testing stages.

**Properties**:
- `name`: "CI" (workflow identifier)
- `triggers`: Array of event types (push, pull_request)
- `concurrency_group`: String for grouping runs
- `cancel_in_progress`: Boolean (true - implements FR-010)
- `jobs`: Array of Job entities

**Relationships**:
- Contains multiple Job entities
- Triggered by GitHub events (push, pull_request)

**Validation Rules**:
- Must trigger on `push` and `pull_request` events (FR-001, FR-002)
- Must enable concurrency cancellation (FR-010)
- Must define at least 4 jobs: lint, unit-test, integration-test, e2e-test

### 2. Job
**Definition**: A discrete execution unit within the workflow (lint, test, etc.).

**Properties**:
- `name`: String (e.g., "lint", "unit-test", "integration-test", "e2e-test")
- `runs_on`: String (e.g., "ubuntu-latest" - implements FR-014)
- `timeout_minutes`: Integer (5, 10, 20, or 30 - implements FR-013)
- `needs`: Array of job names (defines dependencies)
- `steps`: Array of Step entities
- `permissions`: Object (contents: read, pull-requests: read, checks: write)

**State Transitions**:
```
queued → in_progress → completed (success/failure/cancelled)
```

**Relationships**:
- Belongs to CI Workflow
- Contains multiple Step entities
- Depends on other Jobs (via `needs` field)

**Validation Rules**:
- Timeout must match requirements (FR-013):
  - lint: 5 minutes
  - unit-test: 10 minutes
  - integration-test: 20 minutes
  - e2e-test: 30 minutes
- Must run on ubuntu-latest (FR-014)
- Dependencies must be acyclic (no circular job dependencies)

### 3. Step
**Definition**: An individual action or command within a Job.

**Properties**:
- `name`: String (human-readable description)
- `uses`: String (GitHub Action reference) OR `run`: String (shell command)
- `with`: Object (action inputs)
- `env`: Object (environment variables)
- `if`: String (conditional expression, e.g., "always()", "failure()")
- `working_directory`: String (optional)

**Types**:
1. **Setup Steps**: Checkout code, setup Go, setup tools
2. **Execution Steps**: Run lint, run tests, start services
3. **Artifact Steps**: Upload test results, logs, coverage
4. **Cleanup Steps**: Tear down Docker services (always run - FR-009)

**Relationships**:
- Belongs to Job
- Executes sequentially within job
- May produce Artifacts

**Validation Rules**:
- Cleanup steps must use `if: always()` (FR-009)
- Artifact upload steps for logs should use `if: failure()` (FR-020)
- Service startup steps must include health check polling (FR-007)

### 4. Docker Service
**Definition**: External test dependency managed via Docker Compose.

**Properties**:
- `name`: String (e.g., "vfps_db", "vfps", "fhir-pseudonymizer")
- `image`: String (Docker image reference)
- `healthcheck`: Object (health check configuration)
- `depends_on`: Array of service names
- `environment`: Object (environment variables)

**State Transitions**:
```
starting → healthy/unhealthy → running → stopped
```

**Relationships**:
- Defined in `.github/test/docker-compose.yaml`
- Started by integration-test and e2e-test Jobs
- Must reach "healthy" state before tests run (FR-007)

**Validation Rules**:
- Must have health check defined
- Health check must pass within 60 seconds (FR-007)
- Must be stopped after tests, even on failure (FR-009)

### 5. Artifact
**Definition**: Output file(s) from CI run uploaded to GitHub for download.

**Properties**:
- `name`: String (unique identifier for this run)
- `path`: String or Array (file paths to upload)
- `retention_days`: Integer (7-90 days)
- `if_no_files_found`: String ("warn", "error", "ignore")

**Types**:
1. **Test Results**: JSON/XML test output (always uploaded)
2. **Coverage Reports**: coverage.out, coverage.html (always uploaded)
3. **Service Logs**: Docker Compose logs (failure only - FR-020)
4. **E2E Diagnostics**: Screenshots, trace files (failure only)

**Relationships**:
- Produced by Steps
- Belongs to Job/Workflow run
- Stored by GitHub Actions platform

**Validation Rules**:
- Test results must be uploaded regardless of test outcome (`if: always()`)
- Service logs only uploaded on failure (`if: failure()`) (FR-020)
- Retention must be 7-90 days (GitHub Actions limits)

### 6. Cache
**Definition**: Stored data reused across workflow runs for performance.

**Properties**:
- `key`: String (unique cache identifier using hash)
- `paths`: Array of strings (directories to cache)
- `restore_keys`: Array (fallback keys for partial match)

**Types**:
1. **Go Module Cache**: `~/go/pkg/mod`, `~/.cache/go-build`
2. **golangci-lint Analysis Cache**: Automatic via action
3. **Docker Layer Cache**: (Optional, not in initial implementation)

**Relationships**:
- Used by Jobs
- Invalidated by key change (go.sum hash)

**Validation Rules**:
- Key must include `${{ runner.os }}` and `${{ hashFiles('**/go.sum') }}`
- Must achieve 40% speedup (SC-010) - validated through metrics

## Workflow State Machine

```
GitHub Event (push/PR)
  ↓
CI Workflow Triggered
  ↓
[concurrency check: cancel outdated runs]
  ↓
Lint Job (queued → running → success/failure)
  ↓ (if success)
Unit Test Job (parallel or sequential)
  ↓ (if success)
Integration Test Job
  ├─ Start Docker Services
  ├─ Wait for Health Checks (max 60s)
  ├─ Run Integration Tests
  ├─ Capture Logs (if failure)
  └─ Cleanup Services (always)
  ↓ (if success)
E2E Test Job
  ├─ Start Docker Services
  ├─ Wait for Health Checks (max 60s)
  ├─ Run E2E Tests
  ├─ Capture Logs (if failure)
  └─ Cleanup Services (always)
  ↓
Upload Artifacts (test results, logs, coverage)
  ↓
Workflow Complete → Update PR Status
```

## Configuration Files

### Primary Workflow File
**Location**: `.github/workflows/ci.yml`
**Format**: YAML (GitHub Actions schema)
**Content**:
- Workflow metadata (name, triggers, concurrency)
- Job definitions (lint, unit-test, integration-test, e2e-test)
- Steps for each job
- Cache configuration
- Artifact configuration

### Supporting Configuration
**Location**: `.github/test/docker-compose.yaml` (existing)
**Purpose**: Defines Docker services for integration and E2E tests
**Usage**: Started by CI jobs via `docker compose up -d`

### Optional Configuration
**Location**: `.golangci.yml` (may be created for consistency)
**Purpose**: golangci-lint configuration shared between local dev and CI
**Content**: Linter rules, exclusions, severity levels

## Entity Relationships Diagram

```
CI Workflow (1)
  ├─── triggers: Events (push, pull_request)
  ├─── concurrency: ConcurrencyGroup (1)
  └─── jobs: Job (4)
         ├── Lint Job (1)
         │   ├── steps: Step (5)
         │   │   ├── Setup Go
         │   │   ├── Cache Dependencies
         │   │   ├── Run golangci-lint
         │   │   └── Upload Results
         │   └── produces: Artifact (test-results)
         │
         ├── Unit Test Job (1)
         │   ├── needs: Lint Job
         │   ├── steps: Step (4)
         │   └── produces: Artifact (coverage, test-results)
         │
         ├── Integration Test Job (1)
         │   ├── needs: Unit Test Job
         │   ├── steps: Step (7)
         │   │   ├── Start Services
         │   │   ├── Health Check Wait
         │   │   ├── Run Tests
         │   │   ├── Capture Logs (if failure)
         │   │   └── Cleanup (always)
         │   ├── manages: Docker Services (3)
         │   │   ├── vfps_db
         │   │   ├── vfps
         │   │   └── fhir-pseudonymizer
         │   └── produces: Artifact (test-results, logs)
         │
         └── E2E Test Job (1)
             ├── needs: Integration Test Job
             ├── steps: Step (7)
             ├── manages: Docker Services (3)
             └── produces: Artifact (test-results, e2e-artifacts, logs)
```

## Validation Matrix

| Entity | Requirement | Validation Method |
|--------|-------------|-------------------|
| Workflow | Triggers on push/PR | Check `on:` section includes both (FR-001, FR-002) |
| Workflow | Cancels outdated runs | Check `concurrency.cancel-in-progress: true` (FR-010) |
| Lint Job | Timeout 5 min | Check `timeout-minutes: 5` (FR-013) |
| Unit Test Job | Timeout 10 min | Check `timeout-minutes: 10` (FR-013) |
| Integration Job | Timeout 20 min | Check `timeout-minutes: 20` (FR-013) |
| E2E Job | Timeout 30 min | Check `timeout-minutes: 30` (FR-013) |
| All Jobs | Runs on ubuntu-latest | Check `runs-on: ubuntu-latest` (FR-014) |
| Service Steps | Health check wait | Max 60s timeout in polling loop (FR-007) |
| Cleanup Steps | Always runs | Check `if: always()` condition (FR-009) |
| Log Capture | On failure | Check `if: failure()` condition (FR-020) |
| Artifacts | Test results uploaded | Check upload step with `if: always()` (FR-012) |

## Notes

- This is a configuration feature, not a code feature, so entities are workflow components rather than application data structures
- No database or persistent storage involved (CI state is ephemeral)
- The "data model" describes the structure and relationships of the GitHub Actions workflow configuration
- Validation focuses on compliance with functional requirements rather than data integrity constraints
