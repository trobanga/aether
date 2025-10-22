# Implementation Plan: Bundle Splitting

**Branch**: `004-bundle-splitting` | **Date**: 2025-10-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-bundle-splitting/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

This feature adds automatic FHIR Bundle splitting to prevent HTTP 413 "Payload Too Large" errors when processing large Bundles (50MB+) through the DIMP pseudonymization service. The system will detect oversized Bundles, split them into configurable chunks (default 10MB threshold), process each chunk independently, and reassemble the pseudonymized results while maintaining data integrity and Bundle structure.

## Technical Context

**Language/Version**: Go 1.25.1 (minimum Go 1.21)
**Primary Dependencies**:
- encoding/json (FHIR resource parsing and serialization)
- github.com/spf13/viper (configuration management)
- github.com/schollz/progressbar/v3 (progress reporting)
- existing internal packages (pipeline, models, services, lib)

**Storage**: Filesystem (job state as JSON, FHIR NDJSON files organized by job ID)
**Testing**: testify (github.com/stretchr/testify) for unit and integration tests
**Target Platform**: Linux server (CLI application)
**Project Type**: Single project (CLI tool)
**Performance Goals**:
- Process 100MB Bundles within 15 minutes
- Zero performance regression for Bundles <10MB (no splitting overhead)
- Maintain 100% data integrity during split-reassemble operations

**Constraints**:
- Memory: Must handle Bundles in-memory during splitting (assumes chunks <10MB fit in memory)
- HTTP payload limits: Work within typical 30MB server limits (using 10MB chunks for safety margin)
- FHIR compliance: All chunks must be valid FHIR R4 Bundles

**Scale/Scope**:
- Handle Bundles with 100k+ entries
- Support unlimited chunk count (no artificial limits)
- Process mixed resource types (Patient, Observation, Condition, etc.)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Functional Programming ✅

**Compliance**: PASS with minor notes

- **Immutability**: Bundle splitting naturally lends itself to immutable operations. Original Bundle → Split chunks (immutable) → Pseudonymized chunks → Reassembled Bundle. Each transformation creates new data structures.
- **Pure Functions**: Splitting logic (calculate chunks, partition entries) and reassembly can be pure functions taking Bundle and returning chunks/reassembled Bundle.
- **Explicit Side Effects**: Side effects are clearly bounded:
  - I/O: Reading FHIR files, HTTP calls to DIMP, writing output files (existing patterns in pipeline/dimp.go)
  - State: Job state updates (existing pattern via UpdateJob)
- **Function Composition**: Split algorithm = detect size → partition entries → create chunks → process → reassemble (composable pipeline)

**Notes**:
- Follow existing DIMP processing pattern which already separates pure logic from I/O
- Splitting and reassembly logic should be testable without I/O

### II. Test-Driven Development (TDD) ✅

**Compliance**: PASS

- **TDD Workflow**: Will follow RED-GREEN-REFACTOR cycle
- **Test Scenarios**: Comprehensive acceptance scenarios defined in spec (15 scenarios across 3 user stories)
- **Test Types Required**:
  - **Unit tests**: Bundle size calculation, entry partitioning, chunk creation, reassembly logic
  - **Integration tests**: End-to-end splitting with mock DIMP service, configuration validation, error handling
  - **Contract tests**: Ensure chunks are valid FHIR R4 Bundles (structural validation)

**Test Coverage Requirements**:
- FR-001 to FR-020: Each functional requirement maps to specific test cases
- Edge cases: 9 identified scenarios (Bundle references, mixed types, size changes, etc.)
- Success criteria: All 8 SC criteria are testable and measurable

### III. Keep It Simple, Stupid (KISS) ✅

**Compliance**: PASS

**Simplicity Justifications**:

1. **Approach**: In-memory splitting (not streaming)
   - **Rationale**: Chunks are <10MB (manageable in memory). Streaming would add complexity for minimal benefit in this scale.
   - **Alternative Rejected**: Streaming parser - unnecessary complexity given chunk sizes

2. **No premature optimization**:
   - Start with simple sequential chunk processing
   - Add parallelization only if performance testing shows need
   - Defer optimization until measurements justify it

3. **Reuse existing patterns**:
   - Follow processDIMPFile structure in internal/pipeline/dimp.go
   - Use existing config validation patterns
   - Leverage existing retry logic (per-chunk independence is natural extension)

4. **Clear over clever**:
   - Bundle splitting = simple array partitioning
   - Reassembly = concatenate entry arrays + restore metadata
   - No complex algorithms needed

**YAGNI Adherence**:
- No cross-chunk reference resolution (out of scope, not needed for current use case)
- No streaming (not needed for target scale)
- No transaction/batch semantics preservation (document/collection Bundles only)

## Project Structure

### Documentation (this feature)

```
specs/004-bundle-splitting/
├── spec.md              # Feature specification (complete)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (pending)
├── data-model.md        # Phase 1 output (pending)
├── quickstart.md        # Phase 1 output (pending)
├── contracts/           # Phase 1 output (pending)
│   └── bundle-chunk.json # FHIR Bundle chunk schema
├── checklists/
│   └── requirements.md  # Specification quality checklist (complete)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
internal/
├── models/
│   ├── config.go        # Add BundleSplitThresholdMB field to PipelineConfig
│   └── bundle.go        # NEW: Bundle splitting data structures
├── services/
│   ├── dimp_client.go   # Existing DIMP client (unchanged)
│   └── bundle_splitter.go # NEW: Bundle splitting logic
├── pipeline/
│   └── dimp.go          # MODIFY: Integrate bundle splitting
└── lib/
    ├── validation.go    # MODIFY: Add bundle size validation
    └── logger.go        # Existing logger (unchanged)

tests/
├── unit/
│   ├── bundle_splitter_test.go  # NEW: Pure function tests
│   └── bundle_validation_test.go # NEW: Size calculation tests
└── integration/
    └── pipeline_dimp_split_test.go # NEW: End-to-end split tests
```

**Structure Decision**: Using existing single project structure (internal/ + tests/). Bundle splitting integrates directly into the DIMP pipeline step, following established patterns. New code concentrated in:
1. `internal/services/bundle_splitter.go` - core splitting logic (pure functions)
2. `internal/pipeline/dimp.go` - integration point (orchestration with side effects)
3. `internal/models/config.go` - configuration extension

This maintains consistency with existing architecture and avoids unnecessary restructuring.

## Complexity Tracking

*No violations - constitution check passed cleanly. Implementation follows simple, functional, test-driven approach aligned with all three principles.*
