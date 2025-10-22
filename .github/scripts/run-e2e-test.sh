#!/usr/bin/env bash
set -euo pipefail

# Wrapper script to run aether E2E tests with dynamic Docker Compose ports
# Sources set-env.sh to extract dynamically assigned ports and set environment variables

COMPOSE_FILE="${COMPOSE_FILE:-.github/test/compose.yaml}"
CONFIG_FILE="${CONFIG_FILE:-.github/test/aether.yaml}"

# Source the environment setup script to get dynamic Docker Compose ports
source ../scripts/set-env.sh

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
