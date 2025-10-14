# Data Model: TORCH Server Data Import

**Feature**: 002-import-via-torch | **Date**: 2025-10-10

## Overview

This document defines the data structures, entities, and their relationships for TORCH integration. All structures follow Go conventions and align with existing aether data model patterns.

## Core Entities

### 1. TORCHConfig

**Purpose**: Configuration for TORCH server connection and extraction behavior

**Fields**:
| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `BaseURL` | `string` | Yes | `""` | Must be valid HTTP(S) URL |
| `Username` | `string` | Yes | `""` | Non-empty string |
| `Password` | `string` | Yes | `""` | Non-empty string |
| `ExtractionTimeoutMinutes` | `int` | No | `30` | > 0 |
| `PollingIntervalSeconds` | `int` | No | `5` | > 0, <= 60 |
| `MaxPollingIntervalSeconds` | `int` | No | `30` | >= PollingIntervalSeconds |

**Go Struct**:
```go
type TORCHConfig struct {
    BaseURL                   string `yaml:"base_url" json:"base_url"`
    Username                  string `yaml:"username" json:"username"`
    Password                  string `yaml:"password" json:"password"`
    ExtractionTimeoutMinutes  int    `yaml:"extraction_timeout_minutes" json:"extraction_timeout_minutes"`
    PollingIntervalSeconds    int    `yaml:"polling_interval_seconds" json:"polling_interval_seconds"`
    MaxPollingIntervalSeconds int    `yaml:"max_polling_interval_seconds" json:"max_polling_interval_seconds"`
}
```

**Location**: `internal/models/config.go` (add to existing `ServiceConfig`)

**Relationships**:
- Embedded in `ServiceConfig.TORCH`
- Used by `TORCHClient` for connection and behavior configuration

### 2. InputType Constants

**Purpose**: Enumeration of supported input source types

**Values**:
| Constant | String Value | Description |
|----------|-------------|-------------|
| `InputTypeLocalDir` | `"local_directory"` | Existing: local directory with NDJSON files |
| `InputTypeRemoteURL` | `"remote_url"` | Existing: HTTP URL to download NDJSON |
| `InputTypeCRTDL` | `"crtdl_file"` | NEW: Local CRTDL file for TORCH extraction |
| `InputTypeTORCHURL` | `"torch_result_url"` | NEW: Direct TORCH extraction result URL |

**Go Constants**:
```go
const (
    InputTypeLocalDir   = "local_directory"
    InputTypeRemoteURL  = "remote_url"
    InputTypeCRTDL      = "crtdl_file"        // NEW
    InputTypeTORCHURL   = "torch_result_url"  // NEW
)
```

**Location**: `internal/models/job.go` (add to existing constants)

**Usage**: Stored in `PipelineJob.InputType` field

### 3. PipelineJob Extensions

**Purpose**: Track TORCH-specific job state

**New Fields**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `TORCHExtractionURL` | `string` | No | Content-Location URL for polling/resume |

**Modified Struct**:
```go
type PipelineJob struct {
    // ... existing fields ...
    JobID               string    `json:"job_id"`
    InputSource         string    `json:"input_source"`
    InputType           string    `json:"input_type"`  // MODIFIED: new values added
    TORCHExtractionURL  string    `json:"torch_extraction_url,omitempty"`  // NEW
    Status              JobStatus `json:"status"`
    // ... other existing fields ...
}
```

**Location**: `internal/models/job.go` (modify existing struct)

**Relationships**:
- `InputType`: References InputType constants
- `TORCHExtractionURL`: Used for job resumption if process dies during polling

### 4. CRTDL Structure (External)

**Purpose**: Cohort definition and data extraction specification (TORCH input)

**Note**: This is an external format consumed by TORCH, not defined by aether. Included for reference only.

**Required Keys**:
```json
{
  "cohortDefinition": {
    "version": "string",
    "display": "string",
    "inclusionCriteria": [/* array of criterion groups */]
  },
  "dataExtraction": {
    "attributeGroups": [/* array of attribute groups */]
  }
}
```

**Validation**: Syntax validation only (structure checks), semantic validation delegated to TORCH server

**Location**: User-provided file, read as `[]byte` or `map[string]interface{}`

### 5. TORCH Parameters Resource (FHIR)

**Purpose**: FHIR Parameters resource for TORCH $extract-data operation

**Structure**:
```json
{
  "resourceType": "Parameters",
  "parameter": [
    {
      "name": "crtdl",
      "valueBase64Binary": "<base64-encoded CRTDL JSON>"
    }
  ]
}
```

**Go Representation**:
```go
type TORCHExtractionRequest struct {
    ResourceType string             `json:"resourceType"`
    Parameter    []TORCHParameter   `json:"parameter"`
}

type TORCHParameter struct {
    Name            string `json:"name"`
    ValueBase64Binary string `json:"valueBase64Binary,omitempty"`
}
```

**Location**: `internal/services/torch_client.go` (internal to TORCH client)

**Usage**: Marshaled to JSON and sent as POST body to `/fhir/$extract-data`

### 6. TORCH Extraction Response (FHIR)

**Purpose**: TORCH response containing extraction result file URLs

**Structure** (HTTP 200):
```json
{
  "resourceType": "Parameters",
  "parameter": [
    {
      "name": "output",
      "part": [
        {
          "name": "url",
          "valueUrl": "http://torch-server/output/batch-1.ndjson"
        },
        {
          "name": "url",
          "valueUrl": "http://torch-server/output/batch-2.ndjson"
        }
      ]
    }
  ]
}
```

**Go Representation**:
```go
type TORCHExtractionResult struct {
    ResourceType string                  `json:"resourceType"`
    Parameter    []TORCHResultParameter  `json:"parameter"`
}

type TORCHResultParameter struct {
    Name string          `json:"name"`
    Part []TORCHResultPart `json:"part,omitempty"`
}

type TORCHResultPart struct {
    Name     string `json:"name"`
    ValueURL string `json:"valueUrl,omitempty"`
}
```

**Location**: `internal/services/torch_client.go` (internal to TORCH client)

**Usage**: Parsed from HTTP 200 response, URLs extracted for download

## State Transitions

### InputType Detection Flow

```
User Input String
    ↓
DetectInputType()
    ↓
    ├─ os.Stat().IsDir() = true → InputTypeLocalDir
    ├─ HasPrefix("http") + Contains("/fhir/") → InputTypeTORCHURL
    ├─ HasPrefix("http") → InputTypeRemoteURL
    ├─ HasSuffix(".crtdl" | ".json") + ValidCRTDLSyntax() → InputTypeCRTDL
    └─ else → ERROR: unknown input type
```

### TORCH Extraction Lifecycle

```
CRTDL File
    ↓
ValidateCRTDLSyntax()
    ↓
SubmitExtraction() → HTTP 202 + Content-Location URL
    ↓
    │ [Store URL in job.TORCHExtractionURL]
    ↓
PollExtractionStatus(contentLocation)
    ├─ HTTP 202 → Sleep(interval) → Poll again
    ├─ HTTP 200 → Parse result URLs → DownloadFiles()
    └─ Timeout → ERROR
    ↓
DownloadFiles(resultURLs, jobDir)
    ↓
[Files in job directory, ready for import]
```

### Job State with TORCH

```go
// Job created with CRTDL input
job := PipelineJob{
    InputSource: "/path/to/query.crtdl",
    InputType:   InputTypeCRTDL,
    Status:      JobStatusPending,
}

// After submission
job.TORCHExtractionURL = "http://torch/fhir/extraction/abc123"
job.Status = JobStatusInProgress

// During polling (job can be resumed from this state)
// job.TORCHExtractionURL is used to resume polling

// After download completes
job.Status = JobStatusInProgress  // Import step continues
// job.TORCHExtractionURL preserved for audit trail
```

## Validation Rules

### TORCHConfig Validation

```go
func (c *TORCHConfig) Validate() error {
    if c.BaseURL == "" {
        return errors.New("TORCH base_url is required")
    }

    if _, err := url.Parse(c.BaseURL); err != nil {
        return fmt.Errorf("invalid TORCH base_url: %w", err)
    }

    if c.Username == "" {
        return errors.New("TORCH username is required")
    }

    if c.Password == "" {
        return errors.New("TORCH password is required")
    }

    if c.ExtractionTimeoutMinutes <= 0 {
        return fmt.Errorf("extraction_timeout_minutes must be > 0, got %d", c.ExtractionTimeoutMinutes)
    }

    if c.PollingIntervalSeconds <= 0 || c.PollingIntervalSeconds > 60 {
        return fmt.Errorf("polling_interval_seconds must be 1-60, got %d", c.PollingIntervalSeconds)
    }

    if c.MaxPollingIntervalSeconds < c.PollingIntervalSeconds {
        return fmt.Errorf("max_polling_interval_seconds (%d) must be >= polling_interval_seconds (%d)",
            c.MaxPollingIntervalSeconds, c.PollingIntervalSeconds)
    }

    return nil
}
```

### CRTDL Syntax Validation

See `research.md` Section 3 for validation implementation.

**Rules**:
1. Must be valid JSON
2. Must have `cohortDefinition` object with `inclusionCriteria` array
3. Must have `dataExtraction` object with `attributeGroups` array
4. File size < 1MB (sanity check)

## Data Flow Diagram

```
┌─────────────────┐
│   User Input    │
│  (CRTDL path)   │
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│  DetectInputType()  │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ ValidateCRTDLSyntax │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐     ┌──────────────┐
│  Create Job with    │────►│ PipelineJob  │
│  InputTypeCRTDL     │     │  + InputType │
└────────┬────────────┘     │  + TORCHUrl  │
         │                  └──────────────┘
         ▼
┌─────────────────────┐     ┌──────────────┐
│ SubmitExtraction()  │────►│   TORCH      │
│  (POST FHIR Params) │     │   Server     │
└────────┬────────────┘     └──────────────┘
         │                          │
         │  ◄──────────────────────┘
         │  HTTP 202 + Content-Location
         │
         ▼
┌─────────────────────┐
│ PollStatus() loop   │
│  GET Content-Loc    │◄────┐
└────────┬────────────┘     │
         │                  │
         ├─ HTTP 202 ───────┘
         │  (retry)
         │
         ├─ HTTP 200
         ▼
┌─────────────────────┐     ┌──────────────┐
│ ParseResultURLs()   │────►│   File URLs  │
└────────┬────────────┘     │   (NDJSON)   │
         │                  └──────────────┘
         ▼
┌─────────────────────┐
│  DownloadFiles()    │
│  to job directory   │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│  Import Step        │
│  (existing logic)   │
└─────────────────────┘
```

## Immutability Guarantees

Per Constitution Principle I (Functional Programming):

- **TORCHConfig**: Immutable after load (read-only throughout job lifecycle)
- **InputType constants**: Immutable strings
- **CRTDL content**: Read once, never modified
- **TORCHExtractionRequest**: Created immutably, marshaled to JSON
- **TORCHExtractionResult**: Parsed immutably, URLs extracted to slice
- **PipelineJob.TORCHExtractionURL**: Write-once field (set during submission, never modified)

**Mutation Points** (explicitly controlled):
- `PipelineJob.TORCHExtractionURL`: Set exactly once after submission
- `PipelineJob.Status`: Updated through explicit state transition functions
- Job state file: Updated after each state transition (existing pattern)

## Testing Strategy

### Unit Tests
- `DetectInputType()`: All input type branches
- `ValidateCRTDLSyntax()`: Valid/invalid CRTDL structures
- `TORCHConfig.Validate()`: Boundary conditions
- Base64 encoding: Round-trip verification

### Contract Tests
- TORCH API: FHIR Parameters format validation
- HTTP 202/200 response parsing
- Error response handling (4xx, 5xx)

### Integration Tests
- End-to-end: CRTDL file → extraction → download → import
- Polling timeout scenarios
- Job resumption after process restart
- Backward compatibility: local directory still works

See `tests/` structure in `plan.md` for file organization.
