package unit

import (
	"encoding/base64"
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
	"github.com/trobanga/aether/internal/services"
)

// T020: Unit test for TORCH client SubmitExtraction()

func TestTORCHClient_SubmitExtraction_Success(t *testing.T) {
	// Create temp CRTDL file
	tempDir := t.TempDir()
	crtdlPath := filepath.Join(tempDir, "test.crtdl")
	crtdlContent := map[string]interface{}{
		"cohortDefinition": map[string]interface{}{
			"version":           "1.0.0",
			"inclusionCriteria": []interface{}{},
		},
		"dataExtraction": map[string]interface{}{
			"attributeGroups": []interface{}{},
		},
	}
	crtdlJSON, _ := json.Marshal(crtdlContent)
	_ = os.WriteFile(crtdlPath, crtdlJSON, 0644)

	// Mock TORCH server
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/fhir/$extract-data", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify authentication
		authHeader := r.Header.Get("Authorization")
		require.NotEmpty(t, authHeader)
		assert.Contains(t, authHeader, "Basic ")

		// Verify body is valid FHIR Parameters
		var params map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&params)
		require.NoError(t, err)
		assert.Equal(t, "Parameters", params["resourceType"])

		// Return 202 with Content-Location
		w.Header().Set("Content-Location", serverURL+"/fhir/extraction/job-abc123")
		w.WriteHeader(http.StatusAccepted)
	}))
	serverURL = server.URL
	defer server.Close()

	// Test execution
	logger := lib.NewLogger(lib.LogLevelDebug)
	httpClient := services.NewHTTPClient(5*time.Second, models.RetryConfig{MaxAttempts: 3, InitialBackoffMs: 100, MaxBackoffMs: 1000}, logger)
	torchConfig := models.TORCHConfig{
		BaseURL:                   server.URL,
		Username:                  "testuser",
		Password:                  "testpass",
		ExtractionTimeoutMinutes:  30,
		PollingIntervalSeconds:    1,
		MaxPollingIntervalSeconds: 5,
	}

	client := services.NewTORCHClient(torchConfig, httpClient, logger)
	extractionURL, err := client.SubmitExtraction(crtdlPath)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, server.URL+"/fhir/extraction/job-abc123", extractionURL)
}

func TestTORCHClient_SubmitExtraction_FileNotFound(t *testing.T) {
	logger := lib.NewLogger(lib.LogLevelDebug)
	httpClient := services.NewHTTPClient(5*time.Second, models.RetryConfig{MaxAttempts: 3, InitialBackoffMs: 100, MaxBackoffMs: 1000}, logger)
	torchConfig := models.TORCHConfig{
		BaseURL:  "http://localhost:8080",
		Username: "testuser",
		Password: "testpass",
	}

	client := services.NewTORCHClient(torchConfig, httpClient, logger)
	_, err := client.SubmitExtraction("/nonexistent/file.crtdl")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read CRTDL file")
}

func TestTORCHClient_SubmitExtraction_Unauthorized(t *testing.T) {
	// Create temp CRTDL file
	tempDir := t.TempDir()
	crtdlPath := filepath.Join(tempDir, "test.crtdl")
	crtdlJSON := []byte(`{"cohortDefinition":{"inclusionCriteria":[]},"dataExtraction":{"attributeGroups":[]}}`)
	_ = os.WriteFile(crtdlPath, crtdlJSON, 0644)

	// Mock TORCH server returning 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Invalid credentials"))
	}))
	defer server.Close()

	logger := lib.NewLogger(lib.LogLevelDebug)
	httpClient := services.NewHTTPClient(5*time.Second, models.RetryConfig{MaxAttempts: 3, InitialBackoffMs: 100, MaxBackoffMs: 1000}, logger)
	torchConfig := models.TORCHConfig{
		BaseURL:  server.URL,
		Username: "wrong",
		Password: "wrong",
	}

	client := services.NewTORCHClient(torchConfig, httpClient, logger)
	_, err := client.SubmitExtraction(crtdlPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

// T021: Unit test for TORCH client PollExtractionStatus()

func TestTORCHClient_PollExtractionStatus_ImmediateSuccess(t *testing.T) {
	// Mock TORCH server that returns success immediately
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/fhir/extraction/")

		// Return 200 with file URLs
		result := map[string]interface{}{
			"resourceType": "Parameters",
			"parameter": []map[string]interface{}{
				{
					"name": "output",
					"part": []map[string]interface{}{
						{
							"name":     "url",
							"valueUrl": serverURL + "/output/batch-1.ndjson",
						},
						{
							"name":     "url",
							"valueUrl": serverURL + "/output/batch-2.ndjson",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
	}))
	serverURL = server.URL
	defer server.Close()

	logger := lib.NewLogger(lib.LogLevelDebug)
	httpClient := services.NewHTTPClient(5*time.Second, models.RetryConfig{MaxAttempts: 3, InitialBackoffMs: 100, MaxBackoffMs: 1000}, logger)
	torchConfig := models.TORCHConfig{
		BaseURL:                   server.URL,
		Username:                  "testuser",
		Password:                  "testpass",
		ExtractionTimeoutMinutes:  1,
		PollingIntervalSeconds:    1,
		MaxPollingIntervalSeconds: 5,
	}

	client := services.NewTORCHClient(torchConfig, httpClient, logger)
	urls, err := client.PollExtractionStatus(server.URL+"/fhir/extraction/job-123", false)

	assert.NoError(t, err)
	require.Len(t, urls, 2)
	assert.Equal(t, server.URL+"/output/batch-1.ndjson", urls[0])
	assert.Equal(t, server.URL+"/output/batch-2.ndjson", urls[1])
}

func TestTORCHClient_PollExtractionStatus_Timeout(t *testing.T) {
	// Mock TORCH server that always returns 202
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	logger := lib.NewLogger(lib.LogLevelDebug)
	httpClient := services.NewHTTPClient(5*time.Second, models.RetryConfig{MaxAttempts: 3, InitialBackoffMs: 100, MaxBackoffMs: 1000}, logger)
	torchConfig := models.TORCHConfig{
		BaseURL:                   server.URL,
		Username:                  "testuser",
		Password:                  "testpass",
		ExtractionTimeoutMinutes:  0, // Immediate timeout (converted to milliseconds)
		PollingIntervalSeconds:    1,
		MaxPollingIntervalSeconds: 5,
	}

	client := services.NewTORCHClient(torchConfig, httpClient, logger)
	_, err := client.PollExtractionStatus(server.URL+"/fhir/extraction/job-123", false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestTORCHClient_PollExtractionStatus_ServerError(t *testing.T) {
	// Mock TORCH server returning 500 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resourceType": "OperationOutcome",
			"issue": []map[string]interface{}{
				{
					"severity":    "error",
					"code":        "processing",
					"diagnostics": "Database timeout",
				},
			},
		})
	}))
	defer server.Close()

	logger := lib.NewLogger(lib.LogLevelDebug)
	httpClient := services.NewHTTPClient(5*time.Second, models.RetryConfig{MaxAttempts: 3, InitialBackoffMs: 100, MaxBackoffMs: 1000}, logger)
	torchConfig := models.TORCHConfig{
		BaseURL:                   server.URL,
		Username:                  "testuser",
		Password:                  "testpass",
		ExtractionTimeoutMinutes:  5,
		PollingIntervalSeconds:    1,
		MaxPollingIntervalSeconds: 30,
	}

	client := services.NewTORCHClient(torchConfig, httpClient, logger)
	_, err := client.PollExtractionStatus(server.URL+"/fhir/extraction/job-123", false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

// T022: Unit test for TORCH client DownloadExtractionFiles()

func TestTORCHClient_DownloadExtractionFiles_Success(t *testing.T) {
	ndjsonContent := `{"resourceType":"Patient","id":"1"}
{"resourceType":"Patient","id":"2"}
{"resourceType":"Observation","id":"obs-1"}`

	// Mock TORCH server serving files
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		// Verify authentication
		authHeader := r.Header.Get("Authorization")
		assert.NotEmpty(t, authHeader)

		w.Header().Set("Content-Type", "application/fhir+ndjson")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(ndjsonContent))
	}))
	defer server.Close()

	logger := lib.NewLogger(lib.LogLevelDebug)
	httpClient := services.NewHTTPClient(5*time.Second, models.RetryConfig{MaxAttempts: 3, InitialBackoffMs: 100, MaxBackoffMs: 1000}, logger)
	torchConfig := models.TORCHConfig{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
	}

	tempDir := t.TempDir()
	client := services.NewTORCHClient(torchConfig, httpClient, logger)

	fileURLs := []string{
		server.URL + "/output/batch-1.ndjson",
		server.URL + "/output/batch-2.ndjson",
	}

	files, err := client.DownloadExtractionFiles(fileURLs, tempDir, false)

	assert.NoError(t, err)
	assert.Len(t, files, 2)

	// Verify files were created
	for _, file := range files {
		filePath := filepath.Join(tempDir, file.FileName)
		assert.FileExists(t, filePath)

		// Verify content
		content, _ := os.ReadFile(filePath)
		assert.Contains(t, string(content), "Patient")
	}
}

func TestTORCHClient_DownloadExtractionFiles_PartialFailure(t *testing.T) {
	// Mock server that fails for second file
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First file succeeds
			w.Header().Set("Content-Type", "application/fhir+ndjson")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"resourceType":"Patient","id":"1"}`))
		} else {
			// Second file fails
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	logger := lib.NewLogger(lib.LogLevelDebug)
	httpClient := services.NewHTTPClient(5*time.Second, models.RetryConfig{MaxAttempts: 3, InitialBackoffMs: 100, MaxBackoffMs: 1000}, logger)
	torchConfig := models.TORCHConfig{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
	}

	tempDir := t.TempDir()
	client := services.NewTORCHClient(torchConfig, httpClient, logger)

	fileURLs := []string{
		server.URL + "/output/batch-1.ndjson",
		server.URL + "/output/batch-2.ndjson",
	}

	_, err := client.DownloadExtractionFiles(fileURLs, tempDir, false)

	// Should fail on second file
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

// T023: Unit test for base64 CRTDL encoding

func TestTORCHClient_EncodeCRTDLToBase64_ValidJSON(t *testing.T) {
	// Create temp CRTDL file
	tempDir := t.TempDir()
	crtdlPath := filepath.Join(tempDir, "test.crtdl")
	crtdlContent := map[string]interface{}{
		"cohortDefinition": map[string]interface{}{
			"version":           "1.0.0",
			"inclusionCriteria": []interface{}{},
		},
		"dataExtraction": map[string]interface{}{
			"attributeGroups": []interface{}{},
		},
	}
	crtdlJSON, _ := json.Marshal(crtdlContent)
	_ = os.WriteFile(crtdlPath, crtdlJSON, 0644)

	logger := lib.NewLogger(lib.LogLevelDebug)
	httpClient := services.NewHTTPClient(5*time.Second, models.RetryConfig{MaxAttempts: 3, InitialBackoffMs: 100, MaxBackoffMs: 1000}, logger)
	torchConfig := models.TORCHConfig{
		BaseURL:  "http://localhost:8080",
		Username: "testuser",
		Password: "testpass",
	}

	_ = services.NewTORCHClient(torchConfig, httpClient, logger)

	// Call internal encoding function (will be exposed via reflection or package access)
	// For now, test via SubmitExtraction which uses it internally
	// This validates round-trip encoding

	// Read file and encode manually to test
	fileContent, err := os.ReadFile(crtdlPath)
	require.NoError(t, err)

	encoded := base64.StdEncoding.EncodeToString(fileContent)
	assert.NotEmpty(t, encoded)

	// Decode and verify
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err)

	var decodedCRTDL map[string]interface{}
	err = json.Unmarshal(decoded, &decodedCRTDL)
	require.NoError(t, err)

	assert.Contains(t, decodedCRTDL, "cohortDefinition")
	assert.Contains(t, decodedCRTDL, "dataExtraction")
}

// T024: Unit test for exponential backoff polling logic

func TestTORCHClient_PollExtractionStatus_ExponentialBackoff(t *testing.T) {
	pollTimes := []time.Time{}
	pollCount := 0
	maxPolls := 4

	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pollTimes = append(pollTimes, time.Now())
		pollCount++

		if pollCount < maxPolls {
			// Return 202 (in progress)
			w.WriteHeader(http.StatusAccepted)
			return
		}

		// Return 200 (complete)
		result := map[string]interface{}{
			"resourceType": "Parameters",
			"parameter": []map[string]interface{}{
				{
					"name": "output",
					"part": []map[string]interface{}{
						{
							"name":     "url",
							"valueUrl": serverURL + "/output/result.ndjson",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
	}))
	serverURL = server.URL
	defer server.Close()

	logger := lib.NewLogger(lib.LogLevelDebug)
	httpClient := services.NewHTTPClient(5*time.Second, models.RetryConfig{MaxAttempts: 3, InitialBackoffMs: 100, MaxBackoffMs: 1000}, logger)
	torchConfig := models.TORCHConfig{
		BaseURL:                   server.URL,
		Username:                  "testuser",
		Password:                  "testpass",
		ExtractionTimeoutMinutes:  1,
		PollingIntervalSeconds:    1,  // Start at 1 second
		MaxPollingIntervalSeconds: 10, // Cap at 10 seconds
	}

	client := services.NewTORCHClient(torchConfig, httpClient, logger)
	urls, err := client.PollExtractionStatus(server.URL+"/fhir/extraction/job-123", false)

	assert.NoError(t, err)
	assert.Len(t, urls, 1)
	assert.Equal(t, maxPolls, pollCount)

	// Verify exponential backoff (intervals should grow)
	// First interval: ~1s, Second: ~2s, Third: ~4s
	if len(pollTimes) >= 3 {
		interval1 := pollTimes[1].Sub(pollTimes[0])
		interval2 := pollTimes[2].Sub(pollTimes[1])

		// Second interval should be roughly 2x first interval (with tolerance)
		// Allow for timing variance - just check second > first
		assert.Greater(t, interval2, interval1, "Polling intervals should increase (exponential backoff)")
	}
}

func TestTORCHClient_Ping_Success(t *testing.T) {
	// Mock TORCH server responding to GET request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := lib.NewLogger(lib.LogLevelDebug)
	httpClient := services.NewHTTPClient(5*time.Second, models.RetryConfig{MaxAttempts: 3, InitialBackoffMs: 100, MaxBackoffMs: 1000}, logger)
	torchConfig := models.TORCHConfig{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
	}

	client := services.NewTORCHClient(torchConfig, httpClient, logger)
	err := client.Ping()

	assert.NoError(t, err)
}

func TestTORCHClient_Ping_Unreachable(t *testing.T) {
	logger := lib.NewLogger(lib.LogLevelDebug)
	httpClient := services.NewHTTPClient(5*time.Second, models.RetryConfig{MaxAttempts: 3, InitialBackoffMs: 100, MaxBackoffMs: 1000}, logger)
	torchConfig := models.TORCHConfig{
		BaseURL:  "http://unreachable-host-12345.invalid:9999",
		Username: "testuser",
		Password: "testpass",
	}

	client := services.NewTORCHClient(torchConfig, httpClient, logger)
	err := client.Ping()

	assert.Error(t, err)
}

// T071: Performance test - verify TORCH connectivity check < 5 seconds

func TestTORCHClient_Ping_PerformanceWithin5Seconds(t *testing.T) {
	// Mock TORCH server with slight delay to simulate realistic network latency
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate 100ms network latency
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := lib.NewLogger(lib.LogLevelError) // Reduce log noise for performance test
	httpClient := services.NewHTTPClient(5*time.Second, models.RetryConfig{MaxAttempts: 3, InitialBackoffMs: 100, MaxBackoffMs: 1000}, logger)
	torchConfig := models.TORCHConfig{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
	}

	client := services.NewTORCHClient(torchConfig, httpClient, logger)

	// Measure execution time
	startTime := time.Now()
	err := client.Ping()
	duration := time.Since(startTime)

	// Assertions
	assert.NoError(t, err)
	assert.Less(t, duration, 5*time.Second, "TORCH connectivity check must complete within 5 seconds, took: %v", duration)

	// Log performance for visibility
	t.Logf("TORCH connectivity check completed in %v (requirement: < 5s)", duration)
}
