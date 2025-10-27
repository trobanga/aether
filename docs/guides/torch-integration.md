# TORCH Integration

TORCH is a FHIR server for clinical research data. Aether can directly extract data from TORCH using CRTDL (Clinical Resource Transfer Definition Language) query files.

## Overview

**What is TORCH?**

TORCH is a FHIR-based data extraction service that allows researchers to define patient cohorts using CRTDL query files and retrieve matching FHIR data automatically.

**TORCH Integration allows you to:**
- Query cohorts from TORCH FHIR servers using standardized CRTDL queries
- Automatically extract patient data based on clinical criteria with data minimization
- Import extracted FHIR resources into Aether for processing
- Combine with DIMP for pseudonymization of sensitive data (optional)
- Resume and reprocess previously extracted data

## Prerequisites

- Access to a TORCH server with valid credentials
- A CRTDL query file defining your cohort criteria
- TORCH server URL, username, and password

## Configuration

### 1. Add TORCH Credentials

Add your TORCH credentials to `aether.yaml`:

```yaml
services:
  torch:
    base_url: "https://torch.hospital.org"
    username: "researcher-username"
    password: "secure-password"
```

**Security Note:** Consider using environment variables for sensitive credentials:

```bash
export TORCH_USERNAME="researcher"
export TORCH_PASSWORD="secure-password"
```

Then reference in `aether.yaml` (if supported):

```yaml
services:
  torch:
    base_url: "https://torch.hospital.org"
    username: "${TORCH_USERNAME}"
    password: "${TORCH_PASSWORD}"
```

### 2. Enable TORCH in Pipeline

```yaml
pipeline:
  enabled_steps:
    - torch    # Extract from TORCH
    - import   # Import extracted data
    - dimp     # Optional: pseudonymize
```

## Creating CRTDL Query Files

A CRTDL query file defines the cohort selection criteria. Example:

```json
{
  "description": "Patients with diabetes diagnosis",
  "criteria": {
    "resourceType": "Condition",
    "code": {
      "coding": [{
        "system": "http://snomed.info/sct",
        "code": "73211009"
      }]
    }
  }
}
```

## Input Methods

Aether supports multiple ways to work with TORCH data:

### 1. **CRTDL File Extraction** (recommended for new queries)

Submit a CRTDL file to TORCH and automatically download results:

```bash
aether pipeline start cohort-query.crtdl
```

Aether will:
1. Connect to the configured TORCH server
2. Submit the CRTDL query
3. Poll extraction status until complete
4. Download resulting FHIR NDJSON files
5. Continue with pipeline processing

### 2. **TORCH Result URL** (for reusing existing extractions)

Skip extraction and download files directly from a TORCH result URL:

```bash
aether pipeline start http://localhost:8080/fhir/result/abc123
```

This is useful for:
- Resuming or reprocessing previously extracted data
- Sharing extraction results with other researchers
- Testing pipeline changes on existing data

### 3. **Backward Compatibility**

Existing workflows using local directories or HTTP URLs continue to work:

```bash
# Still supported
aether pipeline start ./test-data/
aether pipeline start https://example.com/fhir/export
```

## Using TORCH Integration

### Basic TORCH Query

Run a simple query to extract data from TORCH:

```bash
aether pipeline start my_cohort.crtdl
```

### With Pseudonymization

Extract data and pseudonymize it in one command:

```yaml
# aether.yaml
services:
  torch:
    base_url: "https://torch.hospital.org"
    username: "researcher"
    password: "secret"
  dimp_url: "http://localhost:8083/fhir"

pipeline:
  enabled_steps:
    - torch
    - import
    - dimp  # Apply pseudonymization

jobs:
  jobs_dir: "./jobs"
```

Run:
```bash
aether pipeline start my_cohort.crtdl
```

### Monitoring Progress

Monitor your TORCH extraction:

```bash
# Check job status
aether job list

# Get details for specific job
aether pipeline status <job-id>

# View logs
aether job logs <job-id>
```

## Common CRTDL Patterns

### Patient Demographics

Extract specific patient age range:

```json
{
  "description": "Patients aged 18-65",
  "criteria": {
    "resourceType": "Patient",
    "birthDate": {
      "$gte": "1960-01-01",
      "$lte": "2006-01-01"
    }
  }
}
```

### Diagnosis-Based Cohorts

Extract patients with specific diagnoses:

```json
{
  "description": "Hypertension patients",
  "criteria": {
    "resourceType": "Condition",
    "code": {
      "coding": [{
        "system": "http://snomed.info/sct",
        "code": "59621000"
      }]
    },
    "clinicalStatus": "active"
  }
}
```

### Medication-Based Cohorts

Extract patients on specific medications:

```json
{
  "description": "Patients on Metformin",
  "criteria": {
    "resourceType": "MedicationStatement",
    "medicationReference": {
      "code": {
        "coding": [{
          "system": "http://www.nlm.nih.gov/research/umls/rxnorm",
          "code": "6809"
        }]
      }
    }
  }
}
```

## Advanced Configuration

### Extraction Timeout

Configure how long Aether waits for TORCH extractions to complete:

```yaml
services:
  torch:
    base_url: "https://torch.hospital.org"
    username: "researcher"
    password: "secret"
    extraction_timeout_minutes: 30    # Default: 30
    polling_interval_seconds: 5        # Default: 5
    max_polling_interval_seconds: 30   # Default: 30
```

- `extraction_timeout_minutes`: Maximum wait time for extraction (adjust for large cohorts)
- `polling_interval_seconds`: Initial poll interval (increases exponentially up to max)
- `max_polling_interval_seconds`: Maximum poll interval between checks

### File Server Configuration

If your TORCH has a separate file download server:

```yaml
services:
  torch:
    base_url: "https://torch.hospital.org"
    file_server_url: "http://torch-files.hospital.org"
    username: "researcher"
    password: "secret"
```

## Error Handling

Aether implements robust error handling for TORCH operations:

- **Server unreachable**: Clear error within 5 seconds
- **Authentication failure**: Fails early with credential error
- **Extraction timeout**: Configurable timeout (default 30 minutes)
- **Empty results**: Gracefully handles zero-patient cohorts
- **Malformed CRTDL**: Validates syntax before submission
- **Connection lost during download**: Automatic retry with exponential backoff

## Troubleshooting

### "TORCH server unreachable"
- Verify `services.torch.base_url` is correct and accessible
- Check network connectivity from your machine to TORCH server
- Test with curl: `curl -u username:password https://torch.hospital.org/fhir/`

### "Authentication failed"
- Verify username and password in configuration
- Ensure credentials have appropriate permissions on TORCH server
- Check for password expiration or account lockout
- Verify you have access to the TORCH system

### "CRTDL query syntax error"
- Validate CRTDL query file syntax (should be valid JSON)
- Check field names and data types match TORCH schema
- Review TORCH documentation for query syntax
- Test query manually on TORCH interface

### "No patients matched"
- Review cohort inclusion/exclusion criteria
- Verify TORCH server contains matching data
- Test query manually on TORCH interface
- Ensure query criteria are not too restrictive

### "Extraction takes longer than expected"
- Large cohorts may need higher `extraction_timeout_minutes`
- Check TORCH server logs for processing status
- Verify network connectivity remains stable
- Consider splitting into smaller cohorts

### "How long do TORCH extractions take?"
- Depends on cohort size and server load
- Aether polls every 5 seconds (configurable) with default 30-minute timeout
- Large cohorts may need a higher timeout in configuration
- Monitor progress with `aether job list` and `aether pipeline status <job-id>`

## Next Steps

- [Configuration Guide](../getting-started/configuration.md) - Set up Aether fully
- [DIMP Pseudonymization](./dimp-pseudonymization.md) - Protect sensitive data
- [Pipeline Steps](./pipeline-steps.md) - Understand the full pipeline
- [CLI Commands Reference](../api-reference/cli-commands.md) - All available commands
