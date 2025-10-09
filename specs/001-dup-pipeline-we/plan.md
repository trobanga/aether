# Implementation Plan: DUP Pipeline CLI

**Branch**: `001-dup-pipeline-we` | **Date**: 2025-10-08 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-dup-pipeline-we/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a command-line interface (Aether) that orchestrates a DUP (Data Use Process) pipeline for medical FHIR data. The CLI imports pre-extracted TORCH data and processes it through optional steps: DIMP pseudonymization, validation (placeholder), and format conversion (CSV/Parquet). All processing steps use external HTTP services. The pipeline supports session-independent resumption, hybrid retry strategies (automatic for transient errors, manual for validation failures), and project-based configuration of enabled steps.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Cobra (CLI framework), net/http (HTTP client), encoding/json (FHIR NDJSON parsing), gopkg.in/yaml.v3 (config), testify (testing assertions), progress bar library (schollz/progressbar or cheggaaa/pb for FR-029 requirements)
**Storage**: Filesystem (job state as JSON, FHIR NDJSON files organized by job ID)
**Testing**: go test (built-in), testify for assertions, httptest for mocking HTTP services
**Target Platform**: Linux (primary), macOS (secondary), cross-platform CLI binary
**Project Type**: Single CLI application
**Performance Goals**: Handle 10GB+ FHIR datasets, download throughput limited by network, status queries <2s, progress indicators update every 2s
**Constraints**: Single-user per job, file-based state (no database), HTTP-only external services, session-independent operation
**Scale/Scope**: CLI tool for data engineers, ~10-50 concurrent jobs typical, medical research data volumes (1GB-100GB per job)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Functional Programming ✅

- **Immutability**: Pipeline state will be represented as immutable data structures. Job state transitions create new state objects rather than mutating existing ones.
- **Pure Functions**: Core business logic (FHIR file parsing, retry decision logic, state transitions) will be pure functions.
- **Explicit Side Effects**: HTTP calls, file I/O, and state persistence will be isolated in dedicated modules/services at system boundaries.
- **Function Composition**: Pipeline orchestration will compose small, single-purpose functions (validate input → import files → apply step → persist state).
- **No Hidden State**: All state (job status, retry counts, file lists) will be explicitly passed or loaded from filesystem.

**Status**: PASS - CLI architecture naturally supports functional design with side effects at boundaries.

### II. Test-Driven Development (TDD) ✅

- **Red-Green-Refactor**: All functional requirements will have tests written first.
- **User Approval**: Acceptance scenarios from spec.md will be converted to integration tests before implementation.
- **Test Coverage**: Unit tests for pure functions (retry logic, state transitions), integration tests for pipeline steps, contract tests for HTTP service interactions.
- **Contract Tests**: Required for DIMP, CSV conversion, and Parquet conversion service interactions.
- **Integration Tests**: Required for file import (local and URL), pipeline continuation, concurrent job handling.

**Status**: PASS - Spec provides clear acceptance criteria suitable for test-first approach.

### III. Keep It Simple, Stupid (KISS) ✅

- **Start Simple**: File-based state storage (no database), HTTP-only communication (no message queues), single CLI binary (no microservices).
- **YAGNI**: No multi-user collaboration, no job scheduling/cron, no automatic cleanup policies in v1.
- **Complexity Justification**: Retry logic (transient vs non-transient) is essential for reliability with external services - justified by FR-023 to FR-026. Progress indicator library adds visual feedback complexity for user experience - justified by FR-029 requirements.
- **Clear Over Clever**: Command structure is explicit (`aether pipeline start`, not cryptic abbreviations).

**Status**: PASS - Architecture avoids unnecessary abstraction. External services handle complexity (pseudonymization, conversion). Progress indicators use established Go libraries.

### Complexity Tracking

No violations detected. Progress indicator library (FR-029 refinement) is justified complexity for user experience - provides visual feedback (progress bars, spinners), ETA calculations, and throughput display required by FR-029a-e.

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
src/
├── models/              # Immutable data structures (Job, Step, File, Config)
├── pipeline/            # Pipeline orchestration and step execution
├── services/            # Side effects: HTTP client, file I/O, state persistence
├── cli/                 # Command parsing and dispatch
├── ui/                  # Progress indicators, formatters, visual feedback (FR-029)
└── lib/                 # Pure utilities (retry logic, validation, parsers)

tests/
├── contract/            # HTTP service contract tests (DIMP, conversions)
├── integration/         # End-to-end pipeline tests
└── unit/                # Pure function tests

config/
└── aether.example.yaml  # Example project configuration

jobs/                    # Runtime directory (gitignored)
└── <job-id>/
    ├── state.json       # Job state
    ├── import/          # Imported FHIR files
    ├── pseudonymized/   # DIMP output (if enabled)
    ├── csv/             # CSV conversion output (if enabled)
    └── parquet/         # Parquet conversion output (if enabled)
```

**Structure Decision**: Single CLI project structure with dedicated `ui/` directory for progress indicator components. Job state and artifacts stored in `jobs/` directory for session persistence. Configuration at project root for easy discovery. Clean separation between pure logic (`lib/`, `models/`), side effects (`services/`), visual feedback (`ui/`), and orchestration (`pipeline/`).

## Complexity Tracking

No complexity violations. Progress indicator library adds justified complexity for FR-029 requirements (visual feedback, ETA calculation, throughput display). All design choices align with KISS principle.

---

## Planning Summary

**Status**: ✅ Phase 0 (Research) and Phase 1 (Design) Complete | **Updated**: 2025-10-08 (FR-029 refinement)

### Artifacts Generated

1. **research.md**: Technology stack decisions (Go + Cobra, rationale, alternatives)
2. **data-model.md**: Core domain entities (PipelineJob, PipelineStep, FHIRDataFile, ProjectConfig)
3. **contracts/dimp-service.md**: DIMP pseudonymization HTTP API contract
4. **contracts/conversion-service.md**: CSV/Parquet conversion HTTP API contracts  
5. **quickstart.md**: User guide for installation and basic workflow
6. **CLAUDE.md**: Updated agent context with Go + Cobra stack

### Constitution Re-Check (Post-Design)

**I. Functional Programming**: ✅ PASS
- Data model uses immutable value semantics (Go structs passed by value)
- State transitions return new instances (no mutations)
- Pure functions for validation and business logic
- Progress indicator library encapsulated in `ui/` module (side effects isolated)

**II. Test-Driven Development**: ✅ PASS
- Clear contract tests defined for HTTP services
- Integration test scenarios mapped from spec acceptance criteria
- Go's `go test` + `testify` support fast TDD cycles
- Progress indicators testable via mocked output streams

**III. Keep It Simple, Stupid**: ✅ PASS
- Single binary CLI (no microservices)
- File-based state (no database)
- Standard library-first (minimal external deps)
- External services handle domain complexity (pseudonymization, conversion)
- Progress library is well-established Go library (justified by FR-029 UX requirements)

**Final Verdict**: All gates passed. Ready for Phase 2 (Task Generation).

### Next Steps

1. Run `/speckit.tasks` to generate implementation tasks from this plan
2. Implement in priority order: P1 (Import) → P2 (Resumption) → P3 (DIMP) → P4 (Conversion)
3. Follow TDD: Write tests first for each task

### Key Design Decisions

- **Language**: Go 1.21+ for cross-platform CLI with strong concurrency
- **CLI Framework**: Cobra for hierarchical commands (`aether pipeline start`, `aether job list`)
- **State Management**: JSON files in `jobs/<job-id>/state.json` for session persistence
- **Retry Strategy**: Hybrid - automatic for transient errors (5xx, network), manual for validation errors (4xx)
- **Service Integration**: All processing (DIMP, conversion) via external HTTP services
- **Configuration**: YAML file + CLI flag overrides via Viper
- **Progress Indicators**: Visual feedback library (schollz/progressbar or cheggaaa/pb) for FR-029 requirements (progress bars, spinners, ETA, throughput)

### Architecture Highlights

```
┌─────────────┐
│   CLI User  │
└──────┬──────┘
       │ aether pipeline start --input /data
       ▼
┌─────────────────────────────────────────┐
│          Aether CLI (Go)                │
│  ┌─────────────────────────────────┐   │
│  │  Cobra Commands                 │   │
│  │  - pipeline start/continue/status│  │
│  │  - job list/run                 │   │
│  └──────────┬──────────────────────┘   │
│             ▼                            │
│  ┌─────────────────────────────────┐   │
│  │  UI Components (FR-029)         │   │
│  │  - Progress bars (known size)   │   │
│  │  - Spinners (unknown duration)  │   │
│  │  - ETA calculation              │   │
│  │  - Throughput display           │   │
│  └──────────┬──────────────────────┘   │
│             ▼                            │
│  ┌─────────────────────────────────┐   │
│  │  Pipeline Orchestrator          │   │
│  │  (Pure Functions + State)       │   │
│  └──────────┬──────────────────────┘   │
│             ▼                            │
│  ┌─────────────────────────────────┐   │
│  │  Services (Side Effects)        │   │
│  │  - HTTP Client (DIMP/Convert)  │   │
│  │  - File I/O (Import/Save)      │   │
│  │  - State Persistence (JSON)    │   │
│  └──────────┬──────────────────────┘   │
└─────────────┼──────────────────────────┘
              │
    ┌─────────┼─────────┐
    ▼         ▼         ▼
┌───────┐ ┌───────┐ ┌───────────┐
│ DIMP  │ │ CSV   │ │ Parquet   │
│Service│ │Convert│ │ Convert   │
└───────┘ └───────┘ └───────────┘
```

### FR-029 Progress Indicator Design

Based on refined specification requirements (FR-029a-e):

**Visual Components**:
- Progress bars for operations with known total (file downloads, batch processing)
- Animated spinners for operations with unknown duration (HTTP service calls)
- Update frequency: Minimum every 2 seconds during active operations

**Display Information**:
- Completion percentage (e.g., "45%") for known-size operations
- Elapsed time and ETA with formula: `ETA = (total_items - processed_items) * avg_time_per_item`
  - `avg_time_per_item` computed from last 10 items or last 30 seconds (whichever more recent)
- Current operation name (e.g., "Downloading FHIR files", "Processing Patient resources")
- Items processed/total (e.g., "127/500 files")
- Throughput rate (e.g., "2.3 files/sec" or "5.2 MB/sec")

**Implementation Approach**:
- Encapsulate in `src/ui/` package for isolation
- Use established Go library (schollz/progressbar recommended for rich formatting)
- Wire into service layer operations (download, import, DIMP, conversion steps)
- Mock output streams in tests for verification

