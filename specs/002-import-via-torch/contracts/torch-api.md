# TORCH API Contract Specification

**Feature**: 002-import-via-torch | **Date**: 2025-10-10

## Overview

This document defines the contract between aether and the TORCH server API. Contract tests will verify that aether correctly implements the FHIR $extract-data operation as specified by TORCH.

**Contract Test Location**: `tests/contract/torch_service_test.go`

## Base Configuration

**Endpoint**: `{TORCH_BASE_URL}/fhir/$extract-data`
**Authentication**: HTTP Basic Authentication
**Content-Type**: `application/fhir+json`
**FHIR Version**: R4

## Operation: Submit Extraction

### Request

**Method**: `POST`
**Path**: `/fhir/$extract-data`
**Headers**:
```http
Authorization: Basic <base64(username:password)>
Content-Type: application/fhir+json
```

**Body**: FHIR Parameters resource with base64-encoded CRTDL

```json
{
  "resourceType": "Parameters",
  "parameter": [
    {
      "name": "crtdl",
      "valueBase64Binary": "<base64-encoded CRTDL JSON>"
    }
  ]
}
```

**Example** (with CRTDL encoded):
```json
{
  "resourceType": "Parameters",
  "parameter": [
    {
      "name": "crtdl",
      "valueBase64Binary": "eyJjb2hvcnREZWZpbml0aW9uIjp7InZlcnNpb24iOiIxLjAuMCIsImluY2x1c2lvbkNyaXRlcmlhIjpbXX0sImRhdGFFeHRyYWN0aW9uIjp7ImF0dHJpYnV0ZUdyb3VwcyI6W119fQ=="
    }
  ]
}
```

### Response: Accepted

**Status**: `202 Accepted`
**Headers**:
```http
Content-Location: {TORCH_BASE_URL}/fhir/extraction/{extraction-id}
```

**Body**: Empty or minimal acknowledgment (implementation-specific)

**Contract Assertions**:
- Status code MUST be 202
- `Content-Location` header MUST be present
- `Content-Location` MUST be a valid absolute URL
- `Content-Location` MUST contain `/fhir/` in path

### Response: Errors

#### 400 Bad Request
**Cause**: Invalid CRTDL structure, malformed FHIR Parameters

**Body**:
```json
{
  "resourceType": "OperationOutcome",
  "issue": [
    {
      "severity": "error",
      "code": "invalid",
      "diagnostics": "<error description>"
    }
  ]
}
```

#### 401 Unauthorized
**Cause**: Missing or invalid authentication credentials

**Headers**:
```http
WWW-Authenticate: Basic realm="TORCH"
```

#### 500 Internal Server Error
**Cause**: TORCH server error

**Body**: OperationOutcome or text error message

**Contract Assertions**:
- Error responses MUST have appropriate status codes
- 400/500 errors SHOULD include OperationOutcome if available

## Operation: Poll Extraction Status

### Request

**Method**: `GET`
**Path**: `{Content-Location from submit response}`
**Headers**:
```http
Authorization: Basic <base64(username:password)>
Accept: application/fhir+json
```

### Response: In Progress

**Status**: `202 Accepted`
**Body**: Empty or status message (implementation-specific)

**Contract Assertions**:
- Status code MUST be 202 while extraction is processing
- Response MAY be empty or contain progress information

### Response: Complete

**Status**: `200 OK`
**Content-Type**: `application/fhir+json`

**Body**: FHIR Parameters resource with output file URLs

```json
{
  "resourceType": "Parameters",
  "parameter": [
    {
      "name": "output",
      "part": [
        {
          "name": "url",
          "valueUrl": "http://torch-server:8080/output/batch-1.ndjson"
        },
        {
          "name": "url",
          "valueUrl": "http://torch-server:8080/output/batch-2.ndjson"
        }
      ]
    }
  ]
}
```

**Contract Assertions**:
- Status code MUST be 200
- Body MUST be valid FHIR Parameters resource
- MUST have `parameter` array with at least one element
- Each parameter with `name: "output"` MUST have `part` array
- Each `part` MUST have `name: "url"` and `valueUrl` fields
- `valueUrl` MUST be a valid HTTP(S) URL

### Response: Failed

**Status**: `500 Internal Server Error` or `410 Gone`
**Body**: OperationOutcome with error details

```json
{
  "resourceType": "OperationOutcome",
  "issue": [
    {
      "severity": "error",
      "code": "exception",
      "diagnostics": "Extraction failed: <reason>"
    }
  ]
}
```

**Contract Assertions**:
- Failed extractions MUST return 4xx or 5xx status
- Body SHOULD include OperationOutcome with diagnostics

## Operation: Download Extraction Files

### Request

**Method**: `GET`
**Path**: `{valueUrl from extraction result}`
**Headers**:
```http
Authorization: Basic <base64(username:password)>
Accept: application/fhir+ndjson
```

### Response

**Status**: `200 OK`
**Content-Type**: `application/fhir+ndjson` or `application/x-ndjson`

**Body**: NDJSON file with FHIR Bundle resources

```ndjson
{"resourceType":"Bundle","type":"transaction","entry":[...]}
{"resourceType":"Bundle","type":"transaction","entry":[...]}
{"resourceType":"Bundle","type":"transaction","entry":[...]}
```

**Contract Assertions**:
- Status code MUST be 200
- Content-Type MUST be NDJSON variant
- Body MUST be valid NDJSON (newline-delimited JSON)
- Each line MUST be a valid FHIR Bundle resource

### Response: Errors

#### 404 Not Found
**Cause**: File no longer available or invalid URL

#### 410 Gone
**Cause**: Extraction results expired

**Contract Assertions**:
- 404/410 MUST be handled gracefully
- Client SHOULD provide clear error message to user

## Authentication

**Scheme**: HTTP Basic Authentication

**Header Format**:
```
Authorization: Basic <base64(username:password)>
```

**Example**:
```
Username: test
Password: test
Header: Authorization: Basic dGVzdDp0ZXN0
```

**Contract Assertions**:
- All requests MUST include `Authorization` header
- Base64 encoding MUST be correct (no newlines, padding preserved)
- Missing auth MUST result in 401 response
- Invalid credentials MUST result in 401 response

## Timeout Behavior

**Client Timeout**: Configurable (default 30 minutes for polling)
**Server Timeout**: Implementation-specific

**Contract Assertions**:
- Client MUST respect configured timeout
- Client MUST stop polling after timeout expires
- Client MUST provide clear timeout error message

## Retry Policy

**Transient Errors** (retryable):
- 5xx server errors
- Network timeouts
- Connection refused

**Permanent Errors** (non-retryable):
- 400 Bad Request
- 401 Unauthorized
- 404 Not Found

**Contract Assertions**:
- Client MUST retry transient errors with exponential backoff
- Client MUST NOT retry permanent errors
- Retry configuration MUST be respected (max attempts, backoff)

## Contract Test Scenarios

### Test 1: Successful Extraction Flow

```
GIVEN valid CRTDL file and TORCH server running
WHEN submitting extraction request
THEN receive 202 with Content-Location
WHEN polling Content-Location
THEN receive 202 (in progress) followed by 200 (complete)
WHEN parsing result
THEN extract valid file URLs
WHEN downloading files
THEN receive valid NDJSON bundles
```

### Test 2: Invalid CRTDL

```
GIVEN invalid CRTDL file (missing cohortDefinition)
WHEN submitting extraction request
THEN receive 400 Bad Request with OperationOutcome
```

### Test 3: Authentication Failure

```
GIVEN invalid credentials
WHEN submitting extraction request
THEN receive 401 Unauthorized
```

### Test 4: Polling Timeout

```
GIVEN extraction that exceeds timeout
WHEN polling for duration > configured timeout
THEN client stops polling and returns timeout error
```

### Test 5: Empty Result

```
GIVEN CRTDL with empty cohort
WHEN extraction completes
THEN receive 200 with empty output parameter array
AND client handles gracefully (zero files)
```

### Test 6: Server Error During Extraction

```
GIVEN extraction that fails server-side
WHEN polling status
THEN receive 500 or 410 with error details
AND client propagates error to user
```

## Mock TORCH Server for Testing

For contract tests, implement mock TORCH server that:

1. **Accepts valid requests**: Returns 202 → 200 with file URLs
2. **Rejects invalid requests**: Returns appropriate 400/401 responses
3. **Simulates delays**: Returns 202 for N polls before 200
4. **Provides test files**: Serves sample NDJSON at file URLs
5. **Tests error scenarios**: Can be configured to return errors

**Implementation**: Use Go `httptest` package or run actual TORCH container

## Conformance

**This contract is based on**:
- FHIR R4 specification for Parameters resource
- TORCH example in `/home/trobanga/development/mii/dse-example/torch/`
- TORCH `execute-crtdl.sh` script behavior

**Deviations from standard FHIR**:
- TORCH-specific `$extract-data` operation (not standard FHIR bulk export)
- Custom CRTDL format in base64 parameter
- Polling model (vs. FHIR Bulk Data async pattern with status endpoint)

**Version**: This contract reflects TORCH behavior as of 2025-10-10. TORCH API may evolve - contract tests will catch breaking changes.
