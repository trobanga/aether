# Feature Specification: DUP Pipeline CLI

**Feature Branch**: `001-dup-pipeline-we`
**Created**: 2025-10-08
**Status**: Draft
**Input**: User description: "dup pipeline - We are going to build a command-line interface to run a DUP Pipeline for medical FHIR data."

## Clarifications

### Session 2025-10-08

- Q: CLI subcommand structure - How should pipeline steps be invoked? → A: Hybrid approach - Primary: pipeline workflow commands (`aether pipeline start`, `aether pipeline continue <job-id>`, `aether pipeline status <job-id>`) for normal usage respecting project config; Secondary: job-centric commands (`aether job run <job-id> --step <step-name>`, `aether job list`) for manual control and failure recovery. Step optionality configured in project settings, not at runtime.
- Q: TORCH data acquisition - How does the pipeline obtain TORCH extraction data? → A: Pipeline imports pre-extracted TORCH output (TORCH extraction already completed externally). User provides either a local directory path or download link to the TORCH output. Pipeline does not call TORCH API directly.
- Q: Data conversion mechanism - How should CSV/Parquet conversion be performed? → A: External conversion service via HTTP. Aether sends FHIR NDJSON to conversion service endpoints and receives flattened CSV or Parquet files in response (similar architecture to DIMP step).
- Q: Pipeline step failure recovery - How should users recover from failed steps? → A: Hybrid retry strategy - Automatic retry with exponential backoff for transient errors (network failures, timeouts, HTTP 5xx); Manual retry via `aether job run --step` required for non-transient errors (validation failures, malformed data, HTTP 4xx).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic Pipeline Execution (Priority: P1)

A data engineer has TORCH-extracted FHIR medical data (from a previous TORCH export) and wants to import it into the pipeline for processing. This is the core functionality that all other steps depend on.

**Why this priority**: Without importing the TORCH data, no other pipeline steps can function. This is the foundational capability that delivers immediate value by organizing TORCH output for downstream processing.

**Independent Test**: Can be fully tested by providing a TORCH output directory path or download URL and verifying that FHIR data files are imported and organized in a job-specific directory with a unique job identifier.

**Acceptance Scenarios**:

1. **Given** no previous pipeline runs exist and user has a local TORCH output directory, **When** user runs `aether pipeline start --input /path/to/torch/output`, **Then** system creates a new job ID, imports all FHIR NDJSON files to a job-specific directory, and marks import step complete
2. **Given** user has a TORCH output download URL, **When** user runs `aether pipeline start --input https://example.com/torch/export/job-123`, **Then** system downloads all FHIR NDJSON files and imports them to a job-specific directory, displaying a progress bar with percentage complete, elapsed time, ETA, and download throughput (MB/sec)
3. **Given** a pipeline import is in progress, **When** user runs `aether pipeline status <job-id>`, **Then** system displays current step (import), job ID, number of files imported, and elapsed time
4. **Given** a pipeline has completed import, **When** user runs `aether pipeline status <job-id>`, **Then** system shows completion status, total files imported, total data size, and completion timestamp
5. **Given** download URL is unreachable, **When** user runs `aether pipeline start --input <invalid-url>`, **Then** system reports connection error with clear guidance and does not create incomplete job data

---

### User Story 2 - Pipeline Resumption Across Sessions (Priority: P2)

A data engineer starts a pipeline but needs to close their terminal or switch computers. They want to resume monitoring and continue the pipeline from where it left off.

**Why this priority**: Long-running pipelines should not require continuous terminal sessions. This enables reliability and flexibility in operational workflows.

**Independent Test**: Can be tested by starting a pipeline, closing the CLI, reopening it later, and verifying that job status is accurately retrieved and the next pipeline step can be initiated.

**Acceptance Scenarios**:

1. **Given** a pipeline job was started in a previous session, **When** user runs `aether job list`, **Then** system displays all jobs with their IDs, current step, status (in-progress/completed/failed), retry count, and last update time
2. **Given** a pipeline completed data import in a previous session, **When** user runs `aether pipeline continue <job-id>`, **Then** system loads the imported data and proceeds with the next configured step (e.g., pseudonymization if enabled)
3. **Given** a pipeline job failed with transient error, **When** user runs `aether pipeline status <job-id>`, **Then** system shows error details, retry attempts made, and indicates automatic retry in progress
4. **Given** a pipeline job failed with non-transient error, **When** user runs `aether pipeline status <job-id>`, **Then** system shows error details (error type, validation message), last successful step, and suggests manual recovery via `aether job run --step`

---

### User Story 3 - Optional Pseudonymization (DIMP) (Priority: P3)

A data engineer needs to de-identify, minimize, and pseudonymize FHIR data before sharing it externally, as required by privacy regulations.

**Why this priority**: Not all projects require pseudonymization (internal analytics may use identified data). This is an enhancement for compliance scenarios, configured per-project.

**Independent Test**: Can be tested by configuring DIMP in project settings, running a pipeline, and verifying that output files contain pseudonymized identifiers while preserving FHIR structure.

**Acceptance Scenarios**:

1. **Given** project configuration has DIMP enabled and data import has completed, **When** pipeline continues automatically or via `aether pipeline continue <job-id>`, **Then** system sends each FHIR resource to the pseudonymization service and saves de-identified versions alongside original files
2. **Given** DIMP processing is configured and running, **When** user runs `aether pipeline status <job-id>`, **Then** system shows pseudonymization progress (files processed / total files) and estimated time remaining
3. **Given** DIMP service is unavailable during pipeline execution, **When** system attempts pseudonymization, **Then** system retries with exponential backoff and reports persistent failures clearly

---

### User Story 4 - Optional Data Format Conversion (Priority: P4)

A data analyst needs FHIR data converted to tabular formats (CSV or Parquet) for analysis in tools like Python pandas, R, or SQL databases.

**Why this priority**: Many downstream analytics tools cannot consume FHIR's hierarchical JSON/NDJSON format. This unlocks broader analytics capabilities.

**Independent Test**: Can be tested by configuring CSV/Parquet conversion in project settings, running a pipeline with FHIR data, and verifying that flattened tables are created with correct schemas.

**Acceptance Scenarios**:

1. **Given** project configuration has CSV conversion enabled and FHIR data is available (original or pseudonymized), **When** pipeline continues to conversion step, **Then** system sends FHIR resources to CSV conversion service and saves returned flattened CSV files per resource type
2. **Given** project configuration has Parquet conversion enabled, **When** pipeline continues to conversion step, **Then** system sends FHIR resources to Parquet conversion service and saves returned columnar Parquet files
3. **Given** conversion is in progress, **When** user runs `aether pipeline status <job-id>`, **Then** system displays format being generated, resource types processed, and output file locations
4. **Given** project configuration enables both CSV and Parquet, **When** conversion completes, **Then** both formats are available in separate subdirectories under the job folder
5. **Given** conversion service is unavailable during pipeline execution, **When** system attempts conversion, **Then** system retries with exponential backoff and reports persistent failures clearly

---

### Edge Cases

- What happens when a pipeline step is interrupted mid-processing (network failure, user cancellation)?
- How does the system handle extremely large FHIR datasets that may exceed disk space?
- What happens if the user tries to run a step out of order (e.g., DIMP before import)?
- How does the system handle conflicting job operations (two users trying to process the same job ID simultaneously)?
- What happens when imported TORCH data contains malformed FHIR resources?
- How does the system behave if configuration for optional services (DIMP, conversion) is missing?
- What happens when a service returns transient errors (HTTP 5xx, network timeout) vs. permanent errors (HTTP 4xx, validation failures)?
- How many automatic retry attempts should be made before requiring manual intervention?
- What happens if a retry succeeds after partial processing (e.g., 500 of 1000 files already processed)?

## Requirements *(mandatory)*

### Functional Requirements

#### Command Structure

- **FR-001**: System MUST provide `aether pipeline start` command to initiate a new pipeline job that generates a unique job identifier and executes configured steps
- **FR-002**: System MUST provide `aether pipeline continue <job-id>` command to resume a pipeline from the last completed step
- **FR-003**: System MUST provide `aether pipeline status <job-id>` command to retrieve detailed status for a specific job (step history, file inventory, progress, error logs)
- **FR-004**: System MUST provide `aether job list` command to display all pipeline jobs with their current status and progress
- **FR-005**: System MUST provide `aether job run <job-id> --step <step-name>` command for manual execution of individual pipeline steps (for failure recovery and advanced control)

#### Project Configuration

- **FR-006**: System MUST read project configuration file that specifies which pipeline steps are enabled (DIMP, validation, CSV conversion, Parquet conversion)
- **FR-007**: System MUST execute only the steps enabled in project configuration when running `aether pipeline start` or `aether pipeline continue`

#### Data Import

- **FR-008**: System MUST accept TORCH output as input via `--input` flag, supporting both local directory paths and HTTP/HTTPS download URLs
- **FR-009**: System MUST import FHIR NDJSON files from local directory path provided by user, copying them to job-specific directory structure (e.g., `jobs/<job-id>/import/`)
- **FR-010**: System MUST download FHIR NDJSON files from HTTP/HTTPS URL provided by user and save them to job-specific directory structure
- **FR-011**: System MUST validate that input source (directory or URL) contains valid FHIR NDJSON files before starting pipeline
- **FR-012**: System MUST report progress during download operations (files downloaded, data transferred, estimated time remaining)

#### Pipeline Execution

- **FR-013**: System MUST store pipeline state (current step, status, timestamps, file counts) that survives process termination
- **FR-014**: System MUST support optional DIMP (pseudonymization) step that sends FHIR resources to a pseudonymization service via HTTP
- **FR-015**: System MUST save pseudonymized FHIR data separately from original data (e.g., `jobs/<job-id>/pseudonymized/`)
- **FR-016**: System MUST support optional validation step placeholder (not implemented in v1.0; reserves position in pipeline sequence between DIMP and conversion steps; future specification will define validation service contract and acceptance criteria)
- **FR-017**: System MUST support optional CSV conversion step that sends FHIR resources to a CSV conversion service via HTTP and saves returned flattened files
- **FR-018**: System MUST support optional Parquet conversion step that sends FHIR resources to a Parquet conversion service via HTTP and saves returned flattened files
- **FR-019**: System MUST organize output files by conversion format (e.g., `jobs/<job-id>/csv/`, `jobs/<job-id>/parquet/`)
- **FR-020**: System MUST enforce correct pipeline step sequencing: Import → DIMP (if enabled) → Validation (if enabled) → Conversion (if enabled)
- **FR-021**: System MUST prevent users from running a step via `aether job run --step` if prerequisite steps have not completed successfully
- **FR-022**: System MUST provide clear error messages when external services (DIMP, conversion) or download URLs are unreachable
- **FR-023**: System MUST distinguish between transient errors (network failures, timeouts, HTTP 5xx) and non-transient errors (validation failures, malformed data, HTTP 4xx)
- **FR-024**: System MUST automatically retry pipeline steps that fail with transient errors using exponential backoff strategy (e.g., retry after 1s, 2s, 4s, 8s, etc.)
- **FR-025**: System MUST stop automatic retries and require manual intervention for non-transient errors
- **FR-026**: System MUST track retry attempts per pipeline step and include retry count in job status output
- **FR-027**: System MUST log all pipeline operations (downloads, file operations, API calls, errors, retry attempts) for troubleshooting
- **FR-028**: System MUST support concurrent execution of different pipeline jobs without data corruption
- **FR-029**: System MUST display human-readable progress indicators during long-running operations (>30 seconds) with the following requirements:
  - **FR-029a**: Progress indicators MUST show completion percentage (e.g., "45%") for operations with known total size (file downloads, file processing)
  - **FR-029b**: Progress indicators MUST show elapsed time and estimated time remaining (ETA) calculated as: `ETA = (total_items - processed_items) * avg_time_per_item` where `avg_time_per_item` is computed from the last 10 processed items or last 30 seconds of throughput (whichever is more recent)
  - **FR-029c**: Progress indicators MUST use visual progress bars for operations with known progress (percentage-based), and animated spinners for operations with unknown duration (service calls, network operations)
  - **FR-029d**: Progress indicators MUST update at least every 2 seconds during active operations
  - **FR-029e**: Progress indicators MUST display current operation name, items processed/total (e.g., "Processing FHIR files: 127/500"), and throughput rate (e.g., "2.3 files/sec" or "5.2 MB/sec") for batch operations
- **FR-030**: System MUST allow users to configure service endpoints (DIMP URL, CSV conversion URL, Parquet conversion URL) via configuration file with CLI flag overrides (defaults read from config file; CLI flags override config file values when provided)

### Assumptions

- TORCH extraction has been performed externally before pipeline execution
- TORCH output follows standard FHIR NDJSON format (newline-delimited JSON with one resource per line)
- DIMP pseudonymization service accepts single FHIR resources via POST and returns pseudonymized versions
- CSV conversion service accepts FHIR NDJSON via HTTP and returns flattened CSV files
- Parquet conversion service accepts FHIR NDJSON via HTTP and returns flattened Parquet files
- Users have sufficient disk space for storing pipeline data (no quota enforcement in v1)
- Pipeline jobs are single-user operations (no multi-user collaboration on same job)
- Job data retention is manual (no automatic cleanup policies in v1)
- If TORCH output is provided as URL, it remains accessible for the duration of the download
- Progress indicators will be implemented using a Go library that supports progress bars and spinners (recommended: `schollz/progressbar` or `cheggaaa/pb` for their cross-platform CLI support and rich formatting options)

### Key Entities

- **Pipeline Job**: Represents a single execution of the DUP pipeline with unique identifier, creation timestamp, current step, overall status, input source (directory path or URL), and directory path for all artifacts
- **Pipeline Step**: A discrete stage in the pipeline (Import, DIMP pseudonymization, Validation, CSV conversion, Parquet conversion) with status, start/end timestamps, input/output file references, and error information
- **FHIR Data File**: An NDJSON file containing FHIR resources with filename, resource type, file size, and source step (imported, pseudonymized, or converted)
- **Project Configuration**: Per-project settings that define which pipeline steps are enabled (DIMP, validation, CSV conversion, Parquet conversion) and service endpoints for each external service
- **Service Configuration**: Connection details for external HTTP services (DIMP, CSV conversion, Parquet conversion) including base URL, timeout settings, and authentication credentials (if applicable)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can complete a full pipeline run (Import → DIMP → conversion to CSV/Parquet) without manual file transfers between steps
- **SC-002**: Users can resume pipeline operations after closing and reopening the CLI with 100% state accuracy
- **SC-003**: Pipeline status queries return complete information (job ID, current step, progress percentage, file counts) in under 2 seconds
- **SC-004**: System successfully processes FHIR datasets of at least 10GB without data loss or corruption
- **SC-005**: Error messages for download failures and service connectivity issues provide actionable guidance (service name, endpoint, suggested troubleshooting) in 90% of failure scenarios
- **SC-006**: Users can track progress of long-running operations (>30 seconds) with visual feedback that updates at least every 2 seconds, showing completion percentage, elapsed time, ETA, and throughput rate
- **SC-007**: Pipeline jobs can be uniquely identified and retrieved from past runs for at least 30 days of operation
