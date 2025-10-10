# Tasks: TORCH Server Data Import

**Feature**: 002-import-via-torch | **Date**: 2025-10-10
**Input**: Design documents from `/specs/002-import-via-torch/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/torch-api.md ✓, quickstart.md ✓

**Tests**: ✅ Tests included - TDD approach committed per Constitution Principle II

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (Setup, Foundation, US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create test data directory structure for TORCH tests (test-data/torch/)
- [X] T002 [P] Copy example CRTDL file from dse-example to test-data/torch/example.crtdl
- [X] T003 [P] Update config/aether.example.yaml with TORCH configuration section

**Checkpoint**: Basic test infrastructure ready

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core models and configuration that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests for Foundation (TDD - Write and approve BEFORE implementation)

- [X] T004 [P] Contract test for TORCH API submission in tests/contract/torch_service_test.go
- [X] T005 [P] Contract test for TORCH API polling in tests/contract/torch_service_test.go
- [X] T006 [P] Contract test for TORCH API file download in tests/contract/torch_service_test.go
- [X] T007 [P] Unit test for TORCHConfig validation in tests/unit/config_loading_test.go
- [X] T008 [P] Unit test for InputType detection in tests/unit/input_detection_test.go
- [X] T009 [P] Unit test for CRTDL syntax validation in tests/unit/crtdl_validation_test.go

### Implementation for Foundation

- [X] T010 Add TORCHConfig struct to internal/models/config.go
- [X] T011 Add TORCHConfig to ServiceConfig struct in internal/models/config.go
- [X] T012 Add TORCHConfig.Validate() method in internal/models/config.go
- [X] T013 [P] Add InputTypeCRTDL and InputTypeTORCHURL constants to internal/models/job.go
- [X] T014 [P] Add TORCHExtractionURL field to PipelineJob struct in internal/models/job.go
- [X] T015 Create DetectInputType() function in internal/lib/validation.go
- [X] T016 Create IsCRTDLFile() helper function in internal/lib/validation.go
- [X] T017 Create ValidateCRTDLSyntax() function in internal/lib/validation.go
- [X] T018 Update DefaultConfig() in internal/models/config.go to include TORCH defaults
- [X] T019 Extend ValidateServiceConnectivity() in internal/models/validation.go to check TORCH connectivity

**Checkpoint**: Foundation ready - all user stories can now proceed

---

## Phase 3: User Story 1 - Data Extraction with CRTDL File (Priority: P1) 🎯 MVP

**Goal**: Enable researchers to extract FHIR data from TORCH using CRTDL files

**Independent Test**: Provide CRTDL file path to `aether pipeline start --input /path/to/query.crtdl` and verify NDJSON data is retrieved from TORCH and stored locally

### Tests for User Story 1 (TDD - Write and approve BEFORE implementation)

- [X] T020 [P] [US1] Unit test for TORCH client SubmitExtraction() in tests/unit/torch_client_test.go
- [X] T021 [P] [US1] Unit test for TORCH client PollExtractionStatus() in tests/unit/torch_client_test.go
- [X] T022 [P] [US1] Unit test for TORCH client DownloadExtractionFiles() in tests/unit/torch_client_test.go
- [X] T023 [P] [US1] Unit test for base64 CRTDL encoding in tests/unit/torch_client_test.go
- [X] T024 [P] [US1] Unit test for exponential backoff polling logic in tests/unit/torch_client_test.go
- [X] T025 [US1] Integration test for CRTDL → extraction → download flow in tests/integration/pipeline_torch_test.go

### Implementation for User Story 1

- [X] T026 [P] [US1] Create TORCHClient struct in internal/services/torch_client.go
- [X] T027 [P] [US1] Create TORCHExtractionRequest struct in internal/services/torch_client.go
- [X] T028 [P] [US1] Create TORCHExtractionResult structs in internal/services/torch_client.go
- [X] T029 [P] [US1] Create TORCHParameter and related structs in internal/services/torch_client.go
- [X] T030 [US1] Implement NewTORCHClient() constructor in internal/services/torch_client.go
- [X] T031 [US1] Implement TORCHClient.SubmitExtraction() method in internal/services/torch_client.go
- [X] T032 [US1] Implement TORCHClient.PollExtractionStatus() with exponential backoff in internal/services/torch_client.go
- [X] T033 [US1] Implement TORCHClient.DownloadExtractionFiles() method in internal/services/torch_client.go
- [X] T034 [US1] Implement TORCHClient.Ping() connectivity check in internal/services/torch_client.go
- [X] T035 [US1] Implement encodeCRTDLToBase64() helper function in internal/services/torch_client.go
- [X] T036 [US1] Implement parseExtractionResult() helper function in internal/services/torch_client.go
- [X] T037 [US1] Implement buildBasicAuthHeader() helper function in internal/services/torch_client.go
- [X] T038 [US1] Create executeTORCHExtraction() function in internal/pipeline/import.go
- [X] T039 [US1] Update ExecuteImportStep() to handle InputTypeCRTDL case in internal/pipeline/import.go
- [X] T040 [US1] Update CreateJob() to call DetectInputType() in internal/pipeline/job.go
- [X] T041 [US1] Update CreateJob() to validate CRTDL syntax if detected in internal/pipeline/job.go
- [X] T042 [US1] Update CLI runPipelineStart() to handle CRTDL input in cmd/pipeline.go
- [X] T043 [US1] Add TORCH error types (ErrExtractionTimeout, ErrInvalidCRTDL, etc.) in internal/services/torch_client.go
- [X] T044 [US1] Add error wrapping with context for TORCH operations in internal/services/torch_client.go
- [X] T045 [US1] Add logging for TORCH extraction lifecycle events in internal/services/torch_client.go

**Checkpoint**: User Story 1 complete - CRTDL extraction works end-to-end

---

## Phase 4: User Story 2 - Backward Compatibility with Local Directories (Priority: P1)

**Goal**: Ensure existing local directory imports continue to work without changes

**Independent Test**: Run `aether pipeline start --input ./test-data/` and verify it works identically to current implementation

### Tests for User Story 2 (TDD - Write and approve BEFORE implementation)

- [X] T046 [P] [US2] Integration test for local directory input still works in tests/integration/pipeline_torch_test.go
- [X] T047 [P] [US2] Integration test for remote HTTP URL input still works in tests/integration/pipeline_torch_test.go
- [X] T048 [US2] Unit test that DetectInputType() correctly identifies directories in tests/unit/input_detection_test.go

### Implementation for User Story 2

- [X] T049 [US2] Verify ExecuteImportStep() preserves existing InputTypeLocal case in internal/pipeline/import.go
- [X] T050 [US2] Verify ExecuteImportStep() preserves existing InputTypeHTTP case in internal/pipeline/import.go
- [X] T051 [US2] Verify DetectInputType() returns InputTypeLocal for directories in internal/lib/validation.go
- [X] T052 [US2] Run full test suite to confirm no regression in existing functionality

**Checkpoint**: User Story 2 complete - backward compatibility verified

---

## Phase 5: User Story 3 - TORCH Server URL Input (Priority: P3)

**Goal**: Support direct TORCH result URLs to reuse existing extractions

**Independent Test**: Provide TORCH result URL (e.g., `http://localhost:8080/result/abc123`) and verify files download without new extraction

### Tests for User Story 3 (TDD - Write and approve BEFORE implementation)

- [X] T053 [P] [US3] Unit test for DetectInputType() detecting TORCH URLs in tests/unit/input_detection_test.go
- [X] T054 [US3] Integration test for direct TORCH URL download in tests/integration/pipeline_torch_test.go

### Implementation for User Story 3

- [X] T055 [US3] Implement executeTORCHDownload() function in internal/pipeline/import.go
- [X] T056 [US3] Update ExecuteImportStep() to handle InputTypeTORCHURL case in internal/pipeline/import.go
- [X] T057 [US3] Update DetectInputType() to recognize TORCH URL pattern (/fhir/) in internal/lib/validation.go
- [X] T058 [US3] Add logic to parse result URLs directly and download files in executeTORCHDownload()

**Checkpoint**: User Story 3 complete - direct TORCH URLs work

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T059 [P] Add documentation for TORCH configuration in README.md or docs/
- [ ] T060 [P] Update example config with detailed TORCH comments in config/aether.example.yaml
- [ ] T061 [P] Add progress bar integration for TORCH download phase (reuse existing progressbar)
- [ ] T062 Code review: verify functional purity per Constitution Principle I
- [ ] T063 Code review: verify all tests pass and follow TDD per Constitution Principle II
- [ ] T064 Code review: verify simplicity (KISS) per Constitution Principle III
- [ ] T065 [P] Add edge case handling for TORCH server unreachable
- [ ] T066 [P] Add edge case handling for TORCH timeout
- [ ] T067 [P] Add edge case handling for empty cohort results
- [ ] T068 [P] Add edge case handling for malformed CRTDL
- [ ] T069 [P] Add edge case handling for authentication failures
- [ ] T070 Run quickstart.md validation workflow
- [ ] T071 Performance test: verify TORCH connectivity check < 5 seconds
- [ ] T072 Performance test: verify CRTDL validation < 1 second
- [ ] T073 Integration test: verify polling timeout works correctly
- [ ] T074 Integration test: verify job resumption after process restart during polling
- [ ] T075 [P] Add telemetry/metrics for TORCH operations
- [ ] T076 Final code cleanup and refactoring
- [ ] T077 Update CLAUDE.md with TORCH-specific commands if needed

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - **BLOCKS all user stories**
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2) - Core TORCH functionality
- **User Story 2 (Phase 4)**: Depends on Foundational (Phase 2) - Can run in parallel with US1
- **User Story 3 (Phase 5)**: Depends on US1 completion (reuses TORCH client) - Lower priority
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Foundation complete → Can start immediately - No dependencies on other stories
- **User Story 2 (P1)**: Foundation complete → Can run in parallel with US1 - Independent testing of existing features
- **User Story 3 (P3)**: US1 complete (needs TORCH client) → Extends US1 with URL shortcut

### Within Each User Story

**CRITICAL TDD ORDERING**:
1. **Tests FIRST**: Write all test scenarios and ensure they FAIL
2. **User Approval**: Tests must be reviewed and approved before implementation
3. **Implementation**: Write code to make tests pass
4. **Verification**: Confirm all tests pass (RED → GREEN)
5. **Refactor**: Clean up code while keeping tests green

**Within Implementation**:
- Models/structs before services
- Services before pipeline integration
- Core logic before error handling
- Independent components marked [P] can run in parallel

### Parallel Opportunities

**Phase 1 (Setup)**: T002 and T003 can run in parallel

**Phase 2 (Foundation)**:
- All test tasks (T004-T009) can run in parallel
- Model extensions (T013-T014) can run in parallel after tests approved
- Validation functions (T015-T017) can run in parallel

**Phase 3 (US1)**:
- All test tasks (T020-T024) can run in parallel
- Struct definitions (T026-T029) can run in parallel after tests approved

**Phase 4 (US2)**:
- All test tasks (T046-T048) can run in parallel
- Can run entire US2 in parallel with US1 (different focus)

**Phase 5 (US3)**:
- Test tasks (T053-T054) can run in parallel

**Phase 6 (Polish)**:
- Documentation (T059-T060) can run in parallel
- Edge case handling (T065-T070) can run in parallel
- Performance tests (T071-T074) can run in parallel
- Telemetry and cleanup (T075-T076) can run in parallel

---

## Parallel Example: User Story 1 Tests

```bash
# Launch all test writing tasks for US1 together (before approval):
Task: "[US1] Unit test for TORCH client SubmitExtraction()"
Task: "[US1] Unit test for TORCH client PollExtractionStatus()"
Task: "[US1] Unit test for TORCH client DownloadExtractionFiles()"
Task: "[US1] Unit test for base64 CRTDL encoding"
Task: "[US1] Unit test for exponential backoff polling logic"

# After tests approved, launch struct definitions together:
Task: "[US1] Create TORCHClient struct"
Task: "[US1] Create TORCHExtractionRequest struct"
Task: "[US1] Create TORCHExtractionResult structs"
Task: "[US1] Create TORCHParameter and related structs"
```

---

## Implementation Strategy

### MVP First (User Story 1 + User Story 2)

1. Complete Phase 1: Setup (~15 minutes)
2. Complete Phase 2: Foundational (~2-3 hours with tests)
   - **STOP**: Get test approval before implementation
   - **VALIDATE**: All foundation tests pass
3. Complete Phase 3: User Story 1 (~4-6 hours with tests)
   - **STOP**: Get test approval before implementation
   - **VALIDATE**: CRTDL extraction works end-to-end
   - **DEPLOY/DEMO**: Core TORCH functionality ready
4. Complete Phase 4: User Story 2 (~1 hour)
   - **VALIDATE**: No regression in existing features
   - **DEPLOY/DEMO**: MVP complete with backward compatibility

**MVP Delivers**: Full TORCH CRTDL extraction + backward compatibility (US1 + US2)

### Incremental Delivery

1. **Foundation** → Tests approved → Implementation complete → All tests pass
2. **MVP (US1 + US2)** → Independent testing → Deploy/Demo (Core value delivered!)
3. **Enhancement (US3)** → Independent testing → Deploy/Demo (Convenience feature)
4. **Polish** → Final validation → Production ready

### Parallel Team Strategy

With multiple developers:

1. **Together**: Complete Setup + Foundational (write tests, get approval, implement)
2. **Split after Foundation**:
   - Developer A: User Story 1 (TORCH extraction)
   - Developer B: User Story 2 (backward compatibility validation)
3. **After US1 complete**:
   - Developer A or B: User Story 3 (TORCH URL shortcut)
4. **Together**: Polish and cross-cutting concerns

---

## Task Count Summary

- **Total Tasks**: 77
- **Setup Tasks**: 3
- **Foundation Tasks**: 16 (10 tests + 6 implementation)
- **User Story 1 Tasks**: 26 (6 tests + 20 implementation)
- **User Story 2 Tasks**: 7 (3 tests + 4 verification)
- **User Story 3 Tasks**: 6 (2 tests + 4 implementation)
- **Polish Tasks**: 19

**Parallel Opportunities**: 45 tasks marked [P] can run concurrently

**Test Tasks**: 21 test tasks (must complete BEFORE implementation per TDD)

**Independent Test Criteria**:
- US1: CRTDL file → TORCH extraction → NDJSON downloaded
- US2: Local directory → files processed (existing flow)
- US3: TORCH URL → NDJSON downloaded (no extraction)

**Suggested MVP Scope**: Phase 1 + Phase 2 + Phase 3 + Phase 4 (US1 + US2) = 52 tasks

---

## Notes

- **[P] tasks**: Different files or independent components, no dependencies
- **[Story] label**: Maps task to specific user story for traceability
- **TDD Critical**: Tests must FAIL before implementation (RED → GREEN → REFACTOR)
- **Constitution**: FP (immutability), TDD (tests first), KISS (simple patterns)
- **Each user story**: Independently completable and testable
- **Stop at checkpoints**: Validate story independently before proceeding
- **Commit frequently**: After each task or logical group
- **File paths**: All paths use internal/ and tests/ per existing project structure
- **Reuse patterns**: TORCH client mirrors dimp_client.go pattern
- **No breaking changes**: Backward compatibility is P1 requirement

---

## Next Steps

1. Review tasks with stakeholders
2. Get test scenarios approved (T004-T009, T020-T025, etc.)
3. Begin Phase 1: Setup
4. Proceed through phases sequentially or with parallel team
5. Stop at each checkpoint for independent validation
6. Target MVP = US1 + US2 for initial release
7. Add US3 as enhancement in follow-up release
