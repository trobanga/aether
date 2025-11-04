package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
)

// FileContext holds file handles and cleanup logic for atomic file writing
type FileContext struct {
	InFile   *os.File
	OutFile  *os.File
	TempFile string
	Cleanup  func()
}

// SetupFileProcessing initializes files for atomic write pattern
// Writes to .part file first, renamed on success
func SetupFileProcessing(inputFile, outputFile string) (*FileContext, error) {
	// Open input file
	inFile, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open input file: %w", err)
	}

	// Create temporary output file with .part extension
	tempOutputFile := outputFile + ".part"
	outFile, err := os.Create(tempOutputFile)
	if err != nil {
		_ = inFile.Close()
		return nil, fmt.Errorf("failed to create temporary output file: %w", err)
	}

	// Track whether operation succeeded for cleanup
	var success bool
	ctx := &FileContext{
		InFile:   inFile,
		OutFile:  outFile,
		TempFile: tempOutputFile,
		Cleanup: func() {
			if !success {
				_ = os.Remove(tempOutputFile)
			}
		},
	}

	return ctx, nil
}

// FinalizeFileProcessing closes files and atomically renames .part to final filename
func FinalizeFileProcessing(ctx *FileContext, outputFile string, markSuccess bool) error {
	defer ctx.Cleanup()

	// Close output file before rename
	if err := ctx.OutFile.Close(); err != nil {
		return fmt.Errorf("failed to close output file: %w", err)
	}

	// Close input file
	if err := ctx.InFile.Close(); err != nil {
		return fmt.Errorf("failed to close input file: %w", err)
	}

	// Only rename if successful
	if !markSuccess {
		return nil
	}

	// Atomic rename: move .part file to final filename
	if err := os.Rename(ctx.TempFile, outputFile); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// WriteProcessedResource marshals and writes a FHIR resource to file
func WriteProcessedResource(resource map[string]any, outFile *os.File) error {
	pseudonymizedJSON, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("failed to marshal pseudonymized resource: %w", err)
	}

	if _, err := outFile.Write(pseudonymizedJSON); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	if _, err := outFile.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}
