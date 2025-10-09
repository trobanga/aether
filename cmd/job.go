package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"github.com/trobanga/aether/internal/lib"
	"github.com/trobanga/aether/internal/pipeline"
	"github.com/trobanga/aether/internal/services"
)

// jobCmd represents the job command group
var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Manage pipeline jobs",
	Long: `Manage pipeline jobs: list, inspect, and control job execution.

Available subcommands:
  list - List all pipeline jobs`,
}

// jobListCmd represents the job list command
var jobListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all pipeline jobs",
	Long: `List all pipeline jobs in the jobs directory.

Shows:
  - Job ID
  - Status
  - Current step
  - Creation time
  - File counts

Example:
  aether job list`,
	RunE: runJobList,
}

func init() {
	rootCmd.AddCommand(jobCmd)
	jobCmd.AddCommand(jobListCmd)
}

func runJobList(cmd *cobra.Command, args []string) error {
	// Load configuration
	config, err := services.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// List all job IDs
	jobIDs, err := services.ListAllJobs(config.JobsDir)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	if len(jobIDs) == 0 {
		fmt.Println("No jobs found")
		return nil
	}

	// Load each job and display summary
	type jobSummary struct {
		ID         string
		Status     string
		Step       string
		Files      int
		RetryCount int
		CreatedAt  time.Time
		ElapsedStr string
	}

	var jobs []jobSummary
	for _, jobID := range jobIDs {
		job, err := pipeline.LoadJob(config.JobsDir, jobID)
		if err != nil {
			lib.DefaultLogger.Warn("Failed to load job", "job_id", jobID, "error", err)
			continue
		}

		elapsed := time.Since(job.CreatedAt)
		elapsedStr := formatDuration(elapsed)

		// Get retry count from current step
		retryCount := 0
		if currentStep, found := pipeline.GetCurrentStep(job); found {
			retryCount = currentStep.RetryCount
		}

		jobs = append(jobs, jobSummary{
			ID:         job.JobID,
			Status:     string(job.Status),
			Step:       job.CurrentStep,
			Files:      job.TotalFiles,
			RetryCount: retryCount,
			CreatedAt:  job.CreatedAt,
			ElapsedStr: elapsedStr,
		})
	}

	// Sort by creation time (newest first)
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.After(jobs[j].CreatedAt)
	})

	// Print table header
	fmt.Printf("%-38s %-15s %-20s %-8s %-8s %s\n", "JOB ID", "STATUS", "STEP", "FILES", "RETRIES", "AGE")
	fmt.Println("------------------------------------------------------------------------------------------------------------------------")

	// Print jobs
	for _, j := range jobs {
		statusSymbol := getJobStatusSymbol(j.Status)
		fmt.Printf("%-38s %s %-13s %-20s %-8d %-8d %s\n",
			j.ID,
			statusSymbol,
			j.Status,
			j.Step,
			j.Files,
			j.RetryCount,
			j.ElapsedStr,
		)
	}

	fmt.Printf("\nTotal: %d jobs\n", len(jobs))

	return nil
}

func getJobStatusSymbol(status string) string {
	switch status {
	case "completed":
		return "✓"
	case "in_progress":
		return "→"
	case "failed":
		return "✗"
	case "pending":
		return "○"
	default:
		return " "
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%dd", days)
}
