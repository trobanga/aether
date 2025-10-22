# Data Model: Bundle Splitting

**Feature**: Bundle Splitting for DIMP Pipeline
**Date**: 2025-10-22
**Language**: Go

## Overview

This document defines the data structures required for FHIR Bundle splitting functionality. All structures follow Go conventions and align with the functional programming principles in the project constitution.

## Core Entities

### 1. BundleMetadata

**Purpose**: Captures original Bundle metadata required for reassembly

**Location**: `internal/models/bundle.go`

```go
// BundleMetadata captures essential metadata from original Bundle for reassembly
// Immutable once created - used to restore Bundle structure after pseudonymization
type BundleMetadata struct {
    ID        string    // Original Bundle.id
    Type      string    // Bundle.type (document, collection, etc.)
    Timestamp time.Time // Bundle.timestamp (if present)
}
```

**Validation Rules**:
- `ID`: Must be non-empty string
- `Type`: Must be valid FHIR Bundle type ("document", "collection", "transaction", "batch", etc.)
- `Timestamp`: Optional (zero value allowed)

**Usage**:
- Extracted before splitting
- Passed through splitting process (immutable)
- Used to reconstruct final Bundle after reassembly

---

### 2. BundleChunk

**Purpose**: Represents a single chunk of a split Bundle with its metadata

**Location**: `internal/models/bundle.go`

```go
// BundleChunk represents one chunk of a split FHIR Bundle
// Contains subset of original entries plus metadata for tracking
type BundleChunk struct {
    ChunkID      string                   // Unique identifier: "{originalID}-chunk-{index}"
    Index        int                      // 0-based chunk index
    TotalChunks  int                      // Total number of chunks in split operation
    OriginalID   string                   // Original Bundle.id (for reassembly)
    Metadata     BundleMetadata           // Original Bundle metadata
    Entries      []map[string]interface{} // Bundle entries (JSON objects)
    EstimatedSize int                     // Estimated serialized size in bytes
}
```

**Validation Rules**:
- `ChunkID`: Must match pattern "{originalID}-chunk-{index}"
- `Index`: Must be >= 0 and < TotalChunks
- `TotalChunks`: Must be > 0
- `OriginalID`: Must be non-empty
- `Entries`: Must be non-empty (at least one entry per chunk)
- `EstimatedSize`: Must be > 0

**Relationships**:
- Contains `BundleMetadata` (composition)
- Multiple `BundleChunk` instances belong to one original Bundle
- Chunks are ordered by `Index`

---

### 3. SplitResult

**Purpose**: Encapsulates result of Bundle splitting operation (pure function output)

**Location**: `internal/models/bundle.go`

```go
// SplitResult contains the outcome of a Bundle splitting operation
// Immutable result structure following functional programming principles
type SplitResult struct {
    Metadata     BundleMetadata  // Original Bundle metadata
    Chunks       []BundleChunk   // Ordered list of Bundle chunks
    WasSplit     bool            // Whether splitting was necessary
    OriginalSize int             // Original Bundle size in bytes
    TotalChunks  int             // Number of chunks created (convenience field)
}
```

**State Transitions**: None (immutable)

**Usage**:
- Returned by `SplitBundle()` pure function
- Passed to chunk processing logic
- Used to track splitting statistics

**Invariants**:
- If `WasSplit == false`, then `TotalChunks == 1` and `Chunks[0]` contains all entries
- If `WasSplit == true`, then `TotalChunks > 1`
- `len(Chunks) == TotalChunks` always

---

### 4. ReassembledBundle

**Purpose**: Result of reassembling pseudonymized chunks

**Location**: `internal/models/bundle.go`

```go
// ReassembledBundle represents the final Bundle after pseudonymization and reassembly
// Contains all pseudonymized entries in original order with restored metadata
type ReassembledBundle struct {
    Bundle       map[string]interface{} // Complete FHIR Bundle (JSON object)
    EntryCount   int                    // Total entries in reassembled Bundle
    OriginalID   string                 // Original Bundle.id
    WasReassembled bool                 // Whether Bundle was reassembled from chunks
}
```

**Validation Rules**:
- `Bundle`: Must be valid FHIR R4 Bundle structure
- `Bundle["entry"]`: Must be array with length == EntryCount
- `Bundle["id"]`: Must match OriginalID
- `EntryCount`: Must be > 0

---

### 5. SplitConfig

**Purpose**: Configuration parameters for Bundle splitting

**Location**: `internal/models/config.go` (extends existing PipelineConfig)

```go
// Extend existing PipelineConfig struct
type PipelineConfig struct {
    EnabledSteps           []StepName `yaml:"enabled_steps" json:"enabled_steps"`
    BundleSplitThresholdMB int        `yaml:"bundle_split_threshold_mb" json:"bundle_split_threshold_mb"` // NEW
}
```

**Default Values**:
- `BundleSplitThresholdMB`: 10 (10 megabytes)

**Validation Rules** (in `Validate()` method):
- Must be > 0
- Must be <= 100 (sanity check - values >100MB likely misconfiguration)
- Warning if > 50 (approaching typical server limits)

**Configuration Example**:
```yaml
pipeline:
  enabled_steps:
    - import
    - dimp
  bundle_split_threshold_mb: 10  # Optional, defaults to 10
```

---

## Helper Structures

### 6. SplitStats

**Purpose**: Statistics for monitoring and logging

**Location**: `internal/models/bundle.go`

```go
// SplitStats captures metrics about Bundle splitting operation
// Used for logging and monitoring purposes
type SplitStats struct {
    BundleID         string
    OriginalSize     int
    OriginalEntries  int
    ChunksCreated    int
    AverageChunkSize int
    LargestChunkSize int
    SmallestChunkSize int
    SplitDuration    time.Duration
}
```

**Usage**:
- Collected during splitting operation
- Logged at INFO level
- Used for performance monitoring

---

### 7. OversizedResourceError

**Purpose**: Error type for individual resources exceeding limits

**Location**: `internal/services/bundle_splitter.go`

```go
// OversizedResourceError indicates a single resource exceeds threshold
// Cannot be split without violating FHIR semantics
type OversizedResourceError struct {
    ResourceType  string
    ResourceID    string
    Size          int
    Threshold     int
    Guidance      string // User-facing guidance message
}

func (e *OversizedResourceError) Error() string {
    return fmt.Sprintf(
        "resource %s/%s (%d bytes) exceeds threshold (%d bytes). %s",
        e.ResourceType, e.ResourceID, e.Size, e.Threshold, e.Guidance,
    )
}
```

**Error Handling**: Non-retryable error, should be logged and resource skipped

---

## Data Flow

### Splitting Flow

```
Original Bundle (map[string]interface{})
    ↓
[calculateSize] → OriginalSize (int)
    ↓
[shouldSplit?] → Decision (bool)
    ↓
[splitBundle] → SplitResult
    ↓
    ├─ BundleMetadata (extracted)
    └─ []BundleChunk (created)
```

### Processing Flow

```
SplitResult
    ↓
for each BundleChunk:
    ↓
    [createFHIRBundle] → map[string]interface{}
    ↓
    [sendToDIMP] → Pseudonymized Bundle
    ↓
    [extractEntries] → []map[string]interface{}
    ↓
    [collect] → [][]map[string]interface{}
    ↓
[reassembleBundle] → ReassembledBundle
```

### State Transitions

**BundleChunk** states (implicit, tracked via processing):
1. **Created**: Chunk exists, not yet sent to DIMP
2. **Processing**: HTTP request to DIMP in progress
3. **Completed**: Pseudonymized entries received
4. **Failed**: DIMP returned error (retryable or permanent)

**No explicit state field** - state tracked implicitly via call stack and error handling

---

## Relationships

```
Original Bundle (1)
    ├─ has → BundleMetadata (1)
    └─ splits into → BundleChunk (*) [0..N]
        └─ contains → BundleMetadata (1) [same as original]

BundleChunk (*) [processed]
    └─ reassembles into → ReassembledBundle (1)
        └─ restores → BundleMetadata (1) [from original]
```

---

## Immutability Guarantee

**Following Constitution Principle I (Functional Programming)**:

All entities are **value types** (structs, not pointers in function signatures where possible):

```go
// GOOD: Pure function with value types
func SplitBundle(bundle map[string]interface{}, thresholdBytes int) (SplitResult, error)

// GOOD: Pure function returning new data
func ReassembleBundle(metadata BundleMetadata, pseudonymizedChunks [][]map[string]interface{}) (ReassembledBundle, error)

// AVOID: Mutation
func (b *BundleChunk) AddEntry(entry map[string]interface{}) // Don't do this
```

**Note**: `map[string]interface{}` for FHIR JSON is mutable by nature (Go limitation), but functions create new maps rather than mutating inputs.

---

## Validation Functions

**Location**: `internal/lib/validation.go`

```go
// ValidateBundleMetadata checks BundleMetadata structure
func ValidateBundleMetadata(m BundleMetadata) error

// ValidateBundleChunk checks BundleChunk structure and constraints
func ValidateBundleChunk(chunk BundleChunk) error

// ValidateSplitConfig checks configuration values
func ValidateSplitConfig(config SplitConfig) error

// IsFHIRBundle checks if JSON object is valid FHIR R4 Bundle
func IsFHIRBundle(obj map[string]interface{}) error
```

**Validation Strategy**:
- Fail fast on invalid data
- Clear error messages with field names
- Used at boundaries (config load, before splitting, after reassembly)

---

## Size Calculation

**Core Function**:

```go
// CalculateJSONSize returns serialized byte count of JSON object
// Used to determine if Bundle exceeds threshold
func CalculateJSONSize(obj map[string]interface{}) (int, error) {
    jsonBytes, err := json.Marshal(obj)
    if err != nil {
        return 0, fmt.Errorf("failed to marshal JSON: %w", err)
    }
    return len(jsonBytes), nil
}
```

**Usage**: Called for entire Bundle and individual entries to make splitting decisions

---

## Testing Data Structures

**Test Fixtures** (for unit tests):

```go
// CreateTestBundle generates FHIR Bundle for testing
func CreateTestBundle(entryCount int, entrySizeKB int) map[string]interface{}

// CreateLargeTestBundle generates 50MB Bundle with 100k entries
func CreateLargeTestBundle() map[string]interface{}

// MockPseudonymizeEntry simulates DIMP pseudonymization
func MockPseudonymizeEntry(entry map[string]interface{}) map[string]interface{}
```

**Location**: `tests/unit/test_helpers.go`

---

## Summary

**New Files**:
- `internal/models/bundle.go`: Core Bundle splitting data structures
- `tests/unit/test_helpers.go`: Test fixtures

**Modified Files**:
- `internal/models/config.go`: Add `BundleSplitThresholdMB` to `PipelineConfig`
- `internal/lib/validation.go`: Add Bundle validation functions

**Design Principles**:
- Immutable data structures (functional programming)
- Clear separation of data and behavior
- Explicit error types for clear error handling
- Value types preferred over pointers for pure functions
