# Research: DUP Pipeline CLI Technology Stack

**Date**: 2025-10-08
**Feature**: DUP (Data Use Process) Pipeline CLI (Aether)
**Phase**: 0 - Technology Selection

## Technology Decisions

### Language: Go 1.21+

**Decision**: Use Go (Golang) version 1.21 or later.

**Rationale**:
- **Cross-platform CLI**: Single statically-linked binary for Linux/macOS, no runtime dependencies
- **Concurrency**: Built-in goroutines for parallel file processing and HTTP requests
- **Standard library**: Excellent `net/http`, `encoding/json`, `os/filepath` packages for our use case
- **Fast compilation**: Quick build times support TDD workflow
- **Memory safety**: Garbage collected but efficient for large file handling
- **Functional programming support**: First-class functions, immutable structs via value semantics
- **Medical domain adoption**: Used in healthcare infrastructure (e.g., FHIR servers, HL7 tools)

**Alternatives Considered**:
- **Python 3.11+**: Slower for large file processing, runtime dependency management complexity, weaker type safety
- **Rust 1.75+**: Steeper learning curve, longer compile times, overkill for CLI orchestration (no unsafe memory requirements)

---

### CLI Framework: Cobra

**Decision**: Use [spf13/cobra](https://github.com/spf13/cobra) for command-line interface.

**Rationale**:
- **Industry standard**: Used by kubectl, Hugo, GitHub CLI - proven for complex command hierarchies
- **Subcommand support**: Natural fit for `aether pipeline start`, `aether job list` structure
- **Flag management**: Built-in support for persistent flags (config file overrides)
- **Auto-generated help**: Documentation generation from command structure
- **Complementary with Viper**: Easy integration for configuration file + CLI flag merging

**Alternatives Considered**:
- **urfave/cli**: Simpler but less suited for deep command hierarchies
- **Standard library flag**: Too basic for nested subcommands

---

### HTTP Client: net/http (standard library)

**Decision**: Use Go's built-in `net/http` package.

**Rationale**:
- **No external dependency**: Standard library reduces supply chain risk (medical data security)
- **Retry control**: Full control over timeouts, retries, exponential backoff
- **HTTP/2 support**: Built-in for performance
- **Testing**: `httptest` package for mocking external services

**Enhancements**:
- **hashicorp/go-retryablehttp**: Optional wrapper for automatic retry logic (transient errors)
- **Custom retry logic**: Implement FR-023 to FR-026 (distinguish transient vs non-transient errors)

---

### Configuration: Viper + YAML

**Decision**: Use [spf13/viper](https://github.com/spf13/viper) with YAML config files.

**Rationale**:
- **Cobra integration**: Designed to work seamlessly with Cobra CLI
- **Hierarchical config**: Supports config file defaults + environment variables + CLI flag overrides (FR-030)
- **YAML format**: Human-readable for medical researchers, supports comments
- **Type-safe unmarshaling**: Direct mapping to Go structs

**Config Schema**:
```yaml
services:
  dimp_url: "http://localhost:8083/fhir"
  csv_conversion_url: "http://localhost:9000/convert/csv"
  parquet_conversion_url: "http://localhost:9000/convert/parquet"

pipeline:
  enabled_steps:
    - dimp
    - csv_conversion
    - parquet_conversion

retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000
```

---

### FHIR Processing

**Decision**: Manual JSON parsing with `encoding/json` (no heavyweight FHIR library).

**Rationale**:
- **NDJSON format**: Simple line-by-line JSON parsing, no complex FHIR validation needed
- **Pass-through architecture**: CLI doesn't interpret FHIR content, just forwards to services
- **KISS principle**: Avoid dependency on full FHIR SDKs (e.g., google/fhir) which add complexity
- **Validation**: External services (DIMP, conversion) handle FHIR schema validation

**Parsing Strategy**:
```go
type FHIRResource map[string]interface{}  // Generic JSON object
```

---

### State Persistence

**Decision**: JSON files in `jobs/<job-id>/state.json`.

**Rationale**:
- **Simple**: No database setup, portable across environments
- **Human-readable**: Easy debugging and manual inspection
- **Atomic writes**: Use `os.Rename` for atomic state updates (concurrency safety)
- **Schema**:
```json
{
  "job_id": "uuid-v4",
  "created_at": "2025-10-08T10:00:00Z",
  "input_source": "/path/or/url",
  "current_step": "import",
  "status": "in_progress",
  "steps": [
    {
      "name": "import",
      "status": "completed",
      "started_at": "2025-10-08T10:00:00Z",
      "completed_at": "2025-10-08T10:05:00Z",
      "files_processed": 150,
      "retry_count": 0
    }
  ],
  "errors": []
}
```

---

### Testing Strategy

**Decision**: Go's built-in `go test` + `testify` for assertions + `httptest` for mocking.

**Rationale**:
- **TDD-friendly**: Fast test execution, table-driven test support
- **Testify**: Cleaner assertions (`assert.Equal`) vs manual `if` checks
- **httptest**: Mock HTTP servers for contract tests (DIMP, conversion services)
- **Coverage**: Built-in `go test -cover` for coverage reports

**Test Structure**:
- **Unit tests**: `lib/retry_test.go`, `models/job_test.go` (pure functions)
- **Integration tests**: `tests/integration/pipeline_test.go` (end-to-end flows)
- **Contract tests**: `tests/contract/dimp_test.go` (HTTP service mocks)

---

### Project Structure (Go-specific)

```
aether/
├── cmd/
│   └── aether/
│       └── main.go              # Entry point
├── internal/
│   ├── models/                  # Domain models (Job, Step, File, Config)
│   ├── pipeline/                # Pipeline orchestration
│   ├── services/                # HTTP client, file I/O, state persistence
│   ├── cli/                     # Cobra command definitions
│   ├── ui/                      # Progress indicators, formatters (FR-029)
│   └── lib/                     # Pure utilities (retry, validation)
├── tests/
│   ├── contract/
│   ├── integration/
│   └── fixtures/
├── config/
│   └── aether.example.yaml
├── go.mod
├── go.sum
└── Makefile
```

**Rationale for `internal/`**:
- Go convention: `internal/` prevents external imports, enforcing encapsulation
- Forces clean API boundaries

---

### Build & Distribution

**Decision**: Makefile + `go build` for cross-compilation.

**Targets**:
```makefile
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/aether-linux cmd/aether/main.go

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/aether-mac cmd/aether/main.go

test:
	go test ./... -v -cover

install:
	go install ./cmd/aether
```

---

### Progress Indicators: schollz/progressbar

**Decision**: Use [schollz/progressbar](https://github.com/schollz/progressbar) for FR-029 visual feedback requirements.

**Rationale**:
- **Rich formatting**: Supports percentage, ETA, throughput rate, custom descriptions (FR-029a-e requirements)
- **Multiple styles**: Progress bars for known-size operations, spinners for unknown duration
- **Update control**: Configurable refresh rate (FR-029d: minimum 2 second updates)
- **Cross-platform**: Works on Linux/macOS terminals
- **Pure Go**: No C dependencies, easy to cross-compile
- **Testable**: Can mock io.Writer for unit tests
- **Constitution compliance**: Isolated side effect (terminal output), functional updates via value methods

**Alternatives Considered**:
- **cheggaaa/pb**: Good but less rich formatting options, older codebase
- **vbauerster/mpb**: Over-engineered for our needs (multi-progress support not required)
- **Custom implementation**: Violates KISS - reinventing well-solved problem

**Usage Pattern**:
```go
// Encapsulated in internal/ui/progress.go
bar := progressbar.NewOptions(totalFiles,
    progressbar.OptionSetDescription("Importing FHIR files"),
    progressbar.OptionShowBytes(true),
    progressbar.OptionShowCount(),
    progressbar.OptionSetWidth(40),
    progressbar.OptionThrottle(2 * time.Second),  // FR-029d: 2s updates
)
```

---

## Dependencies Summary

**Core**:
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `github.com/google/uuid` - Job ID generation
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/schollz/progressbar/v3` - Progress indicators (FR-029)

**Testing**:
- `github.com/stretchr/testify` - Assertions and mocking

**Optional**:
- `github.com/hashicorp/go-retryablehttp` - HTTP retry wrapper (if not implementing custom)

**Standard Library** (no install needed):
- `net/http` - HTTP client
- `encoding/json` - JSON parsing
- `os` / `io` / `path/filepath` - File operations
- `time` - Timestamps, backoff timing
- `context` - Cancellation, timeouts

---

## Functional Programming in Go

Go supports functional paradigms through:

1. **Immutability**: Use value semantics for structs (pass by value, not pointer for read-only)
2. **Pure functions**: Return new state instead of mutating
   ```go
   func UpdateJobStatus(job Job, newStatus string) Job {
       job.Status = newStatus  // Copy, not mutate original
       return job
   }
   ```
3. **First-class functions**: Closures for dependency injection
   ```go
   type HttpClient func(url string, body []byte) ([]byte, error)
   ```
4. **Function composition**: Pipeline of transformations
   ```go
   result := Validate(Parse(ReadFile(path)))
   ```

**Constraints**:
- No immutable-by-default collections (use defensive copying)
- Manual discipline required (Go doesn't enforce immutability)

---

## Constitution Alignment

- **Functional Programming**: ✅ Supported via value semantics, first-class functions
- **TDD**: ✅ Fast test execution, excellent tooling (`go test`, `httptest`)
- **KISS**: ✅ Standard library-first, minimal dependencies, single binary output

---

## Next Steps

1. Initialize Go module: `go mod init github.com/user/aether`
2. Install Cobra CLI: `go install github.com/spf13/cobra-cli@latest`
3. Generate command structure: `cobra-cli init`
4. Define data models in `internal/models/`
5. Proceed to Phase 1: Data Model & Contracts
