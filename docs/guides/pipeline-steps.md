# Pipeline Steps

The Aether pipeline is a modular, configurable series of processing steps that work together to extract, transform, and protect healthcare data.

## Pipeline Architecture

### High-Level Overview

```
Start
  ↓
[Import Step] - One of:
  • torch: Extract from TORCH via CRTDL
  • local_import: Load from local directory
  • http_import: Load from HTTP URL
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
    - torch          # or local_import or http_import
    - dimp
```

**Important**: The first step must always be one of the import step types:
- `torch` - Import from TORCH server via CRTDL
- `local_import` - Import from local directory
- `http_import` - Import from HTTP URL

Only enabled steps execute; others are skipped.

## Available Pipeline Steps

### Import Steps (Step 1 - Required)

**One of the following import steps must be the first step in your pipeline:**

#### 1a. TORCH Import (`torch`)

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
    - dimp  # optional next steps
```

**Input**: CRTDL query file (JSON)
**Output**: FHIR NDJSON data in jobs directory

**Example**:
```bash
aether pipeline start my_cohort.crtdl
```

See [TORCH Integration](./torch-integration.md) for details.

#### 1b. Local Import (`local_import`)

**Purpose**: Load and validate FHIR data from local directory.

**Requires**:
- Local directory containing FHIR NDJSON files

**Configuration**:
```yaml
pipeline:
  enabled_steps:
    - local_import
    - dimp  # optional next steps
```

**Input**: Path to directory with FHIR NDJSON files
**Output**: Validated FHIR data in jobs directory

**Features**:
- Validates FHIR schema compliance
- Handles multiple NDJSON files
- Reports validation errors

**Example**:
```bash
aether pipeline start /path/to/fhir/files/
```

#### 1c. HTTP Import (`http_import`)

**Purpose**: Download and validate FHIR data from HTTP/HTTPS URL.

**Requires**:
- HTTP/HTTPS URL to FHIR NDJSON file or endpoint

**Configuration**:
```yaml
pipeline:
  enabled_steps:
    - http_import
    - dimp  # optional next steps
```

**Input**: HTTP/HTTPS URL to FHIR data
**Output**: Downloaded and validated FHIR data in jobs directory

**Features**:
- Downloads FHIR data from remote URLs
- Validates FHIR schema compliance
- Supports authentication (if configured)

**Example**:
```bash
aether pipeline start https://fhir.server.org/export/Patient.ndjson
```

### 2. DIMP Step

**Purpose**: De-identify and pseudonymize FHIR data via DIMP service.

**Requires**:
- DIMP service running
- One of the import steps (torch, local_import, or http_import) to complete first

**Configuration**:
```yaml
services:
  dimp:
    url: "http://localhost:32861/fhir"
    bundle_split_threshold_mb: 10

pipeline:
  enabled_steps:
    - local_import  # or torch or http_import
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

### 3. Validation Step (Placeholder)

**Purpose**: Validate data quality and FHIR compliance.

**Status**: Not yet implemented

**Configuration**:
```yaml
pipeline:
  enabled_steps:
    - local_import  # or torch or http_import
    - validation
```

**Planned Features**:
- FHIR profile validation
- Data quality checks
- Missing field detection
- Cross-reference validation

### 4. CSV Conversion (Placeholder)

**Purpose**: Convert FHIR data to CSV format for analysis.

**Status**: Not yet implemented

**Requires**: CSV conversion service

**Configuration**:
```yaml
services:
  csv_conversion_url: "http://localhost:9000/convert/csv"

pipeline:
  enabled_steps:
    - local_import  # or torch or http_import
    - csv_conversion
```

### 5. Parquet Conversion (Placeholder)

**Purpose**: Convert FHIR data to Parquet columnar format for big data analysis.

**Status**: Not yet implemented

**Requires**: Parquet conversion service

**Configuration**:
```yaml
services:
  parquet_conversion_url: "http://localhost:9000/convert/parquet"

pipeline:
  enabled_steps:
    - local_import  # or torch or http_import
    - parquet_conversion
```

## Step Dependencies

The order of steps matters:

```
Must be in order:
1. Import Step (torch OR local_import OR http_import) → 2. Transformation (DIMP) → 3-5. Output formats
```

**Valid pipelines**:
```yaml
# Option A: Local files only
- local_import
- dimp

# Option B: TORCH + DIMP
- torch
- dimp

# Option C: HTTP import with format conversion (when available)
- http_import
- dimp
- csv_conversion
- parquet_conversion

# Option D: TORCH to multiple formats
- torch
- dimp
- csv_conversion
- parquet_conversion
```

**Invalid pipelines**:
```yaml
# ❌ DIMP without import step
- dimp

# ❌ Multiple import steps
- torch
- local_import

# ❌ Conversion without import step
- csv_conversion

# ❌ No import step first
- validation
- local_import
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
aether pipeline continue <job-id>

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
