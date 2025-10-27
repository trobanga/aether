# CLI Commands

Complete reference for all Aether CLI commands and options.

## Global Options

All commands support these options:

```bash
aether [global-options] <command> [command-options]
```

**Global Options:**
- `--config, -c FILE` - Path to aether.yaml configuration file
- `--jobs-dir DIR` - Override jobs directory
- `--help, -h` - Show command help
- `--version, -v` - Show Aether version
- `--debug` - Enable debug logging

## Commands

### aether pipeline start

Start a new pipeline execution.

**Syntax:**
```bash
aether pipeline start [options] <input>
```

**Arguments:**
- `<input>` - Path to FHIR directory or CRTDL query file

**Options:**
- `--config, -c FILE` - Configuration file (default: aether.yaml)
- `--jobs-dir DIR` - Override jobs directory
- `--steps STEP1,STEP2` - Override enabled steps

**Examples:**
```bash
# Start from local FHIR files
aether pipeline start /data/fhir/

# Start from CRTDL query
aether pipeline start my_cohort.crtdl

# Override configuration
aether pipeline start --config prod.yaml query.crtdl

# Run specific steps only
aether pipeline start --steps import,dimp /data/fhir/
```

### aether pipeline status

Check the status of a running or completed pipeline.

**Syntax:**
```bash
aether pipeline status [options] <job-id>
```

**Arguments:**
- `<job-id>` - Job identifier

**Options:**
- `--json` - Output as JSON
- `--jobs-dir DIR` - Override jobs directory

**Examples:**
```bash
# Check job status
aether pipeline status abc123

# Get JSON output for scripting
aether pipeline status --json abc123
```

### aether pipeline continue

Resume a failed pipeline from where it stopped.

**Syntax:**
```bash
aether pipeline continue [options] <job-id>
```

**Arguments:**
- `<job-id>` - Job identifier of failed pipeline

**Options:**
- `--config, -c FILE` - Configuration file
- `--jobs-dir DIR` - Override jobs directory

**Examples:**
```bash
# Resume failed job
aether pipeline continue abc123

# Resume with specific configuration
aether pipeline continue --config prod.yaml abc123
```

### aether job list

List all jobs.

**Syntax:**
```bash
aether job list [options]
```

**Options:**
- `--jobs-dir DIR` - Override jobs directory
- `--status STATUS` - Filter by status (running, completed, failed)
- `--json` - Output as JSON
- `--limit N` - Show last N jobs (default: 10)

**Examples:**
```bash
# List all jobs
aether job list

# Show failed jobs only
aether job list --status failed

# Get as JSON for scripting
aether job list --json
```

### aether job logs

View logs for a specific job.

**Syntax:**
```bash
aether job logs [options] <job-id>
```

**Arguments:**
- `<job-id>` - Job identifier

**Options:**
- `--jobs-dir DIR` - Override jobs directory
- `--follow, -f` - Stream logs continuously
- `--step STEP` - Show logs for specific step only
- `--errors-only` - Show only error lines

**Examples:**
```bash
# View job logs
aether job logs abc123

# Stream logs as they happen
aether job logs --follow abc123

# View only DIMP step logs
aether job logs --step dimp abc123

# Show only errors
aether job logs --errors-only abc123
```

### aether job delete

Delete a completed job and its data.

**Syntax:**
```bash
aether job delete [options] <job-id>
```

**Arguments:**
- `<job-id>` - Job identifier

**Options:**
- `--jobs-dir DIR` - Override jobs directory
- `--force, -f` - Skip confirmation prompt

**Examples:**
```bash
# Delete job (with confirmation)
aether job delete abc123

# Force delete without confirmation
aether job delete --force abc123
```

### aether completion

Generate shell completion scripts.

**Syntax:**
```bash
aether completion <shell>
```

**Arguments:**
- `<shell>` - Shell type: bash, zsh, fish, powershell

**Examples:**
```bash
# Generate bash completions
aether completion bash

# Install for zsh
aether completion zsh | sudo tee /etc/zsh/completions/_aether

# Install for bash
aether completion bash | sudo tee /etc/bash_completion.d/aether
```

### aether version

Show Aether version and build information.

**Syntax:**
```bash
aether version
```

**Output:**
```
Aether v1.0.0
Build: abc123def456
Go: 1.21.0
```

### aether help

Show help information.

**Syntax:**
```bash
aether help [command]
```

**Arguments:**
- `[command]` - Optional specific command to show help for

**Examples:**
```bash
# Show general help
aether help

# Show help for specific command
aether help pipeline start
```

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Configuration error
- `3` - Invalid input
- `4` - Service unavailable
- `5` - Pipeline failed (retryable)
- `6` - Pipeline failed (fatal)

## Output Formats

### Default (Human-Readable)

```
Job: abc123
Status: running
Steps: torch, import, dimp
Progress: 45%
Elapsed: 2m 30s
ETA: 3m 15s
```

### JSON Format

```json
{
  "job_id": "abc123",
  "status": "running",
  "steps": ["torch", "import", "dimp"],
  "progress": 0.45,
  "elapsed_seconds": 150,
  "eta_seconds": 195
}
```

## Environment Variables

- `AETHER_CONFIG` - Default configuration file path
- `AETHER_JOBS_DIR` - Default jobs directory
- `AETHER_LOG_LEVEL` - Logging level (debug, info, warn, error)
- `TORCH_USERNAME` - TORCH username
- `TORCH_PASSWORD` - TORCH password
- `DIMP_URL` - DIMP service URL

## Next Steps

- [Configuration Reference](./config-reference.md) - Configuration file options
- [Pipeline Steps](../guides/pipeline-steps.md) - Pipeline architecture
- [Troubleshooting](../getting-started/installation.md) - Common issues
