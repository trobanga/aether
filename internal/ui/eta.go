package ui

import (
	"fmt"
	"time"
)

// ETACalculator computes estimated time of arrival for operations
// Formula: ETA = (total_items - processed_items) * avg_time_per_item
// Average time per item is computed from last 10 samples or last 30 seconds (whichever more recent)
type ETACalculator struct {
	samples       []TimestampedProgress
	maxSamples    int           // Max number of samples to keep (default 10)
	maxTimeWindow time.Duration // Max time window for samples (default 30s)
}

// TimestampedProgress records a progress measurement at a specific time
type TimestampedProgress struct {
	Timestamp time.Time
	Items     int64
}

// NewETACalculator creates an ETA calculator with default settings
// Uses last 10 samples or 30 second time window for averaging
func NewETACalculator() *ETACalculator {
	return &ETACalculator{
		samples:       make([]TimestampedProgress, 0),
		maxSamples:    10,
		maxTimeWindow: 30 * time.Second,
	}
}

// NewETACalculatorCustom creates an ETA calculator with custom settings
func NewETACalculatorCustom(maxSamples int, maxTimeWindow time.Duration) *ETACalculator {
	return &ETACalculator{
		samples:       make([]TimestampedProgress, 0),
		maxSamples:    maxSamples,
		maxTimeWindow: maxTimeWindow,
	}
}

// RecordProgress records a progress measurement
func (e *ETACalculator) RecordProgress(itemsProcessed int64) {
	now := time.Now()

	// Add new sample
	e.samples = append(e.samples, TimestampedProgress{
		Timestamp: now,
		Items:     itemsProcessed,
	})

	// Remove old samples (keep only maxSamples most recent)
	if len(e.samples) > e.maxSamples {
		e.samples = e.samples[len(e.samples)-e.maxSamples:]
	}

	// Remove samples outside time window
	e.pruneOldSamples(now)
}

// pruneOldSamples removes samples older than maxTimeWindow
func (e *ETACalculator) pruneOldSamples(now time.Time) {
	cutoff := now.Add(-e.maxTimeWindow)

	// Find first sample within time window
	firstValid := 0
	for i, sample := range e.samples {
		if sample.Timestamp.After(cutoff) {
			firstValid = i
			break
		}
	}

	// Keep only samples within time window
	if firstValid > 0 && firstValid < len(e.samples) {
		e.samples = e.samples[firstValid:]
	}
}

// CalculateETA computes estimated time to completion
// Returns:
//   - eta: estimated time remaining
//   - valid: whether ETA can be reliably computed (need at least 2 samples)
//
// Formula: ETA = (total_items - processed_items) * avg_time_per_item
func (e *ETACalculator) CalculateETA(totalItems int64, currentItems int64) (time.Duration, bool) {
	if len(e.samples) < 2 {
		return 0, false // Not enough data
	}

	if currentItems >= totalItems {
		return 0, true // Already complete
	}

	// Calculate average time per item from recent samples
	avgTimePerItem, valid := e.getAverageTimePerItem()
	if !valid {
		return 0, false
	}

	// Apply formula: ETA = (total_items - processed_items) * avg_time_per_item
	remainingItems := totalItems - currentItems
	eta := time.Duration(float64(remainingItems) * avgTimePerItem.Seconds() * float64(time.Second))

	return eta, true
}

// getAverageTimePerItem calculates average time per item from recent samples
func (e *ETACalculator) getAverageTimePerItem() (time.Duration, bool) {
	if len(e.samples) < 2 {
		return 0, false
	}

	// Use first and last sample in the window
	first := e.samples[0]
	last := e.samples[len(e.samples)-1]

	// Calculate time and items delta
	timeDelta := last.Timestamp.Sub(first.Timestamp)
	itemsDelta := last.Items - first.Items

	if itemsDelta <= 0 || timeDelta <= 0 {
		return 0, false // No progress or negative progress
	}

	// Average time per item = total_time / items_processed
	avgTimePerItem := timeDelta / time.Duration(itemsDelta)

	return avgTimePerItem, true
}

// GetThroughput returns current throughput (items per second)
// This complements the ETA calculation
func (e *ETACalculator) GetThroughput() (float64, bool) {
	avgTimePerItem, valid := e.getAverageTimePerItem()
	if !valid {
		return 0, false
	}

	// Items per second = 1 / (seconds per item)
	itemsPerSecond := 1.0 / avgTimePerItem.Seconds()

	return itemsPerSecond, true
}

// Reset clears all recorded samples
func (e *ETACalculator) Reset() {
	e.samples = make([]TimestampedProgress, 0)
}

// FormatETA formats an ETA duration as a human-readable string
func FormatETA(eta time.Duration) string {
	if eta < time.Second {
		return "< 1s"
	}

	if eta < time.Minute {
		return eta.Round(time.Second).String()
	}

	if eta < time.Hour {
		minutes := int(eta.Minutes())
		seconds := int(eta.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}

	hours := int(eta.Hours())
	minutes := int(eta.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", hours, minutes)
}

// FormatDuration formats a duration as a human-readable string
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(time.Second).String()
}
