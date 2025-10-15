<p align="center">
  <img src="aether.png" alt="Aether Logo" width="200"/>
</p>

# Aether

A command-line interface for orchestrating Data Use Process (DUP) pipelines for medical FHIR data.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://go.dev/)
[![codecov](https://codecov.io/gh/trobanga/aether/branch/main/graph/badge.svg)](https://codecov.io/gh/trobanga/aether)

## Overview

Aether is a CLI tool designed for medical researchers and data engineers to process FHIR (Fast Healthcare Interoperability Resources) data through configurable pipeline steps. It provides session-independent job management, automatic retry mechanisms, and real-time progress tracking.

### Key Features

- **TORCH Integration**: Direct FHIR data extraction from TORCH servers using CRTDL query files
- **FHIR Data Import**: Import from local directories, HTTP URLs, or TORCH extractions with progress indicators
- **Session-Independent**: Resume pipelines across terminal sessions with file-based state persistence
- **DIMP Pseudonymization**: Optional de-identification and pseudonymization via DIMP HTTP service
- **Configurable Pipelines**: Enable/disable processing steps per project requirements
- **Hybrid Retry Strategy**: Automatic retries for transient errors, manual intervention for validation failures
- **Real-time Progress**: Progress bars with ETA, throughput, and completion percentages
- **Medical Data Focus**: Purpose-built for TORCH FHIR extractions and medical research workflows

## Quick Start

<p align="center">
  <img src="https://private-user-images.githubusercontent.com/8888869/500970240-8544ec9c-1f76-4775-9c1f-ae9f160fda2b.gif?jwt=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTUiLCJleHAiOjE3NjA0NDkxNjEsIm5iZiI6MTc2MDQ0ODg2MSwicGF0aCI6Ii84ODg4ODY5LzUwMDk3MDI0MC04NTQ0ZWM5Yy0xZjc2LTQ3NzUtOWMxZi1hZTlmMTYwZmRhMmIuZ2lmP1gtQW16LUFsZ29yaXRobT1BV1M0LUhNQUMtU0hBMjU2JlgtQW16LUNyZWRlbnRpYWw9QUtJQVZDT0RZTFNBNTNQUUs0WkElMkYyMDI1MTAxNCUyRnVzLWVhc3QtMSUyRnMzJTJGYXdzNF9yZXF1ZXN0JlgtQW16LURhdGU9MjAyNTEwMTRUMTMzNDIxWiZYLUFtei1FeHBpcmVzPTMwMCZYLUFtei1TaWduYXR1cmU9YjcwYTE3YWNlYmY2NDY0NWE0YzVmODliMjJlOTRiMmExOWRiMmFjZjJjMjNjZGRiMmNkZTZiYzQ1OWYyZTI0ZiZYLUFtei1TaWduZWRIZWFkZXJzPWhvc3QifQ.XP8HJwLZJwhDP24SjPy660GWEzIzStZbGawitHO3nBA" alt="Aether Quick Start Demo" width="800"/>
</p>

### Installation

**From source:**
```bash
git clone https://github.com/trobanga/aether.git
cd aether
make build
sudo make install  # installs to /usr/local/bin
```

**Alternative: Install to user directory (no sudo):**
```bash
make build
make install-local  # installs to ~/.local/bin
```

**Verify installation:**
```bash
aether --help
```

### Basic Usage

**1. Create configuration:**
```bash
cp config/aether.example.yaml aether.yaml
# Edit aether.yaml to configure services and enabled steps
```

**2. Start a pipeline:**
```bash
# From CRTDL file (TORCH extraction)
aether pipeline start query.crtdl

# From local directory
aether pipeline start /path/to/fhir/data

# From HTTP URL
aether pipeline start https://example.com/fhir/Patient.ndjson

# From TORCH result URL (reuse existing extraction)
aether pipeline start http://torch-server/fhir/extraction/result-123
```

**3. Monitor progress:**
```bash
aether pipeline status <job-id>
```

**4. List all jobs:**
```bash
aether job list
```

**5. Resume a failed pipeline:**
```bash
aether pipeline continue <job-id>
```

See the [Quickstart Guide](specs/001-dup-pipeline-we/quickstart.md) for detailed usage instructions.

### TORCH Integration

Aether supports direct data extraction from TORCH servers using CRTDL (Cohort Representation for Trial Data Linking) files, eliminating manual file downloads.

**What is TORCH?**
TORCH is a FHIR-based data extraction service that allows researchers to define patient cohorts using CRTDL query files and retrieve matching FHIR data automatically.

**Input Methods:**

1. **CRTDL File Extraction** (recommended for new queries):
   ```bash
   aether pipeline start cohort-query.crtdl
   ```
   - Submits CRTDL to TORCH server
   - Polls extraction status until complete
   - Downloads resulting FHIR NDJSON files
   - Continues with pipeline processing

2. **TORCH Result URL** (for reusing existing extractions):
   ```bash
   aether pipeline start http://localhost:8080/fhir/result/abc123
   ```
   - Skips extraction submission
   - Downloads files directly from result URL
   - Useful for resuming or reprocessing data

**Configuration:**

Add TORCH settings to `aether.yaml`:
```yaml
services:
  torch:
    base_url: "http://localhost:8080"
    username: "your-username"
    password: "your-password"
    extraction_timeout_minutes: 30
    polling_interval_seconds: 5
```

**CRTDL File Format:**

CRTDL files must contain:
- `cohortDefinition`: Patient inclusion/exclusion criteria
- `dataExtraction`: FHIR resource types and attributes to extract

See [TORCH quickstart](specs/002-import-via-torch/quickstart.md) for examples and detailed workflow.

**Error Handling:**

- **Server unreachable**: Clear error within 5 seconds
- **Authentication failure**: Fails early with credential error
- **Extraction timeout**: Configurable timeout (default 30 minutes)
- **Empty results**: Gracefully handles zero-patient cohorts
- **Malformed CRTDL**: Validates syntax before submission

**Backward Compatibility:**

Existing workflows using local directories or HTTP URLs continue to work without changes:
```bash
# Still supported
aether pipeline start ./test-data/
aether pipeline start https://example.com/fhir/export
```

## Architecture

Aether follows functional programming principles with clear separation of concerns:

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
│  │  - pipeline start/continue/status│  │
│  │  - job list                     │   │
│  └──────────┬──────────────────────┘   │
│             ▼                            │
│  ┌─────────────────────────────────┐   │
│  │  Pipeline Orchestrator          │   │
│  │  (Pure Functions + State)       │   │
│  └──────────┬──────────────────────┘   │
│             ▼                            │
│  ┌─────────────────────────────────┐   │
│  │  Services (Side Effects)        │   │
│  │  - HTTP Client (DIMP)           │   │
│  │  - File I/O (Import/Save)       │   │
│  │  - State Persistence (JSON)     │   │
│  └──────────┬──────────────────────┘   │
└─────────────┼──────────────────────────┘
              │
    ┌─────────┴─────────┐
    ▼                   ▼
┌───────┐           ┌───────────┐
│ DIMP  │           │ Filesystem│
│Service│           │  (Jobs)   │
└───────┘           └───────────┘
```

### Project Structure

```
aether/
├── cmd/                      # CLI entry points
│   ├── root.go               # Root command
│   ├── pipeline.go           # Pipeline commands
│   └── job.go                # Job management commands
├── internal/
│   ├── models/               # Domain models (immutable)
│   │   ├── job.go            # PipelineJob, JobStatus
│   │   ├── step.go           # PipelineStep, StepStatus
│   │   ├── config.go         # ProjectConfig
│   │   └── validation.go     # Model validation
│   ├── pipeline/             # Pipeline orchestration
│   │   ├── job.go            # Job initialization
│   │   ├── import.go         # Import step
│   │   └── dimp.go           # DIMP step
│   ├── services/             # Side effects (I/O, HTTP)
│   │   ├── importer.go       # Local file import
│   │   ├── downloader.go     # HTTP download
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

## Configuration

Aether uses YAML configuration files with support for CLI flag overrides:

```yaml
services:
  dimp_url: "http://localhost:8083/fhir"
  csv_conversion_url: "http://localhost:9000/convert/csv"
  parquet_conversion_url: "http://localhost:9000/convert/parquet"

pipeline:
  enabled_steps:
    - import
    - dimp
    # - validation  (placeholder, not implemented)
    # - csv_conversion  (service not available yet)
    # - parquet_conversion  (service not available yet)

retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000

jobs_dir: "./jobs"
```

**Key Configuration Options:**
- `services.*_url`: HTTP service endpoints for processing steps
- `pipeline.enabled_steps`: List of steps to execute (order matters)
- `retry.max_attempts`: Maximum automatic retry attempts for transient errors
- `jobs_dir`: Directory for job state and data storage

## Development

### Prerequisites

- Go 1.21 or later
- Docker (for DIMP test service)

### Building

```bash
# Build for current platform
make build

# Cross-compile for all platforms
make build-all

# Build for specific platform
make build-linux      # Linux amd64
make build-mac        # macOS amd64
make build-mac-arm    # macOS arm64 (M1/M2)
```

### Testing

```bash
# Run all tests
make test

# Run specific test suites
make test-unit
make test-integration
make test-contract

# Run with coverage report
make coverage

# Format code and run checks
make check
```

### Running Tests with DIMP Service

```bash
# Start DIMP test service
cd .github/test
make dimp-up

# Run DIMP integration tests
make dimp-test

# Stop service
make dimp-down

# Or run all tests with services in one command
make test-with-services
```

See `.github/test/Makefile` for all test infrastructure targets.

## Design Principles

Aether follows three core principles defined in the [project constitution](.specify/memory/constitution.md):

### I. Functional Programming
- **Immutability**: Data structures are immutable by default
- **Pure Functions**: Business logic uses pure functions whenever possible
- **Explicit Side Effects**: I/O and mutations isolated to service boundaries
- **Function Composition**: Complex logic built from small, composable functions

### II. Test-Driven Development (TDD)
- Tests written first, implementation follows
- Red-Green-Refactor cycle strictly enforced
- Comprehensive coverage (unit, integration, contract tests)

### III. Keep It Simple, Stupid (KISS)
- Single binary, no microservices
- File-based state, no database
- Standard library-first approach
- External services handle domain complexity

## Documentation

### Core Pipeline (001-dup-pipeline-we)
- **[Feature Specification](specs/001-dup-pipeline-we/spec.md)**: Complete functional requirements
- **[Implementation Plan](specs/001-dup-pipeline-we/plan.md)**: Technical architecture and decisions
- **[Quickstart Guide](specs/001-dup-pipeline-we/quickstart.md)**: Detailed usage instructions
- **[Data Model](specs/001-dup-pipeline-we/data-model.md)**: Domain entities and schemas
- **[API Contracts](specs/001-dup-pipeline-we/contracts/)**: HTTP service specifications

### TORCH Integration (002-import-via-torch)
- **[TORCH Specification](specs/002-import-via-torch/spec.md)**: TORCH extraction requirements
- **[TORCH Implementation Plan](specs/002-import-via-torch/plan.md)**: TORCH integration design
- **[TORCH Quickstart](specs/002-import-via-torch/quickstart.md)**: CRTDL extraction workflow
- **[TORCH API Contract](specs/002-import-via-torch/contracts/torch-api.md)**: TORCH API specification

## Roadmap

### Completed ✅
- **FHIR Data Import**: Local directory and HTTP URL support with progress tracking
- **Session-Independent Operation**: File-based state persistence for cross-session resumption
- **DIMP Pseudonymization**: Integration with DIMP service for de-identification
- **Hybrid Retry Strategy**: Automatic retries for transient errors, manual for validation failures
- **Progress Indicators**: Real-time progress bars with ETA, throughput, and completion percentages
- **Job Management**: List, status, and continue operations for all jobs
- **Manual Step Execution**: Run individual pipeline steps via `job run --step` command
- **Concurrent Job Safety**: File locks prevent multiple processes from corrupting job state
- **Functional Programming Compliance**: Immutable data structures with pure functions

### Planned 📋
- **Format Conversion**: CSV and Parquet output (requires external services)
- **Enhanced Validation**: FHIR schema validation step (requires external service)

See `specs/*/tasks.md` for detailed implementation tracking.

## Contributing

Contributions welcome! Please ensure:

1. **Test**: All new code must be covered by tests
2. **Functional style**: Prefer immutability and pure functions
3. **Keep it simple**: Justify any added complexity
4. **Code review**: All changes go through pull request review

See [CLAUDE.md](CLAUDE.md) for development guidelines.

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Acknowledgments

Built for medical research workflows with TORCH FHIR extractions. Designed for session-independent operation to support long-running data processing tasks.

---

**Status**: Active Development | **Branch**: `001-dup-pipeline-we` | **Core Features**: Complete ✅
