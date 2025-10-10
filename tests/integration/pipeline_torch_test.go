package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trobanga/aether/internal/lib"
	"github.com/trobanga/aether/internal/models"
	"github.com/trobanga/aether/internal/pipeline"
	"github.com/trobanga/aether/internal/services"
)

// T025: Integration test for CRTDL → extraction → download flow

func TestPipeline_TORCHExtraction_EndToEnd(t *testing.T) {
	// Setup: Create test environment
	tempDir := t.TempDir()
	jobsDir := filepath.Join(tempDir, "jobs")
	os.MkdirAll(jobsDir, 0755)

	// Create CRTDL file
	crtdlPath := filepath.Join(tempDir, "test.crtdl")
	crtdlContent := map[string]interface{}{
		"cohortDefinition": map[string]interface{}{
			"version": "1.0.0",
			"display": "Test cohort",
			"inclusionCriteria": []map[string]interface{}{
				{
					"name": "age_criteria",
					"type": "age",
					"min":  18,
					"max":  65,
				},
			},
		},
		"dataExtraction": map[string]interface{}{
			"attributeGroups": []map[string]interface{}{
				{
					"name":        "demographics",
					"resourceType": "Patient",
					"attributes":   []string{"birthDate", "gender"},
				},
			},
		},
	}
	crtdlJSON, _ := json.Marshal(crtdlContent)
	os.WriteFile(crtdlPath, crtdlJSON, 0644)

	// Mock NDJSON content to be returned
	ndjsonContent := `{"resourceType":"Bundle","type":"transaction","entry":[{"resource":{"resourceType":"Patient","id":"test-patient-1","birthDate":"1990-01-01","gender":"male"}}]}
{"resourceType":"Bundle","type":"transaction","entry":[{"resource":{"resourceType":"Patient","id":"test-patient-2","birthDate":"1985-05-15","gender":"female"}}]}`

	// Mock TORCH server with full workflow
	extractionJobPath := "/fhir/extraction/job-xyz"
	pollCount := 0
	maxPollsBeforeComplete := 2

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle extraction submission
		if r.Method == "POST" && r.URL.Path == "/fhir/$extract-data" {
			// Verify FHIR Parameters format
			var params map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&params)
			require.NoError(t, err)
			assert.Equal(t, "Parameters", params["resourceType"])

			// Return 202 with Content-Location
			w.Header().Set("Content-Location", server.URL+extractionJobPath)
			w.WriteHeader(http.StatusAccepted)
			return
		}

		// Handle polling
		if r.Method == "GET" && r.URL.Path == extractionJobPath {
			pollCount++
			if pollCount < maxPollsBeforeComplete {
				// Still processing
				w.WriteHeader(http.StatusAccepted)
				return
			}

			// Extraction complete - return file URLs
			result := map[string]interface{}{
				"resourceType": "Parameters",
				"parameter": []map[string]interface{}{
					{
						"name": "output",
						"part": []map[string]interface{}{
							{
								"name":     "url",
								"valueUrl": server.URL + "/output/Patient.ndjson",
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(result)
			return
		}

		// Handle file download
		if r.Method == "GET" && r.URL.Path == "/output/Patient.ndjson" {
			w.Header().Set("Content-Type", "application/fhir+ndjson")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(ndjsonContent))
			return
		}

		// Handle ping/connectivity check
		if r.Method == "GET" && r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create configuration
	config := models.ProjectConfig{
		Services: models.ServiceConfig{
			TORCH: models.TORCHConfig{
				BaseURL:                   server.URL,
				Username:                  "testuser",
				Password:                  "testpass",
				ExtractionTimeoutMinutes:  1,
				PollingIntervalSeconds:    1,
				MaxPollingIntervalSeconds: 5,
			},
		},
		Pipeline: models.PipelineConfig{
			EnabledSteps: []models.StepName{models.StepImport},
		},
		Retry: models.RetryConfig{
			MaxAttempts:      3,
			InitialBackoffMs: 100,
			MaxBackoffMs:     1000,
		},
		JobsDir: jobsDir,
	}

	// Create job with CRTDL input
	logger := lib.NewLogger(lib.LogLevelDebug)
	job, err := pipeline.CreateJob(crtdlPath, config, logger)
	require.NoError(t, err)

	// Verify job was created with correct input type
	assert.Equal(t, models.InputTypeCRTDL, job.InputType)
	assert.Equal(t, crtdlPath, job.InputSource)
	assert.NotEmpty(t, job.JobID)

	// Execute import step (which should trigger TORCH extraction)
	httpClient := services.NewHTTPClient(2*time.Second, config.Retry, logger)
	updatedJob, err := pipeline.ExecuteImportStep(job, logger, httpClient, false)

	// Verify successful execution
	require.NoError(t, err)
	assert.NotNil(t, updatedJob)

	// Verify TORCH extraction URL was stored
	assert.NotEmpty(t, updatedJob.TORCHExtractionURL)
	assert.Contains(t, updatedJob.TORCHExtractionURL, "/fhir/extraction/")

	// Verify import step completed
	importStep, found := models.GetStepByName(*updatedJob, models.StepImport)
	require.True(t, found)
	assert.Equal(t, models.StepStatusCompleted, importStep.Status)

	// Verify files were downloaded
	assert.Greater(t, updatedJob.TotalFiles, 0)
	assert.Greater(t, updatedJob.TotalBytes, int64(0))

	// Verify NDJSON file exists in job directory
	importDir := services.GetJobOutputDir(jobsDir, job.JobID, models.StepImport)
	files, err := os.ReadDir(importDir)
	require.NoError(t, err)
	assert.NotEmpty(t, files, "Expected downloaded NDJSON files in import directory")

	// Verify file content
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".ndjson" {
			content, err := os.ReadFile(filepath.Join(importDir, file.Name()))
			require.NoError(t, err)
			assert.Contains(t, string(content), "Patient", "Downloaded file should contain FHIR Patient resources")
		}
	}

	// Verify polling happened multiple times (exponential backoff)
	assert.GreaterOrEqual(t, pollCount, maxPollsBeforeComplete, "Should have polled until completion")
}

func TestPipeline_TORCHExtraction_EmptyResult(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	jobsDir := filepath.Join(tempDir, "jobs")
	os.MkdirAll(jobsDir, 0755)

	// Create CRTDL file
	crtdlPath := filepath.Join(tempDir, "empty-cohort.crtdl")
	crtdlJSON := []byte(`{"cohortDefinition":{"version":"1.0.0","inclusionCriteria":[]},"dataExtraction":{"attributeGroups":[]}}`)
	os.WriteFile(crtdlPath, crtdlJSON, 0644)

	// Mock TORCH server returning empty result
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/fhir/$extract-data" {
			w.Header().Set("Content-Location", server.URL+"/fhir/extraction/empty-job")
			w.WriteHeader(http.StatusAccepted)
			return
		}

		if r.Method == "GET" && r.URL.Path == "/fhir/extraction/empty-job" {
			// Return result with no output files
			result := map[string]interface{}{
				"resourceType": "Parameters",
				"parameter":    []map[string]interface{}{},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(result)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := models.ProjectConfig{
		Services: models.ServiceConfig{
			TORCH: models.TORCHConfig{
				BaseURL:                   server.URL,
				Username:                  "testuser",
				Password:                  "testpass",
				ExtractionTimeoutMinutes:  1,
				PollingIntervalSeconds:    1,
				MaxPollingIntervalSeconds: 5,
			},
		},
		Pipeline: models.PipelineConfig{
			EnabledSteps: []models.StepName{models.StepImport},
		},
		Retry: models.RetryConfig{
			MaxAttempts:      3,
			InitialBackoffMs: 100,
			MaxBackoffMs:     1000,
		},
		JobsDir: jobsDir,
	}

	logger := lib.NewLogger(lib.LogLevelDebug)
	job, err := pipeline.CreateJob(crtdlPath, config, logger)
	require.NoError(t, err)

	httpClient := services.NewHTTPClient(2*time.Second, config.Retry, logger)
	updatedJob, err := pipeline.ExecuteImportStep(job, logger, httpClient, false)

	// Empty result should be handled gracefully
	require.NoError(t, err)
	assert.NotNil(t, updatedJob)

	// Import step should complete with zero files
	importStep, found := models.GetStepByName(*updatedJob, models.StepImport)
	require.True(t, found)
	assert.Equal(t, models.StepStatusCompleted, importStep.Status)
	assert.Equal(t, 0, updatedJob.TotalFiles)
}

func TestPipeline_TORCHExtraction_ServerUnavailable(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	jobsDir := filepath.Join(tempDir, "jobs")
	os.MkdirAll(jobsDir, 0755)

	// Create CRTDL file
	crtdlPath := filepath.Join(tempDir, "test.crtdl")
	crtdlJSON := []byte(`{"cohortDefinition":{"version":"1.0.0","inclusionCriteria":[]},"dataExtraction":{"attributeGroups":[]}}`)
	os.WriteFile(crtdlPath, crtdlJSON, 0644)

	config := models.ProjectConfig{
		Services: models.ServiceConfig{
			TORCH: models.TORCHConfig{
				BaseURL:  "http://unreachable-torch-server.invalid:9999",
				Username: "testuser",
				Password: "testpass",
			},
		},
		Pipeline: models.PipelineConfig{
			EnabledSteps: []models.StepName{models.StepImport},
		},
		Retry: models.RetryConfig{
			MaxAttempts:      3,
			InitialBackoffMs: 100,
			MaxBackoffMs:     1000,
		},
		JobsDir: jobsDir,
	}

	logger := lib.NewLogger(lib.LogLevelDebug)
	job, err := pipeline.CreateJob(crtdlPath, config, logger)
	require.NoError(t, err)

	httpClient := services.NewHTTPClient(2*time.Second, config.Retry, logger)
	_, err = pipeline.ExecuteImportStep(job, logger, httpClient, false)

	// Should fail with network error
	assert.Error(t, err)
}

// T054: Integration test for direct TORCH URL download

func TestPipeline_DirectTORCHURL_Download(t *testing.T) {
	// Setup: Create test environment
	tempDir := t.TempDir()
	jobsDir := filepath.Join(tempDir, "jobs")
	os.MkdirAll(jobsDir, 0755)

	// Mock NDJSON content to be returned
	ndjsonContent := `{"resourceType":"Bundle","type":"transaction","entry":[{"resource":{"resourceType":"Patient","id":"patient-1","birthDate":"1990-01-01","gender":"male"}}]}
{"resourceType":"Bundle","type":"transaction","entry":[{"resource":{"resourceType":"Observation","id":"obs-1","status":"final"}}]}`

	// Mock TORCH server with direct result URL access
	resultPath := "/fhir/extraction/result-abc123"

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle direct result URL GET (should return 200 immediately with file URLs)
		if r.Method == "GET" && r.URL.Path == resultPath {
			// Return completed extraction result directly
			result := map[string]interface{}{
				"resourceType": "Parameters",
				"parameter": []map[string]interface{}{
					{
						"name": "output",
						"part": []map[string]interface{}{
							{
								"name":     "url",
								"valueUrl": server.URL + "/output/Patient.ndjson",
							},
							{
								"name":     "url",
								"valueUrl": server.URL + "/output/Observation.ndjson",
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(result)
			return
		}

		// Handle file downloads
		if r.Method == "GET" && (r.URL.Path == "/output/Patient.ndjson" || r.URL.Path == "/output/Observation.ndjson") {
			w.Header().Set("Content-Type", "application/fhir+ndjson")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(ndjsonContent))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Direct TORCH result URL (not CRTDL file)
	torchResultURL := server.URL + resultPath

	// Create configuration
	config := models.ProjectConfig{
		Services: models.ServiceConfig{
			TORCH: models.TORCHConfig{
				BaseURL:                   server.URL,
				Username:                  "testuser",
				Password:                  "testpass",
				ExtractionTimeoutMinutes:  1,
				PollingIntervalSeconds:    1,
				MaxPollingIntervalSeconds: 5,
			},
		},
		Pipeline: models.PipelineConfig{
			EnabledSteps: []models.StepName{models.StepImport},
		},
		Retry: models.RetryConfig{
			MaxAttempts:      3,
			InitialBackoffMs: 100,
			MaxBackoffMs:     1000,
		},
		JobsDir: jobsDir,
	}

	// Create job with TORCH result URL input
	logger := lib.NewLogger(lib.LogLevelDebug)
	job, err := pipeline.CreateJob(torchResultURL, config, logger)
	require.NoError(t, err)

	// Verify job was created with correct input type
	assert.Equal(t, models.InputTypeTORCHURL, job.InputType, "Direct TORCH URL should be detected as InputTypeTORCHURL")
	assert.Equal(t, torchResultURL, job.InputSource)
	assert.NotEmpty(t, job.JobID)

	// Execute import step (should download directly without extraction submission)
	httpClient := services.NewHTTPClient(2*time.Second, config.Retry, logger)
	updatedJob, err := pipeline.ExecuteImportStep(job, logger, httpClient, false)

	// Verify successful execution
	require.NoError(t, err)
	assert.NotNil(t, updatedJob)

	// Verify import step completed
	importStep, found := models.GetStepByName(*updatedJob, models.StepImport)
	require.True(t, found)
	assert.Equal(t, models.StepStatusCompleted, importStep.Status)

	// Verify files were downloaded
	assert.Greater(t, updatedJob.TotalFiles, 0, "Should have downloaded at least one file")
	assert.Greater(t, updatedJob.TotalBytes, int64(0), "Should have non-zero bytes")

	// Verify NDJSON files exist in job directory
	importDir := services.GetJobOutputDir(jobsDir, job.JobID, models.StepImport)
	files, err := os.ReadDir(importDir)
	require.NoError(t, err)
	assert.NotEmpty(t, files, "Expected downloaded NDJSON files in import directory")

	// Verify file content
	ndjsonFileCount := 0
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".ndjson" {
			ndjsonFileCount++
			content, err := os.ReadFile(filepath.Join(importDir, file.Name()))
			require.NoError(t, err)
			assert.NotEmpty(t, content, "Downloaded file should not be empty")
		}
	}
	assert.Greater(t, ndjsonFileCount, 0, "Should have at least one NDJSON file")
}

func TestPipeline_DirectTORCHURL_EmptyResult(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	jobsDir := filepath.Join(tempDir, "jobs")
	os.MkdirAll(jobsDir, 0755)

	// Mock TORCH server returning empty result
	resultPath := "/fhir/extraction/empty-result"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == resultPath {
			// Return result with no output files
			result := map[string]interface{}{
				"resourceType": "Parameters",
				"parameter":    []map[string]interface{}{},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(result)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	torchResultURL := server.URL + resultPath

	config := models.ProjectConfig{
		Services: models.ServiceConfig{
			TORCH: models.TORCHConfig{
				BaseURL:                   server.URL,
				Username:                  "testuser",
				Password:                  "testpass",
				ExtractionTimeoutMinutes:  1,
				PollingIntervalSeconds:    1,
				MaxPollingIntervalSeconds: 5,
			},
		},
		Pipeline: models.PipelineConfig{
			EnabledSteps: []models.StepName{models.StepImport},
		},
		Retry: models.RetryConfig{
			MaxAttempts:      3,
			InitialBackoffMs: 100,
			MaxBackoffMs:     1000,
		},
		JobsDir: jobsDir,
	}

	logger := lib.NewLogger(lib.LogLevelDebug)
	job, err := pipeline.CreateJob(torchResultURL, config, logger)
	require.NoError(t, err)

	httpClient := services.NewHTTPClient(2*time.Second, config.Retry, logger)
	updatedJob, err := pipeline.ExecuteImportStep(job, logger, httpClient, false)

	// Empty result should be handled gracefully
	require.NoError(t, err)
	assert.NotNil(t, updatedJob)

	// Import step should complete with zero files
	importStep, found := models.GetStepByName(*updatedJob, models.StepImport)
	require.True(t, found)
	assert.Equal(t, models.StepStatusCompleted, importStep.Status)
	assert.Equal(t, 0, updatedJob.TotalFiles, "Empty result should have zero files")
}
