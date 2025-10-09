package models

import (
	"path/filepath"
	"strings"
	"time"
)

// FHIRDataFile represents a single FHIR NDJSON file in the pipeline
type FHIRDataFile struct {
	FileName     string    `json:"file_name"`
	FilePath     string    `json:"file_path"`      // Relative to job directory
	ResourceType string    `json:"resource_type"`  // e.g., "Patient", "Observation"
	FileSize     int64     `json:"file_size"`      // Bytes
	SourceStep   StepName  `json:"source_step"`    // Which step produced this file
	LineCount    int       `json:"line_count"`     // Number of FHIR resources
	CreatedAt    time.Time `json:"created_at"`
}

// IsValidFHIRFile checks if the file has valid FHIR NDJSON format
func IsValidFHIRFile(filename string) bool {
	return strings.HasSuffix(strings.ToLower(filename), ".ndjson")
}

// IsSafePath checks if a file path is within job directory boundaries
// Prevents path traversal attacks (e.g., ../../etc/passwd)
func IsSafePath(path string) bool {
	// Clean the path to resolve any .. or . components
	clean := filepath.Clean(path)

	// Reject absolute paths
	if filepath.IsAbs(clean) {
		return false
	}

	// Reject paths that start with .. (parent directory)
	if strings.HasPrefix(clean, "..") {
		return false
	}

	return true
}

// GetResourceTypeFromFilename attempts to extract resource type from filename
// Example: "Patient_001.ndjson" -> "Patient"
func GetResourceTypeFromFilename(filename string) string {
	// Remove .ndjson extension
	base := strings.TrimSuffix(filename, ".ndjson")

	// Split on common delimiters (underscore, dash, dot)
	parts := strings.FieldsFunc(base, func(r rune) bool {
		return r == '_' || r == '-' || r == '.'
	})

	if len(parts) > 0 {
		// First part is typically the resource type
		return parts[0]
	}

	return "Unknown"
}
