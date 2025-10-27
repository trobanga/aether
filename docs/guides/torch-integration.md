# TORCH Integration

TORCH is a FHIR server for clinical research data. Aether can directly extract data from TORCH using CRTDL (Clinical Resource Transfer Definition Language) query files.

## Overview

TORCH Integration allows you to:
- Query cohorts from TORCH FHIR servers using standardized CRTDL queries
- Automatically extract patient data based on clinical criteria
- Import extracted FHIR resources into Aether for processing
- Apply pseudonymization or other transformations to sensitive data

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

## Using TORCH Integration

### Basic TORCH Query

Run a simple query to extract data from TORCH:

```bash
aether pipeline start my_cohort.crtdl
```

Aether will:
1. Connect to the configured TORCH server
2. Execute the CRTDL query to identify matching patients
3. Extract all FHIR resources for those patients
4. Import the data into Aether
5. Apply any enabled pipeline steps

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

## Troubleshooting

### "TORCH server unreachable"
- Verify `services.torch.base_url` is correct and accessible
- Check network connectivity from your machine to TORCH server
- Test with curl: `curl -u username:password https://torch.hospital.org/fhir/`

### "Authentication failed"
- Verify username and password in configuration
- Ensure credentials have appropriate permissions on TORCH server
- Check for password expiration or account lockout

### "CRTDL query syntax error"
- Validate CRTDL query file syntax
- Check field names and data types match TORCH schema
- Review TORCH documentation for query syntax

### "No patients matched"
- Review cohort inclusion/exclusion criteria
- Verify TORCH server contains matching data
- Test query manually on TORCH interface

## Next Steps

- [Configuration Guide](../getting-started/configuration.md) - Set up Aether fully
- [DIMP Pseudonymization](./dimp-pseudonymization.md) - Protect sensitive data
- [Pipeline Steps](./pipeline-steps.md) - Understand the full pipeline
- [CLI Commands Reference](../api-reference/cli-commands.md) - All available commands
