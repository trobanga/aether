# aether Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-08

## Active Technologies
- Go 1.21+ + Cobra (CLI framework), net/http (HTTP client), encoding/json (FHIR NDJSON parsing), gopkg.in/yaml.v3 (config), testify (testing assertions) (001-dup-pipeline-we)
- Go 1.21+ + Cobra (CLI framework), net/http (HTTP client), encoding/json (FHIR NDJSON parsing), gopkg.in/yaml.v3 (config), testify (testing assertions), progress bar library (schollz/progressbar or cheggaaa/pb for FR-029 requirements) (001-dup-pipeline-we)
- Filesystem (job state as JSON, FHIR NDJSON files organized by job ID) (001-dup-pipeline-we)
- Go 1.25.1 (minimum Go 1.21 as per project requirements)
- Filesystem (job state as JSON, FHIR NDJSON files by job ID - existing) (002-import-via-torch)
- Go 1.25.1 (minimum Go 1.21) (004-bundle-splitting)

## Project Structure
```
src/
tests/
```

## Commands
# Add commands for Go 1.21+

## Code Style
Go 1.21+: Follow standard conventions

## Recent Changes
- 004-bundle-splitting: Bundle splitting foundation implemented (Phase 1-2 complete)
  * Core data models: BundleMetadata, BundleChunk, SplitResult, ReassembledBundle, SplitStats
  * Pure functions: ShouldSplit, PartitionEntries, SplitBundle, ReassembleBundle
  * Greedy partitioning algorithm for entry distribution
  * Immutable data structures following functional programming
  * Comprehensive unit tests with TDD approach
  * FHIR R4 Bundle compliance validation
  * Configuration threshold validation (1-100MB)
  * Oversized resource detection with guidance messages
- 003-implement-ci-pipeline: Added Go 1.25.1 (minimum Go 1.21 as per project requirements)
- 002-import-via-torch: Added Go 1.25.1

<!-- MANUAL ADDITIONS START -->
## Bundle Splitting Feature (004-bundle-splitting)

### Current Implementation Status
- ✅ **Phase 1-2 Complete**: Foundation and foundational data structures
- ✅ **Phase 3 Complete**: User Story 1 (Automatic Large Bundle Processing) - Tests & Implementation
- ⏳ **Phase 3 In Progress**: Integration with DIMP pipeline step
- 📋 **Phase 4-6 Pending**: User Story 2-3 and Polish tasks

### Key Files
- `internal/models/bundle.go`: Data structures for Bundle splitting
- `internal/services/bundle_splitter.go`: Pure service functions
- `internal/lib/validation.go`: Configuration and resource validation
- `tests/unit/bundle_splitter_test.go`: Core unit tests
- `tests/unit/test_helpers.go`: Test fixtures for FHIR Bundles

### Architecture
- **Design Pattern**: Functional programming with pure functions
- **Algorithm**: Greedy entry partitioning (KISS principle)
- **Scale**: Handles 100MB+ Bundles, 100k+ entries
- **Threshold**: Configurable 1-100MB (default 10MB)
- **Integrity**: 100% data preservation during split-reassemble

### Next Steps
1. Integrate splitting logic into DIMP pipeline step (T020-T021)
2. Write integration tests with mock DIMP service (T013, T023)
3. Implement User Story 2: Oversized resource handling
4. Implement User Story 3: Configuration flexibility
5. Polish and final validation

<!-- MANUAL ADDITIONS END -->
