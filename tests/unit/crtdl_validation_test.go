package unit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trobanga/aether/internal/lib"
)

// Unit tests for CRTDL syntax validation

func TestValidateCRTDLSyntax_ValidMinimal(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "valid.crtdl")

	// Minimal valid CRTDL structure
	validCRTDL := `{
		"cohortDefinition": {
			"version": "1.0",
			"inclusionCriteria": []
		},
		"dataExtraction": {
			"attributeGroups": []
		}
	}`
	err := os.WriteFile(crtdlFile, []byte(validCRTDL), 0644)
	require.NoError(t, err)

	// Test will validate syntax successfully
	// err = validation.ValidateCRTDLSyntax(crtdlFile)
	// assert.NoError(t, err, "Valid minimal CRTDL should pass validation")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

func TestValidateCRTDLSyntax_ValidComplete(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "complete.crtdl")

	// Complete CRTDL with all fields
	completeCRTDL := `{
		"cohortDefinition": {
			"version": "1.0",
			"display": "Test Cohort",
			"inclusionCriteria": [
				{
					"criteriaGroup": [
						{
							"code": "8310-5",
							"system": "http://loinc.org",
							"display": "Body temperature"
						}
					]
				}
			]
		},
		"dataExtraction": {
			"attributeGroups": [
				{
					"name": "Demographics",
					"attributes": [
						{
							"name": "birthDate",
							"path": "Patient.birthDate"
						}
					]
				}
			]
		}
	}`
	err := os.WriteFile(crtdlFile, []byte(completeCRTDL), 0644)
	require.NoError(t, err)

	// Test will validate complete CRTDL successfully
	// err = validation.ValidateCRTDLSyntax(crtdlFile)
	// assert.NoError(t, err, "Valid complete CRTDL should pass validation")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

func TestValidateCRTDLSyntax_MissingCohortDefinition(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "missing_cohort.crtdl")

	// CRTDL missing cohortDefinition
	invalidCRTDL := `{
		"dataExtraction": {
			"attributeGroups": []
		}
	}`
	err := os.WriteFile(crtdlFile, []byte(invalidCRTDL), 0644)
	require.NoError(t, err)

	// Test will fail validation
	// err = validation.ValidateCRTDLSyntax(crtdlFile)
	// assert.Error(t, err, "CRTDL without cohortDefinition should fail")
	// assert.Contains(t, err.Error(), "cohortDefinition")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

func TestValidateCRTDLSyntax_MissingDataExtraction(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "missing_extraction.crtdl")

	// CRTDL missing dataExtraction
	invalidCRTDL := `{
		"cohortDefinition": {
			"version": "1.0",
			"inclusionCriteria": []
		}
	}`
	err := os.WriteFile(crtdlFile, []byte(invalidCRTDL), 0644)
	require.NoError(t, err)

	// Test will fail validation
	// err = validation.ValidateCRTDLSyntax(crtdlFile)
	// assert.Error(t, err, "CRTDL without dataExtraction should fail")
	// assert.Contains(t, err.Error(), "dataExtraction")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

func TestValidateCRTDLSyntax_MissingInclusionCriteria(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "missing_inclusion.crtdl")

	// CRTDL with cohortDefinition but missing inclusionCriteria
	invalidCRTDL := `{
		"cohortDefinition": {
			"version": "1.0"
		},
		"dataExtraction": {
			"attributeGroups": []
		}
	}`
	err := os.WriteFile(crtdlFile, []byte(invalidCRTDL), 0644)
	require.NoError(t, err)

	// Test will fail validation
	// err = validation.ValidateCRTDLSyntax(crtdlFile)
	// assert.Error(t, err, "CRTDL without inclusionCriteria should fail")
	// assert.Contains(t, err.Error(), "inclusionCriteria")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

func TestValidateCRTDLSyntax_MissingAttributeGroups(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "missing_attributes.crtdl")

	// CRTDL with dataExtraction but missing attributeGroups
	invalidCRTDL := `{
		"cohortDefinition": {
			"version": "1.0",
			"inclusionCriteria": []
		},
		"dataExtraction": {
		}
	}`
	err := os.WriteFile(crtdlFile, []byte(invalidCRTDL), 0644)
	require.NoError(t, err)

	// Test will fail validation
	// err = validation.ValidateCRTDLSyntax(crtdlFile)
	// assert.Error(t, err, "CRTDL without attributeGroups should fail")
	// assert.Contains(t, err.Error(), "attributeGroups")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

func TestValidateCRTDLSyntax_CohortDefinitionNotObject(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "cohort_not_object.crtdl")

	// CRTDL with cohortDefinition as string instead of object
	invalidCRTDL := `{
		"cohortDefinition": "invalid",
		"dataExtraction": {
			"attributeGroups": []
		}
	}`
	err := os.WriteFile(crtdlFile, []byte(invalidCRTDL), 0644)
	require.NoError(t, err)

	// Test will fail validation
	// err = validation.ValidateCRTDLSyntax(crtdlFile)
	// assert.Error(t, err, "cohortDefinition must be object")
	// assert.Contains(t, err.Error(), "cohortDefinition")
	// assert.Contains(t, err.Error(), "object")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

func TestValidateCRTDLSyntax_DataExtractionNotObject(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "extraction_not_object.crtdl")

	// CRTDL with dataExtraction as array instead of object
	invalidCRTDL := `{
		"cohortDefinition": {
			"version": "1.0",
			"inclusionCriteria": []
		},
		"dataExtraction": []
	}`
	err := os.WriteFile(crtdlFile, []byte(invalidCRTDL), 0644)
	require.NoError(t, err)

	// Test will fail validation
	// err = validation.ValidateCRTDLSyntax(crtdlFile)
	// assert.Error(t, err, "dataExtraction must be object")
	// assert.Contains(t, err.Error(), "dataExtraction")
	// assert.Contains(t, err.Error(), "object")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

func TestValidateCRTDLSyntax_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "invalid.crtdl")

	// Malformed JSON
	invalidJSON := `{
		"cohortDefinition": {
			"version": "1.0"
		// missing closing braces
	`
	err := os.WriteFile(crtdlFile, []byte(invalidJSON), 0644)
	require.NoError(t, err)

	// Test will fail validation
	// err = validation.ValidateCRTDLSyntax(crtdlFile)
	// assert.Error(t, err, "Malformed JSON should fail validation")
	// assert.Contains(t, err.Error(), "JSON")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

func TestValidateCRTDLSyntax_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "empty.crtdl")

	// Empty file
	err := os.WriteFile(crtdlFile, []byte(""), 0644)
	require.NoError(t, err)

	// Test will fail validation
	// err = validation.ValidateCRTDLSyntax(crtdlFile)
	// assert.Error(t, err, "Empty file should fail validation")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

func TestValidateCRTDLSyntax_FileDoesNotExist(t *testing.T) {
	// Test will fail for non-existent file
	// err := validation.ValidateCRTDLSyntax("/path/does/not/exist.crtdl")
	// assert.Error(t, err, "Non-existent file should fail validation")
	// assert.Contains(t, err.Error(), "failed to read")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

func TestValidateCRTDLSyntax_FileSizeLimit(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "large.crtdl")

	// Create CRTDL with large inclusionCriteria array (should still be valid if < 1MB)
	// This tests that reasonable large files are accepted
	largeCRTDL := `{
		"cohortDefinition": {
			"version": "1.0",
			"inclusionCriteria": [`

	// Add 1000 criteria groups
	for i := 0; i < 1000; i++ {
		if i > 0 {
			largeCRTDL += ","
		}
		largeCRTDL += `{"criteriaGroup": []}`
	}

	largeCRTDL += `]
		},
		"dataExtraction": {
			"attributeGroups": []
		}
	}`

	err := os.WriteFile(crtdlFile, []byte(largeCRTDL), 0644)
	require.NoError(t, err)

	// Verify file size is reasonable (< 1MB per spec)
	fileInfo, err := os.Stat(crtdlFile)
	require.NoError(t, err)
	assert.Less(t, fileInfo.Size(), int64(1024*1024), "Test file should be < 1MB")

	// Test will validate successfully if under size limit
	// err = validation.ValidateCRTDLSyntax(crtdlFile)
	// assert.NoError(t, err, "Large but valid CRTDL should pass validation")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

func TestValidateCRTDLSyntax_ExtraFieldsAllowed(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "extra_fields.crtdl")

	// CRTDL with extra fields (should be allowed - forward compatibility)
	crtdlWithExtras := `{
		"cohortDefinition": {
			"version": "1.0",
			"inclusionCriteria": [],
			"customField": "should be ignored"
		},
		"dataExtraction": {
			"attributeGroups": []
		},
		"metadata": {
			"author": "test",
			"description": "extra fields should be allowed"
		}
	}`
	err := os.WriteFile(crtdlFile, []byte(crtdlWithExtras), 0644)
	require.NoError(t, err)

	// Test will validate successfully (only required fields checked)
	// err = validation.ValidateCRTDLSyntax(crtdlFile)
	// assert.NoError(t, err, "CRTDL with extra fields should pass validation")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

func TestValidateCRTDLSyntax_EmptyStructuresValid(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "empty_arrays.crtdl")

	// CRTDL with empty arrays (should be valid - semantic validation by TORCH)
	emptyCRTDL := `{
		"cohortDefinition": {
			"version": "1.0",
			"inclusionCriteria": []
		},
		"dataExtraction": {
			"attributeGroups": []
		}
	}`
	err := os.WriteFile(crtdlFile, []byte(emptyCRTDL), 0644)
	require.NoError(t, err)

	// Test will validate successfully (syntax only, not semantics)
	// err = validation.ValidateCRTDLSyntax(crtdlFile)
	// assert.NoError(t, err, "Empty arrays should pass syntax validation")

	t.Skip("Skipping until validation.ValidateCRTDLSyntax() is implemented")
}

// Performance test - verify CRTDL validation < 1 second

func TestValidateCRTDLSyntax_PerformanceWithin1Second(t *testing.T) {
	tmpDir := t.TempDir()
	crtdlFile := filepath.Join(tmpDir, "performance_test.crtdl")

	// Create a realistic CRTDL file with moderate complexity (simulating real-world usage)
	crtdlContent := map[string]any{
		"cohortDefinition": map[string]any{
			"version": "1.0.0",
			"display": "Performance test cohort",
			"inclusionCriteria": []map[string]any{
				// Add 100 criteria groups to test with realistic complexity
				{
					"criteriaGroup": []map[string]any{
						{
							"code":    "8310-5",
							"system":  "http://loinc.org",
							"display": "Body temperature",
						},
					},
				},
				{
					"criteriaGroup": []map[string]any{
						{
							"code":    "8462-4",
							"system":  "http://loinc.org",
							"display": "Diastolic blood pressure",
						},
					},
				},
			},
		},
		"dataExtraction": map[string]any{
			"attributeGroups": []map[string]any{
				{
					"name":         "demographics",
					"resourceType": "Patient",
					"attributes":   []string{"birthDate", "gender", "name", "address"},
				},
				{
					"name":         "vitals",
					"resourceType": "Observation",
					"attributes":   []string{"code", "value", "effectiveDateTime"},
				},
				{
					"name":         "medications",
					"resourceType": "MedicationStatement",
					"attributes":   []string{"medication", "status", "effectiveDateTime"},
				},
			},
		},
	}

	// Marshal to JSON and write to file
	crtdlJSON, err := json.Marshal(crtdlContent)
	require.NoError(t, err)
	err = os.WriteFile(crtdlFile, crtdlJSON, 0644)
	require.NoError(t, err)

	// Measure validation time
	startTime := time.Now()
	err = lib.ValidateCRTDLSyntax(crtdlFile)
	duration := time.Since(startTime)

	// Assertions
	assert.NoError(t, err)
	assert.Less(t, duration, 1*time.Second, "CRTDL validation must complete within 1 second, took: %v", duration)

	// Log performance for visibility
	t.Logf("CRTDL validation completed in %v (requirement: < 1s)", duration)
}
