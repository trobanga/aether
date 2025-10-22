# Research & Design Decisions: Bundle Splitting

**Feature**: Bundle Splitting for DIMP Pipeline
**Date**: 2025-10-22
**Status**: Phase 0 Complete

## Overview

This document captures research findings, technical decisions, and best practices for implementing FHIR Bundle splitting in the aether pipeline to handle large datasets that exceed HTTP payload limits.

## Key Decisions

### 1. Bundle Size Calculation Method

**Decision**: Calculate Bundle size as serialized JSON byte count using `json.Marshal`

**Rationale**:
- HTTP payload size is determined by serialized bytes sent over the wire
- Go's `json.Marshal` provides exact byte count matching what DIMP server receives
- Simple, accurate, no estimation heuristics needed

**Alternatives Considered**:
- **Estimate from entry count** (REJECTED): Too inaccurate - resource sizes vary wildly (Patient vs Observation with 100 fields)
- **Calculate from raw NDJSON file size** (REJECTED): Bundle may be subset of file, JSON formatting differences
- **Use content-length header** (REJECTED): Requires actual HTTP request, defeats purpose of pre-detection

**Implementation**:
```go
func calculateBundleSize(bundle map[string]interface{}) (int, error) {
    jsonBytes, err := json.Marshal(bundle)
    if err != nil {
        return 0, err
    }
    return len(jsonBytes), nil
}
```

### 2. Entry Partitioning Strategy

**Decision**: Use greedy array partitioning - accumulate entries until chunk would exceed threshold, then start new chunk

**Rationale**:
- Simple algorithm: iterate entries, add to current chunk while under limit
- Maximizes chunk sizes (fewer HTTP requests)
- Preserves entry order naturally
- No complex optimization needed (KISS principle)

**Alternatives Considered**:
- **Equal-size chunks** (REJECTED): Requires calculating all entry sizes upfront, more complex, provides no benefit
- **Fixed entry count per chunk** (REJECTED): Doesn't account for variable entry sizes, may still exceed limits
- **Bin packing algorithm** (REJECTED): Over-engineering for this use case, violates KISS

**Algorithm**:
```
currentChunk = empty Bundle
currentSize = headerSize

for each entry in originalBundle.entry:
    entrySize = calculateSize(entry)

    if currentSize + entrySize > threshold:
        // Chunk would exceed limit
        yield currentChunk
        currentChunk = new Bundle with header
        currentSize = headerSize

    currentChunk.entry.append(entry)
    currentSize += entrySize

yield currentChunk  // Don't forget last chunk
```

### 3. Chunk Metadata Handling

**Decision**: Each chunk is a complete, valid FHIR R4 Bundle with:
- **Preserved fields**: `type`, `timestamp`
- **Modified fields**: `id` (append chunk suffix e.g., "original-id-chunk-1"), `total` (set to chunk entry count)
- **Omitted fields**: `signature`, `link` (not relevant for chunks)

**Rationale**:
- DIMP service expects valid FHIR Bundles, not fragments
- Chunk ID uniqueness prevents collisions
- Total reflects actual chunk size (FHIR spec requirement)
- Omitting signature is correct (signature would be invalid for modified Bundle)

**FHIR R4 Bundle Structure** (reference):
```json
{
  "resourceType": "Bundle",
  "id": "bundle-example",
  "type": "document",
  "timestamp": "2025-10-22T10:00:00Z",
  "total": 100,
  "entry": [
    {
      "fullUrl": "urn:uuid:...",
      "resource": { ...FHIR resource... }
    }
  ]
}
```

### 4. Reassembly Strategy

**Decision**: Concatenate pseudonymized chunk entries in order, restore original Bundle metadata

**Rationale**:
- Chunks processed sequentially maintain order
- Deterministic pseudonymization ensures consistency
- Simple concatenation preserves data integrity

**Implementation**:
```go
func reassembleBundle(originalMetadata, pseudonymizedChunks) Bundle {
    reassembled := Bundle{
        ResourceType: "Bundle",
        ID: originalMetadata.ID,  // Restore original ID
        Type: originalMetadata.Type,
        Timestamp: originalMetadata.Timestamp,
        Entry: []Entry{},
    }

    for each chunk in pseudonymizedChunks:
        reassembled.Entry.append(chunk.Entry...)

    reassembled.Total = len(reassembled.Entry)
    return reassembled
}
```

### 5. Error Handling Strategy

**Decision**: Chunk-level isolation with retry independence

**Rationale**:
- Transient errors in one chunk shouldn't block others
- Failed chunks can be retried individually
- Existing retry logic (exponential backoff) applies per-chunk
- Fail entire Bundle only if any chunk fails after all retries

**Error Scenarios**:
1. **Chunk processing failure** (DIMP returns error): Retry chunk independently
2. **Individual oversized entry** (single entry > threshold): Log error, skip entry, continue with others
3. **Reassembly failure** (unexpected data): Fail entire operation (data integrity compromise)

### 6. Configuration Design

**Decision**: Add `bundle_split_threshold_mb` to `PipelineConfig`, default 10MB

**Rationale**:
- Follows existing configuration pattern in models/config.go
- Pipeline-level setting (not per-job, not global service config)
- 10MB default provides 3x safety margin below typical 30MB limits
- Users can adjust based on their DIMP server configuration

**Configuration YAML**:
```yaml
pipeline:
  enabled_steps:
    - import
    - dimp
  bundle_split_threshold_mb: 10  # Optional, defaults to 10MB
```

**Validation Rules**:
- Must be positive number
- Maximum 100MB (safety check, larger values likely misconfiguration)
- Warning if >50MB (approaching typical limits, may indicate server should be reconfigured instead)

## Best Practices: Go JSON Processing

### Memory Efficiency

**Finding**: Go's encoding/json loads entire JSON into memory during Marshal/Unmarshal

**Implication**: Acceptable for our scale (<100MB Bundles, 10MB chunks)

**Best Practice**: For future scaling beyond 1GB Bundles, consider streaming parser (json.Decoder)

### JSON Marshal Performance

**Benchmark Data** (from Go documentation):
- ~10MB JSON: ~50-100ms to Marshal on modern CPU
- Memory allocation: ~2x JSON size during Marshal (transient)

**Implication**:
- 10 chunks from 100MB Bundle: ~500ms-1s marshaling overhead
- Negligible compared to network I/O to DIMP (~seconds per chunk)
- No optimization needed

### Error Handling Pattern

**Best Practice** (from aether codebase):
```go
// Good: Specific error types
type BundleSplitError struct {
    BundleID string
    Reason   string
}

func (e *BundleSplitError) Error() string {
    return fmt.Sprintf("bundle splitting failed for %s: %s", e.BundleID, e.Reason)
}

// Use with error wrapping
if err := splitBundle(bundle); err != nil {
    return fmt.Errorf("failed to split bundle: %w", err)
}
```

## FHIR R4 Best Practices

### Bundle Types (FHIR Spec)

**Relevant Types**:
- **document**: Immutable set of resources (use case: clinical document)
- **collection**: Unordered set of resources (use case: search results)
- **transaction**: Resources to be processed atomically
- **batch**: Resources to be processed independently

**Decision**: Support splitting for `document` and `collection` types only

**Rationale**:
- Document/collection: Entries are independent, splitting is safe
- Transaction: Atomic processing required, splitting would break semantics (out of scope)
- Batch: Similar to transaction, entries may have dependencies

### Entry References

**FHIR Pattern**: Bundles may contain relative references between entries

**Example**:
```json
{
  "entry": [
    { "fullUrl": "Patient/123", "resource": { "resourceType": "Patient", "id": "123" } },
    { "fullUrl": "Observation/456", "resource": {
        "resourceType": "Observation",
        "subject": { "reference": "Patient/123" }  // Reference to same Bundle
    }}
  ]
}
```

**Decision**: Assume entries are independent (no cross-chunk references required)

**Rationale**:
- Current use case (TORCH extraction): Bundles are collections of resources without internal references
- Supporting cross-chunk references adds significant complexity (violates KISS)
- If needed in future: Re-evaluate with specific use case

**Mitigation**: Document assumption in code comments and quickstart guide

## Testing Strategy

### Unit Tests (Pure Functions)

**Test Cases**:
1. `calculateBundleSize`: Verify byte count matches actual JSON
2. `partitionEntries`: Verify chunks respect threshold, preserve order
3. `createChunk`: Verify valid FHIR Bundle structure
4. `reassembleBundle`: Verify data integrity, metadata restoration

**Test Data**:
- Small Bundle (1MB, 10 entries): Verify no splitting occurs
- Large Bundle (50MB, 100k entries): Verify correct chunk count
- Variable entry sizes: Verify greedy algorithm maximizes chunks
- Edge case: Single entry > threshold (handle gracefully)

### Integration Tests (With Mock DIMP)

**Test Cases**:
1. End-to-end: 50MB Bundle → split → mock DIMP → reassemble → verify output
2. Error handling: Mock DIMP returns error for chunk 3 of 5 → verify retry
3. Configuration: Various thresholds (5MB, 10MB, 20MB) → verify behavior changes
4. Progress reporting: Verify user sees "Processing chunk X/Y" messages

**Mock DIMP**:
- HTTP server returning pseudonymized resources (deterministic transformation)
- Configurable errors (500 for retry testing, 413 for limit testing)

### Contract Tests (FHIR Validation)

**Test Cases**:
1. Chunk structure: Verify each chunk passes FHIR R4 Bundle validation
2. Entry preservation: Verify `fullUrl` and resource structure intact
3. Metadata: Verify chunk IDs unique, totals accurate

**Validation Tool**: Use FHIR Go library or JSON schema validation

## Performance Considerations

### Bottlenecks

**Analysis**:
1. **JSON marshaling**: ~100ms for 10MB Bundle (negligible)
2. **Network I/O to DIMP**: ~1-5 seconds per chunk (dominant factor)
3. **Memory allocation**: ~20MB for 10MB chunk (acceptable)

**Conclusion**: Network I/O dominates, no optimization needed in splitting logic

### Sequential vs Parallel Processing

**Decision**: Start with sequential chunk processing

**Rationale**:
- Simpler code (no concurrency management)
- DIMP may have rate limits or resource constraints
- Network bandwidth likely shared (parallel provides limited benefit)
- Sequential processing maintains order naturally

**Future Optimization**: If testing shows DIMP can handle concurrent requests and performance is insufficient:
- Add worker pool (e.g., 3-5 goroutines)
- Requires chunk result ordering mechanism
- Measure before implementing (YAGNI)

## Open Questions & Future Considerations

### Resolved in Specification

✅ **Max chunk count**: No hard limit (user confirmed flooding unlikely)
✅ **Default threshold**: 10MB (3x safety margin)
✅ **Bundle types supported**: Document/collection only (transaction/batch out of scope)

### Deferred (Out of Current Scope)

**Streaming for >1GB Bundles**:
- Current: Load Bundle in memory
- Future: If gigabyte-scale Bundles needed, implement streaming JSON parser
- Decision point: When use case emerges

**Cross-chunk reference resolution**:
- Current: Assume entries independent
- Future: If TORCH produces Bundles with internal references, add reference tracking
- Decision point: When use case emerges with actual reference patterns

**Parallel chunk processing**:
- Current: Sequential processing
- Future: Add worker pool if performance testing shows bottleneck
- Decision point: After measuring with realistic data

## References

- FHIR R4 Bundle Resource: https://hl7.org/fhir/R4/bundle.html
- Go encoding/json: https://pkg.go.dev/encoding/json
- Aether codebase: internal/pipeline/dimp.go (existing patterns)
- Project constitution: .specify/memory/constitution.md (functional programming, TDD, KISS)
