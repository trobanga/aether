package services

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/trobanga/aether/internal/lib"
	"github.com/trobanga/aether/internal/models"
)

// DIMPClient handles communication with the DIMP pseudonymization service
// Per contracts/dimp-service.md
type DIMPClient struct {
	baseURL    string
	httpClient *HTTPClient
	logger     *lib.Logger
}

// NewDIMPClient creates a new DIMP client with the given base URL
func NewDIMPClient(baseURL string, httpClient *HTTPClient, logger *lib.Logger) *DIMPClient {
	return &DIMPClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		logger:     logger,
	}
}

// Pseudonymize sends a FHIR resource to the DIMP service for pseudonymization
// Returns the pseudonymized resource or an error
// Per contract: POST /$de-identify with single FHIR resource
func (c *DIMPClient) Pseudonymize(resource map[string]interface{}) (map[string]interface{}, error) {
	// Extract resource info for logging
	resourceType, _ := resource["resourceType"].(string)
	resourceID, _ := resource["id"].(string)

	c.logger.Debug("Sending resource to DIMP",
		"resourceType", resourceType,
		"id", resourceID,
		"url", c.baseURL+"/$de-identify")

	// Marshal resource to JSON
	jsonBody, err := json.Marshal(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource: %w", err)
	}

	c.logger.Debug("Request body size", "bytes", len(jsonBody))

	// Construct endpoint URL
	url := c.baseURL + "/$de-identify"

	// Send POST request
	resp, err := c.httpClient.PostJSON(url, jsonBody)
	if err != nil {
		c.logger.Error("DIMP HTTP request failed",
			"resourceType", resourceType,
			"id", resourceID,
			"error", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP error status
	if resp.StatusCode >= 400 {
		// Read error response body
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorType := lib.ClassifyHTTPError(resp.StatusCode)

		c.logger.Error("DIMP service returned error",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"resourceType", resourceType,
			"id", resourceID,
			"error_body", string(bodyBytes),
			"retryable", errorType == models.ErrorTypeTransient)

		// Create error with classification
		err := &DIMPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			ErrorType:  errorType,
			Body:       string(bodyBytes),
		}

		return nil, err
	}

	c.logger.Debug("DIMP service responded successfully",
		"status_code", resp.StatusCode,
		"resourceType", resourceType,
		"id", resourceID)

	// Success - parse pseudonymized resource
	var pseudonymized map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&pseudonymized); err != nil {
		c.logger.Error("Failed to decode DIMP response",
			"resourceType", resourceType,
			"id", resourceID,
			"error", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Log what changed
	originalID, _ := resource["id"].(string)
	newID, _ := pseudonymized["id"].(string)
	if originalID != newID {
		c.logger.Debug("Resource ID pseudonymized",
			"resourceType", resourceType,
			"original_id", originalID,
			"new_id", newID)
	}

	return pseudonymized, nil
}

// DIMPError represents an error response from the DIMP service
type DIMPError struct {
	StatusCode int
	Status     string
	ErrorType  models.ErrorType
	Body       string
}

func (e *DIMPError) Error() string {
	return fmt.Sprintf("DIMP service error: HTTP %d: %s", e.StatusCode, e.Status)
}

// IsRetryable returns true if this error should be retried
func (e *DIMPError) IsRetryable() bool {
	return e.ErrorType == models.ErrorTypeTransient
}
