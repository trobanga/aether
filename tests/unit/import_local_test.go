package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trobanga/aether/internal/lib"
	"github.com/trobanga/aether/internal/models"
	"github.com/trobanga/aether/internal/services"
)

// TestImportFromLocalDirectory_Success verifies successful import from a valid local directory
func TestImportFromLocalDirectory_Success(t *testing.T) {
	// Setup temporary test directories
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "dest")

	// Create source directory with test FHIR files
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	// Create test NDJSON files
	testFiles := map[string]string{
		"Patient.ndjson":     `{"resourceType":"Patient","id":"1"}`,
		"Observation.ndjson": `{"resourceType":"Observation","id":"1"}`,
		"Bundle.ndjson":      `{"resourceType":"Bundle","id":"1"}`,
	}

	for filename, content := range testFiles {
		path := filepath.Join(sourceDir, filename)
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	}

	// Create logger
	logger := lib.NewLogger(lib.LogLevelInfo)

	// Execute import
	importedFiles, err := services.ImportFromLocalDirectory(sourceDir, destDir, logger)

	// Verify results
	assert.NoError(t, err, "Import should succeed")
	assert.Len(t, importedFiles, 3, "Should import all 3 FHIR files")

	// Verify files were copied
	for _, imported := range importedFiles {
		destPath := filepath.Join(destDir, imported.FileName)
		assert.FileExists(t, destPath, "File should be copied to destination")

		// Verify metadata
		assert.NotEmpty(t, imported.FileName, "FileName should be set")
		assert.Greater(t, imported.FileSize, int64(0), "FileSize should be > 0")
		assert.Equal(t, models.StepImport, imported.SourceStep, "SourceStep should be import")
		assert.Equal(t, 1, imported.LineCount, "LineCount should be 1 for single-line files")
	}

	// Verify resource types are extracted correctly
	resourceTypes := make(map[string]bool)
	for _, file := range importedFiles {
		resourceTypes[file.ResourceType] = true
	}
	assert.True(t, resourceTypes["Patient"], "Patient resource type should be identified")
	assert.True(t, resourceTypes["Observation"], "Observation resource type should be identified")
	assert.True(t, resourceTypes["Bundle"], "Bundle resource type should be identified")
}

// TestImportFromLocalDirectory_NonexistentDirectory verifies error handling for missing source directory
func TestImportFromLocalDirectory_NonexistentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "nonexistent")
	destDir := filepath.Join(tempDir, "dest")
	logger := lib.NewLogger(lib.LogLevelInfo)

	// Execute import
	importedFiles, err := services.ImportFromLocalDirectory(sourceDir, destDir, logger)

	// Verify error
	assert.Error(t, err, "Should fail for nonexistent directory")
	assert.Contains(t, err.Error(), "does not exist", "Error should mention nonexistent directory")
	assert.Nil(t, importedFiles, "Should not return files on error")
}

// TestImportFromLocalDirectory_NotADirectory verifies error handling when source is a file, not directory
func TestImportFromLocalDirectory_NotADirectory(t *testing.T) {
	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "file.txt")
	destDir := filepath.Join(tempDir, "dest")
	logger := lib.NewLogger(lib.LogLevelInfo)

	// Create source as a file, not directory
	require.NoError(t, os.WriteFile(sourceFile, []byte("test"), 0644))

	// Execute import
	importedFiles, err := services.ImportFromLocalDirectory(sourceFile, destDir, logger)

	// Verify error
	assert.Error(t, err, "Should fail when source is not a directory")
	assert.Contains(t, err.Error(), "not a directory", "Error should mention path is not a directory")
	assert.Nil(t, importedFiles, "Should not return files on error")
}

// TestImportFromLocalDirectory_NoNDJSONFiles verifies error handling when directory has no FHIR files
func TestImportFromLocalDirectory_NoNDJSONFiles(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "dest")
	logger := lib.NewLogger(lib.LogLevelInfo)

	// Create empty source directory
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	// Create non-FHIR files
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "readme.txt"), []byte("test"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "data.json"), []byte("{}"), 0644))

	// Execute import
	importedFiles, err := services.ImportFromLocalDirectory(sourceDir, destDir, logger)

	// Verify error
	assert.Error(t, err, "Should fail when no NDJSON files found")
	assert.Contains(t, err.Error(), "no FHIR NDJSON files found", "Error should mention no FHIR files")
	assert.Nil(t, importedFiles, "Should not return files on error")
}

// TestImportFromLocalDirectory_RecursiveScan verifies that subdirectories are scanned
func TestImportFromLocalDirectory_RecursiveScan(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "dest")
	logger := lib.NewLogger(lib.LogLevelInfo)

	// Create nested directory structure
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "subdir1"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "subdir2"), 0755))

	// Create NDJSON files in different locations
	testFiles := map[string]string{
		"Patient.ndjson":             `{"resourceType":"Patient"}`,
		"subdir1/Observation.ndjson": `{"resourceType":"Observation"}`,
		"subdir2/Encounter.ndjson":   `{"resourceType":"Encounter"}`,
	}

	for relPath, content := range testFiles {
		fullPath := filepath.Join(sourceDir, relPath)
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	// Execute import
	importedFiles, err := services.ImportFromLocalDirectory(sourceDir, destDir, logger)

	// Verify all files are found recursively
	assert.NoError(t, err, "Import should succeed")
	assert.Len(t, importedFiles, 3, "Should find all 3 NDJSON files recursively")
}

// TestImportFromLocalDirectory_MultilineFiles verifies correct line counting
func TestImportFromLocalDirectory_MultilineFiles(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	destDir := filepath.Join(tempDir, "dest")
	logger := lib.NewLogger(lib.LogLevelInfo)

	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	// Create multi-line NDJSON file (3 resources)
	multilineContent := `{"resourceType":"Patient","id":"1"}
{"resourceType":"Patient","id":"2"}
{"resourceType":"Patient","id":"3"}`

	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "Patient.ndjson"), []byte(multilineContent), 0644))

	// Execute import
	importedFiles, err := services.ImportFromLocalDirectory(sourceDir, destDir, logger)

	// Verify line count
	assert.NoError(t, err, "Import should succeed")
	require.Len(t, importedFiles, 1, "Should import 1 file")
	assert.Equal(t, 3, importedFiles[0].LineCount, "Should count 3 lines/resources")
}

// TestValidateImportSource_LocalDirectory tests input validation for local directories
func TestValidateImportSource_LocalDirectory(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		setupFunc   func() string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid directory with NDJSON files",
			setupFunc: func() string {
				dir := filepath.Join(tempDir, "valid")
				_ = os.MkdirAll(dir, 0755)
				_ = os.WriteFile(filepath.Join(dir, "test.ndjson"), []byte("{}"), 0644)
				return dir
			},
			expectError: false,
		},
		{
			name: "Nonexistent directory",
			setupFunc: func() string {
				return filepath.Join(tempDir, "nonexistent")
			},
			expectError: true,
			errorMsg:    "does not exist",
		},
		{
			name: "Path is a file, not directory",
			setupFunc: func() string {
				file := filepath.Join(tempDir, "file.txt")
				_ = os.WriteFile(file, []byte("test"), 0644)
				return file
			},
			expectError: true,
			errorMsg:    "not a directory",
		},
		{
			name: "Directory with no NDJSON files",
			setupFunc: func() string {
				dir := filepath.Join(tempDir, "empty")
				_ = os.MkdirAll(dir, 0755)
				_ = os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("test"), 0644)
				return dir
			},
			expectError: true,
			errorMsg:    "no FHIR NDJSON files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourcePath := tt.setupFunc()
			err := services.ValidateImportSource(sourcePath, models.InputTypeLocal)

			if tt.expectError {
				assert.Error(t, err, "Should return error")
				assert.Contains(t, err.Error(), tt.errorMsg, "Error message should be descriptive")
			} else {
				assert.NoError(t, err, "Should not return error")
			}
		})
	}
}
