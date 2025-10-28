package models

import (
	"fmt"
	"time"
)

// PipelineStep represents a discrete stage in the pipeline
type PipelineStep struct {
	Name           StepName   `json:"name"`
	Status         StepStatus `json:"status"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	FilesProcessed int        `json:"files_processed"`
	BytesProcessed int64      `json:"bytes_processed"`
	RetryCount     int        `json:"retry_count"`
	LastError      *StepError `json:"last_error,omitempty"`
}

// StepName defines the available pipeline steps
type StepName string

const (
	StepTorchImport       StepName = "torch"              // TORCH import via CRTDL or direct TORCH URL
	StepLocalImport       StepName = "local_import"       // Import from local directory
	StepHttpImport        StepName = "http_import"        // Import from HTTP URL
	StepDIMP              StepName = "dimp"
	StepValidation        StepName = "validation"
	StepCSVConversion     StepName = "csv_conversion"
	StepParquetConversion StepName = "parquet_conversion"
)

// StepStatus defines the execution state of a pipeline step
type StepStatus string

const (
	StepStatusPending    StepStatus = "pending"
	StepStatusInProgress StepStatus = "in_progress"
	StepStatusCompleted  StepStatus = "completed"
	StepStatusFailed     StepStatus = "failed"
)

// StepError captures error details for a failed step
type StepError struct {
	Type       ErrorType `json:"type"` // "transient" | "non_transient"
	Message    string    `json:"message"`
	HTTPStatus int       `json:"http_status,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// Error implements the error interface
func (e *StepError) Error() string {
	if e.HTTPStatus > 0 {
		return fmt.Sprintf("HTTP %d: %s", e.HTTPStatus, e.Message)
	}
	return e.Message
}

// ErrorType classifies errors for retry strategy
type ErrorType string

const (
	ErrorTypeTransient    ErrorType = "transient"     // Network, 5xx, timeout - automatic retry
	ErrorTypeNonTransient ErrorType = "non_transient" // 4xx, validation, malformed - manual intervention
)

// IsValidStepName checks if the step name is recognized
func IsValidStepName(name StepName) bool {
	switch name {
	case StepTorchImport, StepLocalImport, StepHttpImport, StepDIMP, StepValidation, StepCSVConversion, StepParquetConversion:
		return true
	default:
		return false
	}
}

// IsValidStepStatus checks if the step status is recognized
func IsValidStepStatus(s StepStatus) bool {
	switch s {
	case StepStatusPending, StepStatusInProgress, StepStatusCompleted, StepStatusFailed:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if step status transition is valid
// Valid transitions:
//
//	pending -> in_progress
//	in_progress -> completed | failed
//	failed -> in_progress (retry if transient error)
func (s StepStatus) CanTransitionTo(next StepStatus) bool {
	switch s {
	case StepStatusPending:
		return next == StepStatusInProgress
	case StepStatusInProgress:
		return next == StepStatusCompleted || next == StepStatusFailed
	case StepStatusFailed:
		return next == StepStatusInProgress // Allow retry
	case StepStatusCompleted:
		return false // Terminal state
	default:
		return false
	}
}

// IsRetryable determines if a step error should trigger automatic retry
func (e StepError) IsRetryable(maxRetries int, currentRetries int) bool {
	return e.Type == ErrorTypeTransient && currentRetries < maxRetries
}

// IsTransientHTTPStatus classifies HTTP status codes for retry logic
func IsTransientHTTPStatus(status int) bool {
	// 5xx server errors are transient (service might recover)
	if status >= 500 && status < 600 {
		return true
	}
	// 408 Request Timeout, 429 Too Many Requests are transient
	if status == 408 || status == 429 {
		return true
	}
	return false
}
