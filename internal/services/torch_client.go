package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/trobanga/aether/internal/lib"
	"github.com/trobanga/aether/internal/models"
	"github.com/trobanga/aether/internal/ui"
)

// TORCHClient handles communication with TORCH server for CRTDL-based data extraction
// Per contracts/torch-api.md
type TORCHClient struct {
	config     models.TORCHConfig
	httpClient *HTTPClient
	logger     *lib.Logger
}

// TORCHExtractionRequest represents the FHIR Parameters resource for extraction submission
type TORCHExtractionRequest struct {
	ResourceType string           `json:"resourceType"`
	Parameter    []TORCHParameter `json:"parameter"`
}

// TORCHParameter represents a parameter in the FHIR Parameters resource
type TORCHParameter struct {
	Name              string `json:"name"`
	ValueBase64Binary string `json:"valueBase64Binary,omitempty"`
}

// TORCHExtractionResult represents the FHIR Parameters response with extraction results
type TORCHExtractionResult struct {
	ResourceType string                 `json:"resourceType"`
	Parameter    []TORCHResultParameter `json:"parameter"`
}

// TORCHResultParameter represents an output parameter containing file URLs
type TORCHResultParameter struct {
	Name string            `json:"name"`
	Part []TORCHResultPart `json:"part,omitempty"`
}

// TORCHResultPart represents a part of an output parameter (e.g., file URL)
type TORCHResultPart struct {
	Name     string `json:"name"`
	ValueURL string `json:"valueUrl,omitempty"`
}

// TORCHError represents errors from TORCH operations
type TORCHError struct {
	Operation  string // "submit", "poll", "download"
	StatusCode int
	Message    string
	ErrorType  models.ErrorType
}

func (e *TORCHError) Error() string {
	return fmt.Sprintf("TORCH %s error: HTTP %d: %s", e.Operation, e.StatusCode, e.Message)
}

// IsRetryable returns true if this error should be retried
func (e *TORCHError) IsRetryable() bool {
	return e.ErrorType == models.ErrorTypeTransient
}

// ErrExtractionTimeout is returned when extraction polling exceeds configured timeout
var ErrExtractionTimeout = fmt.Errorf("TORCH extraction timeout exceeded")

// ErrInvalidCRTDL is returned when CRTDL file is malformed
var ErrInvalidCRTDL = fmt.Errorf("invalid CRTDL file")

// NewTORCHClient creates a new TORCH client with the given configuration
func NewTORCHClient(config models.TORCHConfig, httpClient *HTTPClient, logger *lib.Logger) *TORCHClient {
	return &TORCHClient{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
	}
}

// SubmitExtraction submits a CRTDL file for extraction to TORCH server
// Returns the Content-Location URL for polling extraction status
// Per TORCH API: POST /fhir/$extract-data with base64-encoded CRTDL
func (c *TORCHClient) SubmitExtraction(crtdlPath string) (string, error) {
	c.logger.Info("Submitting CRTDL extraction to TORCH", "file", crtdlPath, "server", c.config.BaseURL)

	// Encode CRTDL to base64
	base64Content, err := c.encodeCRTDLToBase64(crtdlPath)
	if err != nil {
		return "", fmt.Errorf("failed to encode CRTDL: %w", err)
	}

	// Build FHIR Parameters request
	requestBody := TORCHExtractionRequest{
		ResourceType: "Parameters",
		Parameter: []TORCHParameter{
			{
				Name:              "crtdl",
				ValueBase64Binary: base64Content,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	c.logger.Debug("TORCH extraction request", "body_size", len(jsonBody))

	// Construct URL
	url := c.config.BaseURL + "/fhir/$extract-data"

	// Create HTTP request with authentication
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.buildBasicAuthHeader())

	// Send request
	resp, err := c.httpClient.client.Do(req)
	if err != nil {
		c.logger.Error("TORCH submission failed", "error", err)
		return "", &TORCHError{
			Operation:  "submit",
			StatusCode: 0,
			Message:    err.Error(),
			ErrorType:  models.ErrorTypeTransient,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// Check for errors
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorType := lib.ClassifyHTTPError(resp.StatusCode)

		c.logger.Error("TORCH submission returned error",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"error_body", string(bodyBytes))

		return "", &TORCHError{
			Operation:  "submit",
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
			ErrorType:  errorType,
		}
	}

	// Extract Content-Location header
	contentLocation := resp.Header.Get("Content-Location")
	if contentLocation == "" {
		return "", fmt.Errorf("TORCH server did not return Content-Location header")
	}

	// Normalize URL to use our configured base URL (handles Docker internal URLs)
	contentLocation = c.normalizeURL(contentLocation)

	c.logger.Info("TORCH extraction submitted successfully", "extraction_url", contentLocation)

	return contentLocation, nil
}

// PollExtractionStatus polls the extraction status URL until completion or timeout
// Returns the list of file URLs when extraction is complete
// Per TORCH API: GET Content-Location URL until HTTP 200, handle HTTP 202 as in-progress
// FR-029c: Uses spinner for unknown-duration polling
func (c *TORCHClient) PollExtractionStatus(extractionURL string, showProgress bool) ([]string, error) {
	c.logger.Info("Polling TORCH extraction status", "url", extractionURL)

	timeout := time.Duration(c.config.ExtractionTimeoutMinutes) * time.Minute
	startTime := time.Now()
	pollInterval := time.Duration(c.config.PollingIntervalSeconds) * time.Second
	maxPollInterval := time.Duration(c.config.MaxPollingIntervalSeconds) * time.Second

	pollCount := 0

	// Start spinner for polling (FR-029c: unknown duration)
	var spinner *ui.Spinner
	if showProgress {
		spinner = ui.NewSpinner("Waiting for TORCH extraction to complete")
		spinner.Start()
		defer func() {
			if spinner != nil {
				spinner.Stop(true)
			}
		}()
	}

	for {
		// Check timeout
		if time.Since(startTime) > timeout {
			c.logger.Error("TORCH extraction timeout",
				"duration", time.Since(startTime),
				"timeout", timeout,
				"polls", pollCount)
			return nil, ErrExtractionTimeout
		}

		pollCount++
		c.logger.Debug("Polling TORCH extraction", "attempt", pollCount, "interval", pollInterval)

		// Create request with authentication
		req, err := http.NewRequest("GET", extractionURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create poll request: %w", err)
		}

		req.Header.Set("Authorization", c.buildBasicAuthHeader())
		req.Header.Set("Accept", "application/json")

		// Send request
		resp, err := c.httpClient.client.Do(req)
		if err != nil {
			c.logger.Error("TORCH polling failed", "error", err, "attempt", pollCount)
			return nil, &TORCHError{
				Operation:  "poll",
				StatusCode: 0,
				Message:    err.Error(),
				ErrorType:  models.ErrorTypeTransient,
			}
		}

		// Handle response
		if resp.StatusCode == http.StatusAccepted {
			// Still in progress
			_ = resp.Body.Close()
			c.logger.Debug("TORCH extraction in progress", "attempt", pollCount)

			// Wait with exponential backoff
			time.Sleep(pollInterval)

			// Double interval for next poll (exponential backoff)
			pollInterval = pollInterval * 2
			if pollInterval > maxPollInterval {
				pollInterval = maxPollInterval
			}

			continue
		}

		// Read response body
		bodyBytes, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// Extraction complete - parse result
			c.logger.Info("TORCH extraction completed", "polls", pollCount)
			c.logger.Debug("TORCH extraction response body", "body", string(bodyBytes))
			return c.parseExtractionResult(bodyBytes)
		}

		// Error response
		errorType := lib.ClassifyHTTPError(resp.StatusCode)
		c.logger.Error("TORCH extraction failed",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"error_body", string(bodyBytes))

		return nil, &TORCHError{
			Operation:  "poll",
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
			ErrorType:  errorType,
		}
	}
}

// DownloadExtractionFiles downloads all NDJSON files from the extraction result
// Returns list of downloaded files with metadata
// FR-029c: Uses spinner for each file download (unknown size)
func (c *TORCHClient) DownloadExtractionFiles(fileURLs []string, destinationDir string, showProgress bool) ([]models.FHIRDataFile, error) {
	c.logger.Info("Downloading TORCH extraction files",
		"file_count", len(fileURLs),
		"destination", destinationDir)

	if len(fileURLs) == 0 {
		c.logger.Warn("No files to download from TORCH extraction")
		return []models.FHIRDataFile{}, nil
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(destinationDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	downloadedFiles := []models.FHIRDataFile{}

	for i, fileURL := range fileURLs {
		c.logger.Debug("Downloading TORCH file", "index", i+1, "total", len(fileURLs), "url", fileURL)

		// Determine filename
		fileName := filepath.Base(fileURL)
		if fileName == "." || fileName == "/" {
			fileName = fmt.Sprintf("torch-batch-%d.ndjson", i+1)
		}

		// Ensure .ndjson extension
		if !strings.HasSuffix(fileName, ".ndjson") {
			fileName = fileName + ".ndjson"
		}

		destPath := filepath.Join(destinationDir, fileName)

		// Start spinner for this download (FR-029c: unknown size)
		var spinner *ui.Spinner
		if showProgress {
			spinnerMsg := fmt.Sprintf("Downloading file %d/%d: %s", i+1, len(fileURLs), fileName)
			spinner = ui.NewSpinner(spinnerMsg)
			spinner.Start()
		}

		// Download file
		file, err := c.downloadFile(fileURL, destPath)

		// Stop spinner
		if spinner != nil {
			spinner.Stop(err == nil)
		}

		if err != nil {
			c.logger.Error("Failed to download TORCH file", "url", fileURL, "error", err)
			return nil, fmt.Errorf("failed to download file %s: %w", fileURL, err)
		}

		downloadedFiles = append(downloadedFiles, file)
		c.logger.Info("Downloaded TORCH file",
			"file", fileName,
			"size", file.FileSize,
			"resources", file.LineCount)
	}

	c.logger.Info("All TORCH files downloaded successfully", "total_files", len(downloadedFiles))

	return downloadedFiles, nil
}

// downloadFile downloads a single file from URL to destination path
func (c *TORCHClient) downloadFile(fileURL, destPath string) (models.FHIRDataFile, error) {
	// Create request
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return models.FHIRDataFile{}, fmt.Errorf("failed to create download request: %w", err)
	}

	// Only add authentication if downloading from TORCH API server (not file server)
	// File servers (nginx) typically don't require auth
	if c.config.FileServerURL == "" || !strings.Contains(fileURL, c.config.FileServerURL) {
		req.Header.Set("Authorization", c.buildBasicAuthHeader())
	}
	req.Header.Set("Accept", "application/fhir+ndjson")

	// Send request
	resp, err := c.httpClient.client.Do(req)
	if err != nil {
		return models.FHIRDataFile{}, &TORCHError{
			Operation:  "download",
			StatusCode: 0,
			Message:    err.Error(),
			ErrorType:  models.ErrorTypeTransient,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// Check for errors
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorType := lib.ClassifyHTTPError(resp.StatusCode)

		return models.FHIRDataFile{}, &TORCHError{
			Operation:  "download",
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
			ErrorType:  errorType,
		}
	}

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return models.FHIRDataFile{}, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = destFile.Close() }()

	// Copy content
	bytesWritten, err := io.Copy(destFile, resp.Body)
	if err != nil {
		_ = os.Remove(destPath)
		return models.FHIRDataFile{}, fmt.Errorf("failed to write file: %w", err)
	}

	// Count lines (FHIR resources)
	lineCount, _ := lib.CountResourcesInFile(destPath)

	// Extract resource type from filename
	fileName := filepath.Base(destPath)
	resourceType := models.GetResourceTypeFromFilename(fileName)

	return models.FHIRDataFile{
		FileName:     fileName,
		FilePath:     fileName, // Relative to job directory
		ResourceType: resourceType,
		FileSize:     bytesWritten,
		SourceStep:   models.StepImport,
		LineCount:    lineCount,
		CreatedAt:    lib.GetFileModTime(destPath),
	}, nil
}

// Ping checks connectivity to TORCH server
// Used by ValidateServiceConnectivity()
func (c *TORCHClient) Ping() error {
	c.logger.Debug("Checking TORCH server connectivity", "url", c.config.BaseURL)

	// Simple GET request to base URL
	req, err := http.NewRequest("GET", c.config.BaseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create ping request: %w", err)
	}

	req.Header.Set("Authorization", c.buildBasicAuthHeader())

	resp, err := c.httpClient.client.Do(req)
	if err != nil {
		c.logger.Error("TORCH ping failed", "error", err)
		return fmt.Errorf("TORCH server unreachable: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Accept any non-5xx response as "server is up"
	if resp.StatusCode >= 500 {
		return fmt.Errorf("TORCH server error: HTTP %d", resp.StatusCode)
	}

	c.logger.Debug("TORCH server ping successful", "status", resp.StatusCode)
	return nil
}

// encodeCRTDLToBase64 reads CRTDL file and encodes it to base64
func (c *TORCHClient) encodeCRTDLToBase64(crtdlPath string) (string, error) {
	// Read CRTDL file
	crtdlContent, err := os.ReadFile(crtdlPath)
	if err != nil {
		return "", fmt.Errorf("failed to read CRTDL file: %w", err)
	}

	if len(crtdlContent) == 0 {
		return "", fmt.Errorf("CRTDL file is empty")
	}

	// Validate it's valid JSON
	var crtdl map[string]interface{}
	if err := json.Unmarshal(crtdlContent, &crtdl); err != nil {
		return "", fmt.Errorf("CRTDL file is not valid JSON: %w", err)
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(crtdlContent)

	c.logger.Debug("Encoded CRTDL to base64",
		"original_size", len(crtdlContent),
		"encoded_size", len(encoded))

	return encoded, nil
}

// parseExtractionResult parses FHIR Parameters response and extracts file URLs
func (c *TORCHClient) parseExtractionResult(responseBody []byte) ([]string, error) {
	var result TORCHExtractionResult
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse extraction result: %w", err)
	}

	if result.ResourceType != "Parameters" {
		return nil, fmt.Errorf("unexpected response type: %s (expected Parameters)", result.ResourceType)
	}

	fileURLs := []string{}

	// Extract file URLs from output parameters
	for _, param := range result.Parameter {
		if param.Name == "output" {
			for _, part := range param.Part {
				if part.Name == "url" && part.ValueURL != "" {
					// Normalize URL to use our configured base URL (handles Docker internal URLs)
					normalizedURL := c.normalizeURL(part.ValueURL)
					fileURLs = append(fileURLs, normalizedURL)
				}
			}
		}
	}

	c.logger.Debug("Parsed extraction result", "file_count", len(fileURLs))

	return fileURLs, nil
}

// normalizeURL ensures URLs use the configured base URL instead of internal Docker URLs
// For file downloads, uses FileServerURL; for status polling, uses BaseURL
func (c *TORCHClient) normalizeURL(rawURL string) string {
	// Determine which base URL to use based on the URL pattern
	baseURL := c.config.BaseURL

	c.logger.Debug("URL normalization starting",
		"raw_url", rawURL,
		"base_url", c.config.BaseURL,
		"file_server_url", c.config.FileServerURL)

	// If URL looks like a file download (contains file extension or not a status URL), use file server URL
	if c.config.FileServerURL != "" && !strings.Contains(rawURL, "__status") && !strings.Contains(rawURL, "$extract-data") {
		baseURL = c.config.FileServerURL
		c.logger.Debug("Using file server URL for download", "file_server_url", baseURL)
	}

	// If absolute URL, extract path and combine with appropriate base URL
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			c.logger.Warn("Failed to parse URL, using as-is", "url", rawURL, "error", err)
			return rawURL
		}

		// Rebuild URL with appropriate base URL host/port
		normalizedURL := baseURL + parsedURL.Path
		if parsedURL.RawQuery != "" {
			normalizedURL += "?" + parsedURL.RawQuery
		}

		c.logger.Debug("Normalized URL", "original", rawURL, "normalized", normalizedURL, "base_used", baseURL)
		return normalizedURL
	}

	// Relative URL - prepend appropriate base URL
	return baseURL + rawURL
}

// buildBasicAuthHeader creates Basic authentication header
func (c *TORCHClient) buildBasicAuthHeader() string {
	credentials := c.config.Username + ":" + c.config.Password
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return "Basic " + encoded
}
