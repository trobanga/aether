<p align="center">
  <img src="aether.png" alt="Aether Logo" width="200"/>
</p>

# Aether

A command-line interface for orchestrating Data Use Process (DUP) pipelines for medical FHIR data.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://go.dev/)

## Overview

Aether is a CLI tool designed for medical researchers and data engineers to process FHIR (Fast Healthcare Interoperability Resources) data through configurable pipeline steps. It provides session-independent job management, automatic retry mechanisms, and real-time progress tracking.

### Key Features

- **FHIR Data Import**: Import from local directories or HTTP URLs with progress indicators
- **Session-Independent**: Resume pipelines across terminal sessions with file-based state persistence
- **DIMP Pseudonymization**: Optional de-identification and pseudonymization via DIMP HTTP service
- **Configurable Pipelines**: Enable/disable processing steps per project requirements
- **Hybrid Retry Strategy**: Automatic retries for transient errors, manual intervention for validation failures
- **Real-time Progress**: Progress bars with ETA, throughput, and completion percentages
- **Medical Data Focus**: Purpose-built for TORCH FHIR extractions and medical research workflows

## Quick Start

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
# From local directory
aether pipeline start --input /path/to/torch/output

# From HTTP URL
aether pipeline start --input https://example.com/fhir/export/job-123
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

- **[Feature Specification](specs/001-dup-pipeline-we/spec.md)**: Complete functional requirements
- **[Implementation Plan](specs/001-dup-pipeline-we/plan.md)**: Technical architecture and decisions
- **[Quickstart Guide](specs/001-dup-pipeline-we/quickstart.md)**: Detailed usage instructions
- **[Data Model](specs/001-dup-pipeline-we/data-model.md)**: Domain entities and schemas
- **[API Contracts](specs/001-dup-pipeline-we/contracts/)**: HTTP service specifications

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

### In Progress 🚧
- **Performance Validation**: Testing with 10GB+ datasets
- **Documentation**: Enhanced error messages and user guidance

### Planned 📋
- **Format Conversion**: CSV and Parquet output (requires external conversion services)
- **Enhanced Validation**: FHIR schema validation step
- **Additional Output Formats**: Support for more medical data formats

See [tasks.md](specs/001-dup-pipeline-we/tasks.md) for detailed implementation tracking.

## Contributing

Contributions welcome! Please ensure:

1. **Tests first**: Write failing tests before implementation (TDD)
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
