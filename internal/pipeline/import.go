package pipeline

import (
	"fmt"
	"time"

	"github.com/trobanga/aether/internal/lib"
	"github.com/trobanga/aether/internal/models"
	"github.com/trobanga/aether/internal/services"
)

// ExecuteImportStep performs the import step of the pipeline
// Detects input type (local vs HTTP) and delegates to appropriate importer
// Updates job state with progress and imported files
func ExecuteImportStep(job *models.PipelineJob, logger *lib.Logger, httpClient *services.HTTPClient, showProgress bool) (*models.PipelineJob, error) {
	startTime := time.Now()

	lib.LogStepStart(logger, string(models.StepImport), job.JobID)

	// Get import output directory
	importDir := services.GetJobOutputDir(job.Config.JobsDir, job.JobID, models.StepImport)

	// Validate input source
	if err := services.ValidateImportSource(job.InputSource, job.InputType); err != nil {
		// Failed with non-transient error
		updatedJob := failImportStep(job, err, models.ErrorTypeNonTransient, 0)
		lib.LogStepFailed(logger, string(models.StepImport), job.JobID, err, false)
		return &updatedJob, err
	}

	// Execute import based on input type
	var importedFiles []models.FHIRDataFile
	var err error

	switch job.InputType {
	case models.InputTypeLocal:
		logger.Info("Importing from local directory", "source", job.InputSource)
		importedFiles, err = services.ImportFromLocalDirectory(job.InputSource, importDir, logger)

	case models.InputTypeHTTP:
		logger.Info("Downloading from URL", "source", job.InputSource)
		if showProgress {
			importedFiles, err = services.DownloadFromURLWithProgress(job.InputSource, importDir, httpClient, logger)
		} else {
			importedFiles, err = services.DownloadFromURL(job.InputSource, importDir, httpClient, logger, false)
		}

	default:
		err = fmt.Errorf("unsupported input type: %s", job.InputType)
	}

	// Handle errors
	if err != nil {
		// Classify error type
		errorType := classifyImportError(err, job.InputType)
		updatedJob := failImportStep(job, err, errorType, 0)
		lib.LogStepFailed(logger, string(models.StepImport), job.JobID, err, errorType == models.ErrorTypeTransient)
		return &updatedJob, err
	}

	// Calculate total bytes imported
	var totalBytes int64
	for _, file := range importedFiles {
		totalBytes += file.FileSize
	}

	// Update job with imported file metrics
	updatedJob := models.UpdateJobMetrics(*job, len(importedFiles), totalBytes)

	// Complete the import step
	importStep, _ := models.GetStepByName(updatedJob, models.StepImport)
	completedStep := models.CompleteStep(importStep, len(importedFiles), totalBytes)
	updatedJob = models.ReplaceStep(updatedJob, completedStep)

	duration := time.Since(startTime)
	lib.LogStepComplete(logger, string(models.StepImport), job.JobID, len(importedFiles), duration)

	return &updatedJob, nil
}

// failImportStep marks the import step as failed
func failImportStep(job *models.PipelineJob, err error, errorType models.ErrorType, httpStatus int) models.PipelineJob {
	importStep, found := models.GetStepByName(*job, models.StepImport)
	if !found {
		// Step not found - shouldn't happen, but handle gracefully
		return models.AddError(*job, err.Error())
	}

	failedStep := models.FailStep(importStep, errorType, err.Error(), httpStatus)
	updatedJob := models.ReplaceStep(*job, failedStep)
	updatedJob = models.AddError(updatedJob, err.Error())

	return updatedJob
}

// classifyImportError determines if an import error is transient or non-transient
func classifyImportError(err error, inputType models.InputType) models.ErrorType {
	if err == nil {
		return models.ErrorTypeNonTransient
	}

	// For HTTP downloads, network errors are transient
	if inputType == models.InputTypeHTTP {
		if lib.IsNetworkError(err) {
			return models.ErrorTypeTransient
		}
	}

	// For local imports, most errors are non-transient (file not found, permissions, etc.)
	// Default to non-transient
	return models.ErrorTypeNonTransient
}

// RetryImportStep attempts to retry a failed import step
// Should only be called if the error was transient
func RetryImportStep(job *models.PipelineJob, logger *lib.Logger, httpClient *services.HTTPClient, showProgress bool) (*models.PipelineJob, error) {
	// Get current import step
	importStep, found := models.GetStepByName(*job, models.StepImport)
	if !found {
		return nil, fmt.Errorf("import step not found")
	}

	// Check if retry is allowed
	if importStep.LastError == nil {
		return nil, fmt.Errorf("no error to retry")
	}

	if !lib.ShouldRetry(importStep.LastError.Type, importStep.RetryCount, job.Config.Retry.MaxAttempts) {
		return nil, fmt.Errorf("retry not allowed: max attempts reached or non-transient error")
	}

	// Increment retry count
	retriedStep := models.IncrementRetry(importStep)
	updatedJob := models.ReplaceStep(*job, retriedStep)

	lib.LogRetry(logger, "import step", retriedStep.RetryCount, job.Config.Retry.MaxAttempts, importStep.LastError)

	// Calculate backoff
	backoff := lib.CalculateBackoff(retriedStep.RetryCount-1, job.Config.Retry.InitialBackoffMs, job.Config.Retry.MaxBackoffMs)
	logger.Info("Waiting before retry", "backoff", backoff)
	time.Sleep(backoff)

	// Retry the import
	return ExecuteImportStep(&updatedJob, logger, httpClient, showProgress)
}
