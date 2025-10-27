# Quick Start

Get up and running with Aether in just a few minutes.

## Prerequisites

Before starting, ensure you have:
- **Aether installed** - See [Installation Guide](./installation.md)
- **Go 1.21+** (if building from source)
- **Optional**: TORCH server access or DIMP service (for advanced features)

## 5-Minute Setup

### 1. Verify Installation

```bash
aether --help
```

You should see the Aether CLI help output.

### 2. Create a Configuration File

Create `aether.yaml` in your project directory:

For basic local testing (no external services):

```yaml
# aether.yaml
pipeline:
  enabled_steps:
    - import

jobs_dir: "./jobs"
```

For local FHIR data with pseudonymization:

```yaml
# aether.yaml
pipeline:
  enabled_steps:
    - import
    - dimp

services:
  dimp_url: "http://localhost:8083/fhir"

jobs_dir: "./jobs"
```

### 3. Run Your First Pipeline

**Option A: Process local FHIR data**

Create a test data directory:

```bash
mkdir -p test-data
```

Start a pipeline:

```bash
aether pipeline start ./test-data/
```

**Option B: From HTTP URL**

```bash
aether pipeline start https://fhir.server.org/export/Patient.ndjson
```

**Option C: From TORCH with CRTDL query**

First, set TORCH credentials in `aether.yaml`:

```yaml
services:
  torch:
    base_url: "http://torch.hospital.org"
    username: "your-username"
    password: "your-password"
```

Then run:

```bash
aether pipeline start query.crtdl
```

### 4. Monitor Progress

```bash
# List all jobs
aether job list

# Check specific job status
aether pipeline status <job-id>

# Continue a stopped pipeline
aether pipeline continue <job-id>
```

## Common Use Cases

### Processing Local FHIR Data

You have FHIR NDJSON files and want to pseudonymize them:

```bash
# Update aether.yaml:
pipeline:
  enabled_steps:
    - import
    - dimp

services:
  dimp_url: "http://localhost:8083/fhir"

jobs:
  jobs_dir: "./jobs"

# Run the pipeline
aether pipeline start /path/to/fhir/files/
```

### Extracting from TORCH

You want to extract a patient cohort from TORCH:

```yaml
# aether.yaml
services:
  torch:
    base_url: "http://torch.hospital.org"
    username: "researcher"
    password: "secret"

pipeline:
  enabled_steps:
    - torch
    - import
    - dimp
```

Run with CRTDL file:

```bash
aether pipeline start query.crtdl
```

## Next Steps

- [Installation Guide](./installation.md) - Detailed installation instructions
- [Configuration Guide](./configuration.md) - Configure Aether for your environment
- [TORCH Integration](../guides/torch-integration.md) - Learn about TORCH integration
- [Pipeline Guide](../guides/pipeline-steps.md) - Understand the pipeline architecture
