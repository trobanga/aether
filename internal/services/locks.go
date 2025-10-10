package services

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/trobanga/aether/internal/lib"
)

// JobLock represents a file lock for a specific job
// Prevents concurrent modification of job state by multiple processes
type JobLock struct {
	jobID    string
	lockFile *os.File
	lockPath string
	logger   *lib.Logger
}

// AcquireJobLock attempts to acquire an exclusive lock for a job
// Returns a JobLock if successful, error if lock is already held by another process
// The lock is automatically released when the JobLock is closed or the process exits
func AcquireJobLock(jobsDir string, jobID string, logger *lib.Logger) (*JobLock, error) {
	jobDir := GetJobDir(jobsDir, jobID)
	lockPath := filepath.Join(jobDir, ".lock")

	// Ensure job directory exists
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create job directory: %w", err)
	}

	// Open/create lock file
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	// Try to acquire exclusive lock (non-blocking)
	// flock() is advisory - cooperating processes must check the lock
	err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		lockFile.Close()
		if err == syscall.EWOULDBLOCK {
			return nil, fmt.Errorf("job %s is locked by another process", jobID)
		}
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	// Write current process ID and timestamp to lock file for debugging
	lockInfo := fmt.Sprintf("pid=%d\ntime=%s\n", os.Getpid(), time.Now().Format(time.RFC3339))
	lockFile.Truncate(0)
	lockFile.Seek(0, 0)
	lockFile.WriteString(lockInfo)
	lockFile.Sync()

	logger.Debug("Acquired job lock", "job_id", jobID, "pid", os.Getpid())

	return &JobLock{
		jobID:    jobID,
		lockFile: lockFile,
		lockPath: lockPath,
		logger:   logger,
	}, nil
}

// Release releases the job lock
// Should be called when job operations are complete
func (jl *JobLock) Release() error {
	if jl.lockFile == nil {
		return nil
	}

	// Release flock
	err := syscall.Flock(int(jl.lockFile.Fd()), syscall.LOCK_UN)
	if err != nil {
		jl.logger.Warn("Failed to release flock", "job_id", jl.jobID, "error", err)
	}

	// Close lock file
	if err := jl.lockFile.Close(); err != nil {
		jl.logger.Warn("Failed to close lock file", "job_id", jl.jobID, "error", err)
		return err
	}

	jl.logger.Debug("Released job lock", "job_id", jl.jobID, "pid", os.Getpid())
	jl.lockFile = nil

	return nil
}

// IsLocked checks if a job is currently locked by any process
// This is a non-destructive check that doesn't acquire the lock
func IsJobLocked(jobsDir string, jobID string) bool {
	jobDir := GetJobDir(jobsDir, jobID)
	lockPath := filepath.Join(jobDir, ".lock")

	// If lock file doesn't exist, job is not locked
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		return false
	}

	// Try to open lock file
	lockFile, err := os.Open(lockPath)
	if err != nil {
		// Can't open lock file - assume not locked
		return false
	}
	defer lockFile.Close()

	// Try to acquire lock (non-blocking)
	err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		// Lock is held by another process
		return err == syscall.EWOULDBLOCK
	}

	// We acquired the lock - release it immediately
	syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
	return false
}

// WithJobLock executes a function while holding a job lock
// Automatically acquires the lock, executes the function, and releases the lock
// Returns error if lock cannot be acquired or if the function returns an error
func WithJobLock(jobsDir string, jobID string, logger *lib.Logger, fn func() error) error {
	// Acquire lock
	lock, err := AcquireJobLock(jobsDir, jobID, logger)
	if err != nil {
		return err
	}
	defer lock.Release()

	// Execute function with lock held
	return fn()
}
