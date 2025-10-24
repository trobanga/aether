#!/usr/bin/env bash
# Usage: source .github/scripts/set-env.sh
#   (or: . .github/scripts/set-env.sh)
# Note: Must be sourced, not executed, to set environment variables in your shell

# Only set strict mode if being executed (not sourced)
# This prevents issues with interactive shells that reference unset variables
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    set -euo pipefail
fi

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
