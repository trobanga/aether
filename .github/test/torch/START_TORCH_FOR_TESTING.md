# Quick TORCH Setup for Testing Aether Integration

## Current Setup

You're in: `.github/test/torch/`

This TORCH setup requires a FHIR server backend to extract data from.

## Option 1: Start TORCH with Mock Data (Simplest for Testing)

For integration testing, we can configure TORCH to work with test data:

### Step 1: Start TORCH Server
```bash
# From this directory (.github/test/torch/)
docker compose up -d

# Check if containers are running
docker compose ps
```

### Step 2: Check TORCH is accessible
```bash
# TORCH should be accessible at localhost:8086
curl -u test:test http://localhost:8086/fhir/metadata

# Or check if it's responding
docker compose logs torch
```

## Option 2: For Real Testing (With FHIR Server)

TORCH needs a FHIR server at `http://fhir-server:8080/fhir` (configured in .env)

You'll need to either:
1. Add a FHIR server to the compose.yaml
2. Point to an existing FHIR server
3. Use the full Feasibility Triangle setup from dse-example

## Testing Aether TORCH Integration

Once TORCH is running, test the aether integration:

### Update Aether Config
```bash
# Go back to project root
cd ../../../

# Edit config to point to TORCH
cat > config/aether.test.yaml << 'YAML'
services:
  torch:
    base_url: "http://localhost:8086"
    username: "test"
    password: "test"
    extraction_timeout_minutes: 5  # Shorter for testing
    polling_interval_seconds: 2
    max_polling_interval_seconds: 10

pipeline:
  enabled_steps:
    - import

jobs_dir: "./test-jobs"
YAML
```

### Run Tests

```bash
# Build aether
go build -o bin/aether ./cmd/aether

# Test 1: Verify TORCH connectivity (unit tests)
go test -v -run TestTORCHClient_Ping ./tests/unit/

# Test 2: Test input type detection
go test -v -run TestDetectInputType ./tests/unit/

# Test 3: Test CRTDL validation  
go test -v -run TestValidateCRTDL ./tests/unit/

# Test 4: Integration test (requires TORCH running)
# Note: Integration tests use mock servers, not real TORCH
go test -v -run TestPipeline_TORCH ./tests/integration/
```

### Manual End-to-End Test

```bash
# This will fail if TORCH doesn't have a working FHIR backend
./bin/aether pipeline start --input test-data/torch/example.crtdl --verbose

# Expected behavior:
# - Detects input as CRTDL file
# - Validates CRTDL syntax
# - Attempts to connect to TORCH
# - May fail at extraction if no FHIR server available
```

## Troubleshooting

### TORCH Not Starting
```bash
# Check logs
docker compose logs torch

# Common issues:
# - Port 8086 already in use
# - Missing FHIR server connection
# - Memory issues (needs 8GB for JVM)
```

### Connection Refused
```bash
# Check if TORCH is listening
docker compose ps
netstat -an | grep 8086

# Try direct curl
curl -v http://localhost:8086/fhir/metadata
```

### FHIR Server Missing
TORCH needs a FHIR server to extract from. Options:
1. Add HAPI FHIR server to compose.yaml
2. Use synthetic data for testing
3. Mock the TORCH responses in tests (already done in integration tests)

## Recommendation for Quick Testing

**Use the automated tests instead of manual testing:**

```bash
# These tests use mock TORCH servers and don't need real TORCH
cd ../../../  # Back to project root

# Run all TORCH tests
go test ./tests/unit/ ./tests/integration/ -v -run "TORCH|Input|CRTDL"
```

The integration tests already mock the TORCH API responses, so you can verify the aether implementation works correctly without needing a fully functional TORCH server!
