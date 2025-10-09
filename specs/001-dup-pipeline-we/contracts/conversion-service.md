# Service Contract: Format Conversion (CSV / Parquet)

**Service**: FHIR Conversion Service
**Purpose**: Flatten hierarchical FHIR NDJSON to tabular formats (CSV or Parquet)
**Base URLs**:
- CSV: Configurable via `services.csv_conversion_url` (e.g., `http://localhost:9000/convert/csv`)
- Parquet: Configurable via `services.parquet_conversion_url` (e.g., `http://localhost:9000/convert/parquet`)

## Endpoint: Convert to CSV

### Request

**Method**: `POST`
**Path**: `/convert/csv`
**Headers**:
- `Content-Type: application/x-ndjson`

**Body**: FHIR NDJSON file (newline-delimited JSON resources)

```
{"resourceType":"Patient","id":"p1","name":[{"family":"Smith","given":["John"]}],"birthDate":"1980-01-01"}
{"resourceType":"Patient","id":"p2","name":[{"family":"Jones","given":["Jane"]}],"birthDate":"1985-05-15"}
```

### Response

**Success (200 OK)**:
**Headers**:
- `Content-Type: text/csv`
- `Content-Disposition: attachment; filename="Patient.csv"`

**Body**: Flattened CSV

```csv
id,name_family,name_given,birthDate
p1,Smith,John,1980-01-01
p2,Jones,Jane,1985-05-15
```

**Error Responses**:

| Status | Type | Meaning | Retry? |
|--------|------|---------|--------|
| 400 Bad Request | Non-transient | Malformed NDJSON | No |
| 422 Unprocessable Entity | Non-transient | Cannot flatten schema (unsupported FHIR resource type) | No |
| 500 Internal Server Error | Transient | Service error | Yes |
| 503 Service Unavailable | Transient | Service overloaded | Yes |
| 504 Gateway Timeout | Transient | Request timeout (large file) | Yes |

**Error Body**:
```json
{
  "error": {
    "code": "unsupported_resource_type",
    "message": "Resource type 'CustomExtension' not supported for CSV conversion"
  }
}
```

---

## Endpoint: Convert to Parquet

### Request

**Method**: `POST`
**Path**: `/convert/parquet`
**Headers**:
- `Content-Type: application/x-ndjson`

**Body**: FHIR NDJSON file (same as CSV)

### Response

**Success (200 OK)**:
**Headers**:
- `Content-Type: application/octet-stream`
- `Content-Disposition: attachment; filename="Patient.parquet"`

**Body**: Binary Parquet file

**Schema** (example for Patient):
```
message Patient {
  required binary id (UTF8);
  optional binary name_family (UTF8);
  optional binary name_given (UTF8);
  optional binary birthDate (UTF8);
}
```

**Error Responses**: Same as CSV endpoint

---

## Flattening Strategy

### Array Handling

**Option 1: First Element** (default)
- `name[0].family` → `name_family`
- `name[0].given[0]` → `name_given`

**Option 2: Concatenation** (if multiple values)
- `name[].given` → `name_given` = "John,Jane,Bob" (comma-separated)

### Nested Objects

- Flatten with underscore: `name.family` → `name_family`
- Max depth: 3 levels (configurable)

### Resource Type Separation

- Service groups resources by `resourceType`
- One CSV/Parquet file per resource type
- Multiple resource types in input → multiple output files (e.g., `Patient.csv`, `Observation.csv`)

---

## Implementation Notes

**Aether behavior**:

1. **Input**: Read FHIR files from `import/*.ndjson` or `pseudonymized/*.ndjson`
2. **Group by resource type**: Parse each line, extract `resourceType`, group into batches
3. **POST to service**: Send entire file (or batch) to conversion endpoint
4. **Save output**: Write response to `csv/<ResourceType>.csv` or `parquet/<ResourceType>.parquet`

**Retry logic**: Same as DIMP (5xx → retry, 4xx → fail)

**Concurrency**: Process resource types in parallel (Patient.csv and Observation.csv can convert simultaneously)

---

## Testing

**Contract Test** (`tests/contract/conversion_test.go`):

```go
func TestConversionService_CSV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/convert/csv", r.URL.Path)

		// Mock CSV response
		w.Header().Set("Content-Type", "text/csv")
		w.Write([]byte("id,name_family\np1,Smith\n"))
	}))
	defer server.Close()

	client := NewConversionClient(server.URL + "/convert/csv")
	csv, err := client.ConvertToCSV(sampleNDJSON)

	assert.NoError(t, err)
	assert.Contains(t, string(csv), "id,name_family")
}
```

---

## Example Usage

### CSV Conversion

```bash
# Input: jobs/abc-123/import/patients.ndjson
{"resourceType":"Patient","id":"p1","name":[{"family":"Smith"}]}
{"resourceType":"Observation","id":"o1","code":{"text":"Blood Pressure"}}

# Output:
# jobs/abc-123/csv/Patient.csv
id,name_family
p1,Smith

# jobs/abc-123/csv/Observation.csv
id,code_text
o1,Blood Pressure
```

### Parquet Conversion

```bash
# Input: Same as above
# Output:
# jobs/abc-123/parquet/Patient.parquet (binary)
# jobs/abc-123/parquet/Observation.parquet (binary)
```

---

## Service Assumptions

- Service handles NDJSON input (one resource per line)
- Service groups by `resourceType` and returns separate files (or Aether does grouping)
- Service is stateless
- Service can handle multi-MB files (typical FHIR dataset size)
- Timeout: 60 seconds for files up to 100MB

---

## Future Enhancements

**Batch Endpoint** (if service supports):
```
POST /convert/csv/batch
Content-Type: multipart/form-data

[Multiple NDJSON files in a single request]
```

**Schema Customization** (future):
```json
{
  "resource_type": "Patient",
  "fields": ["id", "name.family", "birthDate"],
  "flatten_arrays": "first_element"
}
```
