# Implementation Plan: TORCH Server Data Import

**Branch**: `002-import-via-torch` | **Date**: 2025-10-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-import-via-torch/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Enable researchers to extract FHIR patient cohort data directly from TORCH servers using CRTDL (Cohort Representation for Trial Data Linking) files, eliminating the need for manual file downloads. The system will detect CRTDL file inputs, submit extraction requests to TORCH with proper authentication, poll for completion, and download resulting NDJSON files for pipeline processing. Backward compatibility with existing local directory imports is maintained.

## Technical Context

**Language/Version**: Go 1.25.1
**Primary Dependencies**:
- `github.com/spf13/cobra` (CLI framework - existing)
- `github.com/spf13/viper` (configuration - existing)
- `net/http` (HTTP client - existing)
- `encoding/json` (JSON/FHIR parsing - existing)
- `encoding/base64` (CRTDL encoding for TORCH)
- `github.com/stretchr/testify` (testing - existing)

**Storage**: Filesystem (job state as JSON, FHIR NDJSON files by job ID - existing)
**Testing**: `go test` with testify assertions (existing pattern)
**Target Platform**: Linux/macOS command-line tool
**Project Type**: Single CLI application
**Performance Goals**:
- TORCH connectivity check < 5 seconds
- CRTDL file validation < 1 second
- Polling interval: 5-10 seconds (configurable)
- Download throughput matches existing HTTP downloader performance

**Constraints**:
- TORCH extraction timeout: configurable (default 30 minutes)
- Must maintain backward compatibility with existing local directory imports
- Must reuse existing retry logic and HTTP client infrastructure
- No breaking changes to configuration file structure

**Scale/Scope**:
- Support CRTDL files up to 1MB (typical cohort definitions)
- Handle TORCH extractions with hundreds of patient bundles
- Poll status up to configured timeout duration
- Download multiple NDJSON batch files per extraction

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Functional Programming

**Status**: ✅ PASS

- **Immutability**: TORCH client will use immutable structs for CRTDL, extraction requests, and responses
- **Pure Functions**: Input type detection, CRTDL validation, base64 encoding are pure functions
- **Side Effects Isolated**: HTTP calls to TORCH isolated in dedicated service layer (following existing `httpclient.go` pattern)
- **Function Composition**: Extraction workflow composed from small functions: `detectInputType() → validateCRTDL() → submitExtraction() → pollStatus() → downloadFiles()`
- **No Hidden State**: All state passed explicitly through job struct (existing pattern)

### Principle II: Test-Driven Development (TDD)

**Status**: ✅ PASS with COMMITMENT

**Commitment**: Tests will be written and approved BEFORE implementation:
1. **Contract tests**: TORCH API contract (FHIR Parameters, $extract-data endpoint)
2. **Integration tests**: End-to-end CRTDL → extraction → download flow
3. **Unit tests**: Input type detection, CRTDL validation, base64 encoding, polling logic

**Test Structure** (following existing pattern):
- `tests/contract/torch_service_test.go` - TORCH API contracts
- `tests/integration/pipeline_torch_test.go` - Full extraction workflow
- `tests/unit/torch_client_test.go` - TORCH client logic
- `tests/unit/crtdl_validation_test.go` - CRTDL file validation

**User Approval Required**: Test scenarios must be reviewed before implementation begins.

### Principle III: Keep It Simple, Stupid (KISS)

**Status**: ✅ PASS

**Simplicity Decisions**:
- **Reuse existing infrastructure**: Leverage `httpclient.go`, retry logic, config structure, downloader patterns
- **No new abstractions**: TORCH client follows same pattern as existing `dimp_client.go`
- **Minimal API surface**: Only 3 new public functions: `DetectInputType()`, `SubmitCRTDLExtraction()`, `PollAndDownload()`
- **No premature optimization**: Simple polling with configurable interval (no complex event-driven architecture)
- **Clear over clever**: Straightforward state machine for extraction lifecycle (submitted → polling → downloading → complete)

**YAGNI Adherence**:
- NOT implementing: patient override CLI flag (use adapted CRTDL files instead)
- NOT implementing: parallel batch downloads (use existing sequential downloader)
- NOT implementing: extraction result caching
- NOT implementing: CRTDL templating or generation

**Complexity Justification**: None required - feature adds minimal complexity by following existing patterns.

### Gate Summary (Initial)

**✅ ALL GATES PASS** - Proceed to Phase 0 research.

---

## Constitution Check (Post-Phase 1 Review)

*Re-evaluated after Phase 1 design artifacts completed*

### Principle I: Functional Programming ✅

**Design Artifacts Review**:
- **data-model.md**: All structs immutable, pure validation functions defined
- **contracts/torch-api.md**: Stateless HTTP operations, no hidden mutations
- **quickstart.md**: TDD workflow enforces testable, pure function design

**Confirmation**: Design maintains functional purity. Side effects (HTTP, file I/O) properly isolated in service layer.

### Principle II: Test-Driven Development ✅

**Design Artifacts Review**:
- **quickstart.md**: Explicit RED-GREEN-REFACTOR workflow for each component
- **contracts/torch-api.md**: Contract test scenarios defined before implementation
- **Test Coverage Plan**:
  - Contract: TORCH API compliance
  - Integration: End-to-end CRTDL → extraction → download
  - Unit: Input detection, validation, polling logic, config loading

**Confirmation**: TDD workflow embedded in quickstart. Tests specified before code.

### Principle III: Keep It Simple, Stupid ✅

**Design Artifacts Review**:
- **research.md**: Simpler alternatives documented and justified for rejection
- **data-model.md**: Minimal new entities (3 structs, 2 constants), reuses existing patterns
- **Project structure**: No new directories, follows existing `services/` and `lib/` patterns

**Complexity Analysis**:
- No new abstractions beyond what DIMP client already established
- No framework additions (reusing cobra, viper, testify)
- Polling uses simple loop, not complex event system
- Configuration extends existing structure non-destructively

**Confirmation**: Design adheres to KISS. No unjustified complexity introduced.

### Final Gate Summary

**✅ ALL GATES STILL PASS** - Design is ready for `/speckit.tasks` command to generate implementation tasks.

## Project Structure

### Documentation (this feature)

```
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
internal/
├── models/
│   ├── config.go            # Add TORCHConfig to ServiceConfig
│   └── job.go               # Add InputTypeCRTDL constant
├── services/
│   ├── torch_client.go      # NEW: TORCH API client
│   ├── httpclient.go        # REUSE: existing HTTP client
│   ├── downloader.go        # REUSE: existing download logic
│   └── config.go            # MODIFY: add TORCH config
├── pipeline/
│   └── import.go            # MODIFY: add CRTDL input handling
└── lib/
    └── validation.go        # ADD: CRTDL validation functions

cmd/
└── pipeline.go              # MODIFY: detect CRTDL input type

tests/
├── contract/
│   └── torch_service_test.go       # NEW: TORCH API contract tests
├── integration/
│   └── pipeline_torch_test.go      # NEW: end-to-end TORCH tests
└── unit/
    ├── torch_client_test.go        # NEW: TORCH client unit tests
    └── crtdl_validation_test.go    # NEW: CRTDL validation tests

config/
└── aether.example.yaml      # ADD: TORCH configuration section
```

**Structure Decision**: Single CLI application following existing Go project layout. New TORCH functionality added as:
- **New service**: `internal/services/torch_client.go` (mirrors `dimp_client.go` pattern)
- **Model extensions**: TORCH config added to existing `ProjectConfig` struct
- **Pipeline integration**: CRTDL detection added to existing import step
- **Validation utilities**: CRTDL validation functions in `lib/` package

**Rationale**: Maintains existing architecture, minimizes structural changes, follows established patterns.

## Complexity Tracking

*No violations - this section left empty per Constitution Check results.*
