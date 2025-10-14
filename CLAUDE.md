# aether Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-08

## Active Technologies
- Go 1.21+ + Cobra (CLI framework), net/http (HTTP client), encoding/json (FHIR NDJSON parsing), gopkg.in/yaml.v3 (config), testify (testing assertions) (001-dup-pipeline-we)
- Go 1.21+ + Cobra (CLI framework), net/http (HTTP client), encoding/json (FHIR NDJSON parsing), gopkg.in/yaml.v3 (config), testify (testing assertions), progress bar library (schollz/progressbar or cheggaaa/pb for FR-029 requirements) (001-dup-pipeline-we)
- Filesystem (job state as JSON, FHIR NDJSON files organized by job ID) (001-dup-pipeline-we)
- Go 1.25.1 (minimum Go 1.21 as per project requirements)
- Filesystem (job state as JSON, FHIR NDJSON files by job ID - existing) (002-import-via-torch)

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
- 003-implement-ci-pipeline: Added Go 1.25.1 (minimum Go 1.21 as per project requirements)
- 002-import-via-torch: Added Go 1.25.1
- 001-dup-pipeline-we: Added Go 1.21+ + Cobra (CLI framework), net/http (HTTP client), encoding/json (FHIR NDJSON parsing), gopkg.in/yaml.v3 (config), testify (testing assertions), progress bar library (schollz/progressbar or cheggaaa/pb for FR-029 requirements)
- 001-dup-pipeline-we: Added Go 1.21+ + Cobra (CLI framework), net/http (HTTP client), encoding/json (FHIR NDJSON parsing), gopkg.in/yaml.v3 (config), testify (testing assertions)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
