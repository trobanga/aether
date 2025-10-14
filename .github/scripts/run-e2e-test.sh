#!/usr/bin/env bash
set -euo pipefail

# Wrapper script to run aether E2E tests with dynamic Docker Compose ports
# This script extracts the dynamically assigned ports and sets environment variables
# for the aether.yaml configuration file

COMPOSE_FILE="${COMPOSE_FILE:-.github/test/compose.yaml}"
CONFIG_FILE="${CONFIG_FILE:-.github/test/aether.yaml}"

echo "Extracting dynamic service ports..."

# Get DIMP pseudonymizer port
if DIMP_PORT=$(docker compose port fhir-pseudonymizer 8080 2>/dev/null | cut -d: -f2); then
    export DIMP_URL="http://localhost:${DIMP_PORT}/fhir"
    echo "DIMP service available at: ${DIMP_URL}"
else
    echo "Warning: DIMP service not found, using default URL"
    export DIMP_URL="http://localhost:32861/fhir"
fi

# Get TORCH server port
if TORCH_PORT=$(docker compose port torch 8080 2>/dev/null | cut -d: -f2); then
    export TORCH_URL="http://localhost:${TORCH_PORT}"
    echo "TORCH service available at: ${TORCH_URL}"
else
    echo "Warning: TORCH service not found, using default URL"
    export TORCH_URL="http://localhost:8081"
fi

# Get TORCH file server port
if TORCH_FILE_SERVER_PORT=$(docker compose port torch-file-server 80 2>/dev/null | cut -d: -f2); then
    export TORCH_FILE_SERVER_URL="http://localhost:${TORCH_FILE_SERVER_PORT}"
    echo "TORCH file server available at: ${TORCH_FILE_SERVER_URL}"
else
    echo "Warning: TORCH file server not found, using default URL"
    export TORCH_FILE_SERVER_URL="http://localhost:8082"
fi

# Get VFPS port
if VFPS_PORT=$(docker compose port vfps 8080 2>/dev/null | cut -d: -f2); then
    export VFPS_URL="http://localhost:${VFPS_PORT}"
    echo "VFPS service available at: ${VFPS_URL}"
else
    echo "Warning: VFPS service not found, using default URL"
    export VFPS_URL="http://localhost:8080"
fi

# Add more service port extractions here as needed
# Example for CSV service:
# if CSV_PORT=$(docker compose -f "${COMPOSE_FILE}" port csv-service 9000 2>/dev/null | cut -d: -f2); then
#     export CSV_URL="http://localhost:${CSV_PORT}/convert/csv"
# fi

echo "Initializing VFPS namespace..."
# Create the patient-identifiers namespace required by the anonymization rules
curl --request POST \
    --url "${VFPS_URL}/v1/namespaces" \
    --header 'content-type: application/json' \
    --data '{
      "name": "patient-identifiers",
      "pseudonymGenerationMethod": "PSEUDONYM_GENERATION_METHOD_UNSPECIFIED",
      "pseudonymLength": 32,
      "pseudonymPrefix": "string",
      "pseudonymSuffix": "string",
      "description": "string"
    }'

# curl --silent --show-error --fail --request POST \
#     --url "${VFPS_URL}/api/v1/Namespace" \
#     --header 'content-type: application/json' \
#     --data '{
#   "name": "patient-identifiers",
#   "pseudonymGenerationMethod": "PSEUDONYM_GENERATION_METHOD_SECURE_RANDOM_BASE64URL_ENCODED",
#   "pseudonymLength": 32,
#   "description": "Namespace for patient identifier pseudonymization"
# }' && echo "VFPS namespace 'patient-identifiers' created successfully" || echo "Warning: Failed to create VFPS namespace (may already exist)"

echo ""
echo "Running aether with config: ${CONFIG_FILE}"
echo "Environment variables:"
echo "  DIMP_URL=${DIMP_URL}"
echo "  TORCH_URL=${TORCH_URL}"
echo "  TORCH_FILE_SERVER_URL=${TORCH_FILE_SERVER_URL}"
echo "  VFPS_URL=${VFPS_URL}"

# Change to test directory and run aether with the configuration
../../bin/aether pipeline start torch/queries/example-crtdl.json --config aether.yaml
