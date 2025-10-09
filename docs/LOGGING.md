# Logging in Aether

## Log Levels

Aether supports multiple log levels for controlling output verbosity:

- **ERROR**: Critical errors that cause operations to fail
- **WARN**: Warnings about issues that don't stop execution (e.g., retries)
- **INFO**: General informational messages about pipeline progress
- **DEBUG**: Detailed debug information for troubleshooting

## Enabling Verbose Logging

Use the `--verbose` or `-v` flag to enable DEBUG level logging:

```bash
./bin/aether pipeline start --input test-data/ --config aether.yaml --verbose
```

## What Gets Logged at Each Level

### DEBUG Level (--verbose)
When you use `--verbose`, you'll see:

**DIMP Processing:**
- Each resource being sent to DIMP (resourceType, ID, URL)
- Request body size
- DIMP service responses (status code, resource info)
- Resource ID transformations (original → pseudonymized)
- Each FHIR resource being processed (file, line number, resourceType, ID)

**Import Processing:**
- Individual file imports with details (filename, size, resource count)

**HTTP Requests:**
- Service call details (endpoint, method)
- Service response details (status, duration)

### INFO Level (default)
Without `--verbose`, you'll see:

- Job creation and completion
- Step start/completion with file counts and duration
- DIMP file processing progress (file X of Y)
- Successfully processed files with resource counts
- Major pipeline events

### WARN Level
Always shown:

- Service errors (HTTP 4xx, 5xx)
- Retry attempts with attempt number and error
- Service connectivity issues

### ERROR Level
Always shown:

- Critical failures that stop the pipeline
- Step failures with error details
- HTTP request failures
- Failed to parse/pseudonymize specific resources
- **DIMP errors with full error response body**

## DIMP-Specific Logging

With `--verbose`, for each resource the DIMP step processes, you'll see:

1. **Before pseudonymization:**
   ```
   [DEBUG] Sending resource to DIMP | [resourceType Patient id abc-123 url http://localhost:33048/fhir/$de-identify]
   [DEBUG] Request body size | [bytes 1234]
   ```

2. **On success:**
   ```
   [DEBUG] DIMP service responded successfully | [status_code 200 resourceType Patient id abc-123]
   [DEBUG] Resource ID pseudonymized | [resourceType Patient original_id abc-123 new_id xyz-789-hashed]
   ```

3. **On error:**
   ```
   [ERROR] DIMP service returned error | [status_code 500 status 500 Internal Server Error resourceType Patient id abc-123 error_body {...full error response...} retryable true]
   ```

## Example: Debugging DIMP 500 Errors

Run with verbose logging to see the full error details:

```bash
./bin/aether pipeline start --input test-data/ --config .github/test/aether.yaml --verbose
```

Look for these log entries:
- `[ERROR] DIMP service returned error` - Shows the full error response from DIMP
- `[ERROR] Failed to pseudonymize FHIR resource` - Shows which specific resource failed
- `[INFO] Processing FHIR file through DIMP` - Shows which file is being processed
- `[WARN] Retry attempt X/5` - Shows retry attempts for transient errors

## Example Output (verbose mode)

```
2025/10/09 13:40:41 [INFO] Processing FHIR file through DIMP | [file_number 1 total_files 13 filename test.ndjson job_id abc-123]
2025/10/09 13:40:41 [DEBUG] Processing FHIR resource | [file test.ndjson line_number 1 resourceType Patient id patient-001]
2025/10/09 13:40:41 [DEBUG] Sending resource to DIMP | [resourceType Patient id patient-001 url http://localhost:33048/fhir/$de-identify]
2025/10/09 13:40:41 [DEBUG] Request body size | [bytes 456]
2025/10/09 13:40:41 [ERROR] DIMP service returned error | [status_code 500 status 500 Internal Server Error resourceType Patient id patient-001 error_body {"error":"Invalid resource format"} retryable true]
2025/10/09 13:40:41 [WARN] Retry attempt 1/5 for: http://localhost:33048/fhir/$de-identify | [error HTTP 500: 500 Internal Server Error]
```

## Checking Docker Logs for DIMP Service

If DIMP is returning errors, also check the service logs:

```bash
cd .github/test
docker compose logs fhir-pseudonymizer --tail 50 -f
```

This will show what's happening on the DIMP service side.

## Tips for Troubleshooting

1. **Always use `--verbose` when debugging** - It shows the actual error responses from services
2. **Check both aether logs and Docker logs** - The issue might be in the service itself
3. **Look for the `error_body` field** - This contains the full error response from DIMP
4. **Note the resource that fails** - The logs show exactly which resourceType and ID caused the error
5. **Check retry behavior** - `retryable true/false` tells you if the error will be retried

## Log Format

All logs follow this format:
```
TIMESTAMP [LEVEL] Message | [key1 value1 key2 value2 ...]
```

Example:
```
2025/10/09 13:40:41 [ERROR] DIMP service returned error | [status_code 500 resourceType Patient error_body {...}]
```
