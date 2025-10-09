package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trobanga/aether/internal/models"
	"github.com/trobanga/aether/internal/services"
)

// TestConfigLoading_MultipleEnabledSteps verifies that all enabled steps
// are loaded from YAML config file (regression test for viper.Unmarshal bug)
func TestConfigLoading_MultipleEnabledSteps(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	jobsDir := filepath.Join(tmpDir, "jobs")
	os.MkdirAll(jobsDir, 0755)

	configContent := `
services:
  dimp_url: "http://localhost:32861/fhir"
  csv_conversion_url: "http://localhost:9000/csv"
  parquet_conversion_url: "http://localhost:9000/parquet"

pipeline:
  enabled_steps:
    - import
    - dimp
    - validation
    - csv_conversion
    - parquet_conversion

retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000

jobs_dir: "` + jobsDir + `"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := services.LoadConfig(configFile)
	require.NoError(t, err, "Config should load without error")

	// Verify ALL steps are loaded
	assert.Len(t, config.Pipeline.EnabledSteps, 5, "Should have 5 enabled steps")
	assert.Equal(t, models.StepImport, config.Pipeline.EnabledSteps[0])
	assert.Equal(t, models.StepDIMP, config.Pipeline.EnabledSteps[1])
	assert.Equal(t, models.StepValidation, config.Pipeline.EnabledSteps[2])
	assert.Equal(t, models.StepCSVConversion, config.Pipeline.EnabledSteps[3])
	assert.Equal(t, models.StepParquetConversion, config.Pipeline.EnabledSteps[4])
}

// TestConfigLoading_ServiceURLs verifies all service URLs are loaded correctly
func TestConfigLoading_ServiceURLs(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	jobsDir := filepath.Join(tmpDir, "jobs")
	os.MkdirAll(jobsDir, 0755)

	expectedDIMPUrl := "http://dimp.example.com:8080/fhir"
	expectedCSVUrl := "http://csv.example.com:9000/convert"
	expectedParquetUrl := "http://parquet.example.com:9001/convert"

	configContent := `
services:
  dimp_url: "` + expectedDIMPUrl + `"
  csv_conversion_url: "` + expectedCSVUrl + `"
  parquet_conversion_url: "` + expectedParquetUrl + `"

pipeline:
  enabled_steps:
    - import

retry:
  max_attempts: 3
  initial_backoff_ms: 500
  max_backoff_ms: 10000

jobs_dir: "` + jobsDir + `"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := services.LoadConfig(configFile)
	require.NoError(t, err)

	// Verify all service URLs are loaded exactly as specified
	assert.Equal(t, expectedDIMPUrl, config.Services.DIMPUrl, "DIMP URL should be loaded correctly")
	assert.Equal(t, expectedCSVUrl, config.Services.CSVConversionUrl, "CSV URL should be loaded correctly")
	assert.Equal(t, expectedParquetUrl, config.Services.ParquetConversionUrl, "Parquet URL should be loaded correctly")
}

// TestConfigLoading_RetrySettings verifies retry configuration is loaded
func TestConfigLoading_RetrySettings(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	jobsDir := filepath.Join(tmpDir, "jobs")
	os.MkdirAll(jobsDir, 0755)

	configContent := `
services:
  dimp_url: ""

pipeline:
  enabled_steps:
    - import

retry:
  max_attempts: 7
  initial_backoff_ms: 2000
  max_backoff_ms: 60000

jobs_dir: "` + jobsDir + `"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := services.LoadConfig(configFile)
	require.NoError(t, err)

	assert.Equal(t, 7, config.Retry.MaxAttempts, "MaxAttempts should be loaded")
	assert.Equal(t, int64(2000), config.Retry.InitialBackoffMs, "InitialBackoffMs should be loaded")
	assert.Equal(t, int64(60000), config.Retry.MaxBackoffMs, "MaxBackoffMs should be loaded")
}

// TestConfigLoading_JobsDirectory verifies jobs_dir is loaded
func TestConfigLoading_JobsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	customJobsDir := filepath.Join(tmpDir, "custom_jobs_location")
	os.MkdirAll(customJobsDir, 0755)

	configContent := `
services:
  dimp_url: ""

pipeline:
  enabled_steps:
    - import

retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000

jobs_dir: "` + customJobsDir + `"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := services.LoadConfig(configFile)
	require.NoError(t, err)

	assert.Equal(t, customJobsDir, config.JobsDir, "Custom jobs directory should be loaded")
}

// TestConfigLoading_EmptyServiceURLs verifies empty service URLs are preserved
func TestConfigLoading_EmptyServiceURLs(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	jobsDir := filepath.Join(tmpDir, "jobs")
	os.MkdirAll(jobsDir, 0755)

	configContent := `
services:
  dimp_url: ""
  csv_conversion_url: ""
  parquet_conversion_url: ""

pipeline:
  enabled_steps:
    - import

retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000

jobs_dir: "` + jobsDir + `"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := services.LoadConfig(configFile)
	require.NoError(t, err)

	assert.Empty(t, config.Services.DIMPUrl, "Empty DIMP URL should remain empty")
	assert.Empty(t, config.Services.CSVConversionUrl, "Empty CSV URL should remain empty")
	assert.Empty(t, config.Services.ParquetConversionUrl, "Empty Parquet URL should remain empty")
}

// TestConfigLoading_PartialServiceURLs verifies mixed empty and non-empty URLs
func TestConfigLoading_PartialServiceURLs(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	jobsDir := filepath.Join(tmpDir, "jobs")
	os.MkdirAll(jobsDir, 0755)

	configContent := `
services:
  dimp_url: "http://localhost:32861/fhir"
  csv_conversion_url: ""
  parquet_conversion_url: "http://localhost:9001/parquet"

pipeline:
  enabled_steps:
    - import
    - dimp

retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000

jobs_dir: "` + jobsDir + `"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := services.LoadConfig(configFile)
	require.NoError(t, err)

	assert.Equal(t, "http://localhost:32861/fhir", config.Services.DIMPUrl, "DIMP URL should be loaded")
	assert.Empty(t, config.Services.CSVConversionUrl, "Empty CSV URL should remain empty")
	assert.Equal(t, "http://localhost:9001/parquet", config.Services.ParquetConversionUrl, "Parquet URL should be loaded")
}

// TestConfigLoading_NoConfigFile verifies error when specific config file doesn't exist
func TestConfigLoading_NoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "does-not-exist.yaml")

	// Try to load non-existent config file - should error
	_, err := services.LoadConfig(nonExistentFile)
	require.Error(t, err, "Should error when specified config file doesn't exist")
	assert.Contains(t, err.Error(), "no such file or directory", "Error should indicate file not found")
}

// TestConfigLoading_StepOrder verifies steps are loaded in the order specified
func TestConfigLoading_StepOrder(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	jobsDir := filepath.Join(tmpDir, "jobs")
	os.MkdirAll(jobsDir, 0755)

	// Deliberately specify steps in a specific order
	// Note: Steps that require services must have URLs configured
	configContent := `
services:
  dimp_url: "http://localhost:32861/fhir"
  csv_conversion_url: "http://localhost:9000/csv"
  parquet_conversion_url: "http://localhost:9001/parquet"

pipeline:
  enabled_steps:
    - import
    - validation
    - dimp
    - parquet_conversion
    - csv_conversion

retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000

jobs_dir: "` + jobsDir + `"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := services.LoadConfig(configFile)
	require.NoError(t, err)

	// Verify steps are in the exact order specified in YAML
	require.Len(t, config.Pipeline.EnabledSteps, 5)
	assert.Equal(t, models.StepImport, config.Pipeline.EnabledSteps[0])
	assert.Equal(t, models.StepValidation, config.Pipeline.EnabledSteps[1])
	assert.Equal(t, models.StepDIMP, config.Pipeline.EnabledSteps[2])
	assert.Equal(t, models.StepParquetConversion, config.Pipeline.EnabledSteps[3])
	assert.Equal(t, models.StepCSVConversion, config.Pipeline.EnabledSteps[4])
}

// TestConfigLoading_MinimalConfig verifies minimal valid config loads
func TestConfigLoading_MinimalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	jobsDir := filepath.Join(tmpDir, "jobs")
	os.MkdirAll(jobsDir, 0755)

	// Absolute minimal config
	configContent := `
pipeline:
  enabled_steps:
    - import

retry:
  max_attempts: 3
  initial_backoff_ms: 500
  max_backoff_ms: 5000

jobs_dir: "` + jobsDir + `"
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := services.LoadConfig(configFile)
	require.NoError(t, err, "Minimal config should load successfully")

	assert.Len(t, config.Pipeline.EnabledSteps, 1)
	assert.Equal(t, models.StepImport, config.Pipeline.EnabledSteps[0])
	assert.Empty(t, config.Services.DIMPUrl)
	assert.Empty(t, config.Services.CSVConversionUrl)
	assert.Empty(t, config.Services.ParquetConversionUrl)
}
