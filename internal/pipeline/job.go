package pipeline

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/trobanga/aether/internal/models"
	"github.com/trobanga/aether/internal/services"
)

// CreateJob initializes a new pipeline job
// Returns the created job with generated UUID and initialized steps
func CreateJob(inputSource string, config models.ProjectConfig) (*models.PipelineJob, error) {
	// Generate unique job ID
	jobID := uuid.New().String()

	// Determine input type
	inputType := determineInputType(inputSource)

	// Initialize steps from config
	steps := models.InitializeSteps(config.Pipeline.EnabledSteps)

	// Create job
	job := &models.PipelineJob{
		JobID:        jobID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		InputSource:  inputSource,
		InputType:    inputType,
		CurrentStep:  string(models.StepImport), // Always start with import
		Status:       models.JobStatusPending,
		Steps:        steps,
		Config:       config,
		TotalFiles:   0,
		TotalBytes:   0,
		ErrorMessage: "",
	}

	// Validate the job
	if err := job.Validate(); err != nil {
		return nil, fmt.Errorf("failed to create valid job: %w", err)
	}

	// Create job directory structure
	if _, err := services.EnsureJobDirs(config.JobsDir, jobID); err != nil {
		return nil, fmt.Errorf("failed to create job directories: %w", err)
	}

	// Save initial job state
	if err := services.SaveJobState(config.JobsDir, job); err != nil {
		return nil, fmt.Errorf("failed to save initial job state: %w", err)
	}

	return job, nil
}

// determineInputType detects whether input is a local path or HTTP URL
func determineInputType(inputSource string) models.InputType {
	if strings.HasPrefix(inputSource, "http://") || strings.HasPrefix(inputSource, "https://") {
		return models.InputTypeHTTP
	}
	return models.InputTypeLocal
}

// LoadJob loads an existing job from disk
func LoadJob(jobsDir string, jobID string) (*models.PipelineJob, error) {
	return services.LoadJobState(jobsDir, jobID)
}

// UpdateJob updates job state on disk
// Uses pure functions to create new job instance before saving
func UpdateJob(jobsDir string, job *models.PipelineJob) error {
	job.UpdatedAt = time.Now()
	return services.SaveJobState(jobsDir, job)
}

// StartJob transitions job to in_progress status and starts first step
func StartJob(job *models.PipelineJob) *models.PipelineJob {
	// Update job status
	updatedJob := models.UpdateJobStatus(*job, models.JobStatusInProgress)

	// Start first step (should be import)
	if len(updatedJob.Steps) > 0 {
		firstStep := updatedJob.Steps[0]
		startedStep := models.StartStep(firstStep)
		updatedJob = models.ReplaceStep(updatedJob, startedStep)
	}

	return &updatedJob
}

// CompleteJob marks job as completed
func CompleteJob(job *models.PipelineJob) *models.PipelineJob {
	updatedJob := models.UpdateJobStatus(*job, models.JobStatusCompleted)
	updatedJob.CurrentStep = "" // No current step when complete
	return &updatedJob
}

// FailJob marks job as failed with error message
func FailJob(job *models.PipelineJob, errorMsg string) *models.PipelineJob {
	updatedJob := models.AddError(*job, errorMsg)
	return &updatedJob
}

// GetCurrentStep returns the current step being executed
func GetCurrentStep(job *models.PipelineJob) (models.PipelineStep, bool) {
	if job.CurrentStep == "" {
		return models.PipelineStep{}, false
	}

	stepName := models.StepName(job.CurrentStep)
	return models.GetStepByName(*job, stepName)
}

// AdvanceToNextStep moves job to the next enabled step
func AdvanceToNextStep(job *models.PipelineJob) (*models.PipelineJob, error) {
	currentStepName := models.StepName(job.CurrentStep)

	// Get next step from config
	nextStepName := job.Config.Pipeline.GetNextStep(currentStepName)

	if nextStepName == "" {
		// No more steps - job is complete
		return CompleteJob(job), nil
	}

	// Update current step
	updatedJob := models.UpdateCurrentStep(*job, nextStepName)

	// Start the next step
	nextStep, found := models.GetStepByName(updatedJob, nextStepName)
	if !found {
		return nil, fmt.Errorf("next step not found: %s", nextStepName)
	}

	startedStep := models.StartStep(nextStep)
	updatedJob = models.ReplaceStep(updatedJob, startedStep)

	return &updatedJob, nil
}

// UpdateJobProgress updates total files and bytes processed
func UpdateJobProgress(job *models.PipelineJob, files int, bytes int64) *models.PipelineJob {
	updatedJob := models.UpdateJobMetrics(*job, files, bytes)
	return &updatedJob
}

// IsJobComplete checks if all steps are completed
func IsJobComplete(job *models.PipelineJob) bool {
	return models.IsJobComplete(*job)
}

// GetJobSummary returns a human-readable summary of the job
func GetJobSummary(job *models.PipelineJob) string {
	duration := time.Since(job.CreatedAt)

	summary := fmt.Sprintf("Job %s\n", job.JobID)
	summary += fmt.Sprintf("Status: %s\n", job.Status)
	summary += fmt.Sprintf("Current Step: %s\n", job.CurrentStep)
	summary += fmt.Sprintf("Files: %d\n", job.TotalFiles)
	summary += fmt.Sprintf("Duration: %v\n", duration.Round(time.Second))

	if job.ErrorMessage != "" {
		summary += fmt.Sprintf("Error: %s\n", job.ErrorMessage)
	}

	return summary
}
