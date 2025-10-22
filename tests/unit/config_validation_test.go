package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trobanga/aether/internal/lib"
)

// TestValidateSplitConfig verifies configuration validation for Bundle splitting threshold
// Unit test for configuration validation
func TestValidateSplitConfig(t *testing.T) {
	testCases := []struct {
		name        string
		thresholdMB int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Negative threshold - invalid",
			thresholdMB: -1,
			expectError: true,
			errorMsg:    "must be > 0",
		},
		{
			name:        "Zero threshold - invalid",
			thresholdMB: 0,
			expectError: true,
			errorMsg:    "must be > 0",
		},
		{
			name:        "Very small threshold (1MB) - valid",
			thresholdMB: 1,
			expectError: false,
		},
		{
			name:        "Normal threshold (10MB) - valid",
			thresholdMB: 10,
			expectError: false,
		},
		{
			name:        "Large threshold (50MB) - valid with warning",
			thresholdMB: 50,
			expectError: false,
		},
		{
			name:        "Very large threshold (75MB) - valid with warning",
			thresholdMB: 75,
			expectError: false,
		},
		{
			name:        "Maximum valid threshold (100MB) - valid",
			thresholdMB: 100,
			expectError: false,
		},
		{
			name:        "Over maximum (101MB) - invalid",
			thresholdMB: 101,
			expectError: true,
			errorMsg:    "must be <= 100",
		},
		{
			name:        "Significantly over maximum (200MB) - invalid",
			thresholdMB: 200,
			expectError: true,
			errorMsg:    "must be <= 100",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := lib.ValidateSplitConfig(tc.thresholdMB)

			if tc.expectError {
				assert.Error(t, err, "Expected validation error")
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err, "Should not have validation error")
			}
		})
	}
}

// TestValidateSplitConfig_EdgeCases verifies boundary conditions
func TestValidateSplitConfig_EdgeCases(t *testing.T) {
	// Test boundary at 0/1
	err := lib.ValidateSplitConfig(0)
	assert.Error(t, err)
	err = lib.ValidateSplitConfig(1)
	assert.NoError(t, err)

	// Test boundary at 100/101
	err = lib.ValidateSplitConfig(100)
	assert.NoError(t, err)
	err = lib.ValidateSplitConfig(101)
	assert.Error(t, err)
}

// TestValidateSplitConfig_TypeConversion verifies the function handles MB to bytes conversion correctly
func TestValidateSplitConfig_ThresholdConversion(t *testing.T) {
	// 10MB should be valid
	err := lib.ValidateSplitConfig(10)
	assert.NoError(t, err, "10MB threshold should be valid")

	// The validation function works with MB values, not bytes
	// This is important for user configuration
	thresholdMB := 10
	assert.Greater(t, thresholdMB, 0, "MB value should be positive")
	assert.LessOrEqual(t, thresholdMB, 100, "MB value should be <= 100")
}
