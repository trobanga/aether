# Service Contract: DIMP (De-identification, Minimization, Pseudonymization)

**Service**: DIMP Pseudonymization Service
**Purpose**: De-identify and pseudonymize FHIR resources
**Base URL**: Configurable via `services.dimp_url` (e.g., `http://localhost:8083/fhir`)

## Endpoint: Pseudonymize FHIR Resource

### Request

**Method**: `POST`
**Path**: `/$de-identify`
**Headers**:
- `Content-Type: application/json`

**Body**: Single FHIR resource (JSON object)

```json
{
  "resourceType": "Patient",
  "id": "example-patient-123",
  "identifier": [
    {
      "system": "http://hospital.org/patients",
      "value": "12345"
    }
  ],
  "name": [
    {
      "family": "Doe",
      "given": ["John"]
    }
  ],
  "birthDate": "1980-01-01"
}
```

### Response

**Success (200 OK)**:
**Headers**:
- `Content-Type: application/json`

**Body**: Pseudonymized FHIR resource

```json
{
  "resourceType": "Patient",
  "id": "pseudonym-abc123xyz",
  "identifier": [
    {
      "system": "http://hospital.org/patients",
      "value": "PSEUDO_98765"
    }
  ],
  "name": [
    {
      "family": "REDACTED",
      "given": ["REDACTED"]
    }
  ],
  "birthDate": "1980"
}
```

**Error Responses**:

| Status | Type | Meaning | Retry? |
|--------|------|---------|--------|
| 400 Bad Request | Non-transient | Malformed FHIR resource | No |
| 422 Unprocessable Entity | Non-transient | Invalid FHIR schema | No |
| 500 Internal Server Error | Transient | Service error | Yes |
| 502 Bad Gateway | Transient | Upstream dependency (e.g., VFPS) unavailable | Yes |
| 503 Service Unavailable | Transient | Service overloaded | Yes |
| 504 Gateway Timeout | Transient | Request timeout | Yes |

**Error Body**:
```json
{
  "error": {
    "code": "invalid_resource",
    "message": "Missing required field: resourceType"
  }
}
```

## Implementation Notes

- **Aether behavior**:
  - Read FHIR resources line-by-line from `import/*.ndjson`
  - POST each resource to `$de-identify`
  - Write pseudonymized response to `pseudonymized/dimped_<original-filename>.ndjson`
  - Append newline after each resource (NDJSON format)

- **Retry logic**:
  - 5xx, network errors, timeouts → automatic retry with exponential backoff
  - 4xx → fail step, require manual intervention

- **Concurrency**: Process files sequentially, resources within file can be batched (if service supports batch endpoint in future)

## Testing

**Contract Test** (`tests/contract/dimp_test.go`):
```go
func TestDIMPService_Pseudonymize(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/$de-identify", r.URL.Path)

		// Read request body
		var resource map[string]interface{}
		json.NewDecoder(r.Body).Decode(&resource)

		// Return pseudonymized version
		resource["id"] = "pseudonym-123"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resource)
	}))
	defer server.Close()

	// Test client
	client := NewDIMPClient(server.URL)
	result, err := client.Pseudonymize(samplePatient)

	assert.NoError(t, err)
	assert.Equal(t, "pseudonym-123", result["id"])
}
```

## Example Usage

```bash
# Input: jobs/abc-123/import/patients.ndjson
{"resourceType":"Patient","id":"p1","name":[{"family":"Smith"}]}
{"resourceType":"Patient","id":"p2","name":[{"family":"Jones"}]}

# Output: jobs/abc-123/pseudonymized/dimped_patients.ndjson
{"resourceType":"Patient","id":"pseudo-x1","name":[{"family":"REDACTED"}]}
{"resourceType":"Patient","id":"pseudo-x2","name":[{"family":"REDACTED"}]}
```

## Service Assumptions

- Service accepts single FHIR resources (not bundles)
- Service returns pseudonymized resource with same `resourceType`
- Service is stateless (no session management)
- Service handles its own pseudonym generation and storage (e.g., VFPS integration)
