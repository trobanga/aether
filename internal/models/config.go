package models

// ProjectConfig is the top-level configuration for the Aether pipeline
type ProjectConfig struct {
	Services ServiceConfig  `yaml:"services" json:"services"`
	Pipeline PipelineConfig `yaml:"pipeline" json:"pipeline"`
	Retry    RetryConfig    `yaml:"retry" json:"retry"`
	JobsDir  string         `yaml:"jobs_dir" json:"jobs_dir"`
}

// ServiceConfig contains connection details for external HTTP services
type ServiceConfig struct {
	DIMPUrl              string `yaml:"dimp_url" json:"dimp_url"`
	CSVConversionUrl     string `yaml:"csv_conversion_url" json:"csv_conversion_url"`
	ParquetConversionUrl string `yaml:"parquet_conversion_url" json:"parquet_conversion_url"`
}

// PipelineConfig defines which steps are enabled and their execution order
type PipelineConfig struct {
	EnabledSteps []StepName `yaml:"enabled_steps" json:"enabled_steps"`
}

// RetryConfig controls retry behavior for transient errors
type RetryConfig struct {
	MaxAttempts      int   `yaml:"max_attempts" json:"max_attempts"`
	InitialBackoffMs int64 `yaml:"initial_backoff_ms" json:"initial_backoff_ms"`
	MaxBackoffMs     int64 `yaml:"max_backoff_ms" json:"max_backoff_ms"`
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() ProjectConfig {
	return ProjectConfig{
		Services: ServiceConfig{
			DIMPUrl:              "",
			CSVConversionUrl:     "",
			ParquetConversionUrl: "",
		},
		Pipeline: PipelineConfig{
			EnabledSteps: []StepName{StepImport},
		},
		Retry: RetryConfig{
			MaxAttempts:      5,
			InitialBackoffMs: 1000,
			MaxBackoffMs:     30000,
		},
		JobsDir: "./jobs",
	}
}

// IsStepEnabled checks if a specific step is enabled in the pipeline configuration
func (c *PipelineConfig) IsStepEnabled(step StepName) bool {
	for _, enabled := range c.EnabledSteps {
		if enabled == step {
			return true
		}
	}
	return false
}

// GetNextStep returns the next enabled step after the current one, or empty string if no more steps
func (c *PipelineConfig) GetNextStep(current StepName) StepName {
	foundCurrent := false
	for _, step := range c.EnabledSteps {
		if foundCurrent {
			return step
		}
		if step == current {
			foundCurrent = true
		}
	}
	return "" // No next step
}

// HasServiceURL checks if a service URL is configured for a given step
func (c *ServiceConfig) HasServiceURL(step StepName) bool {
	switch step {
	case StepDIMP:
		return c.DIMPUrl != ""
	case StepCSVConversion:
		return c.CSVConversionUrl != ""
	case StepParquetConversion:
		return c.ParquetConversionUrl != ""
	default:
		return true // Import and validation don't require external services
	}
}

// GetServiceURL returns the service URL for a given step
func (c *ServiceConfig) GetServiceURL(step StepName) string {
	switch step {
	case StepDIMP:
		return c.DIMPUrl
	case StepCSVConversion:
		return c.CSVConversionUrl
	case StepParquetConversion:
		return c.ParquetConversionUrl
	default:
		return ""
	}
}
