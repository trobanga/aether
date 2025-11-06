package integration

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trobanga/aether/internal/lib"
	"github.com/trobanga/aether/internal/models"
	"github.com/trobanga/aether/internal/pipeline"
)

// Integration test for full DIMP step execution

func TestPipelineDIMP_EndToEnd(t *testing.T) {
	// Setup mock DIMP service
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/$de-identify", r.URL.Path)

		var resource map[string]any
		_ = json.NewDecoder(r.Body).Decode(&resource)

		// Pseudonymize the resource
		resource["id"] = "pseudo-" + resource["id"].(string)
		if names, ok := resource["name"].([]any); ok && len(names) > 0 {
			if nameMap, ok := names[0].(map[string]any); ok {
				nameMap["family"] = "REDACTED"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resource)
	}))
	defer server.Close()

	// Create test job with imported FHIR files
	tmpDir := t.TempDir()
	jobID := "test-dimp-job"
	jobDir := filepath.Join(tmpDir, "jobs", jobID)
	importDir := filepath.Join(jobDir, "import")
	err := os.MkdirAll(importDir, 0755)
	assert.NoError(t, err)

	// Write test FHIR NDJSON files
	patientsFile := filepath.Join(importDir, "patients.ndjson")
	patients := []map[string]any{
		{"resourceType": "Patient", "id": "p1", "name": []map[string]any{{"family": "Smith"}}},
		{"resourceType": "Patient", "id": "p2", "name": []map[string]any{{"family": "Jones"}}},
	}
	writeNDJSON(t, patientsFile, patients)

	// Create job state
	job := models.PipelineJob{
		JobID:       jobID,
		CurrentStep: string(models.StepDIMP),
		Status:      models.JobStatusInProgress,
		Config: models.ProjectConfig{
			Services: models.ServiceConfig{
				DIMP: models.DIMPConfig{
					URL: server.URL,
				},
			},
			Pipeline: models.PipelineConfig{
				EnabledSteps: []models.StepName{models.StepDIMP},
			},
		},
	}

	// Create logger
	logger := lib.NewLogger(lib.LogLevelInfo)

	// Execute DIMP step
	err = pipeline.ExecuteDIMPStep(&job, jobDir, logger)
	assert.NoError(t, err)

	// Verify output
	outputFile := filepath.Join(jobDir, "pseudonymized", "dimped_patients.ndjson")
	assert.FileExists(t, outputFile)

	// Read and verify pseudonymized resources
	resources := readNDJSON(t, outputFile)
	assert.Len(t, resources, 2)
	assert.Equal(t, "pseudo-p1", resources[0]["id"])
	assert.Equal(t, "pseudo-p2", resources[1]["id"])
	assert.Equal(t, "REDACTED", resources[0]["name"].([]any)[0].(map[string]any)["family"])
}

func TestPipelineDIMP_MultipleFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resource map[string]any
		_ = json.NewDecoder(r.Body).Decode(&resource)
		resource["id"] = "pseudo-" + resource["id"].(string)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resource)
	}))
	defer server.Close()

	// Test processing multiple FHIR files
	tmpDir := t.TempDir()
	jobID := "test-multi-files"
	jobDir := filepath.Join(tmpDir, "jobs", jobID)
	importDir := filepath.Join(jobDir, "import")
	err := os.MkdirAll(importDir, 0755)
	assert.NoError(t, err)

	// Create multiple NDJSON files
	files := []string{"patients.ndjson", "observations.ndjson", "conditions.ndjson"}
	for _, filename := range files {
		filePath := filepath.Join(importDir, filename)
		resources := []map[string]any{
			{"resourceType": "Patient", "id": "test-1"},
			{"resourceType": "Patient", "id": "test-2"},
		}
		writeNDJSON(t, filePath, resources)
	}

	job := createDIMPJob(jobID, server.URL)
	logger := lib.NewLogger(lib.LogLevelInfo)

	err = pipeline.ExecuteDIMPStep(job, jobDir, logger)
	assert.NoError(t, err)

	// Verify all files were processed
	outputDir := filepath.Join(jobDir, "pseudonymized")
	for _, filename := range files {
		outputFile := filepath.Join(outputDir, "dimped_"+filename)
		assert.FileExists(t, outputFile)
	}
}

func TestPipelineDIMP_ProgressReporting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resource map[string]any
		_ = json.NewDecoder(r.Body).Decode(&resource)
		resource["id"] = "pseudo-" + resource["id"].(string)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resource)
	}))
	defer server.Close()

	// Test that progress is reported during DIMP processing
	tmpDir := t.TempDir()
	jobID := "test-progress"
	jobDir := filepath.Join(tmpDir, "jobs", jobID)
	importDir := filepath.Join(jobDir, "import")
	err := os.MkdirAll(importDir, 0755)
	assert.NoError(t, err)

	// Create file with many resources to observe progress
	filePath := filepath.Join(importDir, "many_patients.ndjson")
	resources := make([]map[string]any, 100)
	for i := 0; i < 100; i++ {
		resources[i] = map[string]any{
			"resourceType": "Patient",
			"id":           fmt.Sprintf("p%d", i),
		}
	}
	writeNDJSON(t, filePath, resources)

	job := createDIMPJob(jobID, server.URL)
	logger := lib.NewLogger(lib.LogLevelInfo)

	err = pipeline.ExecuteDIMPStep(job, jobDir, logger)
	assert.NoError(t, err)

	// Verify job state was updated with progress
	dimpStep := getDIMPStep(job)
	assert.NotNil(t, dimpStep)
	assert.Equal(t, 1, dimpStep.FilesProcessed)
}

func TestPipelineDIMP_ServiceError_NonRetryable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"code": "invalid", "message": "Bad resource"}}`))
	}))
	defer server.Close()

	// Test that 400 errors fail the step without retry
	tmpDir := t.TempDir()
	jobID := "test-400-error"
	jobDir := filepath.Join(tmpDir, "jobs", jobID)
	importDir := filepath.Join(jobDir, "import")
	err := os.MkdirAll(importDir, 0755)
	assert.NoError(t, err)

	filePath := filepath.Join(importDir, "patients.ndjson")
	resources := []map[string]any{
		{"resourceType": "Patient", "id": "p1"},
	}
	writeNDJSON(t, filePath, resources)

	job := createDIMPJob(jobID, server.URL)
	logger := lib.NewLogger(lib.LogLevelInfo)

	err = pipeline.ExecuteDIMPStep(job, jobDir, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "400")

	// Verify job state shows failed step
	dimpStep := getDIMPStep(job)
	assert.NotNil(t, dimpStep)
	assert.Equal(t, models.StepStatusFailed, dimpStep.Status)
	assert.Equal(t, models.ErrorTypeNonTransient, dimpStep.LastError.Type)
}

func TestPipelineDIMP_ServiceError_Retryable(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Succeed on 3rd attempt
		var resource map[string]any
		_ = json.NewDecoder(r.Body).Decode(&resource)
		resource["id"] = "pseudo-" + resource["id"].(string)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resource)
	}))
	defer server.Close()

	// Test that 500 errors are retried
	tmpDir := t.TempDir()
	jobID := "test-retry"
	jobDir := filepath.Join(tmpDir, "jobs", jobID)
	importDir := filepath.Join(jobDir, "import")
	err := os.MkdirAll(importDir, 0755)
	assert.NoError(t, err)

	filePath := filepath.Join(importDir, "patients.ndjson")
	resources := []map[string]any{
		{"resourceType": "Patient", "id": "p1"},
	}
	writeNDJSON(t, filePath, resources)

	job := createDIMPJob(jobID, server.URL)
	logger := lib.NewLogger(lib.LogLevelInfo)

	err = pipeline.ExecuteDIMPStep(job, jobDir, logger)
	assert.NoError(t, err)
	assert.Equal(t, 3, callCount, "Should retry until success")

	// Verify output
	outputFile := filepath.Join(jobDir, "pseudonymized", "dimped_patients.ndjson")
	assert.FileExists(t, outputFile)
}

func TestPipelineDIMP_StepDisabled(t *testing.T) {
	// Test that DIMP step is skipped if not enabled in config
	tmpDir := t.TempDir()
	jobID := "test-disabled"
	jobDir := filepath.Join(tmpDir, "jobs", jobID)

	job := models.PipelineJob{
		JobID:       jobID,
		CurrentStep: string(models.StepDIMP),
		Status:      models.JobStatusInProgress,
		Config: models.ProjectConfig{
			Pipeline: models.PipelineConfig{
				EnabledSteps: []models.StepName{
					models.StepLocalImport,
					// DIMP not enabled
					models.StepValidation,
				},
			},
		},
	}

	logger := lib.NewLogger(lib.LogLevelInfo)

	err := pipeline.ExecuteDIMPStep(&job, jobDir, logger)
	assert.NoError(t, err, "Should skip without error")

	// Verify no output directory created
	pseudonymizedDir := filepath.Join(jobDir, "pseudonymized")
	_, err = os.Stat(pseudonymizedDir)
	assert.True(t, os.IsNotExist(err))
}

// Helper functions

// writeNDJSON writes an array of resources to an NDJSON file
func writeNDJSON(t *testing.T, filePath string, resources []map[string]any) {
	file, err := os.Create(filePath)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, file.Close())
	}()

	encoder := json.NewEncoder(file)
	for _, resource := range resources {
		err := encoder.Encode(resource)
		assert.NoError(t, err)
	}
}

// readNDJSON reads an NDJSON file and returns array of resources
func readNDJSON(t *testing.T, filePath string) []map[string]any {
	file, err := os.Open(filePath)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, file.Close())
	}()

	var resources []map[string]any
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var resource map[string]any
		err := json.Unmarshal(scanner.Bytes(), &resource)
		assert.NoError(t, err)
		resources = append(resources, resource)
	}
	assert.NoError(t, scanner.Err())
	return resources
}

// createDIMPJob creates a test job with DIMP configuration
func createDIMPJob(jobID, dimpURL string) *models.PipelineJob {
	return &models.PipelineJob{
		JobID:       jobID,
		CurrentStep: string(models.StepDIMP),
		Status:      models.JobStatusInProgress,
		Config: models.ProjectConfig{
			Services: models.ServiceConfig{
				DIMP: models.DIMPConfig{
					URL: dimpURL,
				},
			},
			Pipeline: models.PipelineConfig{
				EnabledSteps: []models.StepName{models.StepDIMP},
			},
		},
	}
}

// getDIMPStep finds and returns the DIMP step from a job
func getDIMPStep(job *models.PipelineJob) *models.PipelineStep {
	for i := range job.Steps {
		if job.Steps[i].Name == models.StepDIMP {
			return &job.Steps[i]
		}
	}
	return nil
}
