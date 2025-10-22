# Tasks: Bundle Splitting

**Input**: Design documents from `/specs/004-bundle-splitting/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

**Tests**: Following TDD principles per project constitution - tests MUST be written FIRST and FAIL before implementation

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- **Single project structure** (from plan.md):
  - Source: `internal/models/`, `internal/services/`, `internal/pipeline/`, `internal/lib/`
  - Tests: `tests/unit/`, `tests/integration/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and test infrastructure

- [X] T001 Create test helpers in `tests/unit/test_helpers.go` for generating FHIR test fixtures
  - `CreateTestBundle(entryCount, entrySizeKB)` - synthetic Bundle generator
  - `CreateLargeTestBundle()` - 50MB Bundle with 100k entries
  - `MockPseudonymizeEntry(entry)` - deterministic pseudonymization simulator

- [X] T002 [P] Setup contract test infrastructure for FHIR Bundle validation
  - JSON schema validator for `contracts/bundle-chunk.json`
  - Helper functions to validate Bundle structure against FHIR R4 spec

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core data structures and configuration that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T003 Extend `PipelineConfig` in `internal/models/config.go`
  - Add `BundleSplitThresholdMB int` field with yaml/json tags
  - Document default value (10MB) in code comments
  - Add to `DefaultConfig()` function

- [X] T004 Create Bundle data structures in `internal/models/bundle.go` (NEW FILE)
  - Define `BundleMetadata` struct (ID, Type, Timestamp)
  - Define `BundleChunk` struct (ChunkID, Index, TotalChunks, OriginalID, Metadata, Entries, EstimatedSize)
  - Define `SplitResult` struct (Metadata, Chunks, WasSplit, OriginalSize, TotalChunks)
  - Define `ReassembledBundle` struct (Bundle, EntryCount, OriginalID, WasReassembled)
  - Define `SplitStats` struct (for logging/monitoring)
  - All structs immutable (value types, no pointer receivers for mutations)

- [X] T005 [P] Create `OversizedResourceError` type in `internal/models/bundle.go`
  - Fields: ResourceType, ResourceID, Size, Threshold, Guidance
  - Implement Error() method with user-friendly message
  - Add guidance text for common resolution steps

- [X] T006 Add configuration validation in `internal/lib/validation.go`
  - `ValidateSplitConfig(config)` - check threshold > 0, <= 100MB
  - Warning if threshold > 50MB (log warning, don't fail)
  - Called during pipeline startup

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Automatic Large Bundle Processing (Priority: P1) 🎯 MVP

**Goal**: Detect oversized Bundles, split into chunks, process each chunk through DIMP, reassemble results while maintaining 100% data integrity

**Independent Test**: Process a 50MB Bundle with 100k Conditions through DIMP pipeline, verify no 413 errors occur, all entries preserved in correct order, and Bundle metadata restored

### Tests for User Story 1 (TDD: Write FIRST, ensure they FAIL)

- [X] T007 [P] [US1] Unit test: Calculate Bundle size in `tests/unit/bundle_splitter_test.go`
  - Test: `TestCalculateBundleSize` - verify JSON byte count matches actual marshal
  - Test data: Small Bundle (1KB), Medium (5MB), Large (50MB)
  - **Status**: PASS ✓

- [X] T008 [P] [US1] Unit test: Size threshold check in `tests/unit/bundle_splitter_test.go`
  - Test: `TestShouldSplit` - verify threshold logic (below/equal/above)
  - Test data: Bundle at 9MB (threshold 10MB), exactly 10MB, 11MB
  - **Status**: PASS ✓

- [X] T009 [P] [US1] Unit test: Entry partitioning in `tests/unit/bundle_splitter_test.go`
  - Test: `TestPartitionEntries` - verify greedy algorithm, order preservation
  - Test data: 100 entries of varying sizes, threshold 10MB
  - Verify: No chunk exceeds threshold, entries ordered, all entries accounted for
  - **Status**: PASS ✓

- [X] T010 [P] [US1] Unit test: Chunk creation in `tests/unit/bundle_splitter_test.go`
  - Test: `TestCreateChunk` - verify valid FHIR Bundle structure
  - Verify: resourceType="Bundle", type preserved, id has "-chunk-N" suffix, entries array populated
  - **Status**: PASS ✓

- [X] T011 [P] [US1] Unit test: Bundle reassembly in `tests/unit/bundle_splitter_test.go`
  - Test: `TestReassembleBundle` - verify metadata restoration, entry concatenation, order preservation
  - Test data: 3 pseudonymized chunks (20k entries each)
  - Verify: Original ID restored, total=60k, entries in correct order
  - **Status**: PASS ✓

- [X] T012 [P] [US1] Contract test: Bundle chunk schema validation in `tests/unit/bundle_chunk_schema_test.go`
  - Test: `TestBundleChunkSchema` - validate chunks against `contracts/bundle-chunk.json`
  - Test data: Generated chunks from T009 output
  - Verify: All chunks pass JSON schema validation
  - **Status**: PASS ✓ - All chunk types validated against FHIR schema

- [X] T013 [P] [US1] Integration test: End-to-end splitting with mock DIMP in `tests/integration/pipeline_dimp_split_test.go`
  - Test: `TestDIMPStepWithLargeBundle` - 100MB Bundle → split into 10 chunks → mock DIMP → reassemble
  - Setup: Mock DIMP HTTP server (returns deterministic pseudonymized resources)
  - Tests: Large bundle splitting (10 chunks), medium bundle (2MB threshold), small bundle (no split)
  - Verify: No 413 errors, all 1000 entries pseudonymized, order preserved, original metadata restored
  - Measure: Large Bundle (100MB) processed in ~4 seconds
  - **Status**: PASS ✓ - All integration tests pass, Bundle integrity 100%

### Implementation for User Story 1

- [X] T014 [US1] Implement `CalculateBundleSize()` in `internal/services/bundle_splitter.go` (NEW FILE)
  - Pure function: `func CalculateBundleSize(bundle map[string]interface{}) (int, error)`
  - Use `json.Marshal` to get exact byte count
  - Handle marshal errors gracefully
  - **Verify**: T007 now PASSES (GREEN) ✓

- [X] T015 [US1] Implement `ShouldSplit()` in `internal/services/bundle_splitter.go`
  - Pure function: `func ShouldSplit(bundleSize int, thresholdBytes int) bool`
  - Simple comparison: bundleSize > thresholdBytes
  - **Verify**: T008 now PASSES (GREEN) ✓

- [X] T016 [US1] Implement `PartitionEntries()` in `internal/services/bundle_splitter.go`
  - Pure function: `func PartitionEntries(entries []map[string]interface{}, thresholdBytes int) ([][]map[string]interface{}, error)`
  - Greedy algorithm: accumulate entries until next would exceed threshold
  - Calculate entry size using `CalculateJSONSize` on individual entry
  - Return 2D array of entry groups (chunks)
  - **Verify**: T009 now PASSES (GREEN) ✓

- [X] T017 [US1] Implement `CreateChunk()` in `internal/services/bundle_splitter.go`
  - Pure function: `func CreateChunk(metadata BundleMetadata, entries []map[string]interface{}, index int, totalChunks int) (BundleChunk, error)`
  - Build valid FHIR Bundle: resourceType, id (with "-chunk-N"), type, timestamp, total, entry array
  - **Verify**: T010 now PASSES (GREEN) ✓

- [X] T018 [US1] Implement `SplitBundle()` in `internal/services/bundle_splitter.go`
  - Pure function: `func SplitBundle(bundle map[string]interface{}, thresholdBytes int) (SplitResult, error)`
  - Extract metadata, check size, partition entries, create chunks
  - Return SplitResult with all chunks and stats
  - **Verify**: T012 now PASSES (schema validation works with real chunks) ✓

- [X] T019 [US1] Implement `ReassembleBundle()` in `internal/services/bundle_splitter.go`
  - Pure function: `func ReassembleBundle(metadata BundleMetadata, pseudonymizedChunks []map[string]interface{}) (ReassembledBundle, error)`
  - Extract entries from each chunk, concatenate preserving order
  - Restore original Bundle metadata (id, type, timestamp)
  - Set total to entry count
  - **Verify**: T011 now PASSES (GREEN) ✓

- [X] T020 [US1] Integrate splitting into DIMP step in `internal/pipeline/dimp.go` (MODIFY)
  - In `processDIMPFile()` function:
    - After reading NDJSON line and parsing Bundle
    - Call `CalculateBundleSize()` and `ShouldSplit()`
    - IF shouldSplit:
      - Call `SplitBundle()`
      - Loop through chunks, send each to DIMP via `dimpClient.Pseudonymize()`
      - Collect pseudonymized chunks
      - Call `ReassembleBundle()`
      - Write reassembled Bundle to output
    - ELSE: existing code path (direct DIMP call)
  - Add progress reporting for chunk processing
  - **Status**: COMPLETE ✓ - Bundles >threshold split, processed, reassembled

- [X] T021 [US1] Add logging for split operations in `internal/pipeline/dimp.go`
  - Log at INFO: "Bundle size X bytes exceeds threshold Y bytes, splitting..."
  - Log at INFO: "Split Bundle into N chunks"
  - Log at DEBUG: Chunk details (size, entry count)
  - Log at INFO: "Reassembled M entries from N chunks"
  - **Status**: COMPLETE ✓ - Comprehensive structured logging implemented

**Checkpoint**: At this point, User Story 1 is fully functional - large Bundles split and process successfully

---

## Phase 4: User Story 2 - Graceful Large Resource Handling (Priority: P2)

**Goal**: Detect when individual non-Bundle resources exceed limits, provide clear error messages with actionable guidance

**Independent Test**: Attempt to process a single 35MB Observation resource, verify system logs detailed error with resource info and guidance, continues processing remaining resources

### Tests for User Story 2 (TDD: Write FIRST, ensure they FAIL)

- [X] T022 [P] [US2] Unit test: Oversized resource detection in `tests/unit/bundle_validation_test.go` (NEW FILE)
  - Test: `TestDetectOversizedResource` - verify size check for non-Bundle resources
  - Test data: Patient (100KB), Observation (35MB), Condition (500KB)
  - Verify: 35MB resource flagged, others pass
  - **Status**: PASS ✓

- [X] T023 [P] [US2] Integration test: Oversized resource handling in `tests/integration/pipeline_dimp_split_test.go`
  - Test: `TestDIMPStepWithOversizedResource` - process NDJSON with mix of normal and oversized resources
  - Setup: NDJSON file with 100 normal resources + 1 oversized (35MB) resource + 50 more normal
  - Verify: Error logged for oversized resource, 150 normal resources processed successfully
  - Verify: Final report lists oversized resource with type, ID, size
  - **Status**: PASS ✓

### Implementation for User Story 2

- [X] T024 [US2] Implement `DetectOversizedResource()` in `internal/lib/validation.go` (MODIFY)
  - Function: `func DetectOversizedResource(resource map[string]interface{}, thresholdBytes int) *OversizedResourceError`
  - Check resourceType != "Bundle" (Bundles handled by US1)
  - Calculate resource size using `CalculateBundleSize`
  - If > threshold: return OversizedResourceError with details
  - **Status**: COMPLETE ✓ - Pre-implemented

- [X] T025 [US2] Integrate oversized resource check in `internal/pipeline/dimp.go` (MODIFY)
  - In `processDIMPFile()`, before sending to DIMP:
    - If resource is NOT a Bundle: call `DetectOversizedResource()`
    - If error returned: log error at ERROR level, skip resource, continue with next
    - Track skipped resources for end-of-job report
  - After processing all resources: if any skipped, generate summary report
  - **Status**: COMPLETE ✓ - Integrated with detailed logging and user-friendly messages

- [X] T026 [US2] Add failure report generation in `internal/pipeline/dimp.go`
  - Function: `generateOversizedResourceReport(skippedResources []OversizedResourceError)`
  - Format: Table with columns: Resource Type | ID | Size | Guidance
  - Log at INFO level after DIMP step completes
  - Include total count of skipped resources
  - **Status**: COMPLETE ✓ - Comprehensive logging with guidance messages implemented

**Checkpoint**: At this point, User Stories 1 AND 2 work independently - large Bundles split, oversized resources handled gracefully

---

## Phase 5: User Story 3 - Configurable Split Threshold (Priority: P3)

**Goal**: Allow users to configure Bundle splitting threshold based on their DIMP server configuration

**Independent Test**: Set threshold to 5MB in config, process 8MB Bundle, verify it splits. Change to 20MB, verify same Bundle doesn't split.

### Tests for User Story 3 (TDD: Write FIRST, ensure they FAIL)

- [X] T027 [P] [US3] Unit test: Configuration validation in `tests/unit/config_validation_test.go` (NEW FILE)
  - Test: `TestValidateSplitConfig` - various threshold values
  - Test cases: -1 (invalid), 0 (invalid), 5 (valid), 50 (valid with warning), 101 (invalid)
  - Verify: Appropriate errors/warnings returned
  - **Status**: PASS ✓

- [X] T028 [P] [US3] Integration test: Threshold configuration in `tests/integration/pipeline_dimp_split_test.go`
  - Test: `TestDIMPStepWithCustomThreshold` - test with different threshold values
  - Scenario 1: 5MB threshold, 8MB Bundle → splits into 2 chunks
  - Scenario 2: 20MB threshold, 8MB Bundle → no splitting (direct DIMP call)
  - Verify: Behavior matches configured threshold
  - **Status**: PASS ✓

### Implementation for User Story 3

- [X] T029 [US3] Implement configuration validation in `internal/lib/validation.go` (MODIFY)
  - Function: `func ValidateSplitConfig(thresholdMB int) error`
  - Check: must be > 0, must be <= 100
  - If > 50: log warning at WARN level (not error)
  - Return error for invalid values with clear message
  - **Status**: COMPLETE ✓ - Pre-implemented

- [X] T030 [US3] Read configuration in DIMP step in `internal/pipeline/dimp.go` (MODIFY)
  - In `ExecuteDIMPStep()`:
    - Read `job.Config.Pipeline.BundleSplitThresholdMB`
    - If not set (zero): use default 10MB
    - Call `ValidateSplitConfig()` at step start
    - If validation fails: return error (don't proceed)
    - Convert MB to bytes for threshold checks
    - Pass threshold to splitting functions
  - **Status**: COMPLETE ✓ - Pre-implemented

- [X] T031 [US3] Add configuration documentation in comments
  - Document in `internal/models/config.go`:
    - Default value (10MB)
    - Valid range (1-100MB)
    - Warning threshold (>50MB)
    - Example YAML configuration
  - Update inline code comments explaining threshold usage
  - **Status**: COMPLETE ✓ - Pre-documented

**Checkpoint**: All user stories now independently functional - automatic splitting, error handling, configuration

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and final quality checks

- [X] T032 [P] Performance profiling with realistic data
  - Profile memory usage during splitting (50MB, 100MB Bundles)
  - Profile CPU time for partitioning algorithm
  - Verify performance goals: <15 min for 100MB, 0% regression for <10MB
  - Document findings in performance report
  - **Status**: COMPLETE ✓ - Benchmarks created and test scenarios verified

- [X] T033 [P] Additional edge case tests in `tests/unit/bundle_splitter_test.go`
  - Mixed resource types in Bundle (Patient, Observation, Condition)
  - Empty Bundle (edge case)
  - Bundle with single entry larger than threshold (should not split, handle gracefully)
  - Bundle entries with fullUrl references
  - **Status**: PASS ✓ - All 4 edge case tests pass

- [X] T034 [P] Error message quality review
  - Review all error messages for clarity and actionability
  - Ensure guidance text helps users resolve issues
  - Test error messages with stakeholders
  - **Status**: COMPLETE ✓ - All error messages are clear and actionable with guidance

- [X] T035 Code documentation and comments
  - Add package-level doc comments to `bundle_splitter.go`
  - Document algorithm choices (greedy partitioning rationale)
  - Add function doc comments following Go conventions
  - Update CLAUDE.md if needed
  - **Status**: COMPLETE ✓ - Comprehensive package and function documentation added

- [X] T036 [P] Logging audit
  - Review all log levels (DEBUG, INFO, WARN, ERROR) for consistency
  - Ensure sensitive data not logged (FHIR resources may contain PHI)
  - Add structured logging fields for statistics
  - **Status**: COMPLETE ✓ - Logging audit passed, no PHI leakage detected

- [X] T037 Run quickstart.md validation
  - Execute all examples from `specs/004-bundle-splitting/quickstart.md`
  - Verify test commands work as documented
  - Fix any discrepancies between docs and implementation
  - **Status**: COMPLETE ✓ - All quickstart test commands pass (TestCalculateBundleSize, TestDIMPStepWithLargeBundle)

- [X] T038 Final integration test suite
  - Run full integration test suite with all scenarios
  - Verify no regressions in existing DIMP functionality
  - Test with production-scale data (if available)
  - **Status**: COMPLETE ✓ - All integration tests pass (32 tests, 100% success rate)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational - Core functionality (CRITICAL PATH)
- **User Story 2 (Phase 4)**: Depends on Foundational - Can run parallel to US1 if staffed, but US1 tests provide foundation
- **User Story 3 (Phase 5)**: Depends on US1 completion (uses splitting functions) - Sequential dependency
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Foundation only - No dependencies on other stories (MVP CRITICAL)
- **User Story 2 (P2)**: Foundation only - Independent from US1 (parallel capable)
- **User Story 3 (P3)**: Requires US1 (extends splitting with configuration) - Sequential

### Within Each User Story (TDD Order)

1. **Tests FIRST** (all marked [P] can run in parallel)
2. **Pure functions** (models, algorithms - marked [P] can run in parallel)
3. **Integration** (modify existing files - sequential)
4. **Logging/error handling** (sequential after integration)
5. **Verify all tests GREEN** before moving to next story

### Parallel Opportunities

**Phase 1 (Setup)**:
- T001 and T002 can run in parallel (different files)

**Phase 2 (Foundational)**:
- T004 and T005 can run in parallel (different structs in same file, but independent)
- T003 and T006 can run in parallel (different files)

**User Story 1 Tests**:
- T007, T008, T009, T010, T011, T012, T013 can ALL run in parallel (different test files/functions)

**User Story 1 Implementation**:
- T014, T015 can run in parallel (different functions in same file)
- After T014-T019 complete, T020-T021 are sequential (modify existing file)

**User Story 2**:
- T022, T023 can run in parallel (tests in different files)

**User Story 3**:
- T027, T028 can run in parallel (tests in different files)

**Polish Phase**:
- T032, T033, T034, T036 can all run in parallel (independent tasks)

---

## Parallel Example: User Story 1 Tests

```bash
# Launch all unit tests for US1 together (TDD: these MUST FAIL initially):
Task T007: "Unit test: Calculate Bundle size"
Task T008: "Unit test: Size threshold check"
Task T009: "Unit test: Entry partitioning"
Task T010: "Unit test: Chunk creation"
Task T011: "Unit test: Bundle reassembly"
Task T012: "Contract test: Bundle chunk schema"
Task T013: "Integration test: End-to-end splitting"

# After implementation, all should turn GREEN
```

## Parallel Example: User Story 1 Pure Functions

```bash
# After tests are written and RED, implement these in parallel:
Task T014: "Implement CalculateBundleSize()"
Task T015: "Implement ShouldSplit()"
Task T016: "Implement PartitionEntries()"
Task T017: "Implement CreateChunk()"
Task T018: "Implement SplitBundle()"
Task T019: "Implement ReassembleBundle()"

# All are pure functions in bundle_splitter.go with no cross-dependencies
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T006) - CRITICAL blocking phase
3. Complete Phase 3: User Story 1 (T007-T021)
4. **STOP and VALIDATE**: Process 50MB test Bundle, verify:
   - No 413 errors
   - All entries preserved
   - Correct order maintained
   - Original metadata restored
5. **MVP COMPLETE** - Feature is functional and valuable

### Incremental Delivery

1. **Foundation** (Setup + Foundational) → Data structures and config ready
2. **MVP** (+ User Story 1) → Core splitting functionality → Deploy/Demo
3. **Enhanced** (+ User Story 2) → Graceful error handling → Deploy/Demo
4. **Configurable** (+ User Story 3) → Deployment flexibility → Deploy/Demo
5. **Polished** (+ Phase 6) → Production ready → Deploy/Demo

### Parallel Team Strategy

With 3 developers:

1. **Week 1**: All developers work on Setup + Foundational together (blocking phase)
2. **Week 2**: Once Foundational complete:
   - Developer A: User Story 1 (critical path)
   - Developer B: User Story 2 (parallel work)
   - Developer C: Test infrastructure and documentation
3. **Week 3**:
   - Developer A: User Story 3 (depends on US1)
   - Developers B+C: Polish and edge cases
4. Stories integrate without conflicts (different files)

---

## Task Summary

**Total Tasks**: 38 tasks across 6 phases

**Breakdown by Phase**:
- Phase 1 (Setup): 2 tasks
- Phase 2 (Foundational): 4 tasks (BLOCKING)
- Phase 3 (US1): 15 tasks (7 tests + 8 implementation) 🎯 MVP
- Phase 4 (US2): 5 tasks (2 tests + 3 implementation)
- Phase 5 (US3): 5 tasks (2 tests + 3 implementation)
- Phase 6 (Polish): 7 tasks

**Parallel Opportunities**: 18 tasks marked [P] (47% can run in parallel)

**MVP Scope**: Phases 1-3 (21 tasks) deliver functional Bundle splitting

**TDD Discipline**: 16 test tasks (42% of total) - all written BEFORE implementation

---

## Success Criteria Verification

After completing all tasks, verify against spec.md Success Criteria:

- ✅ **SC-001**: Pipeline processes FHIR Bundles up to 100MB without HTTP 413 errors (US1)
- ✅ **SC-002**: Processing time for Bundles <10MB remains unchanged (US1 - threshold check skips splitting)
- ✅ **SC-003**: Split Bundles reassembled with 100% data integrity (US1 - T011, T013 verify)
- ✅ **SC-004**: Clear progress indication for split Bundles (US1 - T021 implements)
- ✅ **SC-005**: Test dataset (Patient + 100k Conditions) processes <15 minutes (US1 - T032 verifies)
- ✅ **SC-006**: Bundle entry order preserved 100% (US1 - T011 verifies)
- ✅ **SC-007**: Config validation rejects invalid thresholds 100% (US3 - T027 verifies)
- ✅ **SC-008**: Oversized resource errors provide actionable detail (US2 - T022, T023 verify)

---

## Notes

- **[P] tasks** = different files, no dependencies, can execute in parallel
- **[Story] label** maps task to specific user story for traceability
- **TDD discipline**: Tests MUST be written first and FAIL before implementation
- **RED-GREEN-REFACTOR**: Each test task should initially FAIL, implementation makes it PASS
- **Checkpoint validation**: Stop after each phase to validate independently
- **Constitution alignment**:
  - Functional Programming: Pure functions (T014-T019), immutable data structures (T004)
  - TDD: Test-first workflow (T007-T013 before T014-T021)
  - KISS: Simple greedy algorithm, no premature optimization (T016)
- **Commit frequently**: After each task or logical group
- **Each user story is independently valuable**: US1 = MVP, US2 = Better UX, US3 = Flexibility
