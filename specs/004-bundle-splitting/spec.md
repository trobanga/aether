# Feature Specification: Bundle Splitting

**Feature Branch**: `004-bundle-splitting`
**Created**: 2025-10-22
**Status**: Draft
**Input**: User description: "Bundle Splitting"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automatic Large Bundle Processing (Priority: P1)

A data engineer runs the pipeline with FHIR data containing large Bundles (50MB+ with 100k+ resources). The system detects oversized Bundles and automatically splits them before sending to DIMP for pseudonymization, preventing HTTP 413 "Payload Too Large" errors.

**Why this priority**: This is the core problem that causes pipeline failures with large datasets. Without automatic splitting, users face cryptic HTTP errors and pipeline failures. This delivers immediate value by ensuring pipeline robustness with any dataset size.

**Independent Test**: Can be fully tested by processing a FHIR Bundle containing 100k Conditions (50MB+) and verifying that: (1) the Bundle is automatically detected as oversized, (2) split into smaller chunks, (3) each chunk is successfully pseudonymized, (4) pseudonymized results are reassembled, and (5) no 413 errors occur.

**Acceptance Scenarios**:

1. **Given** a FHIR Bundle with 100k Conditions (~50MB), **When** the DIMP step processes this Bundle, **Then** system detects size exceeds threshold (10MB), splits Bundle into chunks of ~5MB each, processes each chunk through DIMP, and writes reassembled pseudonymized Bundle to output
2. **Given** a FHIR Bundle smaller than the threshold (5MB), **When** the DIMP step processes this Bundle, **Then** system sends Bundle directly to DIMP without splitting (no performance overhead for normal-sized Bundles)
3. **Given** a large Bundle is being split and processed, **When** user checks pipeline status, **Then** system displays splitting progress (e.g., "Processing 15/20 Bundle chunks")
4. **Given** DIMP returns error for one chunk during split Bundle processing, **When** system processes chunks, **Then** system retries the failed chunk according to retry policy and reports which chunk failed
5. **Given** a FHIR resource that is not a Bundle (e.g., individual Patient resource), **When** the DIMP step processes it, **Then** system sends it to DIMP without modification (splitting only applies to Bundles)

---

### User Story 2 - Graceful Large Resource Handling (Priority: P2)

A data engineer processes FHIR data where individual non-Bundle resources are extremely large (e.g., a single Observation with massive nested data). The system detects when individual resources exceed payload limits and reports clear error messages with guidance.

**Why this priority**: While Bundles can be split, individual resources cannot be meaningfully split without violating FHIR semantics. Users need clear feedback when data quality issues prevent processing.

**Independent Test**: Can be tested by attempting to process a single Observation resource >30MB and verifying that system provides actionable error message directing user to data quality issues.

**Acceptance Scenarios**:

1. **Given** a single non-Bundle FHIR resource exceeds 30MB, **When** system attempts to send to DIMP, **Then** system logs detailed error with resource type, ID, size, and guidance to review data quality or increase server limits
2. **Given** a large individual resource fails to process, **When** pipeline continues, **Then** system marks that specific resource as failed, continues processing remaining resources, and generates failure report at job completion
3. **Given** multiple oversized individual resources in a batch, **When** DIMP step completes, **Then** system generates summary report listing all oversized resources with their IDs, types, and sizes

---

### User Story 3 - Configurable Split Threshold (Priority: P3)

A data engineer operating in environments with different DIMP server configurations needs to adjust the Bundle splitting threshold to match their server's payload limits (e.g., 50MB instead of default 10MB).

**Why this priority**: Different deployment environments have different constraints. Making the threshold configurable provides flexibility without requiring code changes.

**Independent Test**: Can be tested by setting split threshold to 5MB in configuration, processing a 7MB Bundle, and verifying it gets split. Then set threshold to 20MB and verify the same 7MB Bundle is not split.

**Acceptance Scenarios**:

1. **Given** project configuration sets `bundle_split_threshold_mb: 5`, **When** system processes an 8MB Bundle, **Then** system splits the Bundle before sending to DIMP
2. **Given** project configuration omits bundle_split_threshold_mb, **When** system processes Bundles, **Then** system uses default threshold of 10MB
3. **Given** invalid threshold configuration (e.g., negative number), **When** pipeline starts, **Then** system validates configuration and reports error with guidance before processing any data

---

### Edge Cases

- What happens when a Bundle entry references other entries within the same Bundle (entry relationships)?
- How does the system handle Bundle entries with different resource types (mixed Patient, Observation, Condition)?
- What happens if DIMP pseudonymization changes the size of resources (making them larger)?
- How does the system calculate Bundle size (raw JSON bytes, compressed size, or logical size)?
- What happens when splitting results in a chunk that is still too large (pathological case)?
- How does the system handle Bundle.entry.fullUrl references that span across split chunks?
- What happens if the original Bundle has specific Bundle.type (transaction, batch, document) - does splitting preserve semantics?
- How does reassembly maintain original Bundle metadata (id, timestamp, signature)?
- What happens when memory is constrained and the system cannot load the entire Bundle to split it?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST detect when a FHIR Bundle exceeds configured size threshold before sending to DIMP service
- **FR-002**: System MUST split oversized Bundles into smaller chunks, each containing a subset of the original Bundle.entry array
- **FR-003**: System MUST send each Bundle chunk to DIMP for pseudonymization independently
- **FR-004**: System MUST reassemble pseudonymized Bundle chunks into a single Bundle maintaining original Bundle structure (type, identifier, timestamp)
- **FR-005**: System MUST preserve Bundle.entry order during split-and-reassemble operations
- **FR-006**: System MUST calculate Bundle size based on serialized JSON byte count (the actual HTTP payload size)
- **FR-007**: System MUST use a default split threshold of 10MB when not specified in configuration
- **FR-008**: System MUST allow users to configure split threshold via project configuration file (bundle_split_threshold_mb parameter)
- **FR-009**: System MUST validate split threshold configuration at pipeline startup and reject invalid values (negative, zero, unreasonably large >100MB)
- **FR-010**: System MUST skip splitting for Bundles smaller than threshold (no performance penalty for normal-sized data)
- **FR-011**: System MUST skip splitting for non-Bundle FHIR resources (splitting only applies to resourceType: Bundle)
- **FR-012**: System MUST log detailed information when splitting occurs (original size, number of chunks, chunk sizes)
- **FR-013**: System MUST report progress during split Bundle processing (e.g., "Processing chunk 5/12")
- **FR-014**: System MUST detect when individual non-Bundle resources exceed 30MB and generate actionable error messages
- **FR-015**: System MUST continue processing remaining resources when individual oversized resources fail, collecting failures for end-of-job report
- **FR-016**: System MUST apply retry logic to each Bundle chunk independently (transient failures in one chunk don't affect others)
- **FR-017**: System MUST preserve Bundle.entry.fullUrl during splitting and reassembly
- **FR-018**: System MUST handle Bundles with mixed resource types (Patient, Observation, Condition, etc.) in entries
- **FR-019**: System MUST ensure each split chunk is a valid FHIR Bundle (proper Bundle structure with meta, type, entry)
- **FR-020**: System MAY split Bundles into unlimited chunks as needed (no hard limit on chunk count, since flooding with chunks is unlikely in practice)

### Key Entities

- **FHIR Bundle**: A container resource holding multiple FHIR resources in its entry array. Bundles can be arbitrarily large (megabytes to gigabytes) depending on entry count. Split candidates when total serialized size exceeds threshold.
- **Bundle Entry**: Individual FHIR resource within a Bundle's entry array. Each entry contains resource, fullUrl, and optional request/response metadata. Preserved intact during splitting (entries are not subdivided).
- **Bundle Chunk**: Temporary Bundle created during splitting containing a subset of original entries. Maintains original Bundle metadata (type, identifier) but with fewer entries. Valid FHIR Bundle sent to DIMP independently.
- **Split Threshold**: Configurable size limit (in MB) triggering Bundle splitting. Default 10MB. Represents serialized JSON byte count matching HTTP payload size.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Pipeline processes FHIR Bundles up to 100MB without HTTP 413 errors
- **SC-002**: Processing time for Bundles <10MB remains unchanged (no performance regression for normal-sized data)
- **SC-003**: Split Bundles are reassembled with 100% data integrity (no lost or duplicated entries)
- **SC-004**: System provides clear progress indication when processing split Bundles (user sees chunk progress within 2 seconds of chunk completion)
- **SC-005**: System successfully processes test dataset with Patient + 100k Conditions (current blocker scenario) within 15 minutes
- **SC-006**: Bundle entry order is preserved in 100% of split-and-reassemble operations (verifiable by comparing original and final Bundle.entry arrays)
- **SC-007**: Configuration validation detects and rejects 100% of invalid threshold values before processing begins
- **SC-008**: Individual oversized resource errors are reported with sufficient detail for users to identify the problem resource within 30 seconds of reading error message

## Assumptions

- FHIR Bundle structure follows standard FHIR R4 specification (Bundle.entry is an array)
- DIMP service can process split Bundles independently without cross-Bundle context
- Pseudonymization is deterministic (same input always produces same output, enabling chunk reassembly)
- Bundle entries do not have complex inter-dependencies requiring atomic processing
- Serialized JSON size is a reasonable proxy for HTTP payload size (no significant compression)
- Default 10MB threshold provides ~3x safety margin below typical 30MB server limits
- Memory constraints allow loading at least one chunk into memory (chunks are typically <10MB)
- Bundle splitting is only required for DIMP step (other pipeline steps handle large files differently)

## Dependencies

- Existing DIMP client integration (internal/services/dimp_client.go)
- Existing job state management (internal/pipeline/job.go)
- Existing configuration system (internal/models/config.go)
- FHIR resource parsing and serialization capabilities (encoding/json)

## Out of Scope

- Splitting non-Bundle resources (not feasible without breaking FHIR semantics)
- Streaming processing of extremely large Bundles (>1GB) that don't fit in memory
- Optimizing DIMP service itself to accept larger payloads (this is a client-side solution)
- Splitting Bundles for steps other than DIMP (validation, CSV conversion handled separately)
- Transaction/batch Bundle semantic preservation (assumes document/collection Bundle types)
- Bundle entry cross-references resolution (assumes entries are independent)
