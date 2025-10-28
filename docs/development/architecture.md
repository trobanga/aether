# Architecture

System architecture and design overview of Aether.

## System Overview

Aether is a command-line tool for orchestrating FHIR data processing pipelines. The system follows functional programming principles with clear separation of concerns between data models, business logic, and side effects (I/O, HTTP).

### High-Level Architecture

```
┌─────────────┐
│   CLI User  │
└──────┬──────┘
       │ Commands (pipeline, job)
       ▼
┌─────────────────────────────────────────┐
│          Aether CLI (Go)                │
│  ┌─────────────────────────────────┐   │
│  │  Cobra Commands                 │   │
│  │  - pipeline start/continue      │   │
│  │  - job list/status/logs         │   │
│  └──────────┬──────────────────────┘   │
│             ▼                            │
│  ┌─────────────────────────────────┐   │
│  │  Pipeline Orchestrator          │   │
│  │  (Pure Functions + State)       │   │
│  └──────────┬──────────────────────┘   │
│             ▼                            │
│  ┌─────────────────────────────────┐   │
│  │  Services (Side Effects)        │   │
│  │  - HTTP Client (TORCH, DIMP)    │   │
│  │  - File I/O (Import/Save)       │   │
│  │  - State Persistence (JSON)     │   │
│  └──────────┬──────────────────────┘   │
└─────────────┼──────────────────────────┘
              │
    ┌─────────┴──────────┐
    ▼                    ▼
┌──────────┐        ┌──────────────┐
│ TORCH    │        │ DIMP         │
│ Server   │        │ Service      │
└──────────┘        └──────────────┘
    ▲                    ▲
    │                    │
    └────────────────────┘
     External Services
```

## Project Structure

The codebase is organized for clarity and maintainability:

```
aether/
├── cmd/                      # CLI entry points
│   ├── root.go               # Root command (aether)
│   ├── pipeline.go           # Pipeline commands (start, continue, status)
│   └── job.go                # Job management (list, logs, delete)
├── internal/
│   ├── models/               # Domain models (immutable)
│   │   ├── job.go            # PipelineJob, JobStatus
│   │   ├── step.go           # PipelineStep, StepStatus
│   │   ├── config.go         # ProjectConfig
│   │   └── validation.go     # Model validation
│   ├── pipeline/             # Pipeline orchestration (pure)
│   │   ├── job.go            # Job initialization
│   │   ├── import.go         # Import step dispatcher (torch/local/http)
│   │   └── dimp.go           # DIMP pseudonymization step
│   ├── services/             # Side effects (I/O, HTTP)
│   │   ├── importer.go       # Local file import
│   │   ├── downloader.go     # HTTP download
│   │   ├── torch_client.go   # TORCH HTTP client
│   │   ├── dimp_client.go    # DIMP HTTP client
│   │   ├── state.go          # State persistence
│   │   └── config.go         # Configuration loader
│   ├── ui/                   # Progress indicators
│   │   ├── progress.go       # Progress bars
│   │   ├── eta.go            # ETA calculation
│   │   └── throughput.go     # Throughput display
│   └── lib/                  # Pure utilities
│       ├── retry.go          # Retry logic
│       ├── fhir.go           # FHIR parsing
│       └── logging.go        # Logging
├── tests/
│   ├── unit/                 # Unit tests
│   ├── integration/          # Integration tests
│   └── contract/             # HTTP service contracts
├── config/
│   └── aether.example.yaml   # Example configuration
└── jobs/                     # Runtime job data (gitignored)
```

## Design Principles

Aether follows three core principles defined in the project constitution:

### I. Functional Programming

- **Immutability**: Data structures are immutable by default
- **Pure Functions**: Business logic uses pure functions whenever possible
- **Explicit Side Effects**: I/O and mutations isolated to service boundaries
- **Function Composition**: Complex logic built from small, composable functions

Benefits:
- Easier to test (no hidden state)
- Easier to reason about (input → output)
- Easier to refactor (no side effects to track)
- Concurrent safety (no shared mutable state)

### II. Test-Driven Development (TDD)

- Tests written first, implementation follows
- Red-Green-Refactor cycle strictly enforced
- Comprehensive coverage (unit, integration, contract tests)

Benefits:
- Specifications written as tests
- Faster feedback loop
- Fewer bugs in production
- Documentation via examples

### III. Keep It Simple, Stupid (KISS)

- Single binary, no microservices
- File-based state, no database
- Standard library-first approach
- External services handle domain complexity

Benefits:
- Easy to deploy (single binary)
- Easy to understand (clear dependencies)
- Easy to extend (add new service clients)
- No infrastructure overhead

## Data Flow

### Pipeline Execution Flow

```
1. User Input
   ↓
2. Load Configuration (aether.yaml)
   ↓
3. Initialize Job (UUID, state directory)
   ↓
4. Execute Pipeline Steps (in order)
   ├─→ Import Step (required, one of):
   │   ├─→ torch: Extract FHIR from TORCH via CRTDL
   │   ├─→ local_import: Load FHIR from local directory
   │   └─→ http_import: Download FHIR from HTTP URL
   │   └─→ Save results to job directory
   ├─→ DIMP Step (if enabled): Pseudonymization
   │   └─→ Save de-identified data
   └─→ [Future steps...]
   ↓
5. Persist Job State
   ├─→ Step status (completed/failed)
   ├─→ Output data (NDJSON)
   └─→ Logs
   ↓
6. Return Results to User
```

### State Persistence

Job state is persisted to JSON files in the jobs directory:

```
jobs/
└── {job-id}/
    ├── status.json           # Job metadata and step status
    ├── config.json           # Configuration snapshot
    ├── import_results.ndjson # Imported FHIR data
    ├── dimp_results.ndjson   # De-identified data
    └── logs.txt              # Execution logs
```

This enables:
- **Resume capability**: Continue failed pipelines without reprocessing
- **Audit trail**: Full history of what was processed
- **Debugging**: Inspect intermediate results

## Service Integration

### TORCH Integration

```
User Command (with .crtdl file)
    ↓
Aether (torch import step) → TORCH Server
    ├─→ Submit CRTDL query
    ├─→ Poll extraction status (5s intervals)
    └─→ Download FHIR NDJSON results
    ↓
Save to job directory
```

### DIMP Integration

```
Import Results (FHIR Bundles)
    ↓
Aether → DIMP Service
    ├─→ Split large bundles (if needed)
    ├─→ Send for pseudonymization
    └─→ Receive de-identified results
    ↓
Persisted Results
```

## Performance Characteristics

### Memory Usage

- Streams FHIR NDJSON line-by-line (no buffering entire files)
- Job state loaded only when needed
- Progress bars updated incrementally

### Disk Usage

- One job directory per execution
- Can clean up old jobs with `aether job delete`
- NDJSON format is space-efficient

### Network Usage

- Exponential backoff retry strategy (reduces unnecessary requests)
- Configurable polling intervals for TORCH
- Automatic bundle splitting for DIMP

## Extensibility

Adding new pipeline steps:

1. **Define the step** in `internal/models/step.go`
2. **Implement the logic** in `internal/pipeline/{step_name}.go`
3. **Add tests** in `tests/unit/{step_name}_test.go`
4. **Update CLI** to recognize new step
5. **Update configuration** documentation

Example: Adding a "validation" step would involve:
- Create `internal/pipeline/validation.go`
- Implement `ValidateStep(ctx, job, config) error`
- Add tests
- Update step list in help/docs

## Next Steps

- [Design Principles](../getting-started/installation.md) - More details on principles
- [Testing](./testing.md) - Testing strategies and examples
- [Contributing](./contributing.md) - How to contribute to Aether
- [Coding Guidelines](./coding-guidelines.md) - Code style and standards
