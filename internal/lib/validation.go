package lib

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

// DetectInputType determines the input source type from the input string
// Returns InputTypeLocal for directories, InputTypeHTTP for HTTP URLs,
// InputTypeTORCHURL for TORCH result URLs, InputTypeCRTDL for CRTDL files
func DetectInputType(inputSource string) (models.InputType, error) {
	if inputSource == "" {
		return "", fmt.Errorf("input source cannot be empty")
	}

	// Check if directory
	if stat, err := os.Stat(inputSource); err == nil && stat.IsDir() {
		return models.InputTypeLocal, nil
	}

	// Check if HTTP URL
	if strings.HasPrefix(inputSource, "http://") || strings.HasPrefix(inputSource, "https://") {
		// Check if TORCH result URL pattern (contains /fhir/extraction/ or /fhir/result/)
		if strings.Contains(inputSource, "/fhir/extraction/") || strings.Contains(inputSource, "/fhir/result/") {
			return models.InputTypeTORCHURL, nil
		}
		return models.InputTypeHTTP, nil
	}

	// Check if CRTDL file
	if strings.HasSuffix(inputSource, ".crtdl") || strings.HasSuffix(inputSource, ".json") {
		if IsCRTDLFile(inputSource) {
			return models.InputTypeCRTDL, nil
		}
	}

	// Default to local path (backward compatibility)
	// Validation of path existence happens later in ValidateImportSource
	return models.InputTypeLocal, nil
}

// IsCRTDLFile checks if the file at the given path is a valid CRTDL file
// by verifying it contains required cohortDefinition and dataExtraction keys
func IsCRTDLFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var crtdl map[string]interface{}
	if err := json.Unmarshal(data, &crtdl); err != nil {
		return false
	}

	// Verify required CRTDL structure
	_, hasCohort := crtdl["cohortDefinition"]
	_, hasExtraction := crtdl["dataExtraction"]
	return hasCohort && hasExtraction
}

// ValidateCRTDLSyntax validates the syntax of a CRTDL file
// Performs structural validation only - semantic validation is handled by TORCH server
func ValidateCRTDLSyntax(crtdlPath string) error {
	data, err := os.ReadFile(crtdlPath)
	if err != nil {
		return fmt.Errorf("failed to read CRTDL file: %w", err)
	}

	if len(data) == 0 {
		return fmt.Errorf("CRTDL file is empty")
	}

	var crtdl map[string]interface{}
	if err := json.Unmarshal(data, &crtdl); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Check required keys
	cohort, hasCohort := crtdl["cohortDefinition"]
	if !hasCohort {
		return fmt.Errorf("missing required key: cohortDefinition")
	}

	extraction, hasExtraction := crtdl["dataExtraction"]
	if !hasExtraction {
		return fmt.Errorf("missing required key: dataExtraction")
	}

	// Validate cohortDefinition structure
	cohortMap, ok := cohort.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cohortDefinition must be an object")
	}
	if _, hasInclusion := cohortMap["inclusionCriteria"]; !hasInclusion {
		return fmt.Errorf("cohortDefinition missing inclusionCriteria")
	}

	// Validate dataExtraction structure
	extractionMap, ok := extraction.(map[string]interface{})
	if !ok {
		return fmt.Errorf("dataExtraction must be an object")
	}
	if _, hasGroups := extractionMap["attributeGroups"]; !hasGroups {
		return fmt.Errorf("dataExtraction missing attributeGroups")
	}

	return nil
}
