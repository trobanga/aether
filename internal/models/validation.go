package models

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Validate checks if a PipelineJob has valid fields
func (j *PipelineJob) Validate() error {
	// Validate JobID is a valid UUID
	if j.JobID == "" {
		return errors.New("job_id is required")
	}
	if _, err := uuid.Parse(j.JobID); err != nil {
		return fmt.Errorf("invalid job_id: must be a valid UUID: %w", err)
	}

	// Validate InputSource is not empty
	if j.InputSource == "" {
		return errors.New("input_source is required")
	}

	// Validate InputType matches InputSource
	if j.InputType == InputTypeHTTP {
		if !strings.HasPrefix(j.InputSource, "http://") && !strings.HasPrefix(j.InputSource, "https://") {
			return errors.New("input_source must be a valid HTTP(S) URL when input_type is http_url")
		}
		if _, err := url.Parse(j.InputSource); err != nil {
			return fmt.Errorf("invalid input_source URL: %w", err)
		}
	}

	// Validate InputType is recognized
	if !IsValidInputType(j.InputType) {
		return fmt.Errorf("invalid input_type: %s", j.InputType)
	}

	// Validate JobStatus is recognized
	if !IsValidJobStatus(j.Status) {
		return fmt.Errorf("invalid status: %s", j.Status)
	}

	// Validate CurrentStep matches one of the steps in the job
	if j.CurrentStep != "" {
		stepFound := false
		for _, step := range j.Steps {
			if string(step.Name) == j.CurrentStep {
				stepFound = true
				break
			}
		}
		if !stepFound {
			return fmt.Errorf("current_step '%s' not found in job steps", j.CurrentStep)
		}
	}

	// Validate TotalFiles and TotalBytes are non-negative
	if j.TotalFiles < 0 {
		return errors.New("total_files cannot be negative")
	}
	if j.TotalBytes < 0 {
		return errors.New("total_bytes cannot be negative")
	}

	return nil
}

// Validate checks if a PipelineStep has valid fields
func (s *PipelineStep) Validate() error {
	// Validate step name is recognized
	if !IsValidStepName(s.Name) {
		return fmt.Errorf("invalid step name: %s", s.Name)
	}

	// Validate step status is recognized
	if !IsValidStepStatus(s.Status) {
		return fmt.Errorf("invalid step status: %s", s.Status)
	}

	// Validate retry count is non-negative
	if s.RetryCount < 0 {
		return errors.New("retry_count cannot be negative")
	}

	// Validate FilesProcessed and BytesProcessed are non-negative
	if s.FilesProcessed < 0 {
		return errors.New("files_processed cannot be negative")
	}
	if s.BytesProcessed < 0 {
		return errors.New("bytes_processed cannot be negative")
	}

	// Validate StartedAt is set when status is in_progress or completed
	if (s.Status == StepStatusInProgress || s.Status == StepStatusCompleted) && s.StartedAt == nil {
		return errors.New("started_at must be set when step is in_progress or completed")
	}

	// Validate CompletedAt is set when status is completed
	if s.Status == StepStatusCompleted && s.CompletedAt == nil {
		return errors.New("completed_at must be set when step is completed")
	}

	return nil
}

// Validate checks if a FHIRDataFile has valid fields
func (f *FHIRDataFile) Validate() error {
	// Validate file name ends with .ndjson
	if !IsValidFHIRFile(f.FileName) {
		return errors.New("file_name must end with .ndjson")
	}

	// Validate file path is safe (no path traversal)
	if !IsSafePath(f.FilePath) {
		return fmt.Errorf("unsafe file_path detected: %s", f.FilePath)
	}

	// Validate file size is positive
	if f.FileSize <= 0 {
		return errors.New("file_size must be greater than 0")
	}

	// Validate line count is non-negative
	if f.LineCount < 0 {
		return errors.New("line_count cannot be negative")
	}

	// Validate source step is recognized
	if !IsValidStepName(f.SourceStep) {
		return fmt.Errorf("invalid source_step: %s", f.SourceStep)
	}

	return nil
}

// Validate checks if a ProjectConfig has valid fields
func (c *ProjectConfig) Validate() error {
	// Validate at least one pipeline step is enabled
	if len(c.Pipeline.EnabledSteps) == 0 {
		return errors.New("at least one pipeline step must be enabled")
	}

	// Validate first step is always 'import'
	if c.Pipeline.EnabledSteps[0] != StepImport {
		return errors.New("first enabled step must be 'import'")
	}

	// Validate all enabled steps are recognized
	for _, step := range c.Pipeline.EnabledSteps {
		if !IsValidStepName(step) {
			return fmt.Errorf("unrecognized step in enabled_steps: %s", step)
		}
	}

	// Validate service URLs for enabled steps
	for _, step := range c.Pipeline.EnabledSteps {
		if !c.Services.HasServiceURL(step) {
			switch step {
			case StepDIMP, StepCSVConversion, StepParquetConversion:
				return fmt.Errorf("service URL required for enabled step '%s'", step)
			}
		}
	}

	// Validate service URLs are well-formed (if provided)
	if c.Services.DIMP.URL != "" {
		if _, err := url.Parse(c.Services.DIMP.URL); err != nil {
			return fmt.Errorf("invalid dimp url: %w", err)
		}
	}
	if c.Services.CSVConversion.URL != "" {
		if _, err := url.Parse(c.Services.CSVConversion.URL); err != nil {
			return fmt.Errorf("invalid csv_conversion url: %w", err)
		}
	}
	if c.Services.ParquetConversion.URL != "" {
		if _, err := url.Parse(c.Services.ParquetConversion.URL); err != nil {
			return fmt.Errorf("invalid parquet_conversion url: %w", err)
		}
	}

	// Validate retry configuration
	if c.Retry.MaxAttempts < 1 || c.Retry.MaxAttempts > 10 {
		return errors.New("max_attempts must be between 1 and 10")
	}
	if c.Retry.InitialBackoffMs <= 0 {
		return errors.New("initial_backoff_ms must be positive")
	}
	if c.Retry.MaxBackoffMs <= 0 {
		return errors.New("max_backoff_ms must be positive")
	}
	if c.Retry.InitialBackoffMs >= c.Retry.MaxBackoffMs {
		return errors.New("initial_backoff_ms must be less than max_backoff_ms")
	}

	// Validate jobs_dir is not empty
	if c.JobsDir == "" {
		return errors.New("jobs_dir is required")
	}

	return nil
}

// ValidateJobsDir checks if the jobs directory exists and is writable
// Creates the directory automatically if it doesn't exist
func ValidateJobsDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create directory with appropriate permissions
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create jobs directory: %w", err)
			}
			// Directory created successfully, no need to check further
			return nil
		}
		return fmt.Errorf("cannot access jobs directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("jobs_dir is not a directory: %s", path)
	}

	// Check write permission by attempting to create a temp file
	testFile := fmt.Sprintf("%s/.write_test_%s", path, uuid.New().String())
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("jobs directory is not writable: %w", err)
	}
	_ = f.Close()
	_ = os.Remove(testFile)

	return nil
}

// ValidateServiceConnectivity checks if required service URLs are reachable
// This performs a lightweight HTTP HEAD request to verify connectivity
// Validates that configured service URLs are reachable
func (c *ProjectConfig) ValidateServiceConnectivity() error {
	client := &http.Client{
		Timeout: 5 * time.Second, // Quick connectivity check
	}

	// Check TORCH connectivity if base URL is configured
	// TORCH is used by InputTypeCRTDL and InputTypeTORCHURL
	if c.Services.TORCH.BaseURL != "" {
		parsedURL, err := url.Parse(c.Services.TORCH.BaseURL)
		if err != nil {
			return fmt.Errorf("invalid TORCH service URL: %w", err)
		}

		checkURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

		req, err := http.NewRequest("HEAD", checkURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request for TORCH service: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("TORCH service unreachable at %s: %w", checkURL, err)
		}
		defer func() { _ = resp.Body.Close() }()
	}

	for _, step := range c.Pipeline.EnabledSteps {
		var serviceURL string
		var serviceName string

		switch step {
		case StepDIMP:
			serviceURL = c.Services.DIMP.URL
			serviceName = "DIMP"
		case StepCSVConversion:
			serviceURL = c.Services.CSVConversion.URL
			serviceName = "CSV Conversion"
		case StepParquetConversion:
			serviceURL = c.Services.ParquetConversion.URL
			serviceName = "Parquet Conversion"
		default:
			continue // Skip steps that don't require external services
		}

		if serviceURL == "" {
			continue // Already validated in Validate()
		}

		// Parse the base URL
		parsedURL, err := url.Parse(serviceURL)
		if err != nil {
			return fmt.Errorf("invalid %s service URL: %w", serviceName, err)
		}

		// Construct a simple health check URL (just the base)
		checkURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

		// Attempt to connect
		req, err := http.NewRequest("HEAD", checkURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request for %s service: %w", serviceName, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("%s service unreachable at %s: %w", serviceName, checkURL, err)
		}
		_ = resp.Body.Close()

		// Any response (even 404) means the host is reachable
		// We're just checking connectivity, not full service health
	}

	return nil
}
