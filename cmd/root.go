/*
Copyright © 2025 Aether Contributors

Aether is a CLI tool for orchestrating Data Use Process (DUP) pipelines.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aether",
	Short: "Aether - Data Use Process (DUP) Pipeline CLI",
	Long: `Aether orchestrates Data Use Process (DUP) pipelines for medical FHIR data.

The CLI imports TORCH-extracted FHIR data and processes it through optional steps:
  - DIMP pseudonymization
  - Validation (placeholder)
  - Format conversion (CSV/Parquet)

All processing steps use external HTTP services. The pipeline supports
session-independent resumption and hybrid retry strategies (automatic for
transient errors, manual for validation failures).

Example:
  aether pipeline start --input /data/torch/output
  aether pipeline status <job-id>
  aether job list

For more information, visit: https://github.com/trobanga/aether`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./aether.yaml, ~/.config/aether/aether.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")

	// Add version template
	rootCmd.SetVersionTemplate("Aether version {{.Version}}\n")
}


