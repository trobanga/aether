# Quickstart: Aether DUP Pipeline CLI

**Purpose**: Get started with Aether to process FHIR medical data through the Data Use Process (DUP) pipeline.

## Prerequisites

- Go 1.21+ installed
- TORCH extraction output (local directory or download URL)
- Optional: DIMP service running (for pseudonymization)
- Optional: Conversion services running (for CSV/Parquet output)

## Installation

### From Source

```bash
git clone https://github.com/user/aether.git
cd aether
go build -o aether cmd/aether/main.go
sudo mv aether /usr/local/bin/
```

### From Release

```bash
wget https://github.com/user/aether/releases/latest/download/aether-linux-amd64
chmod +x aether-linux-amd64
sudo mv aether-linux-amd64 /usr/local/bin/aether
```

### Verify Installation

```bash
aether --version
# Output: Aether v1.0.0
```

---

## Configuration

### 1. Create Project Configuration

```bash
cp config/aether.example.yaml aether.yaml
```

### 2. Edit `aether.yaml`

```yaml
services:
  dimp_url: "http://localhost:8083/fhir"
  csv_conversion_url: "http://localhost:9000/convert/csv"
  parquet_conversion_url: "http://localhost:9000/convert/parquet"

pipeline:
  enabled_steps:
    - import
    - dimp
    - csv_conversion
    - parquet_conversion

retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000

jobs_dir: "./jobs"
```

**Notes**:
- Comment out steps you don't need (e.g., remove `dimp` if no pseudonymization required)
- Service URLs can be left empty if corresponding step is disabled
- `jobs_dir` can be absolute or relative path

---

## Basic Workflow

### Step 1: Start a Pipeline

**From local directory**:
```bash
aether pipeline start --input /path/to/torch/output
```

**From download URL**:
```bash
aether pipeline start --input https://example.com/torch/export/job-123
```

**Output**:
```
✓ Created pipeline job: abc-123-def
✓ Importing FHIR data...
  [=============>    ] 75% (150/200 files, 2.3GB)
```

The command returns immediately. The job ID is displayed for tracking.

### Step 2: Monitor Progress

```bash
aether pipeline status abc-123-def
```

**Output**:
```
Job ID: abc-123-def
Status: in_progress
Current Step: dimp
Created: 2025-10-08 10:00:00

Steps:
  ✓ import        - completed (200 files, 3.5GB) [0 retries]
  → dimp          - in_progress (150/200 files) [1 retry]
    csv_conversion - pending
    parquet_conversion - pending

Last Updated: 2025-10-08 10:15:30
```

### Step 3: Wait for Completion

**Option A: Poll manually**
```bash
watch -n 5 aether pipeline status abc-123-def
```

**Option B: Blocking wait** (future feature)
```bash
aether pipeline wait abc-123-def
```

### Step 4: Access Results

```bash
ls jobs/abc-123-def/
# import/          - Original TORCH files
# pseudonymized/   - DIMP output
# csv/             - CSV conversion output
# parquet/         - Parquet conversion output
# state.json       - Job metadata
```

---

## Advanced Usage

### Resume a Failed Pipeline

If a step fails due to transient error and automatic retries are exhausted:

```bash
# Check error details
aether pipeline status abc-123-def

# Output:
# Step: dimp - failed (non-transient error)
# Error: HTTP 400 - Malformed FHIR resource at line 42

# Fix the issue (e.g., correct malformed data), then:
aether pipeline continue abc-123-def
```

### Manual Step Execution

Run a specific step manually (bypasses config, expert mode):

```bash
aether job run abc-123-def --step dimp
```

**Use cases**:
- Retry failed step after manual data correction
- Skip validation step: run conversion directly
- Re-run conversion with different service

### List All Jobs

```bash
aether job list
```

**Output**:
```
JOB ID          STATUS      STEP            CREATED             RETRIES
abc-123-def     completed   -               2025-10-08 10:00    0
xyz-456-ghi     in_progress csv_conversion  2025-10-08 11:30    2
old-789-jkl     failed      dimp            2025-10-07 14:20    5
```

**Filters**:
```bash
aether job list --status failed
aether job list --since 2025-10-01
```

### Override Configuration

**Override service URL at runtime**:
```bash
aether pipeline start --input /data \
  --dimp-url http://staging-dimp.example.com/fhir
```

**Enable/disable steps**:
```bash
aether pipeline start --input /data \
  --skip dimp \
  --enable validation
```

---

## Troubleshooting

### Pipeline Stuck in "in_progress"

**Check logs**:
```bash
cat jobs/abc-123-def/state.json
```

**Look for**:
- `last_error` field
- `retry_count` approaching max attempts

**Solution**:
- If transient error: wait for automatic retry
- If non-transient: fix data/config, then `aether pipeline continue <job-id>`

### Service Connection Errors

**Error**: `Error: failed to connect to DIMP service at http://localhost:8083`

**Solutions**:
1. Verify service is running: `curl http://localhost:8083/fhir/$de-identify`
2. Check `aether.yaml` has correct URL
3. Override at runtime: `--dimp-url http://correct-host:8083/fhir`

### Large File Performance

For 10GB+ datasets:

1. **Increase timeouts** (in `aether.yaml`):
```yaml
retry:
  max_attempts: 3  # Reduce retries
  timeout_seconds: 300  # 5 minutes per request
```

2. **Monitor disk space**:
```bash
df -h jobs/
```

3. **Process in batches**: Split TORCH output into smaller jobs

### Disk Space Issues

**Error**: `Error: no space left on device`

**Solutions**:
1. Clean old jobs: `rm -rf jobs/old-job-id`
2. Change jobs directory: `aether --jobs-dir /mnt/large-disk pipeline start ...`
3. Stream mode (future): Process without storing intermediate files

---

## Testing the Setup

### 1. Start Test Services (Docker Compose)

```bash
cd test-environment/
docker-compose up -d
# Starts: DIMP service, conversion services, test FHIR server
```

### 2. Download Test Data

```bash
wget https://example.com/test-torch-export.tar.gz
tar -xzf test-torch-export.tar.gz
```

### 3. Run Test Pipeline

```bash
aether pipeline start --input ./test-torch-export
```

### 4. Verify Results

```bash
# Check final status
aether pipeline status <job-id>

# Verify CSV output
head jobs/<job-id>/csv/Patient.csv

# Count pseudonymized resources
wc -l jobs/<job-id>/pseudonymized/*.ndjson
```

---

## Next Steps

- **Configure for production**: Update `aether.yaml` with production service URLs
- **Automate with cron**: Schedule regular TORCH extractions + Aether processing
- **Monitor with scripts**: Integrate `aether job list` into monitoring dashboards
- **Backup job data**: Regularly backup `jobs/` directory

---

## Getting Help

```bash
aether --help
aether pipeline --help
aether pipeline start --help
```

**Documentation**: https://github.com/user/aether/docs
**Issues**: https://github.com/user/aether/issues
