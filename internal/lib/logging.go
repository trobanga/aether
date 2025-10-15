package lib

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel defines the severity of log messages
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger provides structured logging for the application
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stderr, "", log.LstdFlags),
	}
}

// DefaultLogger returns a logger with INFO level
var DefaultLogger = NewLogger(LogLevelInfo)

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...interface{}) {
	if l.level <= LogLevelDebug {
		l.log("DEBUG", message, fields...)
	}
}

// Info logs an informational message
func (l *Logger) Info(message string, fields ...interface{}) {
	if l.level <= LogLevelInfo {
		l.log("INFO", message, fields...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...interface{}) {
	if l.level <= LogLevelWarn {
		l.log("WARN", message, fields...)
	}
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...interface{}) {
	if l.level <= LogLevelError {
		l.log("ERROR", message, fields...)
	}
}

// log formats and writes a log message with optional fields
func (l *Logger) log(level string, message string, fields ...interface{}) {
	var fieldsStr string
	if len(fields) > 0 {
		fieldsStr = fmt.Sprintf(" | %v", fields)
	}
	l.logger.Printf("[%s] %s%s", level, message, fieldsStr)
}

// LogOperation logs the start and completion of an operation
func LogOperation(logger *Logger, operation string, fn func() error) error {
	logger.Info(fmt.Sprintf("Starting: %s", operation))
	start := time.Now()

	err := fn()

	duration := time.Since(start)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed: %s", operation), "duration", duration, "error", err)
		return err
	}

	logger.Info(fmt.Sprintf("Completed: %s", operation), "duration", duration)
	return nil
}

// LogRetry logs retry attempts
func LogRetry(logger *Logger, operation string, attempt int, maxAttempts int, err error) {
	// Remove line breaks from operation to prevent log spoofing
	safeOperation := strings.ReplaceAll(operation, "\n", "")
	safeOperation = strings.ReplaceAll(safeOperation, "\r", "")
	logger.Warn(
		fmt.Sprintf("Retry attempt %d/%d for: %s", attempt+1, maxAttempts, safeOperation),
		"error", err,
	)
}

// LogStepStart logs the start of a pipeline step
func LogStepStart(logger *Logger, stepName string, jobID string) {
	logger.Info(
		"Step started",
		"step", stepName,
		"job_id", jobID,
	)
}

// LogStepComplete logs the completion of a pipeline step
func LogStepComplete(logger *Logger, stepName string, jobID string, filesProcessed int, duration time.Duration) {
	logger.Info(
		"Step completed",
		"step", stepName,
		"job_id", jobID,
		"files", filesProcessed,
		"duration", duration,
	)
}

// LogStepFailed logs a failed pipeline step
func LogStepFailed(logger *Logger, stepName string, jobID string, err error, retryable bool) {
	logger.Error(
		"Step failed",
		"step", stepName,
		"job_id", jobID,
		"error", err,
		"retryable", retryable,
	)
}

// LogJobCreated logs job creation
func LogJobCreated(logger *Logger, jobID string, inputSource string) {
	logger.Info(
		"Job created",
		"job_id", jobID,
		"input_source", inputSource,
	)
}

// LogJobCompleted logs job completion
func LogJobCompleted(logger *Logger, jobID string, totalFiles int, duration time.Duration) {
	logger.Info(
		"Job completed",
		"job_id", jobID,
		"total_files", totalFiles,
		"duration", duration,
	)
}

// LogServiceCall logs HTTP service calls
func LogServiceCall(logger *Logger, service string, endpoint string, method string) {
	logger.Debug(
		"Service call",
		"service", service,
		"endpoint", endpoint,
		"method", method,
	)
}

// LogServiceResponse logs HTTP service responses
func LogServiceResponse(logger *Logger, service string, statusCode int, duration time.Duration) {
	if statusCode >= 400 {
		logger.Warn(
			"Service response",
			"service", service,
			"status", statusCode,
			"duration", duration,
		)
	} else {
		logger.Debug(
			"Service response",
			"service", service,
			"status", statusCode,
			"duration", duration,
		)
	}
}

// SetLevel changes the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// ParseLogLevel converts a string to LogLevel
func ParseLogLevel(levelStr string) LogLevel {
	switch levelStr {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}
