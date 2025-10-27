# DIMP Pseudonymization

DIMP (Data Integration and Management Platform) provides de-identification and pseudonymization services for FHIR healthcare data using the **FHIR Pseudonymizer**, protecting patient privacy while preserving data utility for research.

## Overview

**What is DIMP?**

DIMP (Data Integration and Management Platform) provides de-identification and pseudonymization services for FHIR healthcare data using the **FHIR Pseudonymizer**, protecting patient privacy while preserving data utility for research.

**DIMP integration in Aether enables you to:**

- **De-identify** sensitive patient information (names, addresses, birthdates, etc.)
- **Pseudonymize** records with consistent, reversible identifiers
- **Maintain data utility** for research purposes
- **Comply** with healthcare privacy regulations (GDPR, HIPAA, etc.)
- **Generate** audit trails of all modifications
- **Scale** pseudonymization for large datasets automatically

## Prerequisites

- DIMP service running and accessible
- FHIR data in NDJSON format
- DIMP HTTP endpoint configured in Aether

## Configuration

### 1. Configure DIMP Endpoint

Add the DIMP service URL to `aether.yaml`:

```yaml
services:
  dimp:
    url: "http://localhost:32861/fhir"
    bundle_split_threshold_mb: 10
```

For production environments:

```yaml
services:
  dimp:
    url: "https://dimp.healthcare.org/fhir"
    bundle_split_threshold_mb: 50
```

The `bundle_split_threshold_mb` setting controls automatic splitting of large FHIR Bundles to prevent HTTP 413 errors when sending to DIMP (range: 1-100 MB).

### 2. Enable DIMP in Pipeline

```yaml
pipeline:
  enabled_steps:
    - import  # Import FHIR data
    - dimp    # Apply pseudonymization
```

The order matters: DIMP should run after data import.

### 3. Full Configuration Example

```yaml
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

## How Pseudonymization Works

### 1. Data Processing Flow

```
Raw FHIR Data
    ↓
Import Step (parse, validate)
    ↓
DIMP Pseudonymization
  - Extract identifiable elements
  - Generate pseudonyms
  - Create mapping
  - Apply transformations
    ↓
De-identified Data
```

### 2. What Gets Pseudonymized

DIMP typically de-identifies:

- **Patient names** → Pseudonyms (PT_00001, PT_00002, etc.)
- **Birth dates** → Year of birth or age ranges
- **Contact information** → Removed
- **Street addresses** → Postal codes or general location
- **Medical record numbers** → New identifiers
- **Social security numbers** → Removed
- **Phone numbers** → Removed
- **Email addresses** → Removed

### 3. What Is Preserved

DIMP preserves:

- Patient demographics (age, gender for research)
- Diagnosis codes (ICD-10, SNOMED CT)
- Procedure codes
- Medication information
- Laboratory values
- Clinical narratives (with terms removed)
- Relationships between records for same patient

## Using DIMP Pseudonymization

### Basic Pseudonymization Workflow

1. **Prepare your data:**

```bash
# Ensure FHIR NDJSON files are ready
ls /data/fhir/*.ndjson
```

2. **Configure aether.yaml:**

```yaml
services:
  dimp:
    url: "http://localhost:32861/fhir"
    bundle_split_threshold_mb: 10

pipeline:
  enabled_steps:
    - import
    - dimp

jobs_dir: "./jobs"
```

3. **Run the pipeline:**

```bash
aether pipeline start /data/fhir/
```

4. **Monitor progress:**

```bash
# Check job status
aether job list

# Get details
aether pipeline status <job-id>

# View logs
aether job logs <job-id>
```

5. **Access results:**

De-identified data is stored in the jobs directory:

```
jobs/
└── <job-id>/
    ├── status.json           # Job metadata
    ├── import_results.ndjson # Imported data
    └── dimp_results.ndjson   # De-identified data
```

### Understanding TORCH vs DIMP

**TORCH** (Data extraction service):
- Extracts FHIR data from TORCH servers based on CRTDL queries
- Applies **TORCH minimization** (extracts only needed fields)
- Returns raw identifiable data

**DIMP** (De-identification service):
- Applies **DIMP pseudonymization** (removes/replaces PII)
- De-identifies already-extracted data
- Returns pseudonymized, de-identified data

**Combined workflow**: Extract with TORCH → Pseudonymize with DIMP

### TORCH + DIMP Workflow

Combine TORCH extraction with pseudonymization:

```yaml
services:
  torch:
    base_url: "https://torch.hospital.org"
    username: "researcher"
    password: "secret"
    extraction_timeout_minutes: 30
    polling_interval_seconds: 5
    max_polling_interval_seconds: 30
  dimp:
    url: "http://localhost:32861/fhir"
    bundle_split_threshold_mb: 10

pipeline:
  enabled_steps:
    - import  # Import extracted TORCH data (minimized but identifiable)
    - dimp    # Apply DIMP pseudonymization (de-identify)

retry:
  max_attempts: 5
  initial_backoff_ms: 1000
  max_backoff_ms: 30000

jobs_dir: "./jobs"
```

Run:

```bash
aether pipeline start my_cohort.crtdl
```

This automatically:
1. Extracts data from TORCH using CRTDL query
2. Applies TORCH minimization (extracts only specified fields)
3. Imports the minimized data
4. Applies DIMP pseudonymization (removes/replaces PII)
5. Outputs de-identified, pseudonymized data

This provides **defense-in-depth** privacy: TORCH minimization reduces initial exposure, DIMP pseudonymization provides additional privacy protection.

## Best Practices

### 1. Always Test First

Test with a small sample before processing large datasets:

```bash
# Use a sample of data
aether pipeline start /data/fhir/sample/
```

### 2. Preserve Mappings

Keep pseudonym mappings in secure storage:

```bash
# The job directory contains mappings
# Back them up securely
cp -r jobs/<job-id>/ /secure/backup/
```

### 3. Version Control

Track your pseudonymization configurations:

```bash
git add aether.yaml
git commit -m "Update DIMP configuration"
```

### 4. Audit Trail

Review logs for compliance auditing:

```bash
# View full audit trail
aether job logs <job-id> | grep -i "audit\|processed\|error"
```

### 5. Data Retention

Plan data lifecycle:

```yaml
# Example: Keep job data for 90 days
jobs:
  jobs_dir: "./jobs"
  retention_days: 90  # Clean up after retention period
```

## Privacy Considerations

### 1. Secure Storage

- Store pseudonym mappings in encrypted storage
- Restrict access to sensitive files
- Use file permissions: `chmod 600 mappings.json`

### 2. Secure Transmission

- Use HTTPS for DIMP communication
- Enable TLS/SSL verification
- Use strong authentication

### 3. Regulatory Compliance

DIMP helps comply with:

- **GDPR**: Right to be forgotten, data minimization
- **HIPAA**: Safe harbor de-identification standards
- **FHIR**: Security and privacy profiles
- **HIPAA Breach Notification Rule**: Protect against re-identification

### 4. Secondary Use

Even with pseudonymization, secondary use requires:

- Explicit research protocol approval
- IRB/Ethics committee review
- Data use agreements
- Publication restrictions

## Troubleshooting

### "DIMP service unavailable"
- Verify DIMP is running: `curl http://localhost:8083/health`
- Check `services.dimp_url` in configuration
- Check network connectivity and firewall rules

### "Pseudonymization failed"
- Check DIMP logs for errors
- Verify FHIR data is valid: Use a FHIR validator first
- Check DIMP has sufficient resources (disk space, memory)

### "Performance is slow"
- DIMP may need tuning for large datasets
- Consider processing in batches
- Check system resources (CPU, RAM, disk I/O)

### "Inconsistent pseudonyms"
- Ensure same mapping is used consistently
- Use the same job for all related data
- Do not re-process the same data with different settings

## Next Steps

- [Configuration Guide](../getting-started/configuration.md) - Full configuration reference
- [TORCH Integration](./torch-integration.md) - Combine TORCH + DIMP
- [Pipeline Steps](./pipeline-steps.md) - Understand the pipeline
- [CLI Commands](../api-reference/cli-commands.md) - Available commands
