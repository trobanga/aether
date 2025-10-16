package pipeline

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/trobanga/aether/internal/lib"
	"github.com/trobanga/aether/internal/models"
	"github.com/trobanga/aether/internal/services"
	"github.com/trobanga/aether/internal/ui"
)

// ExecuteDIMPStep processes FHIR resources through the DIMP pseudonymization service
// Reads from import/ directory, writes to pseudonymized/ directory
// Per spec: T057 - DIMP step orchestration
func ExecuteDIMPStep(job *models.PipelineJob, jobDir string, logger *lib.Logger) error {
	stepName := models.StepDIMP

	// Check if DIMP step is enabled
	if !isStepEnabled(job.Config, stepName) {
		logger.Info("DIMP step not enabled, skipping", "job_id", job.JobID)
		return nil
	}

	// Log step start (DEBUG level to avoid polluting progress bar display)
	logger.Debug("DIMP step starting", "job_id", job.JobID)

	// Get or create DIMP step in job
	step := getOrCreateStep(job, stepName)
	step.Status = models.StepStatusInProgress
	now := time.Now()
	step.StartedAt = &now

	// Validate DIMP service URL is configured
	if job.Config.Services.DIMPUrl == "" {
		err := fmt.Errorf("DIMP service URL not configured")
		lib.LogStepFailed(logger, string(stepName), job.JobID, err, false)
		recordStepError(step, err, models.ErrorTypeNonTransient)
		return err
	}

	// Create DIMP client
	httpClient := services.DefaultHTTPClient()
	dimpClient := services.NewDIMPClient(job.Config.Services.DIMPUrl, httpClient, logger)

	// Setup directories
	importDir := filepath.Join(jobDir, "import")
	outputDir := filepath.Join(jobDir, "pseudonymized")

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		lib.LogStepFailed(logger, string(stepName), job.JobID, err, false)
		recordStepError(step, err, models.ErrorTypeNonTransient)
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Find all NDJSON files in import directory
	files, err := filepath.Glob(filepath.Join(importDir, "*.ndjson"))
	if err != nil {
		lib.LogStepFailed(logger, string(stepName), job.JobID, err, false)
		recordStepError(step, err, models.ErrorTypeNonTransient)
		return fmt.Errorf("failed to list import files: %w", err)
	}

	if len(files) == 0 {
		err := fmt.Errorf("no FHIR NDJSON files found in import directory")
		lib.LogStepFailed(logger, string(stepName), job.JobID, err, false)
		recordStepError(step, err, models.ErrorTypeNonTransient)
		return err
	}

	// Print user-friendly message instead of logger (logger pollutes progress bar)
	fmt.Printf("Processing %d FHIR file(s) through DIMP...\n\n", len(files))

	// Clean up any stale .part files from previous interrupted runs
	partFiles, _ := filepath.Glob(filepath.Join(outputDir, "*.part"))
	for _, partFile := range partFiles {
		logger.Debug("Removing stale partial file from previous run", "file", filepath.Base(partFile))
		_ = os.Remove(partFile)
	}

	// Process each file
	totalResourcesProcessed := 0
	filesProcessed := 0
	for fileIdx, inputFile := range files {
		// Create output filename: dimped_<original-filename>
		baseName := filepath.Base(inputFile)
		outputFile := filepath.Join(outputDir, "dimped_"+baseName)

		// Check if output file already exists (resume support)
		if _, err := os.Stat(outputFile); err == nil {
			// Output file exists - skip processing
			fmt.Printf("  ⊙ %s (already processed, skipping)\n", baseName)
			logger.Debug("Skipping already processed file",
				"filename", baseName,
				"output_file", outputFile,
				"job_id", job.JobID)
			filesProcessed++

			// Count resources in existing file for accurate totals
			if lineCount, err := lib.CountResourcesInFile(outputFile); err == nil {
				totalResourcesProcessed += lineCount
			}
			continue
		}

		// Process file through DIMP using atomic write (writes to .part first)
		resourcesProcessed, err := processDIMPFile(inputFile, outputFile, dimpClient, logger)
		if err != nil {
			logger.Error("Failed to process FHIR file",
				"filename", baseName,
				"file_number", fileIdx+1,
				"total_files", len(files),
				"resources_processed_so_far", totalResourcesProcessed,
				"error", err,
				"job_id", job.JobID)
			lib.LogStepFailed(logger, string(stepName), job.JobID, err, isDIMPErrorRetryable(err))
			recordStepError(step, err, classifyDIMPError(err))
			return fmt.Errorf("failed to process %s: %w", baseName, err)
		}

		// Log completion for this file
		fmt.Printf("  ✓ %s (%d resources)\n", baseName, resourcesProcessed)

		totalResourcesProcessed += resourcesProcessed
		filesProcessed++
	}

	// Update step status
	step.Status = models.StepStatusCompleted
	step.FilesProcessed = len(files)
	completedAt := time.Now()
	step.CompletedAt = &completedAt

	duration := completedAt.Sub(*step.StartedAt)

	// Log to structured logger at DEBUG level
	logger.Debug("DIMP step completed",
		"files_processed", len(files),
		"resources_processed", totalResourcesProcessed,
		"duration", duration,
		"job_id", job.JobID,
	)

	return nil
}

// processDIMPFile processes a single NDJSON file through DIMP
// Returns the number of resources processed
// Uses atomic write pattern: writes to .part file, renames on success
func processDIMPFile(inputFile, outputFile string, dimpClient *services.DIMPClient, logger *lib.Logger) (int, error) {
	// Open input file
	inFile, err := os.Open(inputFile)
	if err != nil {
		return 0, fmt.Errorf("failed to open input file: %w", err)
	}
	defer func() { _ = inFile.Close() }()

	// Create temporary output file with .part extension
	tempOutputFile := outputFile + ".part"
	outFile, err := os.Create(tempOutputFile)
	if err != nil {
		return 0, fmt.Errorf("failed to create temporary output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	// Clean up .part file on any error (will be overridden if rename succeeds)
	var success bool
	defer func() {
		if !success {
			_ = os.Remove(tempOutputFile)
		}
	}()

	// Count resources for progress tracking
	totalResources := countResourcesInFile(inputFile)

	// Create progress bar if we know total count
	var progressBar *ui.ProgressBar
	if totalResources > 0 {
		progressBar = ui.NewProgressBar(int64(totalResources), fmt.Sprintf("Pseudonymizing %s", filepath.Base(inputFile)))
	} else {
		// Use spinner for unknown count
		logger.Info("Processing FHIR resources (unknown count)", "file", filepath.Base(inputFile))
	}

	// Process line by line
	scanner := bufio.NewScanner(inFile)
	// Increase buffer to 10MB to handle large FHIR resources (default is 64KB)
	// This prevents "token too long" errors when processing large Bundles or resources
	buf := make([]byte, 0, 10*1024*1024)
	scanner.Buffer(buf, 10*1024*1024)
	resourcesProcessed := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse FHIR resource
		var resource map[string]interface{}
		if err := json.Unmarshal([]byte(line), &resource); err != nil {
			// Clear progress bar before logging error
			if progressBar != nil {
				_ = progressBar.Clear()
			}
			logger.Error("Failed to parse FHIR resource",
				"file", filepath.Base(inputFile),
				"line_number", resourcesProcessed+1,
				"error", err)
			return resourcesProcessed, fmt.Errorf("failed to parse resource at line %d: %w", resourcesProcessed+1, err)
		}

		resourceType, _ := resource["resourceType"].(string)
		resourceID, _ := resource["id"].(string)

		// Only log individual resources at DEBUG level to avoid interfering with progress bar
		logger.Debug("Processing FHIR resource",
			"file", filepath.Base(inputFile),
			"line_number", resourcesProcessed+1,
			"resourceType", resourceType,
			"id", resourceID)

		// Send to DIMP for pseudonymization
		pseudonymized, err := dimpClient.Pseudonymize(resource)
		if err != nil {
			// Clear progress bar before logging error
			if progressBar != nil {
				_ = progressBar.Clear()
			}
			logger.Error("Failed to pseudonymize FHIR resource",
				"file", filepath.Base(inputFile),
				"line_number", resourcesProcessed+1,
				"resourceType", resourceType,
				"id", resourceID,
				"error", err)

			// Print user-friendly error message
			fmt.Printf("\n✗ DIMP pseudonymization failed\n")
			fmt.Printf("  File: %s (line %d)\n", filepath.Base(inputFile), resourcesProcessed+1)
			fmt.Printf("  Resource: %s/%s\n", resourceType, resourceID)
			fmt.Printf("  Error: %v\n\n", err)

			return resourcesProcessed, fmt.Errorf("failed to pseudonymize resource at line %d: %w", resourcesProcessed+1, err)
		}

		// Write pseudonymized resource to output
		pseudonymizedJSON, err := json.Marshal(pseudonymized)
		if err != nil {
			return resourcesProcessed, fmt.Errorf("failed to marshal pseudonymized resource: %w", err)
		}

		if _, err := outFile.Write(pseudonymizedJSON); err != nil {
			return resourcesProcessed, fmt.Errorf("failed to write output: %w", err)
		}
		if _, err := outFile.Write([]byte("\n")); err != nil {
			return resourcesProcessed, fmt.Errorf("failed to write newline: %w", err)
		}

		resourcesProcessed++

		// Update progress
		if progressBar != nil {
			_ = progressBar.Add(1)
		}
	}

	if err := scanner.Err(); err != nil {
		return resourcesProcessed, fmt.Errorf("error reading file: %w", err)
	}

	// Finish progress bar
	if progressBar != nil {
		_ = progressBar.Finish()
	}

	// Close output file before rename (defer will also close, but explicit close ensures flush)
	if err := outFile.Close(); err != nil {
		return resourcesProcessed, fmt.Errorf("failed to close output file: %w", err)
	}

	// Atomic rename: move .part file to final filename
	// This ensures the output file only exists when complete
	if err := os.Rename(tempOutputFile, outputFile); err != nil {
		return resourcesProcessed, fmt.Errorf("failed to rename temporary file: %w", err)
	}

	// Mark as successful - prevents cleanup defer from removing the file
	success = true

	return resourcesProcessed, nil
}

// countResourcesInFile counts the number of non-empty lines in an NDJSON file
func countResourcesInFile(filename string) int {
	file, err := os.Open(filename)
	if err != nil {
		return 0
	}
	defer func() { _ = file.Close() }()

	count := 0
	scanner := bufio.NewScanner(file)
	// Increase buffer to 10MB to handle large FHIR resources (default is 64KB)
	buf := make([]byte, 0, 10*1024*1024)
	scanner.Buffer(buf, 10*1024*1024)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}

	return count
}

// isDIMPErrorRetryable checks if a DIMP error should be retried
func isDIMPErrorRetryable(err error) bool {
	if dimpErr, ok := err.(*services.DIMPError); ok {
		return dimpErr.IsRetryable()
	}
	// Network errors are retryable
	return lib.IsNetworkError(err)
}

// classifyDIMPError classifies a DIMP error as transient or non-transient
func classifyDIMPError(err error) models.ErrorType {
	if dimpErr, ok := err.(*services.DIMPError); ok {
		return dimpErr.ErrorType
	}
	// Network errors are transient
	if lib.IsNetworkError(err) {
		return models.ErrorTypeTransient
	}
	return models.ErrorTypeNonTransient
}

// Helper functions

func isStepEnabled(config models.ProjectConfig, stepName models.StepName) bool {
	for _, enabled := range config.Pipeline.EnabledSteps {
		if enabled == stepName {
			return true
		}
	}
	return false
}

func getOrCreateStep(job *models.PipelineJob, stepName models.StepName) *models.PipelineStep {
	for i := range job.Steps {
		if job.Steps[i].Name == stepName {
			return &job.Steps[i]
		}
	}

	// Create new step
	step := models.PipelineStep{
		Name:   stepName,
		Status: models.StepStatusPending,
	}
	job.Steps = append(job.Steps, step)
	return &job.Steps[len(job.Steps)-1]
}

func recordStepError(step *models.PipelineStep, err error, errorType models.ErrorType) {
	step.Status = models.StepStatusFailed
	step.LastError = &models.StepError{
		Type:      errorType,
		Message:   err.Error(),
		Timestamp: time.Now(),
	}
}
