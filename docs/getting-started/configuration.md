# Configuration

Aether uses YAML configuration files for flexible setup. Configuration can be overridden via CLI flags.

## Configuration File Format

Create an `aether.yaml` file in your project directory:

```yaml
# Service endpoints
services:
  # TORCH FHIR server (optional)
  torch:
    base_url: "http://localhost:8080"
    username: "researcher"
    password: "password"
    extraction_timeout_minutes: 30
    polling_interval_seconds: 5
    max_polling_interval_seconds: 30

  # DIMP Pseudonymization (optional)
  dimp:
    url: "http://localhost:32861/fhir"
    bundle_split_threshold_mb: 10

  # CSV Conversion Service (optional)
  csv_conversion:
    url: "http://localhost:9000/convert/csv"

  # Parquet Conversion Service (optional)
  parquet_conversion:
    url: "http://localhost:9000/convert/parquet"

# Pipeline configuration
pipeline:
  # NOTE: One of the import step types (torch, local_import, http_import) must always be first
  enabled_steps:
    - local_import  # or torch or http_import
    - dimp
    # - validation    (not yet implemented)
    # - csv_conversion (service not available)
    # - parquet_conversion (service not available)

# Retry strategy for transient errors
retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000

# Job state and data storage directory
jobs_dir: "./jobs"
```

## Configuration Options

### Services Section

**DIMP Pseudonymization:**
```yaml
services:
  dimp:
    url: "http://localhost:32861/fhir"
    bundle_split_threshold_mb: 10
```
- `url`: FHIR Pseudonymizer endpoint
- `bundle_split_threshold_mb`: Auto-split large bundles (1-100 MB, default: 10 MB)

**Data Conversion Services:**
```yaml
services:
  csv_conversion:
    url: "http://localhost:9000/convert/csv"
  parquet_conversion:
    url: "http://localhost:9000/convert/parquet"
```
Endpoints for converting FHIR data to other formats (CSV, Parquet).

**TORCH Integration:**
```yaml
services:
  torch:
    base_url: "http://torch.hospital.org"
    username: "researcher-name"
    password: "secure-password"
    extraction_timeout_minutes: 30
    polling_interval_seconds: 5
    max_polling_interval_seconds: 30
```
Credentials and configuration for TORCH FHIR server integration:
- `base_url`: TORCH API endpoint
- `extraction_timeout_minutes`: Max wait for extraction (default: 30)
- `polling_interval_seconds`: Initial poll interval (default: 5)
- `max_polling_interval_seconds`: Max poll interval (default: 30)

### Pipeline Section

**Enabled Steps:**
```yaml
pipeline:
  enabled_steps:
    - import        # Import FHIR data
    - dimp          # Pseudonymization
```

Steps are executed in order. Available steps:
- `import`: Load FHIR data from local files or TORCH
- `dimp`: Apply pseudonymization via DIMP service
- `validation`: (Placeholder, not implemented)
- `csv_conversion`: Convert to CSV format
- `parquet_conversion`: Convert to Parquet format

### Retry Section

**Automatic Retry Configuration:**
```yaml
retry:
  max_attempts: 5              # Max retry attempts
  initial_backoff_ms: 1000     # Starting wait time (1 second)
  max_backoff_ms: 30000        # Maximum wait time (30 seconds)
```

Exponential backoff is applied between retries for transient errors.

### Jobs Directory

**Job Storage:**
```yaml
jobs_dir: "./jobs"
```

Directory where Aether stores job state and processed data. Must be writable. Each job gets a UUID subdirectory containing state and processed data.

## Usage Examples

### Local FHIR File Processing

Process FHIR NDJSON files from a local directory with pseudonymization:

```yaml
# aether.yaml
services:
  dimp:
    url: "http://localhost:32861/fhir"
    bundle_split_threshold_mb: 10

pipeline:
  enabled_steps:
    - import
    - dimp

retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000

jobs_dir: "./jobs"
```

Run:
```bash
aether pipeline start /path/to/fhir/files/
```

### TORCH Extraction with DIMP Pseudonymization

Extract minimized data from TORCH using a CRTDL query, then apply pseudonymization via FHIR Pseudonymizer:

```yaml
# aether.yaml
services:
  torch:
    base_url: "http://torch.hospital.org"
    username: "researcher"
    password: "secret"
    extraction_timeout_minutes: 30
    polling_interval_seconds: 5
  dimp:
    url: "http://localhost:32861/fhir"
    bundle_split_threshold_mb: 10

pipeline:
  enabled_steps:
    - import     # Import extracted data
    - dimp       # Pseudonymize

retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000

jobs_dir: "./jobs"
```

Run:
```bash
aether pipeline start my_query.crtdl
```

### Custom Retry Policy

Adjust retry behavior for your environment:

```yaml
retry:
  max_attempts: 3              # Fewer retries
  initial_backoff_ms: 500      # Start sooner
  max_backoff_ms: 5000         # Cap at 5 seconds
```

## Overriding via CLI Flags

You can override configuration options via CLI flags:

```bash
aether pipeline start --jobs-dir /data/jobs ./query.crtdl

aether job list --jobs-dir /data/jobs
```

## Environment Variables

Sensitive values like passwords can be set via environment variables:

```bash
# Instead of hardcoding in YAML, use env vars:
export TORCH_PASSWORD="secure-password"
export DIMP_URL="http://localhost:8083/fhir"

aether pipeline start query.crtdl
```

Reference in `aether.yaml` using `${ENV_VAR_NAME}` syntax (if supported by your Aether version).

## Troubleshooting Configuration Issues

**Issue: "jobs directory does not exist"**
- Ensure the jobs directory is created and writable
- Check the `jobs.jobs_dir` path in your configuration

**Issue: "service unavailable" errors**
- Verify service endpoints are correct in `services.*_url`
- Ensure services are running and accessible from your machine
- Check network connectivity and firewall rules

**Issue: TORCH authentication fails**
- Verify `services.torch.username` and `services.torch.password` are correct
- Test TORCH connectivity: `curl -u username:password http://torch.hospital.org/fhir/`

## Next Steps

- [Quick Start](./quick-start.md) - Get started with your first pipeline
- [TORCH Integration](../guides/torch-integration.md) - Learn about TORCH setup
- [API Reference](../api-reference/config-reference.md) - Complete configuration reference
