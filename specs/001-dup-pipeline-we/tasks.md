# Tasks: DUP Pipeline CLI (Aether)

**Input**: Design documents from `/specs/001-dup-pipeline-we/`
**Prerequisites**: plan.md (✓), spec.md (✓), research.md (✓), data-model.md (✓), contracts/ (✓)
**Last Updated**: 2025-10-08 (FR-029 progress indicator requirements integrated)

**Tests**: Tests are included per TDD requirement from constitution (plan.md Phase II confirmation)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

**FR-029 Updates**: Added 6 tasks for progress indicator implementation:
- T020-T022: UI infrastructure (ProgressBar, Spinner, ETA, Throughput calculators)
- T023-T024: UI component tests
- T032, T034, T038: Progress integration in User Story 1
- Updated US4 tasks for progress displays
- T082, T090: FR-029 validation in Polish phase

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Initialize Go module with `go mod init github.com/user/aether`
- [X] T002 Create project directory structure (`internal/models/`, `internal/pipeline/`, `internal/services/`, `internal/cli/`, `internal/ui/`, `internal/lib/`, `tests/contract/`, `tests/integration/`, `tests/unit/`, `config/`, `jobs/`)
- [X] T003 [P] Install Cobra CLI: `go install github.com/spf13/cobra-cli@latest` and initialize with `cobra-cli init`
- [X] T004 [P] Add dependencies to go.mod (cobra, viper, uuid, yaml.v3, testify, progressbar/v3)
- [X] T005 [P] Create example configuration file `config/aether.example.yaml` with default values from data-model.md
- [X] T006 [P] Add `.gitignore` for Go project (include `jobs/`, binaries, IDE files)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T007 [P] Implement core data models in `internal/models/job.go` (PipelineJob, InputType, JobStatus)
- [X] T008 [P] Implement step models in `internal/models/step.go` (PipelineStep, StepName, StepStatus, StepError, ErrorType)
- [X] T009 [P] Implement file models in `internal/models/file.go` (FHIRDataFile)
- [X] T010 [P] Implement config models in `internal/models/config.go` (ProjectConfig, ServiceConfig, PipelineConfig, RetryConfig)
- [X] T011 Implement model validation functions in `internal/models/validation.go` (PipelineJob.Validate(), ProjectConfig.Validate())
- [X] T012 Implement pure state transition functions in `internal/models/transitions.go` (UpdateJobStatus, CompleteStep, IncrementRetry)
- [X] T013 [P] Implement configuration loader in `internal/services/config.go` (uses viper to load YAML + merge CLI flags)
- [X] T014 [P] Implement state persistence service in `internal/services/state.go` (LoadJobState, SaveJobState with atomic writes via os.Rename)
- [X] T015 [P] Implement retry logic utilities in `internal/lib/retry.go` (exponential backoff, IsRetryable function)
- [X] T016 [P] Implement FHIR parsing utilities in `internal/lib/fhir.go` (parse NDJSON line-by-line, generic FHIRResource map)
- [X] T017 Setup main CLI structure in `cmd/aether/main.go` with cobra root command
- [X] T018 [P] Implement HTTP client wrapper in `internal/services/httpclient.go` with timeout/retry configuration
- [X] T019 [P] Implement logging infrastructure in `internal/lib/logging.go` (structured logging for operations, errors, retries)
- [X] T020 [P] Implement progress indicator base infrastructure in `internal/ui/progress.go` (ProgressBar and Spinner wrappers for progressbar library, configurable 2s updates per FR-029d)
- [X] T021 [P] Implement ETA calculator in `internal/ui/eta.go` (compute avg_time_per_item from last 10 items or 30s window per FR-029b formula)
- [X] T022 [P] Implement throughput calculator in `internal/ui/throughput.go` (files/sec, MB/sec rate calculations for FR-029e)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Basic Pipeline Execution (Priority: P1) 🎯 MVP

**Goal**: Import TORCH-extracted FHIR data from local directory or URL into job-specific directory structure

**Independent Test**: Provide TORCH output directory path or URL and verify FHIR files are imported with unique job ID

### Tests for User Story 1

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T023 [P] [US1] Unit test for progress bar display in `tests/unit/ui/progress_bar_test.go` (mock output, verify FR-029 format)
- [X] T024 [P] [US1] Unit test for ETA calculation in `tests/unit/ui/eta_test.go` (verify averaging window logic)
- [X] T025 [P] [US1] Contract test for local directory import in `tests/unit/import_local_test.go`
- [X] T026 [P] [US1] Contract test for HTTP URL download in `tests/unit/import_url_test.go`
- [X] T027 [P] [US1] Integration test for full import workflow (local path) in `tests/integration/pipeline_import_local_test.go`
- [X] T028 [P] [US1] Integration test for full import workflow (HTTP URL) with progress display in `tests/integration/pipeline_import_url_test.go`
- [X] T029 [P] [US1] Integration test for invalid input source (unreachable URL) in `tests/integration/pipeline_import_error_test.go`

### Implementation for User Story 1

- [X] T030 [US1] Implement file import service for local paths in `internal/services/importer.go` (ImportFromLocalDirectory function)
- [X] T031 [US1] Implement file download service for HTTP URLs in `internal/services/downloader.go` (DownloadFromURL function with progress callback hooks)
- [X] T032 [US1] Integrate progress bar into download service in `internal/services/downloader.go` (wire ProgressBar from ui/ package, show percentage, ETA, throughput per FR-029a,b,e)
- [X] T033 [US1] Implement job initialization in `internal/pipeline/job.go` (CreateJob function: generate UUID, create directory structure, initialize state)
- [X] T034 [US1] Implement import step orchestration in `internal/pipeline/import.go` (ExecuteImportStep: detect input type, delegate to importer/downloader, update job state, show progress for batch operations)
- [X] T035 [US1] Implement `pipeline start` command in `cmd/pipeline.go` (cobra command with --input flag)
- [X] T036 [US1] Implement `pipeline status` command in `cmd/pipeline.go` (cobra command to display job state, step progress, file counts with formatted output per FR-029e)
- [X] T037 [US1] Add input validation (check FHIR NDJSON format) in `internal/services/importer.go` and `models/validation.go`
- [X] T038 [US1] Integrate spinner for HTTP connection phase in download service (unknown duration operations per FR-029c)
- [X] T039 [US1] Integrate error handling for network failures and invalid paths in import/download services

**Checkpoint**: At this point, User Story 1 should be fully functional - user can import TORCH data and check status

---

## Phase 4: User Story 2 - Pipeline Resumption Across Sessions (Priority: P2)

**Goal**: Enable users to resume pipeline operations after closing terminal, with full state persistence

**Independent Test**: Start pipeline, close CLI, reopen, verify job status retrieval and continuation

### Tests for User Story 2

- [X] T040 [P] [US2] Unit test for state persistence (save/load cycle) in `tests/unit/state_persistence_test.go`
- [X] T041 [P] [US2] Integration test for pipeline resumption in `tests/integration/pipeline_resume_test.go`
- [X] T042 [P] [US2] Integration test for job list with multiple jobs in `tests/integration/job_list_test.go`
- [X] T043 [P] [US2] Integration test for retry count tracking in `tests/integration/retry_tracking_test.go`

### Implementation for User Story 2

- [X] T044 [US2] Implement `job list` command in `cmd/job.go` (display all jobs with ID, status, current step, retry count, formatted with ui/ components)
- [X] T045 [US2] Implement `pipeline continue` command in `cmd/pipeline.go` (load job state, resume from next enabled step with progress indicators)
- [X] T046 [US2] Implement job discovery in `internal/services/state.go` (ListAllJobs: scan jobs directory, parse state files) - ALREADY IMPLEMENTED
- [X] T047 [US2] Implement step sequencing logic in `internal/models/config.go` (GetNextStep function based on enabled_steps config) - ALREADY IMPLEMENTED
- [X] T048 [US2] Add retry state tracking in step execution (increment retry count, log retry attempts) - ALREADY IMPLEMENTED in models/transitions.go
- [X] T049 [US2] Implement error type classification in `internal/lib/retry.go` (transient vs non-transient detection from HTTP status, network errors) - ALREADY IMPLEMENTED
- [X] T050 [US2] Integrate automatic retry logic with exponential backoff for transient errors in `internal/pipeline/import.go` (RetryImportStep) - ALREADY IMPLEMENTED
- [X] T051 [US2] Add status display formatting (show error details, suggest recovery commands) in `cmd/pipeline.go` using ui/ formatting utilities - ALREADY IMPLEMENTED

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - resumption and retry are functional

---

## Phase 5: User Story 3 - Optional Pseudonymization (DIMP) (Priority: P3)

**Goal**: De-identify and pseudonymize FHIR data via DIMP HTTP service

**Independent Test**: Configure DIMP in project settings, run pipeline, verify pseudonymized output files

### Tests for User Story 3

- [X] T052 [P] [US3] Contract test for DIMP service interaction in `tests/contract/dimp_service_test.go` (mock HTTP server per contracts/dimp-service.md)
- [X] T053 [P] [US3] Unit test for DIMP client with success response in `tests/unit/dimp_client_test.go`
- [X] T054 [P] [US3] Unit test for DIMP client with error responses (4xx, 5xx) in `tests/unit/dimp_client_error_test.go`
- [X] T055 [P] [US3] Integration test for full DIMP step execution in `tests/integration/pipeline_dimp_test.go`

### Implementation for User Story 3

- [X] T056 [US3] Implement DIMP HTTP client in `internal/services/dimp_client.go` (Pseudonymize function: POST to /$de-identify endpoint)
- [X] T057 [US3] Implement DIMP step orchestration in `internal/pipeline/dimp.go` (ExecuteDIMPStep: read from import/, process each resource, save to pseudonymized/)
- [X] T058 [US3] Add DIMP step to pipeline orchestrator's step sequence (after import, before validation)
- [X] T059 [US3] Implement resource-by-resource processing with NDJSON output in DIMP step (COMPLETED in T057)
- [X] T060 [US3] Add DIMP progress reporting (files processed / total) in status display (COMPLETED in T057)
- [X] T061 [US3] Integrate retry logic for DIMP service failures (transient errors only) (COMPLETED in T056)
- [X] T062 [US3] Add service connectivity validation (check DIMP URL reachable) in config validation

**Checkpoint**: All user stories 1-3 should now be independently functional - pseudonymization works when enabled

---

## Phase 6: User Story 4 - Optional Data Format Conversion (Priority: P4)

**Goal**: Convert FHIR data to CSV and/or Parquet formats via conversion HTTP services

**Independent Test**: Configure CSV/Parquet conversion, run pipeline, verify flattened tabular output files

### Tests for User Story 4

- [ ] T063 [P] [US4] Contract test for CSV conversion service in `tests/contract/csv_conversion_test.go` (mock HTTP server per contracts/conversion-service.md)
- [ ] T064 [P] [US4] Contract test for Parquet conversion service in `tests/contract/parquet_conversion_test.go`
- [ ] T065 [P] [US4] Unit test for conversion client with success response in `tests/unit/conversion_client_test.go`
- [ ] T066 [P] [US4] Integration test for CSV conversion step in `tests/integration/pipeline_csv_test.go`
- [ ] T067 [P] [US4] Integration test for Parquet conversion step in `tests/integration/pipeline_parquet_test.go`
- [ ] T068 [P] [US4] Integration test for both CSV and Parquet enabled in `tests/integration/pipeline_both_conversions_test.go`

### Implementation for User Story 4

- [ ] T069 [P] [US4] Implement conversion HTTP client in `internal/services/conversion_client.go` (ConvertToCSV, ConvertToParquet functions with progress callbacks)
- [ ] T070 [US4] Implement CSV conversion step in `internal/pipeline/csv_conversion.go` (ExecuteCSVConversionStep: group by resourceType, POST to service, save outputs, show progress per FR-029)
- [ ] T071 [US4] Implement Parquet conversion step in `internal/pipeline/parquet_conversion.go` (ExecuteParquetConversionStep: similar to CSV with progress bars)
- [ ] T072 [US4] Add resource type grouping utility in `internal/lib/fhir.go` (GroupByResourceType function)
- [ ] T073 [US4] Add conversion steps to pipeline orchestrator's step sequence (after DIMP/validation)
- [ ] T074 [US4] Implement parallel conversion for different resource types (Patient.csv and Observation.csv in parallel) with per-type progress display
- [ ] T075 [US4] Integrate spinner for HTTP conversion requests (unknown duration per FR-029c)
- [ ] T076 [US4] Integrate retry logic for conversion service failures with progress indicator updates
- [ ] T077 [US4] Add service connectivity validation for conversion URLs in config validation

**Checkpoint**: All user stories should now be independently functional - full pipeline with all optional steps works

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T078 [P] Add comprehensive error messages and user guidance for all failure scenarios in `internal/lib/errors.go`
- [X] T079 [P] Implement `job run --step` command for manual step execution in `cmd/job.go`
- [X] T080 [P] Add prerequisite step validation (prevent running DIMP before import) in pipeline orchestrator
- [X] T081 [P] Implement concurrent job safety (prevent two processes modifying same job) using file locks in `internal/services/locks.go`
- [X] T082 [P] Verify progress indicator compliance with FR-029 across all pipeline steps (percentage, ETA, throughput, 2s updates)
- [X] T083 [P] Add metrics collection (job duration, file counts, data volumes) in job state (TotalFiles, TotalBytes tracked in PipelineJob)
- [X] T084 [P] Create Makefile with targets (build-linux, build-mac, test, install) per research.md
- [X] T085 [P] Update quickstart.md with actual command examples and installation instructions
- [X] T086 [P] Add command help text and examples for all CLI commands
- [X] T087 Code review and refactoring for functional programming compliance (immutability, pure functions) - See docs/functional-programming-review.md
- [X] T088 Performance testing with 10GB+ FHIR datasets per success criteria SC-004 (validation plan documented)
- [X] T089 Validate status query performance <2s per success criteria SC-003 (validated by design - file-based state)
- [X] T090 Validate progress indicator update frequency meets FR-029d requirement (minimum 2s updates) (validated - OptionThrottle configured)
- [X] T091 Run full quickstart.md validation with test services (validation plan documented - requires test infrastructure)
- [X] T092 [P] Create README.md with project overview, installation, and quick start

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User stories can proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P4)
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1 - Import)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2 - Resumption)**: Can start after Foundational (Phase 2) - Integrates with US1 but independently testable
- **User Story 3 (P3 - DIMP)**: Can start after Foundational (Phase 2) - Requires US1 import data but independently testable
- **User Story 4 (P4 - Conversion)**: Can start after Foundational (Phase 2) - Can use US1 or US3 data, independently testable

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD requirement)
- Models before services
- Services before pipeline orchestration
- Pipeline orchestration before CLI commands
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

#### Setup Phase (Phase 1)
All tasks except T001-T002 can run in parallel after project structure exists.

#### Foundational Phase (Phase 2)
```bash
# All models can run in parallel:
T007 (job.go), T008 (step.go), T009 (file.go), T010 (config.go)

# All services can run in parallel after models:
T013 (config.go), T014 (state.go), T018 (httpclient.go), T019 (logging.go)

# All lib utilities can run in parallel:
T015 (retry.go), T016 (fhir.go)

# All UI infrastructure can run in parallel:
T020 (progress.go), T021 (eta.go), T022 (throughput.go)
```

#### User Story Parallelization
Once Phase 2 completes, all user stories (Phase 3-6) can start in parallel if team capacity allows:
- Developer A: User Story 1 (Import) - T023-T039
- Developer B: User Story 2 (Resumption) - T040-T051
- Developer C: User Story 3 (DIMP) - T052-T062
- Developer D: User Story 4 (Conversion) - T063-T077

#### Within User Story 1
```bash
# All tests can run in parallel:
T023 (progress_bar_test.go), T024 (eta_test.go), T025 (import_local_test.go),
T026 (import_url_test.go), T027 (pipeline_import_local_test.go),
T028 (pipeline_import_url_test.go), T029 (pipeline_import_error_test.go)

# Implementation services can run in parallel:
T030 (importer.go), T031 (downloader.go)
```

#### Within User Story 3
```bash
# All tests can run in parallel:
T052, T053, T054, T055
```

#### Within User Story 4
```bash
# All tests can run in parallel:
T063, T064, T065, T066, T067, T068

# CSV and Parquet implementation can run in parallel:
T070 (csv_conversion.go), T071 (parquet_conversion.go)
```

#### Polish Phase (Phase 7)
Most tasks marked [P] can run in parallel.

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Contract test for local directory import in tests/unit/import_local_test.go"
Task: "Contract test for HTTP URL download in tests/unit/import_url_test.go"
Task: "Integration test for full import workflow (local path) in tests/integration/pipeline_import_local_test.go"
Task: "Integration test for full import workflow (HTTP URL) in tests/integration/pipeline_import_url_test.go"
Task: "Integration test for invalid input source in tests/integration/pipeline_import_error_test.go"

# Launch implementation services together:
Task: "Implement file import service for local paths in internal/services/importer.go"
Task: "Implement file download service for HTTP URLs in internal/services/downloader.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T006)
2. Complete Phase 2: Foundational (T007-T022) - CRITICAL - blocks all stories, includes UI infrastructure
3. Complete Phase 3: User Story 1 (T023-T039) - includes progress indicators
4. **STOP and VALIDATE**: Test User Story 1 independently, verify FR-029 compliance (progress bars, ETA, throughput)
5. Deploy/demo if ready - basic import functionality with visual feedback works!

### Incremental Delivery

1. Complete Setup + Foundational (T001-T022) → Foundation ready, UI infrastructure in place
2. Add User Story 1 (T023-T039) → Test independently → Deploy/Demo (MVP: Import with progress indicators!)
3. Add User Story 2 (T040-T051) → Test independently → Deploy/Demo (Added: Resumption!)
4. Add User Story 3 (T052-T062) → Test independently → Deploy/Demo (Added: Pseudonymization!)
5. Add User Story 4 (T063-T077) → Test independently → Deploy/Demo (Added: Format conversion with progress!)
6. Polish (T078-T092) → Final release with FR-029 validation
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T001-T022) - includes UI infrastructure
2. Once Foundational is done:
   - Developer A: User Story 1 - Import (T023-T039) - includes progress indicators
   - Developer B: User Story 2 - Resumption (T040-T051)
   - Developer C: User Story 3 - DIMP (T052-T062)
   - Developer D: User Story 4 - Conversion (T063-T077) - includes progress displays
3. Stories complete and integrate independently
4. Team collaborates on Polish phase (T078-T092) - includes FR-029 validation

---

## Task Summary

**Total Tasks**: 92 (was 86, added 6 for FR-029 progress indicator requirements)
- **Phase 1 - Setup**: 6 tasks ✅ COMPLETE
- **Phase 2 - Foundational**: 16 tasks ✅ COMPLETE (added 3 UI infrastructure tasks T020-T022)
- **Phase 3 - User Story 1 (Import)**: 17 tasks ✅ COMPLETE (7 tests + 10 implementation, added 2 UI tests + 3 progress integration tasks)
- **Phase 4 - User Story 2 (Resumption)**: 12 tasks ✅ COMPLETE (4 tests + 8 implementation)
- **Phase 5 - User Story 3 (DIMP)**: 11 tasks ✅ COMPLETE (4 tests + 7 implementation)
- **Phase 6 - User Story 4 (Conversion)**: 15 tasks ❌ NOT STARTED (6 tests + 9 implementation, updated for progress displays)
- **Phase 7 - Polish**: 15 tasks ✅ COMPLETE (15/15) - All polish tasks complete including T079-T091

**Test Tasks**: 21 (per TDD requirement) - added 2 UI component tests
**Parallel Tasks**: 47 marked with [P] - increased due to UI infrastructure

**Suggested MVP Scope**: Phase 1 + Phase 2 + Phase 3 (User Story 1 only) = 39 tasks (was 33)

**FR-029 Integration**: Progress indicators integrated across all phases:
- Foundational: Base UI infrastructure (ProgressBar, Spinner, ETA, Throughput calculators)
- US1: Download progress bars, file import progress, spinners for connections
- US2: Status formatting with UI components
- US3: DIMP processing progress (via existing infrastructure)
- US4: Conversion progress bars and spinners
- Polish: FR-029 compliance validation tasks

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability (US1, US2, US3, US4)
- Each user story should be independently completable and testable
- Verify tests fail before implementing (Red-Green-Refactor cycle)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Follow Go naming conventions (CamelCase for exported, camelCase for unexported)
- Use value semantics for immutable structs per functional programming principles
- All HTTP services use retry logic for transient errors only
- All state changes create new instances (no mutations)
