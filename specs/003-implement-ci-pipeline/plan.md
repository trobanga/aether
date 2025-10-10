# Implementation Plan: GitHub CI Pipeline

**Branch**: `003-implement-ci-pipeline` | **Date**: 2025-10-10 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-implement-ci-pipeline/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement a comprehensive CI/CD pipeline using GitHub Actions that automatically validates code quality through linting, runs unit and integration tests, orchestrates Docker-based test services, and executes end-to-end tests. The pipeline will run on every push and pull request, providing fast feedback (within 3 minutes for lint + unit tests) and ensuring no broken code reaches the main branch through required status checks.

## Technical Context

**Language/Version**: Go 1.25.1 (minimum Go 1.21 as per project requirements)
**Primary Dependencies**:
- GitHub Actions (CI platform)
- Docker & Docker Compose (test service orchestration)
- golangci-lint (linting - NEEDS CLARIFICATION: installation method in CI)
- Existing test infrastructure: `.github/test/` with DIMP stack (PostgreSQL, VFPS, FHIR Pseudonymizer)

**Storage**: N/A (CI pipeline configuration only)
**Testing**:
- Go testing framework (`go test`)
- Existing Makefile targets: `make lint`, `make test-unit`, `make test-integration`
- E2E test script: `.github/test/test-dimp.sh`
- testify library for assertions

**Target Platform**: GitHub Actions runners (ubuntu-latest with Docker support)
**Project Type**: Single Go CLI application with Docker-based test services
**Performance Goals**:
- Lint + unit tests: < 3 minutes
- Full pipeline (lint + unit + integration + E2E): < 25 minutes
- Service startup: < 60 seconds with health checks

**Constraints**:
- Must use existing Makefile commands (no CI-specific test implementations)
- Must reuse `.github/test/` Docker Compose infrastructure
- Must support dependency caching to achieve 40% speedup on subsequent runs
- GitHub Actions free tier limitations (runner concurrency, storage)

**Scale/Scope**:
- 4 pipeline stages (lint, unit, integration, E2E)
- 3 Docker services (vfps_db, vfps, fhir-pseudonymizer)
- Test directories: unit/, integration/, contract/
- Artifact types: test results, logs, coverage reports

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Functional Programming
**Status**: ✅ **PASS** (Not Applicable)

CI pipeline configuration is declarative YAML with imperative shell scripts. No application logic is being written, so functional programming principles don't apply. The feature creates workflow definitions, not Go code.

### Principle II: Test-Driven Development (TDD)
**Status**: ✅ **PASS** (With Validation Strategy)

This feature IS the testing infrastructure itself. TDD compliance will be validated through:
1. **Validation approach**: Create test workflow configurations that intentionally fail, verify they're caught correctly, then ensure they pass
2. **Test the tests**: Each pipeline stage will be validated by:
   - **Lint stage**: Push code with intentional linting violations → verify failure → fix → verify pass
   - **Unit tests**: Introduce failing test → verify CI catches it → fix → verify pass
   - **Integration tests**: Break service integration → verify failure with logs → fix → verify pass
   - **E2E tests**: Break end-to-end flow → verify failure with diagnostics → fix → verify pass
3. **User approval**: Pipeline configuration will be reviewed before implementation

### Principle III: Keep It Simple, Stupid (KISS)
**Status**: ✅ **PASS**

The design prioritizes simplicity:
- **Reuse over recreation**: Uses existing Makefile commands (`make lint`, `make test-unit`, etc.) instead of duplicating logic in CI
- **Existing infrastructure**: Leverages `.github/test/` Docker Compose setup instead of creating new test orchestration
- **Standard patterns**: Uses GitHub Actions' built-in features (caching, artifacts, matrix builds) rather than custom solutions
- **No premature optimization**: Starts with straightforward sequential stages; parallelization only if needed

**Complexity justified**:
- Docker service orchestration is necessary (required by integration/E2E tests)
- Multi-stage pipeline is necessary (each stage serves distinct purpose per requirements)
- Caching is necessary (40% speedup requirement from SC-010)

### Gate Decision: ✅ **PROCEED TO PHASE 0**

All constitution principles are satisfied. This is infrastructure configuration with clear validation strategy, not application code requiring functional purity.

## Project Structure

### Documentation (this feature)

```
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
.github/
└── workflows/
    └── ci.yml              # Main CI workflow (NEW - created by this feature)

.github/test/               # Existing test infrastructure (used by CI)
├── docker-compose.yaml     # DIMP test services
├── dimp/                   # DIMP service definitions
├── test-dimp.sh           # E2E test script
└── aether.yaml            # Test configuration

tests/                      # Existing test suites (executed by CI)
├── contract/              # Contract tests
├── integration/           # Integration tests
└── unit/                  # Unit tests

Makefile                    # Existing build commands (used by CI)
go.mod                      # Go dependencies
```

**Structure Decision**: This feature adds GitHub Actions workflow configuration only. The structure follows GitHub Actions conventions (`.github/workflows/`). All test code, test infrastructure, and build commands already exist - the CI pipeline orchestrates them.

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

No violations. All complexity is justified and aligned with KISS principle (see Constitution Check above).

## Post-Design Constitution Re-evaluation

*Re-check after Phase 1 design complete*

### Principle I: Functional Programming
**Status**: ✅ **PASS** (Still Not Applicable)

The design artifacts (data-model.md, workflow-contract.md, quickstart.md) confirm this is declarative GitHub Actions YAML configuration. No Go code is being written as part of this feature. The workflow orchestrates existing Makefile commands and scripts.

### Principle II: Test-Driven Development (TDD)
**Status**: ✅ **PASS** (Validated)

The design includes comprehensive validation strategy (see quickstart.md "Validation Checklist"):
- Each pipeline stage has specific test scenarios (intentionally failing tests → verify detection → verify pass)
- Workflow contract (contracts/workflow-contract.md) defines testable behaviors with exit codes and expected outcomes
- Quickstart provides step-by-step testing procedures for each job
- Validation can be performed independently for each stage

This aligns with TDD principles: define expected behavior, test it fails when broken, verify it passes when working.

### Principle III: Keep It Simple, Stupid (KISS)
**Status**: ✅ **PASS** (Design Reinforces Simplicity)

The detailed design maintains simplicity:
- **Reuse confirmed**: Research shows we use existing Makefile targets without modification
- **No custom tooling**: Uses official GitHub Actions and golangci-lint action
- **Standard patterns**: Docker Compose orchestration follows documented best practices
- **Minimal abstraction**: 4 jobs with clear dependencies, no complex matrix builds or custom actions
- **Explicit over implicit**: All timeouts, retries, and cleanup steps are explicitly defined

The design adds no unnecessary complexity beyond what's required by the functional requirements.

### Final Gate Decision: ✅ **APPROVED FOR IMPLEMENTATION**

All constitution principles remain satisfied after detailed design. The feature:
1. Is infrastructure configuration, not application code (FP N/A)
2. Has comprehensive validation strategy (TDD satisfied)
3. Maintains simplicity through reuse and standard patterns (KISS satisfied)

**Ready for**: `/speckit.tasks` command to generate implementation tasks.
