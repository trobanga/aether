# Tasks: GitHub CI Pipeline

**Input**: Design documents from `/specs/003-implement-ci-pipeline/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/workflow-contract.md, quickstart.md

**Tests**: This feature IS the testing infrastructure. Validation tasks replace traditional tests.

**Organization**: Tasks are grouped by user story (pipeline stage) to enable independent implementation and validation.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different sections of workflow file, no dependencies)
- **[Story]**: Which user story/pipeline stage this task belongs to (US1, US2, US3, US4)
- File path: `.github/workflows/ci.yml` (single file feature)

## Path Conventions
- Workflow file: `.github/workflows/ci.yml`
- Existing infrastructure: `.github/test/`, `Makefile`, `tests/`
- Documentation: `specs/003-implement-ci-pipeline/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create workflow file structure and common configuration

- [X] **T001** [Setup] Create `.github/workflows/` directory
  - Run: `mkdir -p .github/workflows`

- [X] **T002** [Setup] Create basic `ci.yml` with workflow metadata and triggers
  - File: `.github/workflows/ci.yml`
  - Add workflow name: "CI"
  - Configure triggers for push (all branches) and pull_request events (FR-001, FR-002)
  - Add concurrency control: cancel in-progress runs for same ref (FR-010)
  - Set permissions: `contents: read`, `pull-requests: read`, `checks: write`
  - Reference: quickstart.md Step 2

- [X] **T003** [Setup] Verify workflow file syntax
  - Run: `gh workflow view ci.yml` or use GitHub Actions YAML validator
  - Ensure valid YAML structure

**Checkpoint**: Basic workflow structure created and triggers configured

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core job structure that ALL pipeline stages depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] **T004** [Foundation] Define job structure skeleton in `ci.yml`
  - Add empty job definitions: `lint`, `unit-test`, `integration-test`, `e2e-test`
  - Set `runs-on: ubuntu-latest` for all jobs (FR-014)
  - Define job dependencies: unit-test needs lint, integration-test needs unit-test, e2e-test needs integration-test

- [X] **T005** [Foundation] Commit and push to trigger first workflow run
  - Run: `git add .github/workflows/ci.yml && git commit -m "feat: add CI workflow skeleton" && git push`
  - Verify workflow appears in GitHub Actions UI
  - Expected: All jobs should be skipped (no steps defined yet)

**Checkpoint**: Foundation ready - pipeline stage implementation can now begin

---

## Phase 3: User Story 1 - Automated Code Quality Validation (Priority: P1) 🎯 MVP

**Goal**: Implement lint job that validates code quality on every push/PR

**Independent Test**: Push code with intentional linting violation → verify CI fails with specific error message → fix violation → verify CI passes

### Validation for User Story 1 (Test-the-Tests Approach)

**NOTE: These validation tasks ensure the lint job works correctly by intentionally breaking it**

- [ ] **T006** [US1] Create validation test plan for lint job
  - Document in `specs/003-implement-ci-pipeline/validation-notes.md`:
    - Test 1: Introduce linting violation (e.g., unused variable)
    - Test 2: Verify failure with clear error message
    - Test 3: Fix violation and verify pass
  - Reference: quickstart.md "Lint Validation" checklist

### Implementation for User Story 1

- [ ] **T007** [US1] Implement lint job in `.github/workflows/ci.yml`
  - Set `timeout-minutes: 5` (FR-013)
  - Add checkout step: `actions/checkout@v5`
  - Add Go setup: `actions/setup-go@v6` with go-version '1.25', cache: true (FR-015, FR-016)
  - Add golangci-lint step: `golangci/golangci-lint-action@v8` with version v2.1
  - Add artifact upload for lint results (retention-days: 30) (FR-012)
  - Reference: quickstart.md Step 3, research.md section 1

- [ ] **T008** [US1] Commit and push lint job implementation
  - Run: `git add .github/workflows/ci.yml && git commit -m "feat: implement lint job in CI pipeline" && git push`
  - Verify lint job runs successfully on clean code

- [ ] **T009** [US1] Execute validation test: Introduce linting violation
  - Modify a Go file to add intentional violation (e.g., unused variable)
  - Commit and push: `git commit -m "test: intentional lint violation for CI validation"`
  - Verify: Lint job FAILS with specific violation details
  - Verify: PR shows "checks failed" status
  - Verify: Error message includes file path and line number

- [ ] **T010** [US1] Execute validation test: Fix and verify pass
  - Remove the linting violation
  - Commit and push: `git commit -m "fix: remove lint violation"`
  - Verify: Lint job PASSES
  - Verify: Pipeline proceeds to next stage

**Checkpoint**: User Story 1 complete - Lint job functional and validated. Can merge to main as MVP!

---

## Phase 4: User Story 2 - Automated Unit Test Execution (Priority: P2)

**Goal**: Implement unit-test job that runs all unit tests with coverage reporting

**Independent Test**: Break a unit test → verify CI fails with test details → fix test → verify CI passes

### Validation for User Story 2

- [ ] **T011** [US2] Create validation test plan for unit-test job
  - Document validation scenarios:
    - Test 1: Introduce failing unit test
    - Test 2: Verify pipeline fails with test failure details
    - Test 3: Verify coverage report is generated and uploaded
    - Test 4: Fix test and verify pass

### Implementation for User Story 2

- [ ] **T012** [US2] Implement unit-test job in `.github/workflows/ci.yml`
  - Set `needs: [lint]` dependency
  - Set `timeout-minutes: 10` (FR-013)
  - Add checkout step: `actions/checkout@v5`
  - Add Go setup with caching (same as lint job)
  - Add step to run `make test-unit` (FR-004)
  - Add step to generate coverage: `make coverage`
  - Add artifact upload for test results and coverage (retention-days: 90) (FR-012)
  - Use `if: always()` for artifact upload to capture results even on failure
  - Reference: quickstart.md Step 4

- [ ] **T013** [US2] Commit and push unit-test job implementation
  - Run: `git add .github/workflows/ci.yml && git commit -m "feat: implement unit-test job in CI pipeline" && git push`
  - Verify unit-test job runs after lint passes

- [ ] **T014** [US2] Execute validation test: Introduce failing unit test
  - Modify a unit test to fail (e.g., change assertion)
  - Commit and push: `git commit -m "test: intentional unit test failure for CI validation"`
  - Verify: Unit-test job FAILS
  - Verify: Error details show which test failed and assertion message
  - Verify: Coverage artifacts still uploaded (if: always())

- [ ] **T015** [US2] Execute validation test: Fix and verify pass
  - Fix the failing test
  - Commit and push: `git commit -m "fix: restore passing unit test"`
  - Verify: Unit-test job PASSES
  - Verify: Coverage report uploaded successfully

- [ ] **T016** [US2] Verify concurrency cancellation (FR-010, US2 Scenario 3)
  - Make a trivial change and push commit A
  - Immediately make another change and push commit B (while A's workflow is running)
  - Verify: Commit A's workflow is cancelled
  - Verify: Only commit B's workflow completes

**Checkpoint**: User Story 2 complete - Unit tests running with coverage. Lint + Unit = MVP enhancement!

---

## Phase 5: User Story 3 - Automated Integration Test Execution (Priority: P3)

**Goal**: Implement integration-test job that starts Docker services, runs integration tests, captures logs, and cleans up

**Independent Test**: Break integration test → verify CI starts services, runs tests, captures logs, cleans up containers

### Validation for User Story 3

- [ ] **T017** [US3] Create validation test plan for integration-test job
  - Document validation scenarios:
    - Test 1: Verify Docker services start and health checks pass
    - Test 2: Introduce failing integration test, verify logs captured
    - Test 3: Verify cleanup runs even on failure (no orphaned containers)
    - Test 4: Simulate health check timeout, verify fail-fast behavior

### Implementation for User Story 3

- [ ] **T018** [US3] Implement integration-test job - service startup in `.github/workflows/ci.yml`
  - Set `needs: [unit-test]` dependency
  - Set `timeout-minutes: 20` (FR-013)
  - Add checkout and Go setup steps (same pattern as previous jobs)
  - Add step to start Docker services: `docker compose -f .github/test/docker-compose.yaml up -d` (FR-006)
  - Reference: quickstart.md Step 5

- [ ] **T019** [US3] Implement health check wait logic
  - Add step with timeout 60s to poll for healthy services (FR-007)
  - Script: `timeout 60 bash -c 'until docker compose ps | grep -q "healthy"; do sleep 2; done'`
  - On timeout: Capture docker ps and logs, then exit with code 2 (FR-019)
  - Reference: quickstart.md Step 5, research.md section 3

- [ ] **T020** [US3] Implement test execution and log capture
  - Add step to run `make test-integration` (FR-005)
  - Add step with `if: failure()` to capture Docker logs: `docker compose logs > docker-logs-integration.txt` (FR-020)
  - Add artifact upload for test results (if: always())
  - Add artifact upload for Docker logs (if: failure(), retention-days: 7)

- [ ] **T021** [US3] Implement cleanup with always-run guarantee
  - Add cleanup step: `docker compose -f .github/test/docker-compose.yaml down -v` (FR-009)
  - Set `if: always()` to ensure cleanup runs regardless of test outcome
  - Reference: workflow-contract.md "Resource Cleanup Guarantee"

- [ ] **T022** [US3] Commit and push integration-test job implementation
  - Run: `git add .github/workflows/ci.yml && git commit -m "feat: implement integration-test job with Docker orchestration" && git push`
  - Verify integration-test job runs after unit-test passes

- [ ] **T023** [US3] Execute validation test: Verify service startup and health checks
  - Trigger CI run with clean code
  - Verify: Docker services start successfully
  - Verify: Health checks pass within 60 seconds
  - Verify: Integration tests execute
  - Verify: Cleanup removes all containers (check GitHub Actions logs for `docker compose down -v`)

- [ ] **T024** [US3] Execute validation test: Failing integration test with log capture
  - Modify integration test to fail (e.g., expect wrong status code from DIMP service)
  - Commit and push: `git commit -m "test: intentional integration test failure"`
  - Verify: Integration-test job FAILS
  - Verify: Service logs are captured and uploaded as artifact
  - Verify: Cleanup still runs (if: always())
  - Download artifact and verify log contents include service diagnostics

- [ ] **T025** [US3] Execute validation test: Verify no orphaned containers
  - After T024 completes (failed run):
  - Check GitHub Actions runner (if accessible) or trust cleanup logs
  - Verify: No containers remain (validates SC-003)
  - Alternative: Manually run workflow on self-hosted runner to verify cleanup

- [ ] **T026** [US3] Execute validation test: Fix and verify pass
  - Fix the failing integration test
  - Commit and push: `git commit -m "fix: restore passing integration test"`
  - Verify: Integration-test job PASSES
  - Verify: Pipeline proceeds to E2E stage

**Checkpoint**: User Story 3 complete - Integration tests with Docker orchestration working!

---

## Phase 6: User Story 4 - Automated End-to-End Test Execution (Priority: P4)

**Goal**: Implement e2e-test job that runs full pipeline workflow tests against test environment

**Independent Test**: Create/break E2E test → verify CI orchestrates full workflow, captures diagnostics, cleans up

### Validation for User Story 4

- [ ] **T027** [US4] Create validation test plan for e2e-test job
  - Document validation scenarios:
    - Test 1: Verify E2E script executes successfully
    - Test 2: Verify full pipeline workflow (build → services → test → cleanup)
    - Test 3: Break E2E test, verify comprehensive logs captured
    - Test 4: Verify artifacts accessible within 30 seconds (SC-005)

### Implementation for User Story 4

- [ ] **T028** [US4] Implement e2e-test job - build and service startup in `.github/workflows/ci.yml`
  - Set `needs: [integration-test]` dependency
  - Set `timeout-minutes: 30` (FR-013)
  - Add checkout and Go setup steps
  - Add step to build binary: `make build` (required for E2E test script)
  - Add Docker service startup (same pattern as integration job)
  - Add health check wait (same pattern as integration job)
  - Reference: quickstart.md Step 6

- [ ] **T029** [US4] Implement E2E test execution and artifact collection
  - Add step to make test script executable: `chmod +x .github/test/test-dimp.sh`
  - Add step to run E2E test: `./.github/test/test-dimp.sh` (FR-008)
  - Add step with `if: failure()` to capture Docker logs (FR-020)
  - Add artifact upload for E2E test results (if: always())
  - Add artifact upload for E2E diagnostics/logs (if: failure(), retention-days: 7)
  - Add cleanup step with `if: always()` (FR-009)

- [ ] **T030** [US4] Commit and push e2e-test job implementation
  - Run: `git add .github/workflows/ci.yml && git commit -m "feat: implement e2e-test job with full workflow validation" && git push`
  - Verify e2e-test job runs after integration-test passes

- [ ] **T031** [US4] Execute validation test: Verify E2E workflow execution
  - Trigger CI run with clean code
  - Verify: Binary builds successfully
  - Verify: Docker services start
  - Verify: E2E test script executes (`.github/test/test-dimp.sh`)
  - Verify: Test completes successfully
  - Verify: Artifacts uploaded
  - Verify: Cleanup runs

- [ ] **T032** [US4] Execute validation test: Break E2E test and verify diagnostics
  - Modify `.github/test/test-dimp.sh` to fail (e.g., expect wrong response)
  - Commit and push: `git commit -m "test: intentional E2E test failure"`
  - Verify: E2E-test job FAILS
  - Verify: Comprehensive logs captured (service logs, test output)
  - Verify: Artifacts uploaded with diagnostic information
  - Within 30 seconds: Download artifact from GitHub UI (SC-005)
  - Verify: Cleanup still runs

- [ ] **T033** [US4] Execute validation test: Fix and verify pass
  - Restore `.github/test/test-dimp.sh` to working state
  - Commit and push: `git commit -m "fix: restore passing E2E test"`
  - Verify: E2E-test job PASSES
  - Verify: Full pipeline (lint → unit → integration → E2E) completes successfully

- [ ] **T034** [US4] Verify pipeline completes within performance target
  - Trigger full pipeline run (all tests passing)
  - Measure total time from trigger to completion
  - Verify: Full pipeline < 25 minutes (SC-006)
  - First run (cold cache): Record time
  - Second run (warm cache): Verify 40%+ speedup (SC-010)

**Checkpoint**: User Story 4 complete - Full CI pipeline operational with all 4 stages!

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Finalize pipeline configuration and enable branch protection

- [ ] **T035** [P] [Polish] Configure branch protection rules for main branch
  - Via GitHub UI or CLI: Require status checks before merge (FR-011)
  - Add required checks: `lint`, `unit-test`, `integration-test`, `e2e-test`
  - Enable "Require branches to be up to date before merging"
  - Reference: quickstart.md Step 7

- [ ] **T036** [P] [Polish] Verify branch protection enforcement
  - Create test PR with failing lint
  - Verify: Merge button is disabled
  - Fix lint, push update
  - Verify: Merge button is enabled (SC-004)

- [ ] **T037** [P] [Polish] Add optional golangci-lint configuration file
  - File: `.golangci.yml` (optional, for consistency between local/CI)
  - Configure linter rules if needed
  - Reference: research.md section 1 "Integration with Existing Makefile"

- [ ] **T038** [P] [Polish] Document CI pipeline for team
  - Create `docs/ci-pipeline.md` or update README with:
    - How to view CI runs: `gh run watch`
    - How to download artifacts: `gh run download <run-id>`
    - How to interpret failures
    - Link to quickstart.md for troubleshooting

- [ ] **T039** [Polish] Run full quickstart.md validation checklist
  - Execute all validation scenarios from quickstart.md:
    - ✅ Lint validation (T009-T010)
    - ✅ Unit test validation (T014-T015)
    - ✅ Integration test validation (T023-T026)
    - ✅ E2E test validation (T031-T033)
    - ✅ Performance validation (T034)
    - ✅ Concurrency validation (T016)
    - ✅ Branch protection validation (T036)
  - Confirm all checkboxes pass

- [ ] **T040** [Polish] Performance optimization review
  - Review cache hit rates in workflow logs
  - Verify: Go module cache is working (cache: true in setup-go)
  - Verify: golangci-lint cache is effective (~14s with cache vs ~50s without)
  - If needed: Adjust cache keys or retention

- [ ] **T041** [Polish] Final cleanup and documentation
  - Remove any test code used for validation (intentional failures)
  - Update CLAUDE.md if any CI-specific conventions added
  - Ensure `.github/workflows/ci.yml` has clear comments for each job
  - Commit final state: `git commit -m "docs: finalize CI pipeline documentation"`

**Checkpoint**: CI pipeline fully implemented, validated, and documented!

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - start immediately
  - T001 → T002 → T003 (sequential)

- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
  - T004 → T005 (sequential)

- **User Story 1 (Phase 3)**: Depends on Foundational completion
  - T006 (validation plan) can be done anytime
  - T007 → T008 → T009 → T010 (sequential validation)

- **User Story 2 (Phase 4)**: Depends on Foundational completion (not on US1, but sequential due to single file)
  - Must wait for T010 to avoid merge conflicts in ci.yml
  - T011 → T012 → T013 → T014 → T015 → T016

- **User Story 3 (Phase 5)**: Depends on Foundational completion
  - Must wait for T016 to avoid merge conflicts in ci.yml
  - T017 → T018 → T019 → T020 → T021 → T022 → T023 → T024 → T025 → T026

- **User Story 4 (Phase 6)**: Depends on Foundational completion
  - Must wait for T026 to avoid merge conflicts in ci.yml
  - T027 → T028 → T029 → T030 → T031 → T032 → T033 → T034

- **Polish (Phase 7)**: Depends on all user stories being complete
  - T035, T036, T037, T038 are parallelizable [P]
  - T039 requires all prior validation tasks complete
  - T040 → T041 (final sequential tasks)

### Special Note: Single File Feature

This feature modifies a single file (`.github/workflows/ci.yml`), so most tasks are **sequential** to avoid merge conflicts. However:
- Validation planning tasks can be done in parallel with implementation
- Polish tasks (T035-T038) are independent and can run in parallel
- Different developers can work on different phases by coordinating on the ci.yml file

### Within Each User Story

Each pipeline stage follows this pattern:
1. Create validation plan (can be done early)
2. Implement job in ci.yml
3. Commit and push
4. Execute validation: introduce failure
5. Verify failure behavior
6. Fix and verify pass

### Parallel Opportunities (Limited due to single file)

- **Phase 1**: T001 and T002 could overlap (different operations)
- **Phase 3-6**: Validation planning (T006, T011, T017, T027) can be done concurrently with prior phase implementation
- **Phase 7**: T035, T036, T037, T038 can all run in parallel

---

## Implementation Strategy

### MVP First (User Story 1 Only)

**Minimal Viable Pipeline**: Just the lint job

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T005)
3. Complete Phase 3: User Story 1 (T006-T010)
4. **STOP and VALIDATE**: Full lint validation
5. **DECISION POINT**: This is a working pipeline! Merge to main if desired

**Value**: Immediate code quality enforcement on all PRs

### Incremental Delivery

**Each stage adds incremental value:**

1. **After US1 (Lint)**: Code quality checks on every PR
2. **After US2 (+ Unit Tests)**: Code quality + unit test coverage (3-minute feedback)
3. **After US3 (+ Integration)**: Full testing including Docker services
4. **After US4 (+ E2E)**: Complete pipeline with end-to-end validation

### Recommended Approach (Sequential Stages)

Since this is a single-file feature:

1. **Sprint 1**: Setup + Foundational + US1 (Lint) - MVP
2. **Sprint 2**: US2 (Unit Tests) - Enhanced MVP
3. **Sprint 3**: US3 (Integration Tests) - Full testing
4. **Sprint 4**: US4 (E2E Tests) + Polish - Complete pipeline

Each sprint delivers a working, incrementally better CI pipeline.

### Validation-Driven Development (TDD for Infrastructure)

**For each user story:**
1. Define validation scenarios FIRST (what should fail?)
2. Implement job configuration
3. Execute validation: intentionally break it
4. Verify failure behavior matches expectations
5. Fix and verify pass
6. Move to next story

This ensures the pipeline catches failures correctly before relying on it for main branch protection.

---

## Performance Targets

| Metric | Target | Validation Task |
|--------|--------|-----------------|
| Lint feedback | < 1 minute | T009, T010 |
| Lint + Unit feedback | < 3 minutes | T015 + SC-001 |
| Full pipeline | < 25 minutes | T034 (SC-006) |
| Cache speedup | 40%+ faster | T034, T040 (SC-010) |
| Service startup | < 60 seconds | T023 (FR-007) |
| Resource cleanup | 100% success | T025 (SC-003) |
| Artifact availability | < 30 seconds | T032 (SC-005) |

---

## Summary

- **Total Tasks**: 41
- **Setup Phase**: 3 tasks
- **Foundational Phase**: 2 tasks
- **User Story 1 (Lint)**: 5 tasks - MVP milestone
- **User Story 2 (Unit Tests)**: 6 tasks
- **User Story 3 (Integration Tests)**: 10 tasks
- **User Story 4 (E2E Tests)**: 8 tasks
- **Polish Phase**: 7 tasks

**Parallel Opportunities**: Limited (single file) - mainly validation planning and polish tasks

**Independent Test Criteria**:
- US1: Push lint violation → verify failure → fix → verify pass
- US2: Break unit test → verify failure with details → fix → verify pass
- US3: Break integration test → verify services start, test fails, logs captured, cleanup runs
- US4: Break E2E test → verify full workflow, diagnostics captured, cleanup runs

**Suggested MVP Scope**: User Story 1 (Lint job) - Provides immediate value with code quality enforcement

**Expected Timeline**:
- MVP (US1): ~2-3 hours
- Enhanced MVP (US1+US2): ~4-6 hours
- Full Testing (US1+US2+US3): ~8-12 hours
- Complete Pipeline (All): ~12-16 hours (including thorough validation)

---

## Notes

- Single file feature: Most tasks are sequential to avoid merge conflicts in `ci.yml`
- Validation tasks use "test-the-tests" approach per TDD constitution principle
- Each user story (pipeline stage) can be validated independently
- Checkpoints after each story enable incremental delivery
- All functional requirements (FR-001 through FR-020) mapped to specific tasks
- All success criteria (SC-001 through SC-010) have validation tasks
- Reference documents: quickstart.md (step-by-step), research.md (decisions), workflow-contract.md (behavior specs)
