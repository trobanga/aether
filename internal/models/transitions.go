package models

import "time"

// UpdateJobStatus creates a new PipelineJob with updated status
// Pure function - returns new instance, does not mutate original
func UpdateJobStatus(job PipelineJob, status JobStatus) PipelineJob {
	job.Status = status
	job.UpdatedAt = time.Now()
	return job
}

// UpdateCurrentStep creates a new PipelineJob with updated current step
// Pure function - returns new instance
func UpdateCurrentStep(job PipelineJob, step StepName) PipelineJob {
	job.CurrentStep = string(step)
	job.UpdatedAt = time.Now()
	return job
}

// AddError creates a new PipelineJob with error message
// Pure function - returns new instance
func AddError(job PipelineJob, errorMsg string) PipelineJob {
	job.ErrorMessage = errorMsg
	job.Status = JobStatusFailed
	job.UpdatedAt = time.Now()
	return job
}

// UpdateJobMetrics creates a new PipelineJob with updated file/byte counts
// Pure function - returns new instance
func UpdateJobMetrics(job PipelineJob, files int, bytes int64) PipelineJob {
	job.TotalFiles = files
	job.TotalBytes = bytes
	job.UpdatedAt = time.Now()
	return job
}

// StartStep creates a new PipelineStep with in_progress status
// Pure function - returns new instance
func StartStep(step PipelineStep) PipelineStep {
	now := time.Now()
	step.Status = StepStatusInProgress
	step.StartedAt = &now
	return step
}

// CompleteStep creates a new PipelineStep with completed status
// Pure function - returns new instance
func CompleteStep(step PipelineStep, filesProcessed int, bytesProcessed int64) PipelineStep {
	now := time.Now()
	step.Status = StepStatusCompleted
	step.CompletedAt = &now
	step.FilesProcessed = filesProcessed
	step.BytesProcessed = bytesProcessed
	return step
}

// FailStep creates a new PipelineStep with failed status and error details
// Pure function - returns new instance
func FailStep(step PipelineStep, errorType ErrorType, errorMsg string, httpStatus int) PipelineStep {
	step.Status = StepStatusFailed
	step.LastError = &StepError{
		Type:       errorType,
		Message:    errorMsg,
		HTTPStatus: httpStatus,
		Timestamp:  time.Now(),
	}
	return step
}

// IncrementRetry creates a new PipelineStep with incremented retry count
// Pure function - returns new instance
func IncrementRetry(step PipelineStep) PipelineStep {
	step.RetryCount++
	return step
}

// UpdateStepProgress creates a new PipelineStep with updated progress metrics
// Pure function - returns new instance
func UpdateStepProgress(step PipelineStep, filesProcessed int, bytesProcessed int64) PipelineStep {
	step.FilesProcessed = filesProcessed
	step.BytesProcessed = bytesProcessed
	return step
}

// ReplaceStep replaces a step in the job's step list
// Pure function - returns new job instance with updated steps
func ReplaceStep(job PipelineJob, updatedStep PipelineStep) PipelineJob {
	newSteps := make([]PipelineStep, len(job.Steps))
	copy(newSteps, job.Steps)

	for i, step := range newSteps {
		if step.Name == updatedStep.Name {
			newSteps[i] = updatedStep
			break
		}
	}

	job.Steps = newSteps
	job.UpdatedAt = time.Now()
	return job
}

// InitializeSteps creates initial step list from enabled steps config
// Pure function - creates new step instances
func InitializeSteps(enabledSteps []StepName) []PipelineStep {
	steps := make([]PipelineStep, len(enabledSteps))
	for i, stepName := range enabledSteps {
		steps[i] = PipelineStep{
			Name:           stepName,
			Status:         StepStatusPending,
			StartedAt:      nil,
			CompletedAt:    nil,
			FilesProcessed: 0,
			BytesProcessed: 0,
			RetryCount:     0,
			LastError:      nil,
		}
	}
	return steps
}

// GetStepByName finds a step by name in the job's step list
// Pure function - returns copy of step if found
func GetStepByName(job PipelineJob, name StepName) (PipelineStep, bool) {
	for _, step := range job.Steps {
		if step.Name == name {
			return step, true
		}
	}
	return PipelineStep{}, false
}

// IsJobComplete checks if all steps are completed
// Pure function - no mutations
func IsJobComplete(job PipelineJob) bool {
	if len(job.Steps) == 0 {
		return false
	}

	for _, step := range job.Steps {
		if step.Status != StepStatusCompleted {
			return false
		}
	}
	return true
}

// GetNextPendingStep finds the first pending step in the job
// Pure function - returns copy of step if found
func GetNextPendingStep(job PipelineJob) (PipelineStep, bool) {
	for _, step := range job.Steps {
		if step.Status == StepStatusPending {
			return step, true
		}
	}
	return PipelineStep{}, false
}
