# Pipeline Steps

The Aether pipeline is a modular, configurable series of processing steps that work together to extract, transform, and protect healthcare data.

## Pipeline Architecture

### High-Level Overview

```
Start
  ↓
[TORCH] - Extract from TORCH (optional)
  ↓
[Import] - Load and parse FHIR data
  ↓
[DIMP] - Pseudonymize/de-identify (optional)
  ↓
[Validate] - Verify data quality (placeholder)
  ↓
[CSV/Parquet] - Convert format (placeholders)
  ↓
Output
```

### Execution Model

- **Sequential**: Steps run in order (one completes before the next starts)
- **Resilient**: Failed steps can trigger automatic retries
- **Resumable**: Resume failed pipelines without reprocessing completed steps
- **Monitored**: Real-time progress tracking and logging
- **Configurable**: Enable/disable steps based on requirements

### Configuration

Steps are configured in `aether.yaml`:

```yaml
pipeline:
  enabled_steps:
    - torch
    - import
    - dimp
```

Only enabled steps execute; others are skipped.

## Available Pipeline Steps

### 1. TORCH Step

**Purpose**: Extract FHIR data from TORCH server using CRTDL queries.

**Requires**:
- TORCH server credentials
- CRTDL query file

**Configuration**:
```yaml
services:
  torch:
    base_url: "https://torch.hospital.org"
    username: "researcher"
    password: "secret"

pipeline:
  enabled_steps:
    - torch
```

**Input**: CRTDL query file (JSON)
**Output**: Raw FHIR NDJSON data

**Example**:
```bash
aether pipeline start my_cohort.crtdl
```

See [TORCH Integration](./torch-integration.md) for details.

### 2. Import Step

**Purpose**: Parse, validate, and normalize FHIR data in NDJSON format.

**Requires**:
- FHIR NDJSON files or TORCH extraction

**Configuration**:
```yaml
pipeline:
  enabled_steps:
    - import
```

**Input**: FHIR NDJSON files
**Output**: Validated, normalized FHIR bundles

**Features**:
- Validates FHIR schema compliance
- Normalizes resource identifiers
- Handles duplicates
- Reports validation errors

**Example**:
```bash
# Import from local files
aether pipeline start /path/to/fhir/files/

# Import from TORCH
aether pipeline start query.crtdl
```

### 3. DIMP Step

**Purpose**: De-identify and pseudonymize FHIR data via DIMP service.

**Requires**:
- DIMP service running
- Import step to complete first

**Configuration**:
```yaml
services:
  dimp_url: "http://localhost:8083/fhir"

pipeline:
  enabled_steps:
    - import
    - dimp
```

**Input**: Validated FHIR bundles
**Output**: De-identified FHIR data with pseudonyms

**Features**:
- Removes/masks personally identifiable information
- Generates consistent pseudonyms
- Maintains clinical data utility
- Audit trail of changes

**Example**:
```bash
aether pipeline start /path/to/fhir/
```

See [DIMP Pseudonymization](./dimp-pseudonymization.md) for details.

### 4. Validation Step (Placeholder)

**Purpose**: Validate data quality and FHIR compliance.

**Status**: Not yet implemented

**Configuration**:
```yaml
pipeline:
  enabled_steps:
    - import
    - validation
```

**Planned Features**:
- FHIR profile validation
- Data quality checks
- Missing field detection
- Cross-reference validation

### 5. CSV Conversion (Placeholder)

**Purpose**: Convert FHIR data to CSV format for analysis.

**Status**: Not yet implemented

**Requires**: CSV conversion service

**Configuration**:
```yaml
services:
  csv_conversion_url: "http://localhost:9000/convert/csv"

pipeline:
  enabled_steps:
    - import
    - csv_conversion
```

### 6. Parquet Conversion (Placeholder)

**Purpose**: Convert FHIR data to Parquet columnar format for big data analysis.

**Status**: Not yet implemented

**Requires**: Parquet conversion service

**Configuration**:
```yaml
services:
  parquet_conversion_url: "http://localhost:9000/convert/parquet"

pipeline:
  enabled_steps:
    - import
    - parquet_conversion
```

## Step Dependencies

The order of steps matters:

```
Must be in order:
1. TORCH (if using) → 2. Import → 3. Transformation (DIMP) → 4-6. Output formats
```

**Valid pipelines**:
```yaml
# Option A: Local files only
- import
- dimp

# Option B: TORCH + DIMP
- torch
- import
- dimp

# Option C: Local files with format conversion (when available)
- import
- dimp
- csv_conversion
- parquet_conversion

# Option D: TORCH to multiple formats
- torch
- import
- dimp
- csv_conversion
- parquet_conversion
```

**Invalid pipelines**:
```yaml
# ❌ DIMP before Import
- dimp
- import

# ❌ Conversion without Import
- csv_conversion

# ❌ TORCH + Import skipped
- torch
- dimp
```

## Error Handling & Retries

### Automatic Retries

Transient errors (network timeouts, temporary service unavailability) trigger automatic retries:

```yaml
retry:
  max_attempts: 5              # Maximum retry attempts
  initial_backoff_ms: 1000     # Start with 1 second wait
  max_backoff_ms: 30000        # Cap at 30 seconds
```

Exponential backoff: 1s → 2s → 4s → 8s → 16s → 30s

### Manual Intervention

Permanent errors require manual intervention:
- Invalid CRTDL query
- Missing input files
- Service configuration errors

### Resuming Failed Pipelines

Resume without reprocessing completed steps:

```bash
# See failed job
aether job list

# Resume from where it failed
aether pipeline resume <job-id>

# The pipeline will skip already-completed steps
```

## Performance Considerations

### Large Dataset Processing

For datasets >1GB:

1. **Monitor resources**:
   ```bash
   # Check progress and memory usage
   aether pipeline status <job-id>
   ```

2. **Batch processing**:
   ```bash
   # Process in smaller batches
   aether pipeline start /data/fhir/batch1/
   aether pipeline start /data/fhir/batch2/
   ```

3. **Parallel jobs**:
   ```bash
   # Run multiple pipelines in parallel
   aether pipeline start /data/fhir/batch1/ &
   aether pipeline start /data/fhir/batch2/ &
   ```

### Step Performance Tips

**Import**:
- Large NDJSON files process incrementally
- No additional tuning needed

**DIMP**:
- Scales with data size
- May need service tuning for 100MB+ datasets
- Consider batch processing

## Monitoring Pipeline Execution

```bash
# List all jobs
aether job list

# Get detailed status
aether pipeline status <job-id>

# View real-time logs
aether job logs <job-id>

# Stream logs continuously
aether job logs <job-id> --follow
```

## Next Steps

- [TORCH Integration](./torch-integration.md) - Set up TORCH extraction
- [DIMP Pseudonymization](./dimp-pseudonymization.md) - Protect patient privacy
- [Configuration](../getting-started/configuration.md) - Configure your pipeline
- [CLI Commands](../api-reference/cli-commands.md) - All available commands
