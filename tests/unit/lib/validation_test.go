package lib

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trobanga/aether/internal/lib"
	"github.com/trobanga/aether/internal/models"
)

func TestValidateStepPrerequisites_ImportHasNone(t *testing.T) {
	job := createTestJob([]models.StepName{models.StepImport})

	prerequisite, canRun := lib.ValidateStepPrerequisites(job, models.StepImport)

	assert.True(t, canRun)
	assert.Equal(t, models.StepName(""), prerequisite)
}

func TestValidateStepPrerequisites_DIMPRequiresImport(t *testing.T) {
	job := createTestJob([]models.StepName{models.StepImport, models.StepDIMP})

	// Import not completed yet
	prerequisite, canRun := lib.ValidateStepPrerequisites(job, models.StepDIMP)

	assert.False(t, canRun)
	assert.Equal(t, models.StepImport, prerequisite)
}

func TestValidateStepPrerequisites_DIMPAllowedAfterImport(t *testing.T) {
	job := createTestJob([]models.StepName{models.StepImport, models.StepDIMP})

	// Mark import as completed
	importStep, _ := models.GetStepByName(job, models.StepImport)
	importStep = models.CompleteStep(importStep, 10, 1000)
	job = models.ReplaceStep(job, importStep)

	// Now DIMP should be allowed
	prerequisite, canRun := lib.ValidateStepPrerequisites(job, models.StepDIMP)

	assert.True(t, canRun)
	assert.Equal(t, models.StepName(""), prerequisite)
}

func TestValidateStepPrerequisites_CSVRequiresImport(t *testing.T) {
	job := createTestJob([]models.StepName{models.StepImport, models.StepCSVConversion})

	// Import not completed
	prerequisite, canRun := lib.ValidateStepPrerequisites(job, models.StepCSVConversion)

	assert.False(t, canRun)
	assert.Equal(t, models.StepImport, prerequisite)
}

func TestValidateStepPrerequisites_CSVAllowedAfterImport(t *testing.T) {
	job := createTestJob([]models.StepName{models.StepImport, models.StepCSVConversion})

	// Mark import as completed
	importStep, _ := models.GetStepByName(job, models.StepImport)
	importStep = models.CompleteStep(importStep, 5, 500)
	job = models.ReplaceStep(job, importStep)

	// CSV conversion should be allowed
	prerequisite, canRun := lib.ValidateStepPrerequisites(job, models.StepCSVConversion)

	assert.True(t, canRun)
	assert.Equal(t, models.StepName(""), prerequisite)
}

func TestValidateStepPrerequisites_ValidationRequiresImport(t *testing.T) {
	job := createTestJob([]models.StepName{models.StepImport, models.StepValidation})

	// Import not completed
	prerequisite, canRun := lib.ValidateStepPrerequisites(job, models.StepValidation)

	assert.False(t, canRun)
	assert.Equal(t, models.StepImport, prerequisite)
}

func TestValidateStepPrerequisites_PrerequisiteNotEnabled(t *testing.T) {
	// Job without import step (unusual but possible)
	job := createTestJob([]models.StepName{models.StepDIMP})

	// DIMP requires import, but import is not in the step list
	// Validation should allow this (trust config - if import is disabled, DIMP has no prerequisites)
	prerequisite, canRun := lib.ValidateStepPrerequisites(job, models.StepDIMP)

	assert.True(t, canRun) // Prerequisite not enabled means it doesn't block
	assert.Equal(t, models.StepName(""), prerequisite)
}

func TestValidateStepPrerequisites_UnknownStep(t *testing.T) {
	job := createTestJob([]models.StepName{models.StepImport})

	// Unknown step name
	prerequisite, canRun := lib.ValidateStepPrerequisites(job, models.StepName("unknown"))

	assert.True(t, canRun) // Unknown step has no prerequisites defined
	assert.Equal(t, models.StepName(""), prerequisite)
}

func TestCanRunStep_Wrapper(t *testing.T) {
	job := createTestJob([]models.StepName{models.StepImport, models.StepDIMP})

	// Import not completed
	canRun, prerequisite := lib.CanRunStep(job, models.StepDIMP)

	assert.False(t, canRun)
	assert.Equal(t, models.StepImport, prerequisite)
}

func TestGetStepDependencies_Import(t *testing.T) {
	deps := lib.GetStepDependencies(models.StepImport)
	assert.Empty(t, deps)
}

func TestGetStepDependencies_DIMP(t *testing.T) {
	deps := lib.GetStepDependencies(models.StepDIMP)
	require.Len(t, deps, 1)
	assert.Equal(t, models.StepImport, deps[0])
}

func TestGetStepDependencies_CSV(t *testing.T) {
	deps := lib.GetStepDependencies(models.StepCSVConversion)
	require.Len(t, deps, 1)
	assert.Equal(t, models.StepImport, deps[0])
}

func TestGetStepDependencies_Parquet(t *testing.T) {
	deps := lib.GetStepDependencies(models.StepParquetConversion)
	require.Len(t, deps, 1)
	assert.Equal(t, models.StepImport, deps[0])
}

func TestGetStepDependencies_UnknownStep(t *testing.T) {
	deps := lib.GetStepDependencies(models.StepName("nonexistent"))
	assert.Empty(t, deps)
}

// Integration test: Full pipeline flow
func TestValidation_PipelineSequence(t *testing.T) {
	// Full pipeline: Import -> DIMP -> CSV -> Parquet
	job := createTestJob([]models.StepName{
		models.StepImport,
		models.StepDIMP,
		models.StepCSVConversion,
		models.StepParquetConversion,
	})

	// Initially, only import can run
	canRun, _ := lib.CanRunStep(job, models.StepImport)
	assert.True(t, canRun)

	canRun, prereq := lib.CanRunStep(job, models.StepDIMP)
	assert.False(t, canRun)
	assert.Equal(t, models.StepImport, prereq)

	canRun, prereq = lib.CanRunStep(job, models.StepCSVConversion)
	assert.False(t, canRun)
	assert.Equal(t, models.StepImport, prereq)

	// Complete import
	importStep, _ := models.GetStepByName(job, models.StepImport)
	importStep = models.CompleteStep(importStep, 10, 1000)
	job = models.ReplaceStep(job, importStep)

	// Now DIMP and CSV can run
	canRun, _ = lib.CanRunStep(job, models.StepDIMP)
	assert.True(t, canRun)

	canRun, _ = lib.CanRunStep(job, models.StepCSVConversion)
	assert.True(t, canRun)

	canRun, _ = lib.CanRunStep(job, models.StepParquetConversion)
	assert.True(t, canRun)

	// Complete DIMP (CSV and Parquet should still be allowed)
	dimpStep, _ := models.GetStepByName(job, models.StepDIMP)
	dimpStep = models.CompleteStep(dimpStep, 10, 900)
	job = models.ReplaceStep(job, dimpStep)

	canRun, _ = lib.CanRunStep(job, models.StepCSVConversion)
	assert.True(t, canRun) // Still allowed (doesn't depend on DIMP)

	canRun, _ = lib.CanRunStep(job, models.StepParquetConversion)
	assert.True(t, canRun)
}

// Helper: Create a test job with given steps
func createTestJob(steps []models.StepName) models.PipelineJob {
	now := time.Now()
	return models.PipelineJob{
		JobID:       "test-job-123",
		CreatedAt:   now,
		UpdatedAt:   now,
		InputSource: "/test/input",
		InputType:   models.InputTypeLocal,
		CurrentStep: string(steps[0]),
		Status:      models.JobStatusInProgress,
		Steps:       models.InitializeSteps(steps),
		Config: models.ProjectConfig{
			Pipeline: models.PipelineConfig{
				EnabledSteps: steps,
			},
		},
	}
}
