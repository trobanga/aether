#!/usr/bin/env bash
set -euo pipefail

# Wrapper script to run aether E2E tests with dynamic Docker Compose ports
# This script extracts the dynamically assigned ports and sets environment variables
# for the aether.yaml configuration file

COMPOSE_FILE="${COMPOSE_FILE:-.github/test/compose.yaml}"
CONFIG_FILE="${CONFIG_FILE:-.github/test/aether.yaml}"

echo "Extracting dynamic service ports..."

# Get DIMP pseudonymizer port
if DIMP_PORT=$(docker compose -f "${COMPOSE_FILE}" port fhir-pseudonymizer 8080 2>/dev/null | cut -d: -f2); then
    export DIMP_URL="http://localhost:${DIMP_PORT}/fhir"
    echo "DIMP service available at: ${DIMP_URL}"
else
    echo "Warning: DIMP service not found, using default URL"
    export DIMP_URL="http://localhost:32861/fhir"
fi

# Add more service port extractions here as needed
# Example for CSV service:
# if CSV_PORT=$(docker compose -f "${COMPOSE_FILE}" port csv-service 9000 2>/dev/null | cut -d: -f2); then
#     export CSV_URL="http://localhost:${CSV_PORT}/convert/csv"
# fi

echo "Running aether with config: ${CONFIG_FILE}"
echo "Environment variables:"
echo "  DIMP_URL=${DIMP_URL}"

# Run aether with the configuration
# TODO: Add actual aether command here
# Example: aether import <job-id> --config "${CONFIG_FILE}"

echo "E2E test execution placeholder - add aether commands here"
