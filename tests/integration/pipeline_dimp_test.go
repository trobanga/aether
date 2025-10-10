package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// T055: Integration test for full DIMP step execution

func TestPipelineDIMP_EndToEnd(t *testing.T) {
	// Setup mock DIMP service
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/$de-identify", r.URL.Path)

		var resource map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&resource)

		// Pseudonymize the resource
		resource["id"] = "pseudo-" + resource["id"].(string)
		if names, ok := resource["name"].([]interface{}); ok && len(names) > 0 {
			if nameMap, ok := names[0].(map[string]interface{}); ok {
				nameMap["family"] = "REDACTED"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resource)
	}))
	defer server.Close()

	// Create test job with imported FHIR files
	// tmpDir := t.TempDir()
	// jobID := "test-dimp-job"
	// jobDir := filepath.Join(tmpDir, "jobs", jobID)
	// importDir := filepath.Join(jobDir, "import")
	// os.MkdirAll(importDir, 0755)
	//
	// // Write test FHIR NDJSON files
	// patientsFile := filepath.Join(importDir, "patients.ndjson")
	// patients := []map[string]interface{}{
	//     {"resourceType": "Patient", "id": "p1", "name": []map[string]string{{"family": "Smith"}}},
	//     {"resourceType": "Patient", "id": "p2", "name": []map[string]string{{"family": "Jones"}}},
	// }
	// writeNDJSON(t, patientsFile, patients)
	//
	// // Create job state
	// job := models.PipelineJob{
	//     JobID:       jobID,
	//     CurrentStep: "dimp",
	//     Status:      models.JobStatusInProgress,
	//     Config: models.ProjectConfig{
	//         Services: models.ServiceConfig{
	//             DIMPUrl: server.URL,
	//         },
	//     },
	// }
	//
	// // Execute DIMP step
	// err := pipeline.ExecuteDIMPStep(&job, jobDir)
	// assert.NoError(t, err)
	//
	// // Verify output
	// outputFile := filepath.Join(jobDir, "pseudonymized", "dimped_patients.ndjson")
	// assert.FileExists(t, outputFile)
	//
	// // Read and verify pseudonymized resources
	// resources := readNDJSON(t, outputFile)
	// assert.Len(t, resources, 2)
	// assert.Equal(t, "pseudo-p1", resources[0]["id"])
	// assert.Equal(t, "pseudo-p2", resources[1]["id"])
	// assert.Equal(t, "REDACTED", resources[0]["name"].([]interface{})[0].(map[string]interface{})["family"])

	t.Skip("Skipping until internal/pipeline/dimp.go is implemented")
}

func TestPipelineDIMP_MultipleFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resource map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&resource)
		resource["id"] = "pseudo-" + resource["id"].(string)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resource)
	}))
	defer server.Close()

	// Test processing multiple FHIR files
	// tmpDir := t.TempDir()
	// jobID := "test-multi-files"
	// jobDir := filepath.Join(tmpDir, "jobs", jobID)
	// importDir := filepath.Join(jobDir, "import")
	// os.MkdirAll(importDir, 0755)
	//
	// // Create multiple NDJSON files
	// files := []string{"patients.ndjson", "observations.ndjson", "conditions.ndjson"}
	// for _, filename := range files {
	//     filepath := filepath.Join(importDir, filename)
	//     resources := []map[string]interface{}{
	//         {"resourceType": "Patient", "id": "test-1"},
	//         {"resourceType": "Patient", "id": "test-2"},
	//     }
	//     writeNDJSON(t, filepath, resources)
	// }
	//
	// job := createDIMPJob(t, jobID, server.URL)
	//
	// err := pipeline.ExecuteDIMPStep(&job, jobDir)
	// assert.NoError(t, err)
	//
	// // Verify all files were processed
	// outputDir := filepath.Join(jobDir, "pseudonymized")
	// for _, filename := range files {
	//     outputFile := filepath.Join(outputDir, "dimped_"+filename)
	//     assert.FileExists(t, outputFile)
	// }

	t.Skip("Skipping until internal/pipeline/dimp.go is implemented")
}

func TestPipelineDIMP_ProgressReporting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resource map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&resource)
		resource["id"] = "pseudo-" + resource["id"].(string)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resource)
	}))
	defer server.Close()

	// Test that progress is reported during DIMP processing
	// tmpDir := t.TempDir()
	// jobID := "test-progress"
	// jobDir := filepath.Join(tmpDir, "jobs", jobID)
	// importDir := filepath.Join(jobDir, "import")
	// os.MkdirAll(importDir, 0755)
	//
	// // Create file with many resources to observe progress
	// filepath := filepath.Join(importDir, "many_patients.ndjson")
	// resources := make([]map[string]interface{}, 100)
	// for i := 0; i < 100; i++ {
	//     resources[i] = map[string]interface{}{
	//         "resourceType": "Patient",
	//         "id":           fmt.Sprintf("p%d", i),
	//     }
	// }
	// writeNDJSON(t, filepath, resources)
	//
	// job := createDIMPJob(t, jobID, server.URL)
	//
	// err := pipeline.ExecuteDIMPStep(&job, jobDir)
	// assert.NoError(t, err)
	//
	// // Verify job state was updated with progress
	// assert.Equal(t, 100, job.Steps[0].FilesProcessed)

	t.Skip("Skipping until internal/pipeline/dimp.go is implemented")
}

func TestPipelineDIMP_ServiceError_NonRetryable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"code": "invalid", "message": "Bad resource"}}`))
	}))
	defer server.Close()

	// Test that 400 errors fail the step without retry
	// tmpDir := t.TempDir()
	// jobID := "test-400-error"
	// jobDir := filepath.Join(tmpDir, "jobs", jobID)
	// importDir := filepath.Join(jobDir, "import")
	// os.MkdirAll(importDir, 0755)
	//
	// filepath := filepath.Join(importDir, "patients.ndjson")
	// resources := []map[string]interface{}{
	//     {"resourceType": "Patient", "id": "p1"},
	// }
	// writeNDJSON(t, filepath, resources)
	//
	// job := createDIMPJob(t, jobID, server.URL)
	//
	// err := pipeline.ExecuteDIMPStep(&job, jobDir)
	// assert.Error(t, err)
	// assert.Contains(t, err.Error(), "400")
	//
	// // Verify job state shows failed step
	// assert.Equal(t, models.StepStatusFailed, job.Steps[0].Status)
	// assert.Equal(t, models.ErrorTypeNonTransient, job.Steps[0].LastError.Type)

	t.Skip("Skipping until internal/pipeline/dimp.go is implemented")
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
		var resource map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&resource)
		resource["id"] = "pseudo-" + resource["id"].(string)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resource)
	}))
	defer server.Close()

	// Test that 500 errors are retried
	// tmpDir := t.TempDir()
	// jobID := "test-retry"
	// jobDir := filepath.Join(tmpDir, "jobs", jobID)
	// importDir := filepath.Join(jobDir, "import")
	// os.MkdirAll(importDir, 0755)
	//
	// filepath := filepath.Join(importDir, "patients.ndjson")
	// resources := []map[string]interface{}{
	//     {"resourceType": "Patient", "id": "p1"},
	// }
	// writeNDJSON(t, filepath, resources)
	//
	// job := createDIMPJob(t, jobID, server.URL)
	//
	// err := pipeline.ExecuteDIMPStep(&job, jobDir)
	// assert.NoError(t, err)
	// assert.Equal(t, 3, callCount, "Should retry until success")
	//
	// // Verify output
	// outputFile := filepath.Join(jobDir, "pseudonymized", "dimped_patients.ndjson")
	// assert.FileExists(t, outputFile)

	t.Skip("Skipping until internal/pipeline/dimp.go is implemented")
}

func TestPipelineDIMP_StepDisabled(t *testing.T) {
	// Test that DIMP step is skipped if not enabled in config
	// tmpDir := t.TempDir()
	// jobID := "test-disabled"
	// jobDir := filepath.Join(tmpDir, "jobs", jobID)
	//
	// job := models.PipelineJob{
	//     JobID:       jobID,
	//     CurrentStep: "dimp",
	//     Status:      models.JobStatusInProgress,
	//     Config: models.ProjectConfig{
	//         Pipeline: models.PipelineConfig{
	//             EnabledSteps: []models.StepName{
	//                 models.StepImport,
	//                 // DIMP not enabled
	//                 models.StepValidation,
	//             },
	//         },
	//     },
	// }
	//
	// err := pipeline.ExecuteDIMPStep(&job, jobDir)
	// assert.NoError(t, err, "Should skip without error")
	//
	// // Verify no output directory created
	// pseudonymizedDir := filepath.Join(jobDir, "pseudonymized")
	// _, err = os.Stat(pseudonymizedDir)
	// assert.True(t, os.IsNotExist(err))

	t.Skip("Skipping until internal/pipeline/dimp.go is implemented")
}

// Helper functions removed - were unused as tests are currently skipped
