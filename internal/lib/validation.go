package lib

import (
	"github.com/trobanga/aether/internal/models"
)

// StepPrerequisites defines which steps must complete before a given step can run
var StepPrerequisites = map[models.StepName][]models.StepName{
	models.StepImport:            {}, // No prerequisites - can always run
	models.StepDIMP:              {models.StepImport},
	models.StepValidation:        {models.StepImport}, // Can validate after import (regardless of DIMP)
	models.StepCSVConversion:     {models.StepImport}, // Can convert original or pseudonymized data
	models.StepParquetConversion: {models.StepImport}, // Can convert original or pseudonymized data
}

// ValidateStepPrerequisites checks if all prerequisite steps have completed successfully
// Returns the first missing prerequisite, or empty string if all prerequisites are met
func ValidateStepPrerequisites(job models.PipelineJob, stepName models.StepName) (models.StepName, bool) {
	prerequisites, exists := StepPrerequisites[stepName]
	if !exists {
		// Unknown step - no prerequisites defined
		return "", true
	}

	// Check each prerequisite
	for _, prerequisite := range prerequisites {
		step, found := models.GetStepByName(job, prerequisite)
		if !found {
			// Prerequisite step doesn't exist in job's step list
			// This means it's not enabled in config, which is fine
			continue
		}

		// Prerequisite exists but hasn't completed
		if step.Status != models.StepStatusCompleted {
			return prerequisite, false
		}
	}

	// All prerequisites met
	return "", true
}

// CanRunStep checks if a step can be executed based on current job state
// Returns true if the step can run, false otherwise with the blocking prerequisite
func CanRunStep(job models.PipelineJob, stepName models.StepName) (bool, models.StepName) {
	prerequisite, canRun := ValidateStepPrerequisites(job, stepName)
	return canRun, prerequisite
}

// GetStepDependencies returns the list of steps that must complete before the given step
func GetStepDependencies(stepName models.StepName) []models.StepName {
	deps, exists := StepPrerequisites[stepName]
	if !exists {
		return []models.StepName{}
	}
	return deps
}
