# Data Model: DUP Pipeline CLI

**Date**: 2025-10-08
**Feature**: DUP (Data Use Process) Pipeline CLI (Aether)
**Phase**: 1 - Data Model Design

## Overview

This document defines the core domain entities for the Aether Data Use Process pipeline. All entities are designed as immutable value objects to support functional programming principles.

---

## Entity: PipelineJob

Represents a single execution of the Data Use Process pipeline.

### Go Struct

```go
type PipelineJob struct {
	JobID        string        `json:"job_id"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	InputSource  string        `json:"input_source"`   // Local path or HTTP(S) URL
	InputType    InputType     `json:"input_type"`     // "local_directory" | "http_url"
	CurrentStep  string        `json:"current_step"`   // "import" | "dimp" | "validation" | "csv_conversion" | "parquet_conversion"
	Status       JobStatus     `json:"status"`         // "pending" | "in_progress" | "completed" | "failed"
	Steps        []PipelineStep `json:"steps"`
	Config       ProjectConfig  `json:"config"`
	TotalFiles   int           `json:"total_files"`
	TotalBytes   int64         `json:"total_bytes"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

type InputType string
const (
	InputTypeLocal InputType = "local_directory"
	InputTypeHTTP  InputType = "http_url"
)

type JobStatus string
const (
	JobStatusPending    JobStatus = "pending"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)
```

### Validation Rules

- **JobID**: Must be valid UUID v4
- **InputSource**: Must be non-empty, valid file path or HTTP(S) URL
- **CurrentStep**: Must match one of enabled steps in Config
- **Status**: Valid state transitions:
  - `pending` → `in_progress`
  - `in_progress` → `completed` | `failed`
  - `failed` → `in_progress` (manual retry)

### State Transitions

```
[Created] → pending
   ↓
[Start] → in_progress (Import step begins)
   ↓
[Each step completes] → in_progress (Next step begins)
   ↓
[All steps done] → completed
   ↓
[Error occurs] → failed
   ↓
[Manual retry] → in_progress
```

### Storage

- **File**: `jobs/<job-id>/state.json`
- **Format**: JSON with ISO8601 timestamps
- **Atomicity**: Write to temp file + `os.Rename` for atomic updates

---

## Entity: PipelineStep

Represents a discrete stage in the pipeline.

### Go Struct

```go
type PipelineStep struct {
	Name           StepName     `json:"name"`
	Status         StepStatus   `json:"status"`
	StartedAt      *time.Time   `json:"started_at,omitempty"`
	CompletedAt    *time.Time   `json:"completed_at,omitempty"`
	FilesProcessed int          `json:"files_processed"`
	BytesProcessed int64        `json:"bytes_processed"`
	RetryCount     int          `json:"retry_count"`
	LastError      *StepError   `json:"last_error,omitempty"`
}

type StepName string
const (
	StepImport           StepName = "import"
	StepDIMP             StepName = "dimp"
	StepValidation       StepName = "validation"
	StepCSVConversion    StepName = "csv_conversion"
	StepParquetConversion StepName = "parquet_conversion"
)

type StepStatus string
const (
	StepStatusPending    StepStatus = "pending"
	StepStatusInProgress StepStatus = "in_progress"
	StepStatusCompleted  StepStatus = "completed"
	StepStatusFailed     StepStatus = "failed"
)

type StepError struct {
	Type       ErrorType `json:"type"`          // "transient" | "non_transient"
	Message    string    `json:"message"`
	HTTPStatus int       `json:"http_status,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

type ErrorType string
const (
	ErrorTypeTransient    ErrorType = "transient"    // Network, 5xx, timeout
	ErrorTypeNonTransient ErrorType = "non_transient" // 4xx, validation, malformed data
)
```

### Validation Rules

- **Name**: Must match one of StepName constants
- **Status**: Valid transitions:
  - `pending` → `in_progress`
  - `in_progress` → `completed` | `failed`
  - `failed` → `in_progress` (retry, if transient error)
- **RetryCount**: Must be ≥ 0, incremented on each retry attempt
- **StartedAt**: Must be set when Status = `in_progress`
- **CompletedAt**: Must be set when Status = `completed`

### Retry Logic

```go
func (e StepError) IsRetryable(maxRetries int, currentRetries int) bool {
	return e.Type == ErrorTypeTransient && currentRetries < maxRetries
}
```

---

## Entity: FHIRDataFile

Represents a single FHIR NDJSON file in the pipeline.

### Go Struct

```go
type FHIRDataFile struct {
	FileName     string   `json:"file_name"`
	FilePath     string   `json:"file_path"`      // Relative to job directory
	ResourceType string   `json:"resource_type"`  // e.g., "Patient", "Observation"
	FileSize     int64    `json:"file_size"`      // Bytes
	SourceStep   StepName `json:"source_step"`    // Which step produced this file
	LineCount    int      `json:"line_count"`     // Number of FHIR resources
	CreatedAt    time.Time `json:"created_at"`
}
```

### Validation Rules

- **FileName**: Must end with `.ndjson`
- **FilePath**: Must be within job directory boundaries (prevent path traversal)
- **FileSize**: Must be > 0
- **LineCount**: Must be ≥ 0

### File Organization

```
jobs/<job-id>/
├── import/
│   ├── abc123.ndjson        # SourceStep = "import"
│   └── def456.ndjson
├── pseudonymized/
│   ├── dimped_abc123.ndjson # SourceStep = "dimp"
│   └── dimped_def456.ndjson
├── csv/
│   ├── Patient.csv          # SourceStep = "csv_conversion"
│   └── Observation.csv
└── parquet/
    ├── Patient.parquet      # SourceStep = "parquet_conversion"
    └── Observation.parquet
```

---

## Entity: ProjectConfig

Project-wide configuration for pipeline execution.

### Go Struct

```go
type ProjectConfig struct {
	Services     ServiceConfig `yaml:"services" json:"services"`
	Pipeline     PipelineConfig `yaml:"pipeline" json:"pipeline"`
	Retry        RetryConfig   `yaml:"retry" json:"retry"`
	JobsDir      string        `yaml:"jobs_dir" json:"jobs_dir"`
}

type ServiceConfig struct {
	DIMPUrl               string `yaml:"dimp_url" json:"dimp_url"`
	CSVConversionUrl      string `yaml:"csv_conversion_url" json:"csv_conversion_url"`
	ParquetConversionUrl  string `yaml:"parquet_conversion_url" json:"parquet_conversion_url"`
}

type PipelineConfig struct {
	EnabledSteps []StepName `yaml:"enabled_steps" json:"enabled_steps"`
}

type RetryConfig struct {
	MaxAttempts      int   `yaml:"max_attempts" json:"max_attempts"`
	InitialBackoffMs int64 `yaml:"initial_backoff_ms" json:"initial_backoff_ms"`
	MaxBackoffMs     int64 `yaml:"max_backoff_ms" json:"max_backoff_ms"`
}
```

### Validation Rules

- **EnabledSteps**: Must contain at least `StepImport`, order matters (defines execution sequence)
- **Service URLs**: Must be valid HTTP(S) URLs, can be empty if corresponding step not enabled
- **MaxAttempts**: Must be 1-10
- **Backoff timing**: InitialBackoffMs < MaxBackoffMs

### Default Values

```yaml
services:
  dimp_url: ""
  csv_conversion_url: ""
  parquet_conversion_url: ""

pipeline:
  enabled_steps:
    - import
    # Optional: dimp, validation, csv_conversion, parquet_conversion

retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000

jobs_dir: "./jobs"
```

---

## Entity: ServiceConfiguration

Connection details for external HTTP services (embedded in ProjectConfig).

### HTTP Client Configuration

```go
type HTTPClientConfig struct {
	Timeout       time.Duration
	MaxRetries    int
	RetryBackoff  func(attempt int) time.Duration
}
```

### Authentication

**Phase 1**: No authentication (services on localhost/trusted network)
**Future**: Support for:
- Bearer tokens (via environment variables)
- mTLS certificates
- API keys in headers

---

## Relationships

```
PipelineJob (1) ──< (many) PipelineStep
PipelineJob (1) ──< (many) FHIRDataFile
PipelineJob (1) ─── (1) ProjectConfig
```

---

## Pure Functions for State Transitions

All state mutations return new instances:

```go
// Update job status (immutable)
func UpdateJobStatus(job PipelineJob, status JobStatus) PipelineJob {
	job.Status = status
	job.UpdatedAt = time.Now()
	return job
}

// Add completed step (immutable)
func CompleteStep(job PipelineJob, stepName StepName, filesProcessed int) PipelineJob {
	for i, step := range job.Steps {
		if step.Name == stepName {
			now := time.Now()
			job.Steps[i].Status = StepStatusCompleted
			job.Steps[i].CompletedAt = &now
			job.Steps[i].FilesProcessed = filesProcessed
			break
		}
	}
	job.UpdatedAt = time.Now()
	return job
}

// Increment retry count (immutable)
func IncrementRetry(step PipelineStep) PipelineStep {
	step.RetryCount++
	return step
}
```

---

## Validation Functions

```go
func (j PipelineJob) Validate() error {
	if j.JobID == "" {
		return errors.New("job_id required")
	}
	if _, err := uuid.Parse(j.JobID); err != nil {
		return fmt.Errorf("invalid job_id: %w", err)
	}
	if j.InputSource == "" {
		return errors.New("input_source required")
	}
	// ... additional validations
	return nil
}

func (c ProjectConfig) Validate() error {
	if len(c.Pipeline.EnabledSteps) == 0 {
		return errors.New("at least one pipeline step must be enabled")
	}
	if c.Pipeline.EnabledSteps[0] != StepImport {
		return errors.New("first step must be 'import'")
	}
	// ... additional validations
	return nil
}
```

---

## Next Steps

1. Implement models in `internal/models/`
2. Add JSON/YAML marshaling tests
3. Define HTTP service contracts (see `contracts/` directory)
4. Implement state persistence in `internal/services/state.go`
