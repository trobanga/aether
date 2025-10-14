# Feature Specification: TORCH Server Data Import

**Feature Branch**: `002-import-via-torch`
**Created**: 2025-10-10
**Status**: Draft
**Input**: User description: "Import via Torch Input

The current state is that we use input data from a directory, ie @test-data/
However, the idea is that we get a CRTDL file and use it to get that data from a Torch server.
/home/development/mii/dse-example/torch shows how to use Torch"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Data Extraction with CRTDL File (Priority: P1)

A researcher has a CRTDL (Cohort Representation for Trial Data Linking) file defining their patient cohort and wants to extract FHIR data directly from a TORCH server instead of manually downloading files to a local directory.

**Why this priority**: This is the core value proposition - enabling direct data extraction from TORCH without manual file handling, which is the primary workflow for research data use cases.

**Independent Test**: Can be fully tested by providing a CRTDL file path/URL as input to the pipeline start command and verifying that FHIR NDJSON data is successfully retrieved from TORCH and stored locally for processing.

**Acceptance Scenarios**:

1. **Given** a valid CRTDL file path and configured TORCH server, **When** user executes `aether pipeline start --input /path/to/query.crtdl`, **Then** the system authenticates with TORCH, submits the extraction request, and downloads the resulting NDJSON files
2. **Given** an active TORCH extraction job, **When** TORCH returns HTTP 202 (in progress), **Then** the system polls the extraction status until completion (HTTP 200) before downloading files
3. **Given** a completed TORCH extraction, **When** multiple NDJSON files are returned, **Then** all files are downloaded and stored in the job directory for subsequent pipeline steps
4. **Given** TORCH authentication credentials in configuration, **When** submitting extraction request, **Then** the system includes Basic authentication headers with base64-encoded credentials

---

### User Story 2 - Backward Compatibility with Local Directories (Priority: P1)

Existing users who currently use local directories (`test-data/`) should continue to work without any changes to their workflows.

**Why this priority**: Breaking existing functionality would disrupt current users and workflows. Backward compatibility is essential for smooth adoption.

**Independent Test**: Can be tested by running the existing command `aether pipeline start --input ./test-data/` and verifying it works identically to the current implementation.

**Acceptance Scenarios**:

1. **Given** a local directory with FHIR NDJSON files, **When** user executes `aether pipeline start --input ./test-data/`, **Then** the system processes files from the directory as it currently does
2. **Given** the system detects input is a directory (not a CRTDL file), **When** processing begins, **Then** no TORCH server communication is attempted

---

### User Story 3 - TORCH Server URL Input (Priority: P3)

A researcher wants to provide a direct TORCH extraction result URL instead of initiating a new extraction, allowing them to resume or reuse a previous extraction.

**Why this priority**: This is a convenience feature that enables reusing existing extractions but is not critical for the initial implementation.

**Independent Test**: Can be tested by providing a TORCH result URL (e.g., `http://localhost:8080/result/abc123`) and verifying the system downloads the NDJSON files without submitting a new extraction request.

**Acceptance Scenarios**:

1. **Given** a TORCH extraction result URL, **When** user executes `aether pipeline start --input http://localhost:8080/result/abc123`, **Then** the system downloads files directly from that URL without submitting a new CRTDL extraction
2. **Given** a result URL pointing to multiple NDJSON files, **When** downloading, **Then** all referenced files are retrieved and stored locally

---

### Edge Cases

- What happens when TORCH server is unreachable during extraction submission? System should fail with clear error message indicating connectivity issue
- What happens when TORCH extraction times out or gets stuck in "in progress" status indefinitely? System should timeout after configurable duration and fail with actionable error message
- What happens when TORCH returns no results (empty cohort)? System should handle gracefully, potentially completing the import step with zero files
- What happens when CRTDL file is malformed or invalid? System should validate file format before submission and return clear validation error
- What happens when authentication credentials are missing or invalid? System should fail early with authentication error before attempting extraction
- What happens when user provides a file path that could be either a CRTDL JSON or a directory? System should detect file type intelligently (check if path is directory, then check file extension/content)
- What happens when TORCH returns partial results or fails mid-download? System should use existing retry logic to recover from transient failures

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept CRTDL file paths as input to `pipeline start --input` command
- **FR-002**: System MUST detect input type automatically (local directory vs CRTDL file vs HTTP URL)
- **FR-003**: System MUST submit CRTDL extraction requests to configured TORCH server endpoint
- **FR-004**: System MUST encode CRTDL file content as base64 in Parameters resource according to TORCH API specification
- **FR-005**: System MUST authenticate with TORCH server using Basic authentication with configured credentials
- **FR-006**: System MUST poll TORCH extraction status when receiving HTTP 202 response
- **FR-007**: System MUST download all NDJSON files from TORCH extraction results
- **FR-008**: System MUST store downloaded TORCH data in job directory maintaining same structure as local directory imports
- **FR-009**: System MUST maintain backward compatibility with existing local directory input method
- **FR-010**: System MUST support direct TORCH result URL input for downloading previously-extracted data
- **FR-011**: System MUST validate TORCH server connectivity before starting extraction (leveraging existing connectivity validation)
- **FR-012**: System MUST use configured retry policy for TORCH API requests
- **FR-013**: System MUST timeout TORCH extraction polling after configurable duration
- **FR-014**: System MUST log all TORCH interactions (submission, polling, download) at appropriate log levels

### Key Entities *(include if feature involves data)*

- **CRTDL File**: JSON file containing cohort definition and data extraction specification with `cohortDefinition` and `dataExtraction` sections
- **TORCH Extraction Job**: Server-side extraction process identified by Content-Location URL, with states: submitted (202), completed (200), or failed
- **TORCH Parameters Resource**: FHIR Parameters resource containing base64-encoded CRTDL
- **Input Source Type**: Enumeration distinguishing between local directory, CRTDL file, TORCH URL, or remote HTTP URL (existing)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Researchers can submit CRTDL-based extractions and receive data without manual file downloading
- **SC-002**: TORCH extraction workflow completes end-to-end (submit → poll → download → process) with single command execution
- **SC-003**: System handles TORCH server unavailability gracefully with clear error messages within 5 seconds of detection
- **SC-004**: Existing users experience zero disruption - local directory imports continue working identically
- **SC-005**: TORCH extraction polling completes within reasonable timeframe (configurable timeout, default 30 minutes)
- **SC-006**: Downloaded TORCH data is indistinguishable from local directory imports for all downstream pipeline steps

## Assumptions *(optional)*

### Technical Assumptions

- TORCH server API follows specification shown in `/home/development/mii/dse-example/torch` example
- TORCH server returns extraction results as NDJSON files accessible via HTTP
- TORCH authentication uses Basic authentication with username/password
- Extraction status polling frequency is acceptable (reasonable default: every 5-10 seconds)
- TORCH extraction timeout default of 30 minutes is reasonable for typical cohorts

### Configuration Assumptions

- TORCH server base URL, username, and password are configured in `aether.yaml` (similar to existing FHIR server config)
- TORCH configuration section added to existing config structure without breaking changes
- Retry policy configuration applies to TORCH requests (reuse existing retry configuration)

### Workflow Assumptions

- Downloaded TORCH files have same NDJSON structure as existing test-data files
- Job directory structure remains unchanged - TORCH files stored in same input directory as local imports
- Progress tracking works identically for TORCH downloads as for local file imports
- For testing purposes, adapted CRTDL files (with specific patient selections) can be used instead of requiring separate patient override mechanism

## Dependencies *(optional)*

### External Dependencies

- TORCH server must be running and accessible at configured URL
- TORCH server must support FHIR `$extract-data` operation as documented
- Network connectivity to TORCH server required during extraction submission and file download

### Internal Dependencies

- Existing HTTP client service (`internal/services/httpclient.go`) can be leveraged for TORCH API calls
- Existing retry logic applies to TORCH requests
- Existing input type detection logic needs extension to recognize CRTDL files
- Existing job creation flow needs modification to handle TORCH input type

### Configuration Dependencies

- New configuration section for TORCH server settings (base URL, credentials, timeout)
- CRTDL file path/URL must be provided via `--input` flag (existing mechanism)
