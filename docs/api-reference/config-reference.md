# Configuration Reference

Comprehensive reference for all Aether configuration options.

For an introduction, see [Configuration Guide](../getting-started/configuration.md).

## Complete Configuration Schema

```yaml
# Service endpoints
services:
  dimp_url: string              # DIMP pseudonymization service
  csv_conversion_url: string    # CSV conversion service (future)
  parquet_conversion_url: string # Parquet conversion service (future)
  torch:
    base_url: string            # TORCH FHIR server URL
    username: string            # TORCH username
    password: string            # TORCH password

# Pipeline configuration
pipeline:
  enabled_steps:
    - string                    # List of steps: torch, import, dimp, validation, csv_conversion, parquet_conversion

# Retry strategy
retry:
  max_attempts: integer         # Max retry attempts (1-10, default: 5)
  initial_backoff_ms: integer   # Initial backoff in milliseconds (default: 1000)
  max_backoff_ms: integer       # Maximum backoff in milliseconds (default: 30000)

# Job configuration
jobs:
  jobs_dir: string              # Directory for job state and data (default: ./jobs)
```

## Service Options

### DIMP URL

**Key**: `services.dimp_url`
**Type**: String (URL)
**Required**: Yes (if DIMP step enabled)
**Default**: None

Endpoint for DIMP de-identification service.

```yaml
services:
  dimp_url: "http://localhost:8083/fhir"
```

For production:
```yaml
services:
  dimp_url: "https://dimp.prod.healthcare.org/api/fhir"
```

### CSV Conversion URL

**Key**: `services.csv_conversion_url`
**Type**: String (URL)
**Required**: No
**Default**: None
**Status**: Placeholder for future feature

Endpoint for CSV conversion service.

```yaml
services:
  csv_conversion_url: "http://localhost:9000/convert/csv"
```

### Parquet Conversion URL

**Key**: `services.parquet_conversion_url`
**Type**: String (URL)
**Required**: No
**Default**: None
**Status**: Placeholder for future feature

Endpoint for Parquet conversion service.

```yaml
services:
  parquet_conversion_url: "http://localhost:9000/convert/parquet"
```

### TORCH Configuration

**Key**: `services.torch`
**Type**: Object
**Required**: Yes (if TORCH step enabled)
**Default**: None

TORCH FHIR server connection details.

**Nested Options:**

- `base_url` (String): TORCH server URL
- `username` (String): TORCH username
- `password` (String): TORCH password

```yaml
services:
  torch:
    base_url: "https://torch.hospital.org"
    username: "researcher-name"
    password: "secure-password"
```

**Security**: Use environment variables for sensitive credentials:

```bash
export TORCH_USERNAME="researcher"
export TORCH_PASSWORD="secret"
```

Then in config:
```yaml
services:
  torch:
    base_url: "https://torch.hospital.org"
    username: "${TORCH_USERNAME}"
    password: "${TORCH_PASSWORD}"
```

## Pipeline Options

### Enabled Steps

**Key**: `pipeline.enabled_steps`
**Type**: Array of strings
**Required**: Yes
**Default**: None

List of pipeline steps to execute in order.

```yaml
pipeline:
  enabled_steps:
    - torch      # Optional: Extract from TORCH
    - import     # Required: Import FHIR data
    - dimp       # Optional: Pseudonymization
```

**Available Steps** (must be in order):
- `torch` - Extract from TORCH server
- `import` - Parse and validate FHIR data
- `dimp` - Pseudonymization via DIMP
- `validation` - Data quality validation (placeholder)
- `csv_conversion` - Convert to CSV (placeholder)
- `parquet_conversion` - Convert to Parquet (placeholder)

**Valid Sequences**:
```yaml
# Option A: Local files + DIMP
- import
- dimp

# Option B: TORCH + DIMP
- torch
- import
- dimp

# Option C: Full pipeline
- torch
- import
- dimp
- validation
- csv_conversion
```

## Retry Options

### Max Attempts

**Key**: `retry.max_attempts`
**Type**: Integer
**Range**: 1-10
**Default**: 5

Maximum number of automatic retry attempts for transient errors.

```yaml
retry:
  max_attempts: 3  # Fewer retries for fast-fail
```

Higher values = more resilience but longer wait times.

### Initial Backoff

**Key**: `retry.initial_backoff_ms`
**Type**: Integer (milliseconds)
**Range**: 100-5000
**Default**: 1000 (1 second)

Initial wait time before first retry.

```yaml
retry:
  initial_backoff_ms: 500  # Start with 500ms
```

### Max Backoff

**Key**: `retry.max_backoff_ms`
**Type**: Integer (milliseconds)
**Range**: 1000-60000
**Default**: 30000 (30 seconds)

Maximum wait time between retries.

```yaml
retry:
  max_backoff_ms: 10000  # Cap at 10 seconds
```

**Exponential Backoff Formula**:
```
wait_time = min(initial * (2 ^ attempt), max_backoff)
```

Example with defaults:
- Attempt 1: 1s
- Attempt 2: 2s
- Attempt 3: 4s
- Attempt 4: 8s
- Attempt 5: 16s
- Attempt 6+: 30s (capped)

## Job Options

### Jobs Directory

**Key**: `jobs.jobs_dir`
**Type**: String (directory path)
**Default**: `./jobs`

Directory for storing job state and data.

```yaml
jobs:
  jobs_dir: "./jobs"
```

For network storage:
```yaml
jobs:
  jobs_dir: "/mnt/shared/aether/jobs"
```

**Requirements**:
- Must be writable by Aether process
- Sufficient disk space for processed data
- Should be backed up regularly

## Complete Example Configurations

### Development Setup

```yaml
# aether.yaml - local development
services:
  dimp_url: "http://localhost:8083/fhir"

pipeline:
  enabled_steps:
    - import
    - dimp

retry:
  max_attempts: 3
  initial_backoff_ms: 500
  max_backoff_ms: 5000

jobs:
  jobs_dir: "./jobs"
```

### Production TORCH + DIMP

```yaml
# aether.yaml - production
services:
  torch:
    base_url: "https://torch.prod.healthcare.org"
    username: "${TORCH_USERNAME}"
    password: "${TORCH_PASSWORD}"
  dimp_url: "https://dimp.prod.healthcare.org/api/fhir"

pipeline:
  enabled_steps:
    - torch
    - import
    - dimp

retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000

jobs:
  jobs_dir: "/data/aether/jobs"
```

### Local Files Only

```yaml
# aether.yaml - local processing
pipeline:
  enabled_steps:
    - import

jobs:
  jobs_dir: "./output"
```

### High-Volume Processing

```yaml
# aether.yaml - optimized for large datasets
services:
  dimp_url: "http://dimp-cluster:8083/fhir"

pipeline:
  enabled_steps:
    - import
    - dimp

retry:
  max_attempts: 3
  initial_backoff_ms: 2000
  max_backoff_ms: 10000

jobs:
  jobs_dir: "/mnt/fast-storage/jobs"
```

## Environment Variable References

Configuration supports environment variable substitution:

```yaml
services:
  torch:
    base_url: "${TORCH_BASE_URL}"
    username: "${TORCH_USERNAME}"
    password: "${TORCH_PASSWORD}"
  dimp_url: "${DIMP_URL}"

jobs:
  jobs_dir: "${AETHER_DATA_DIR}/jobs"
```

Set environment variables:
```bash
export TORCH_BASE_URL="https://torch.hospital.org"
export TORCH_USERNAME="researcher"
export TORCH_PASSWORD="secret"
export DIMP_URL="http://localhost:8083/fhir"
export AETHER_DATA_DIR="/data/aether"
```

## Configuration Validation

Aether validates configuration on startup:

```bash
# Validate configuration without running
aether validate-config aether.yaml
```

**Common Validation Errors**:
- Missing required services for enabled steps
- Invalid directory paths
- Invalid retry values
- Malformed YAML

## Troubleshooting

### Config file not found

```
Error: configuration file not found: aether.yaml
```

Solution: Ensure `aether.yaml` exists in the working directory or specify path:
```bash
aether pipeline start --config /etc/aether/config.yaml query.crtdl
```

### Service unreachable

```
Error: DIMP service unreachable: connection refused
```

Solution: Verify service URL and connectivity:
```bash
curl http://localhost:8083/fhir/
```

### Validation failed

```
Error: configuration validation failed: DIMP URL required for enabled step 'dimp'
```

Solution: Add required service configuration:
```yaml
services:
  dimp_url: "http://localhost:8083/fhir"
```

## Next Steps

- [Configuration Guide](../getting-started/configuration.md) - Configuration introduction
- [CLI Commands](./cli-commands.md) - Command reference
- [Getting Started](../getting-started/quick-start.md) - Quick start guide
