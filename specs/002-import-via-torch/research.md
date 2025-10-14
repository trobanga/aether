# Research: TORCH Server Data Import

**Feature**: 002-import-via-torch | **Date**: 2025-10-10

## Overview

This document captures technical research and design decisions for integrating TORCH (FHIR data extraction server) with the aether pipeline. Research focuses on TORCH API patterns, input type detection strategies, and integration points with existing infrastructure.

## Technical Decisions

### 1. Input Type Detection Strategy

**Decision**: File extension + content inspection hybrid approach

**Rationale**:
- **File extension check first**: `.crtdl`, `.json` extensions trigger CRTDL candidate check
- **Content validation**: Parse JSON and verify required CRTDL structure (`cohortDefinition`, `dataExtraction` keys)
- **URL pattern matching**: TORCH result URLs match pattern `*/fhir/*` (from TORCH API spec)
- **Directory detection**: Use `os.Stat().IsDir()` to identify local directory inputs (existing behavior)

**Implementation**:
```go
func DetectInputType(inputSource string) (InputType, error) {
    // 1. Check if directory
    if stat, err := os.Stat(inputSource); err == nil && stat.IsDir() {
        return InputTypeLocalDir, nil
    }

    // 2. Check if HTTP URL
    if strings.HasPrefix(inputSource, "http://") || strings.HasPrefix(inputSource, "https://") {
        // Check if TORCH result URL pattern
        if strings.Contains(inputSource, "/fhir/") {
            return InputTypeTORCHURL, nil
        }
        return InputTypeRemoteURL, nil  // existing HTTP download
    }

    // 3. Check if CRTDL file
    if strings.HasSuffix(inputSource, ".crtdl") || strings.HasSuffix(inputSource, ".json") {
        if isCRTDLFile(inputSource) {
            return InputTypeCRTDL, nil
        }
    }

    return InputTypeUnknown, fmt.Errorf("unable to determine input type: %s", inputSource)
}

func isCRTDLFile(path string) bool {
    data, err := os.ReadFile(path)
    if err != nil {
        return false
    }

    var crtdl map[string]interface{}
    if err := json.Unmarshal(data, &crtdl); err != nil {
        return false
    }

    // Verify required CRTDL structure
    _, hasCohort := crtdl["cohortDefinition"]
    _, hasExtraction := crtdl["dataExtraction"]
    return hasCohort && hasExtraction
}
```

**Alternatives Considered**:
- **Magic number/MIME type**: Rejected - CRTDL files are standard JSON without unique magic bytes
- **Explicit flag (`--type crtdl`)**: Rejected - violates KISS principle, adds UI complexity
- **Extension-only detection**: Rejected - too fragile, many `.json` files aren't CRTDL

### 2. TORCH API Integration Pattern

**Decision**: Follow existing `dimp_client.go` service pattern

**Rationale**:
- **Consistency**: Matches established codebase patterns
- **Proven pattern**: DIMP client already handles HTTP requests, auth, retries successfully
- **Minimal learning curve**: Developers familiar with existing pattern
- **Reuse infrastructure**: Leverages `httpclient.go` and retry logic

**TORCH Client Interface**:
```go
type TORCHClient struct {
    baseURL    string
    username   string
    password   string
    httpClient *HTTPClient
    logger     *Logger
}

// SubmitExtraction submits CRTDL to TORCH and returns Content-Location URL
func (c *TORCHClient) SubmitExtraction(crtdlPath string) (string, error)

// PollExtractionStatus polls until extraction completes, returns result URL or error
func (c *TORCHClient) PollExtractionStatus(contentLocation string, timeout time.Duration) (string, error)

// DownloadExtractionFiles downloads all NDJSON files from extraction result to job directory
func (c *TORCHClient) DownloadExtractionFiles(resultURL string, jobDir string) error
```

**API Flow** (from dse-example/torch):
1. **Submit**: POST to `/fhir/$extract-data` with FHIR Parameters resource containing base64-encoded CRTDL
2. **Response**: HTTP 202 with `Content-Location` header pointing to status endpoint
3. **Poll**: GET `Content-Location` URL repeatedly
   - HTTP 202: Still processing, continue polling
   - HTTP 200: Complete, response contains links to NDJSON files
4. **Download**: GET each NDJSON file URL from result response

**Alternatives Considered**:
- **Generic HTTP service**: Rejected - TORCH has specific workflow (submit → poll → download), warrants dedicated client
- **Webhook-based**: Rejected - adds infrastructure complexity, TORCH doesn't support webhooks
- **Async/channel-based polling**: Rejected - unnecessary complexity, simple loop sufficient

### 3. CRTDL Validation Strategy

**Decision**: Two-tier validation (syntax + semantic)

**Tier 1 - Syntax Validation** (fast, synchronous):
- Valid JSON structure
- Required top-level keys present: `cohortDefinition`, `dataExtraction`
- `cohortDefinition` has required structure (array of inclusion criteria)
- `dataExtraction` has `attributeGroups` array

**Tier 2 - Semantic Validation** (TORCH server-side):
- Let TORCH server validate cohort logic, attribute references, code systems
- Parse TORCH error responses and surface to user

**Rationale**:
- **Fast feedback**: Syntax errors caught immediately before network call
- **Avoid duplication**: TORCH already validates semantics, don't reimplement
- **Flexibility**: TORCH validation rules may evolve, keeping validation server-side avoids client updates
- **Error clarity**: TORCH errors are authoritative, client shouldn't guess

**Implementation**:
```go
func ValidateCRTDLSyntax(crtdlPath string) error {
    data, err := os.ReadFile(crtdlPath)
    if err != nil {
        return fmt.Errorf("failed to read CRTDL file: %w", err)
    }

    var crtdl map[string]interface{}
    if err := json.Unmarshal(data, &crtdl); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }

    // Check required keys
    cohort, hasCohort := crtdl["cohortDefinition"]
    if !hasCohort {
        return fmt.Errorf("missing required key: cohortDefinition")
    }

    extraction, hasExtraction := crtdl["dataExtraction"]
    if !hasExtraction {
        return fmt.Errorf("missing required key: dataExtraction")
    }

    // Validate cohortDefinition structure
    cohortMap, ok := cohort.(map[string]interface{})
    if !ok {
        return fmt.Errorf("cohortDefinition must be an object")
    }
    if _, hasInclusion := cohortMap["inclusionCriteria"]; !hasInclusion {
        return fmt.Errorf("cohortDefinition missing inclusionCriteria")
    }

    // Validate dataExtraction structure
    extractionMap, ok := extraction.(map[string]interface{})
    if !ok {
        return fmt.Errorf("dataExtraction must be an object")
    }
    if _, hasGroups := extractionMap["attributeGroups"]; !hasGroups {
        return fmt.Errorf("dataExtraction missing attributeGroups")
    }

    return nil
}
```

**Alternatives Considered**:
- **JSON Schema validation**: Rejected - CRTDL schema may not be stable, adds dependency
- **Full semantic validation**: Rejected - duplicates TORCH logic, brittle to changes
- **No validation**: Rejected - poor user experience, network round-trip for syntax errors

### 4. Polling Strategy

**Decision**: Exponential backoff with configurable timeout

**Configuration**:
```yaml
torch:
  base_url: "http://localhost:8080"
  username: "test"
  password: "test"
  extraction_timeout_minutes: 30
  polling_interval_seconds: 5
  max_polling_interval_seconds: 30
```

**Algorithm**:
```go
func (c *TORCHClient) PollExtractionStatus(contentLocation string, timeout time.Duration) (string, error) {
    deadline := time.Now().Add(timeout)
    interval := c.config.PollingIntervalSeconds * time.Second
    maxInterval := c.config.MaxPollingIntervalSeconds * time.Second

    for {
        if time.Now().After(deadline) {
            return "", fmt.Errorf("extraction timeout exceeded: %v", timeout)
        }

        resp, err := c.httpClient.Get(contentLocation)
        if err != nil {
            return "", fmt.Errorf("failed to poll extraction status: %w", err)
        }

        switch resp.StatusCode {
        case 200:
            // Extraction complete, parse result URL
            return parseExtractionResult(resp)
        case 202:
            // Still processing, continue polling
            time.Sleep(interval)
            // Exponential backoff
            interval = min(interval * 2, maxInterval)
        default:
            return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
        }
    }
}
```

**Rationale**:
- **Exponential backoff**: Reduces server load for long-running extractions
- **Configurable timeout**: Different cohorts may take different times
- **Max interval cap**: Prevents excessively long waits between checks
- **Simple loop**: No complex event system, easy to understand and debug

**Alternatives Considered**:
- **Fixed interval**: Rejected - inefficient for long extractions
- **Adaptive polling (based on cohort size)**: Rejected - premature optimization, no data to guide
- **WebSocket/SSE**: Rejected - TORCH doesn't support, adds complexity

### 5. Configuration Structure

**Decision**: Add TORCH config to existing `ServiceConfig` structure

**Implementation**:
```go
type ServiceConfig struct {
    DIMPUrl              string      `yaml:"dimp_url" json:"dimp_url"`
    CSVConversionUrl     string      `yaml:"csv_conversion_url" json:"csv_conversion_url"`
    ParquetConversionUrl string      `yaml:"parquet_conversion_url" json:"parquet_conversion_url"`
    TORCH                TORCHConfig `yaml:"torch" json:"torch"`  // NEW
}

type TORCHConfig struct {
    BaseURL                    string `yaml:"base_url" json:"base_url"`
    Username                   string `yaml:"username" json:"username"`
    Password                   string `yaml:"password" json:"password"`
    ExtractionTimeoutMinutes   int    `yaml:"extraction_timeout_minutes" json:"extraction_timeout_minutes"`
    PollingIntervalSeconds     int    `yaml:"polling_interval_seconds" json:"polling_interval_seconds"`
    MaxPollingIntervalSeconds  int    `yaml:"max_polling_interval_seconds" json:"max_polling_interval_seconds"`
}
```

**Default Values**:
```go
TORCH: TORCHConfig{
    BaseURL:                   "",
    Username:                  "",
    Password:                  "",
    ExtractionTimeoutMinutes:  30,
    PollingIntervalSeconds:    5,
    MaxPollingIntervalSeconds: 30,
},
```

**Rationale**:
- **Grouped under services**: TORCH is an external service like DIMP
- **Non-breaking**: Empty TORCH config doesn't affect existing workflows
- **Sensible defaults**: 30min timeout, 5s initial poll covers typical use cases
- **Validation**: Existing `ValidateServiceConnectivity()` can be extended to check TORCH

**Alternatives Considered**:
- **Top-level TORCH key**: Rejected - inconsistent with existing DIMP/CSV/Parquet under `services`
- **Environment variables**: Rejected - config file pattern already established
- **Separate config file**: Rejected - adds complexity, existing single-file pattern works well

### 6. Error Handling Strategy

**Decision**: Wrap TORCH errors with context, preserve original messages

**Pattern**:
```go
// Submission error
if err := torchClient.SubmitExtraction(crtdlPath); err != nil {
    return fmt.Errorf("TORCH extraction submission failed: %w\n\nCRTDL file: %s\nTORCH server: %s",
        err, crtdlPath, config.TORCH.BaseURL)
}

// Timeout error
if err := torchClient.PollExtractionStatus(url, timeout); err != nil {
    if errors.Is(err, ErrExtractionTimeout) {
        return fmt.Errorf("TORCH extraction exceeded timeout (%v):\n"+
            "Try increasing 'extraction_timeout_minutes' in config or reducing cohort size.\n"+
            "Extraction URL: %s", timeout, url)
    }
    return fmt.Errorf("TORCH extraction polling failed: %w", err)
}
```

**Error Types**:
- `ErrExtractionTimeout`: Exceeded configured timeout
- `ErrInvalidCRTDL`: Syntax validation failed
- `ErrTORCHServerError`: TORCH returned 4xx/5xx
- `ErrAuthenticationFailed`: 401 from TORCH

**Rationale**:
- **Actionable errors**: Include hints for resolution
- **Preserve context**: File paths, URLs help debugging
- **Error wrapping**: Go 1.13+ error wrapping for proper chains
- **Typed errors**: Allow callers to handle specific cases

**Alternatives Considered**:
- **Generic errors**: Rejected - poor user experience
- **Error codes**: Rejected - error wrapping more idiomatic in Go
- **Panic on errors**: Rejected - violates Go conventions, non-recoverable

## Integration Points

### 1. Import Step Integration

**Current Flow**:
```
CreateJob() → StartJob() → ExecuteImportStep()
  ↓
  InputTypeLocalDir → FindFHIRFiles() → ImportLocalFiles()
  InputTypeRemoteURL → DownloadFile() → ImportLocalFiles()
```

**New Flow with TORCH**:
```
CreateJob() → StartJob() → ExecuteImportStep()
  ↓
  DetectInputType()
    ├─ InputTypeLocalDir → FindFHIRFiles() → ImportLocalFiles()
    ├─ InputTypeRemoteURL → DownloadFile() → ImportLocalFiles()
    ├─ InputTypeCRTDL → ValidateCRTDL() → SubmitExtraction() → PollStatus() → DownloadFiles() → ImportLocalFiles()
    └─ InputTypeTORCHURL → DownloadFiles() → ImportLocalFiles()
```

**Key Changes**:
- `CreateJob()`: Call `DetectInputType()` to set `job.InputType`
- `ExecuteImportStep()`: Add case for `InputTypeCRTDL` and `InputTypeTORCHURL`
- TORCH-downloaded files stored in job directory, then processed like local files

### 2. Job State Tracking

**Add to Job struct**:
```go
type PipelineJob struct {
    // ... existing fields ...
    InputType           string `json:"input_type"`  // Existing field, new values
    TORCHExtractionURL  string `json:"torch_extraction_url,omitempty"`  // NEW: Content-Location URL for resume
}

// New InputType constants
const (
    InputTypeLocalDir   = "local_directory"   // existing
    InputTypeRemoteURL  = "remote_url"        // existing
    InputTypeCRTDL      = "crtdl_file"        // NEW
    InputTypeTORCHURL   = "torch_result_url"  // NEW
)
```

**Rationale**:
- `TORCHExtractionURL`: Allows job continuation if process dies during polling
- `InputType` values: Clear distinction for telemetry and debugging

### 3. Progress Tracking

**TORCH extraction phases**:
1. **Validating**: CRTDL syntax validation
2. **Submitting**: POST to TORCH
3. **Polling**: Waiting for extraction completion
4. **Downloading**: Fetching NDJSON files

**Integration with existing progress bar**:
- Reuse `progressbar` for download phase (existing pattern)
- Log messages for validation/submission/polling phases
- Estimated time remaining based on polling duration (no cohort size estimate available)

## Open Questions

None - all technical unknowns resolved through research.

## References

- TORCH API documentation: `/home/trobanga/development/mii/dse-example/torch/TORCH.md`
- Example CRTDL: `/home/trobanga/development/mii/dse-example/torch/queries/example-crtdl.json`
- Existing DIMP client pattern: `internal/services/dimp_client.go`
- HTTP client infrastructure: `internal/services/httpclient.go`
- Downloader pattern: `internal/services/downloader.go`
