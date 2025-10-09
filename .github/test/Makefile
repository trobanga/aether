# Makefile for Aether Test Infrastructure
# Located in .github/test/

.PHONY: help services dimp-up dimp-down dimp-test test-with-services clean-test-data

# Default target
.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "Aether Test Infrastructure"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available Services:"
	@echo "  dimp              DIMP pseudonymization service (http://localhost:8083)"
	@echo ""
	@echo "Targets:"
	@echo "  dimp-up           Start DIMP test service"
	@echo "  dimp-down         Stop DIMP test service"
	@echo "  dimp-logs         Show DIMP service logs"
	@echo "  dimp-restart      Restart DIMP service"
	@echo "  dimp-test         Run integration tests with DIMP service"
	@echo "  test-with-services Run all tests with required services"
	@echo "  clean-test-data   Clean test data and temporary files"

## services: Alias for help
services: help

## dimp-up: Start DIMP test service
dimp-up:
	@echo "Starting DIMP test service..."
	cd dimp && docker compose up -d
	@echo "✓ DIMP service running at http://localhost:8083"
	@echo ""
	@echo "Test the service:"
	@echo "  curl http://localhost:8083/fhir/$$de-identify"

## dimp-down: Stop DIMP test service
dimp-down:
	@echo "Stopping DIMP test service..."
	cd dimp && docker compose down
	@echo "✓ DIMP service stopped"

## dimp-logs: Show DIMP service logs
dimp-logs:
	@echo "Showing DIMP service logs..."
	cd dimp && docker compose logs -f

## dimp-restart: Restart DIMP test service
dimp-restart: dimp-down dimp-up
	@echo "✓ DIMP service restarted"

## dimp-test: Run integration tests with DIMP service
dimp-test:
	@echo "Running integration tests with DIMP service..."
	@echo "Starting DIMP service..."
	@cd dimp && docker compose up -d
	@echo "Waiting for service to be ready..."
	@sleep 3
	@echo "Running tests..."
	@cd ../.. && go test -v ./tests/integration/pipeline_dimp_test.go ./tests/integration/pipeline_dimp_resume_test.go
	@echo "Stopping DIMP service..."
	@cd dimp && docker compose down
	@echo "✓ DIMP tests complete"

## test-with-services: Run all tests with required services
test-with-services:
	@echo "Starting all test services..."
	@cd dimp && docker compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 3
	@echo "Running full test suite..."
	@cd ../.. && go test -v ./tests/integration/...
	@echo "Stopping test services..."
	@cd dimp && docker compose down
	@echo "✓ All tests complete"

## clean-test-data: Clean test data and temporary files
clean-test-data:
	@echo "Cleaning test data..."
	@rm -rf ../../jobs/*
	@echo "✓ Test data cleaned"
