#!/usr/bin/env bash
set -e

if ! hds_base_url="http://$(docker compose port torch-hds 8080)/fhir"; then
    >&2 echo "Unable to find health data store URL"
    exit 2
fi

blazectl upload torch/testdata --server "${hds_base_url}" -c 8
