# Test Coverage for Multi-Step Pipeline Execution

This document describes the comprehensive test suite created to verify that the pipeline correctly executes all enabled steps and that configuration is loaded properly.

## Test Files

### 1. `tests/integration/pipeline_multistep_test.go`
Integration tests for multi-step pipeline execution.

#### TestPipelineMultiStep_AutomaticExecution ⭐
**Critical regression test** - Verifies that `pipeline start` automatically executes ALL enabled steps, not just import.

- Creates a job with both `import` and `dimp` steps enabled
- Executes import step
- **Verifies that the next step is DIMP** (not empty)
- Advances to DIMP and executes it
- Verifies pseudonymized files are created
- **This test would have caught the original bug**

#### TestPipelineMultiStep_StepSequencing
Verifies that `GetNextStep()` returns steps in the correct order.

- Tests that import → dimp → validation → csv_conversion works correctly
- Tests that the last step returns empty string (no more steps)

#### TestPipelineMultiStep_OnlyImportEnabled
Verifies pipeline works correctly when only import is enabled.

- Creates job with only import step
- Verifies no next step exists after import
- Verifies job is marked as complete when no more steps

#### TestPipelineMultiStep_ConfigLoadingPreservesSteps
**Regression test for viper.Unmarshal bug** - Verifies config loading doesn't drop enabled steps.

- Writes a YAML config with multiple steps
- Loads config via `LoadConfig()`
- Verifies ALL steps are present (import, dimp, csv_conversion)
- Verifies service URLs are loaded correctly

#### TestPipelineMultiStep_JobStatePersistedBetweenSteps
Verifies job state is correctly saved after each step execution.

- Executes import step and saves state
- Loads job from disk and verifies import step is complete
- Executes DIMP step and saves state
- Loads job again and verifies both steps are complete

---

### 2. `tests/unit/config_loading_test.go`
Unit tests specifically for configuration loading (the root cause of the bug).

#### TestConfigLoading_MultipleEnabledSteps ⭐
**Critical regression test** - Verifies viper doesn't drop enabled steps when loading YAML.

- Tests loading 5 enabled steps: import, dimp, validation, csv_conversion, parquet_conversion
- Verifies ALL 5 steps are present in loaded config
- **This test would have caught the viper.Unmarshal bug**

#### TestConfigLoading_ServiceURLs
Verifies all service URLs are loaded correctly from YAML.

- Tests DIMP URL, CSV URL, and Parquet URL
- Verifies exact values match what's in the config file

#### TestConfigLoading_RetrySettings
Verifies retry configuration values are loaded.

- Tests max_attempts, initial_backoff_ms, max_backoff_ms
- Ensures retry config isn't using defaults when values are specified

#### TestConfigLoading_JobsDirectory
Verifies custom jobs_dir path is loaded.

- Tests that non-default jobs directory is preserved

#### TestConfigLoading_EmptyServiceURLs
Verifies empty service URLs remain empty (not replaced with defaults).

- Important for optional services

#### TestConfigLoading_PartialServiceURLs
Verifies mixed empty and non-empty service URLs.

- Tests DIMP URL present, CSV empty, Parquet present
- Ensures selective service configuration works

#### TestConfigLoading_NoConfigFile
Verifies error handling when specified config file doesn't exist.

- Tests that explicit file path that doesn't exist returns error

#### TestConfigLoading_StepOrder
Verifies steps are loaded in the exact order specified in YAML.

- Tests non-standard order: import → validation → dimp → parquet → csv
- Ensures order is preserved exactly as in config

#### TestConfigLoading_MinimalConfig
Verifies minimal valid configuration loads successfully.

- Tests config with only required fields
- Ensures defaults work for optional fields

---

## Running the Tests

```bash
# Run all multi-step integration tests
go test ./tests/integration/pipeline_multistep_test.go -v

# Run all config loading unit tests
go test ./tests/unit/config_loading_test.go -v

# Run all tests in the test suite
go test ./tests/... -v
```

## Test Summary

| Test File | Tests | Status |
|-----------|-------|--------|
| `pipeline_multistep_test.go` | 5 tests | ✅ All Pass |
| `config_loading_test.go` | 9 tests | ✅ All Pass |
| **Total** | **14 tests** | ✅ **All Pass** |

## Critical Tests (Regression Prevention)

These tests specifically target the bugs that were fixed:

1. **TestPipelineMultiStep_AutomaticExecution** - Ensures pipeline executes all enabled steps, not just import
2. **TestConfigLoading_MultipleEnabledSteps** - Ensures viper doesn't drop enabled steps during config loading
3. **TestPipelineMultiStep_ConfigLoadingPreservesSteps** - Integration test combining both fixes

## Test Coverage

The test suite covers:

- ✅ Automatic multi-step execution
- ✅ Step sequencing and order
- ✅ Config loading (all fields)
- ✅ Service URLs (empty, partial, full)
- ✅ Job state persistence between steps
- ✅ Edge cases (only import, no config file, minimal config)
- ✅ YAML parsing correctness
- ✅ Step advancement logic

## Continuous Integration

These tests should be run in CI/CD pipelines to prevent regression of the fixed bugs.

Recommended CI command:
```bash
go test ./tests/unit/config_loading_test.go ./tests/integration/pipeline_multistep_test.go
```
