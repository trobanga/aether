#!/bin/bash
# Test script for DIMP service

set -e

echo "=== Aether DIMP Test Environment ==="
echo ""

# Check if services are running
echo "1. Checking Docker Compose services..."
cd "$(dirname "$0")"
docker compose ps

echo ""
echo "2. Getting DIMP service port..."
PORT=$(docker compose port fhir-pseudonymizer 8080 2>/dev/null | cut -d: -f2)

if [ -z "$PORT" ]; then
    echo "❌ Error: fhir-pseudonymizer service is not running"
    echo "Run: docker compose up -d"
    exit 1
fi

echo "✓ DIMP service is running on port: $PORT"
DIMP_URL="http://localhost:$PORT/fhir"
echo "✓ DIMP URL: $DIMP_URL"

echo ""
echo "3. Testing DIMP service connectivity..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$DIMP_URL/\$de-identify" -X POST -H "Content-Type: application/json" -d '{}')

if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "422" ]; then
    echo "✓ DIMP service is responding (HTTP $HTTP_CODE)"
else
    echo "❌ DIMP service returned unexpected status: HTTP $HTTP_CODE"
    exit 1
fi

echo ""
echo "4. Testing DIMP with sample Patient resource..."
RESPONSE=$(curl -s -X POST "$DIMP_URL/\$de-identify" \
  -H "Content-Type: application/json" \
  -d '{
    "resourceType": "Patient",
    "id": "test-patient-123",
    "name": [{
      "family": "TestFamily",
      "given": ["TestGiven"]
    }],
    "gender": "male",
    "birthDate": "1990-01-01"
  }')

if echo "$RESPONSE" | grep -q '"resourceType"'; then
    echo "✓ DIMP service processed the request successfully"
    echo ""
    echo "Response:"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
else
    echo "❌ DIMP service did not return a valid FHIR resource"
    echo "Response: $RESPONSE"
    exit 1
fi

echo ""
echo "5. Checking aether.yaml configuration..."
if [ -f "aether.yaml" ]; then
    CONFIGURED_PORT=$(grep -oP 'dimp_url:.*localhost:\K\d+' aether.yaml || echo "not found")
    if [ "$CONFIGURED_PORT" = "$PORT" ]; then
        echo "✓ aether.yaml is configured correctly with port $PORT"
    else
        echo "⚠ Warning: aether.yaml has port $CONFIGURED_PORT, but service is on port $PORT"
        echo "  Update line 8 in aether.yaml to use port $PORT"
    fi
else
    echo "❌ aether.yaml not found in .github/test/"
    exit 1
fi

echo ""
echo "=== All tests passed! ==="
echo ""
echo "You can now run the pipeline with:"
echo "  cd ../../"
echo "  ./bin/aether pipeline start --input test-data/ --config .github/test/aether.yaml"
