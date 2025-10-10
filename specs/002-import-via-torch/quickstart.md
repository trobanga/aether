# Quickstart: TORCH Server Data Import

**Feature**: 002-import-via-torch | **Date**: 2025-10-10

## Overview

This guide helps developers get started implementing TORCH integration for aether. Follow these steps to build and test the feature incrementally, adhering to TDD principles.

## Prerequisites

- Go 1.25.1+ installed
- TORCH server running (or mock server for testing)
- Existing aether codebase on branch `002-import-via-torch`
- Familiarity with existing aether pipeline architecture

## Development Workflow (TDD)

**CRITICAL**: All tests must be written and approved BEFORE implementation. See Constitution Principle II.

### Step 0: Environment Setup

**Goal**: Configure TORCH server access

1. **Start TORCH server** (or use existing instance):
   ```bash
   # Option 1: Use dse-example TORCH
   cd /home/trobanga/development/mii/dse-example/torch
   docker compose up -d

   # Option 2: Configure existing TORCH server
   # (ensure you have access credentials)
   ```

2. **Create test CRTDL file**:
   ```bash
   # Copy example CRTDL for testing
   cp /home/trobanga/development/mii/dse-example/torch/queries/example-crtdl.json \
      test-data/test.crtdl
   ```

3. **Configure aether for TORCH**:
   ```yaml
   # config/aether.example.yaml
   services:
     torch:
       base_url: "http://localhost:8080"
       username: "test"
       password: "test"
       extraction_timeout_minutes: 30
       polling_interval_seconds: 5
       max_polling_interval_seconds: 30
   ```

### Step 1: Data Model (Tests First!)

**Goal**: Define TORCH configuration and extend job model

**Test File**: `tests/unit/config_loading_test.go`

**Write Tests** (RED phase):
```go
func TestLoadConfigWithTORCH(t *testing.T) {
    // Test: TORCH config loads from YAML
    // Test: Default values applied when not specified
    // Test: Validation catches invalid URLs, empty credentials, etc.
}
```

**Implement** (GREEN phase):
```go
// internal/models/config.go

type TORCHConfig struct {
    BaseURL                   string `yaml:"base_url" json:"base_url"`
    Username                  string `yaml:"username" json:"username"`
    Password                  string `yaml:"password" json:"password"`
    ExtractionTimeoutMinutes  int    `yaml:"extraction_timeout_minutes" json:"extraction_timeout_minutes"`
    PollingIntervalSeconds    int    `yaml:"polling_interval_seconds" json:"polling_interval_seconds"`
    MaxPollingIntervalSeconds int    `yaml:"max_polling_interval_seconds" json:"max_polling_interval_seconds"`
}

// Add to ServiceConfig
type ServiceConfig struct {
    // ... existing fields ...
    TORCH TORCHConfig `yaml:"torch" json:"torch"`
}

// Update DefaultConfig() to include TORCH defaults
```

**Verify**: Run tests → all pass → commit

---

### Step 2: Input Type Detection (Tests First!)

**Goal**: Detect CRTDL files and TORCH URLs

**Test File**: `tests/unit/input_detection_test.go` (new file)

**Write Tests** (RED phase):
```go
func TestDetectInputType(t *testing.T) {
    // Test: Local directory detected
    // Test: CRTDL file detected by extension and structure
    // Test: TORCH URL detected by pattern
    // Test: Regular HTTP URL detected
    // Test: Invalid input returns error
}

func TestIsCRTDLFile(t *testing.T) {
    // Test: Valid CRTDL file returns true
    // Test: Missing cohortDefinition returns false
    // Test: Missing dataExtraction returns false
    // Test: Invalid JSON returns false
}
```

**Implement** (GREEN phase):
```go
// internal/lib/validation.go (new functions)

func DetectInputType(inputSource string) (string, error) {
    // Implementation from research.md Section 1
}

func IsCRTDLFile(path string) bool {
    // Implementation from research.md Section 1
}

// internal/models/job.go (new constants)

const (
    InputTypeLocalDir   = "local_directory"
    InputTypeRemoteURL  = "remote_url"
    InputTypeCRTDL      = "crtdl_file"
    InputTypeTORCHURL   = "torch_result_url"
}
```

**Verify**: Run tests → all pass → commit

---

### Step 3: CRTDL Validation (Tests First!)

**Goal**: Validate CRTDL syntax before submission

**Test File**: `tests/unit/crtdl_validation_test.go` (new file)

**Write Tests** (RED phase):
```go
func TestValidateCRTDLSyntax(t *testing.T) {
    // Test: Valid CRTDL passes
    // Test: Missing cohortDefinition fails
    // Test: Missing dataExtraction fails
    // Test: Missing inclusionCriteria fails
    // Test: Missing attributeGroups fails
    // Test: Invalid JSON fails
    // Test: File too large (>1MB) fails
}
```

**Implement** (GREEN phase):
```go
// internal/lib/validation.go

func ValidateCRTDLSyntax(crtdlPath string) error {
    // Implementation from research.md Section 3
}
```

**Verify**: Run tests → all pass → commit

---

### Step 4: TORCH Client - Extraction Submission (Tests First!)

**Goal**: Submit CRTDL to TORCH and get extraction URL

**Test File**: `tests/contract/torch_service_test.go` (new file)

**Write Tests** (RED phase):
```go
func TestTORCHClient_SubmitExtraction(t *testing.T) {
    // Test: Valid CRTDL submission returns 202 + Content-Location
    // Test: Invalid auth returns 401
    // Test: Malformed request returns 400
    // Test: CRTDL base64 encoding is correct
    // Test: Request includes proper FHIR Parameters structure
}
```

**Implement** (GREEN phase):
```go
// internal/services/torch_client.go (new file)

type TORCHClient struct {
    baseURL    string
    username   string
    password   string
    httpClient *HTTPClient
    logger     *Logger
}

func NewTORCHClient(config TORCHConfig, httpClient *HTTPClient, logger *Logger) *TORCHClient {
    // Constructor
}

func (c *TORCHClient) SubmitExtraction(crtdlPath string) (string, error) {
    // 1. Read CRTDL file
    // 2. Encode as base64
    // 3. Build FHIR Parameters request
    // 4. POST to /fhir/$extract-data with Basic auth
    // 5. Parse Content-Location from response
    // 6. Return extraction URL
}
```

**Verify**: Run contract tests → all pass → commit

---

### Step 5: TORCH Client - Status Polling (Tests First!)

**Goal**: Poll extraction status until complete

**Test File**: `tests/unit/torch_client_test.go` (new file)

**Write Tests** (RED phase):
```go
func TestTORCHClient_PollExtractionStatus(t *testing.T) {
    // Test: HTTP 202 → continue polling
    // Test: HTTP 200 → parse result URLs
    // Test: Timeout exceeded → return error
    // Test: Server error → return error
    // Test: Exponential backoff intervals
}
```

**Implement** (GREEN phase):
```go
// internal/services/torch_client.go

func (c *TORCHClient) PollExtractionStatus(contentLocation string, timeout time.Duration) ([]string, error) {
    // Implementation from research.md Section 4
    // Returns array of file URLs
}
```

**Verify**: Run unit tests → all pass → commit

---

### Step 6: TORCH Client - File Download (Tests First!)

**Goal**: Download NDJSON files from TORCH

**Test File**: `tests/unit/torch_client_test.go` (add to existing)

**Write Tests** (RED phase):
```go
func TestTORCHClient_DownloadExtractionFiles(t *testing.T) {
    // Test: Multiple files downloaded to job directory
    // Test: Progress tracking for downloads
    // Test: Retry on transient errors
    // Test: Authentication included in download requests
}
```

**Implement** (GREEN phase):
```go
// internal/services/torch_client.go

func (c *TORCHClient) DownloadExtractionFiles(fileURLs []string, jobDir string, showProgress bool) error {
    // Reuse downloader.go logic
    // Download each URL to jobDir
}
```

**Verify**: Run unit tests → all pass → commit

---

### Step 7: Pipeline Integration (Tests First!)

**Goal**: Integrate TORCH extraction into import step

**Test File**: `tests/integration/pipeline_torch_test.go` (new file)

**Write Tests** (RED phase):
```go
func TestPipeline_TORCHExtraction_EndToEnd(t *testing.T) {
    // Test: CRTDL input triggers TORCH extraction
    // Test: Files downloaded and imported
    // Test: Job state tracks TORCH URL
    // Test: Job can resume if interrupted during polling
}

func TestPipeline_TORCHResultURL_Direct(t *testing.T) {
    // Test: Direct TORCH URL downloads files
    // Test: Skips submission/polling steps
}

func TestPipeline_BackwardCompatibility(t *testing.T) {
    // Test: Local directory input still works
    // Test: Remote URL input still works
}
```

**Implement** (GREEN phase):
```go
// internal/pipeline/import.go (modify existing)

func ExecuteImportStep(job *PipelineJob, logger *Logger, httpClient *HTTPClient, showProgress bool) (*PipelineJob, error) {
    // Add cases for InputTypeCRTDL and InputTypeTORCHURL

    switch job.InputType {
    case InputTypeLocalDir:
        // existing logic
    case InputTypeRemoteURL:
        // existing logic
    case InputTypeCRTDL:
        return executeTORCHExtraction(job, logger, httpClient, showProgress)
    case InputTypeTORCHURL:
        return executeTORCHDownload(job, logger, httpClient, showProgress)
    }
}

func executeTORCHExtraction(job *PipelineJob, ...) (*PipelineJob, error) {
    // 1. Validate CRTDL
    // 2. Create TORCH client
    // 3. Submit extraction
    // 4. Store extraction URL in job
    // 5. Poll status
    // 6. Download files
    // 7. Update job state
}

func executeTORCHDownload(job *PipelineJob, ...) (*PipelineJob, error) {
    // Direct download from TORCH URL
}
```

**Verify**: Run integration tests → all pass → commit

---

### Step 8: CLI Integration (Tests First!)

**Goal**: Update CLI to detect CRTDL input

**Test File**: `tests/integration/pipeline_torch_test.go` (add CLI tests)

**Write Tests** (RED phase):
```go
func TestCLI_TORCHInput(t *testing.T) {
    // Test: `aether pipeline start --input file.crtdl` works
    // Test: Input type detected automatically
    // Test: Error messages for invalid CRTDL
}
```

**Implement** (GREEN phase):
```go
// cmd/pipeline.go (modify runPipelineStart)

func runPipelineStart(cmd *cobra.Command, args []string) error {
    // ... existing code ...

    // Before CreateJob, detect input type
    inputType, err := lib.DetectInputType(inputSource)
    if err != nil {
        return fmt.Errorf("invalid input source: %w", err)
    }

    // If CRTDL, validate syntax
    if inputType == models.InputTypeCRTDL {
        if err := lib.ValidateCRTDLSyntax(inputSource); err != nil {
            return fmt.Errorf("invalid CRTDL file: %w", err)
        }
    }

    // CreateJob with detected type
    job, err := pipeline.CreateJobWithType(inputSource, inputType, *config)

    // ... rest of existing code ...
}
```

**Verify**: Run integration tests → all pass → commit

---

### Step 9: Configuration Validation (Tests First!)

**Goal**: Validate TORCH connectivity on startup

**Test File**: `tests/integration/torch_connectivity_test.go` (new file)

**Write Tests** (RED phase):
```go
func TestTORCH_ConnectivityValidation(t *testing.T) {
    // Test: Valid TORCH server passes connectivity check
    // Test: Invalid URL fails with clear error
    // Test: Invalid credentials fail with 401 error
    // Test: Unreachable server fails with timeout error
}
```

**Implement** (GREEN phase):
```go
// internal/models/config.go (modify ValidateServiceConnectivity)

func (c *ProjectConfig) ValidateServiceConnectivity() error {
    // ... existing DIMP check ...

    // Add TORCH check
    if c.Services.TORCH.BaseURL != "" {
        torchClient := services.NewTORCHClient(c.Services.TORCH, ...)
        if err := torchClient.Ping(); err != nil {
            return fmt.Errorf("TORCH connectivity check failed: %w", err)
        }
    }

    return nil
}

// internal/services/torch_client.go

func (c *TORCHClient) Ping() error {
    // Simple GET to base URL or metadata endpoint
}
```

**Verify**: Run integration tests → all pass → commit

---

### Step 10: End-to-End Testing

**Goal**: Verify complete workflow with real TORCH server

**Manual Test Steps**:

1. **Start TORCH server**:
   ```bash
   cd /home/trobanga/development/mii/dse-example/torch
   docker compose up -d
   ```

2. **Run extraction**:
   ```bash
   ./bin/aether pipeline start --input test-data/test.crtdl --verbose
   ```

3. **Verify**:
   - CRTDL validation passes
   - Extraction submitted to TORCH
   - Polling shows progress
   - Files downloaded to job directory
   - Import step processes files

4. **Test backward compatibility**:
   ```bash
   # Should still work
   ./bin/aether pipeline start --input test-data/
   ```

5. **Test error scenarios**:
   ```bash
   # Invalid CRTDL
   echo '{}' > invalid.crtdl
   ./bin/aether pipeline start --input invalid.crtdl
   # Should fail with validation error

   # Wrong credentials
   # (modify config with bad password)
   ./bin/aether pipeline start --input test-data/test.crtdl
   # Should fail with authentication error
   ```

---

## Common Issues & Solutions

### Issue: CRTDL validation fails
**Solution**: Check that CRTDL file has `cohortDefinition` and `dataExtraction` keys. Use example CRTDL as template.

### Issue: TORCH connection refused
**Solution**: Verify TORCH server is running (`docker ps`). Check `base_url` in config matches TORCH port.

### Issue: Authentication fails (401)
**Solution**: Verify `username` and `password` in config. Test with `curl` to TORCH endpoint.

### Issue: Extraction timeout
**Solution**: Increase `extraction_timeout_minutes` in config. Check TORCH logs for server-side issues.

### Issue: Tests failing with "address already in use"
**Solution**: Mock TORCH server port conflict. Use random port or teardown cleanly in tests.

## Next Steps

After completing quickstart:

1. Run full test suite: `go test ./...`
2. Check test coverage: `go test -cover ./...`
3. Review code against Constitution principles
4. Submit PR with tests + implementation
5. Proceed to `/speckit.tasks` for production tasks

## Reference Documentation

- [research.md](./research.md) - Technical decisions and alternatives
- [data-model.md](./data-model.md) - Data structures and relationships
- [contracts/torch-api.md](./contracts/torch-api.md) - TORCH API specification
- [plan.md](./plan.md) - Overall implementation plan
- [spec.md](./spec.md) - Feature requirements
