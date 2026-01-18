# Implementation Plan: Backend Initialization & Conference Listing

**Branch**: `002-backend-conference-list` | **Date**: 2026-01-17 | **Spec**: [specs/002-backend-conference-list/spec.md](spec.md)
**Input**: Feature specification from `specs/002-backend-conference-list/spec.md`

## Summary

Initialize the UNCONF backend service as a containerized Go application using the Gin framework. Implement the first version of the conference listing API following OpenAPI standards. Create the corresponding CLI command to consume this API. The communication between the two components will be governed by explicit pact contracts (Consumer-Driven Contract Testing).

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: 
- `github.com/gin-gonic/gin` (Backend API)
- `github.com/spf13/cobra` (CLI)
- `github.com/go-resty/resty/v2` (HTTP Client)
- `github.com/pact-foundation/pact-go/v2` (Contract Testing)
- `go.uber.org/zap` (Structured Logging)
**Storage**: Static JSON file (`internal/backend/data/conferences.json`)
**Testing**: Go testing + testify + Pact Go
**Target Platform**: Docker (Linux/macOS), Unix-like CLI
**Project Type**: Web Application (Backend + CLI Client)
**Performance Goals**: <5s container startup, <2s CLI command response
**Constraints**: OpenAPI 3.0, TDD with Pact, Structured JSON logging to stdout
**Scale/Scope**: 1 Service, 1 CLI command, 2 Endpoints (`/conferences`, `/health`)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Repository Pattern**: **REQUIRED**. Backend will use a `ConferenceRepository` interface to decouple the JSON file storage.
- [x] **Privacy by Design**: N/A (Conference list is public information).
- [x] **Real-time State**: CLI fetches fresh data from the API on every call.
- [x] **Email Automation**: N/A.
- [x] **Test-First Development**: **CRITICAL**. We will write the Pact consumer test (CLI) first, generate the contract, and then implement the backend provider to fulfill it.
- [x] **CLI-First Design**: **CRITICAL**. `unconf list` command will support the `--json` flag.

## Project Structure

### Documentation (this feature)

```text
specs/002-backend-conference-list/
├── plan.md              # This file
├── research.md          # Technology decisions
├── data-model.md        # Entity structure
├── quickstart.md        # Build/Run instructions
├── contracts/           # OpenAPI schema
└── tasks.md             # Implementation tasks (Phase 2)
```

### Source Code (repository root)

```text
cmd/
├── unconf/              # CLI entry point
└── backend/             # Backend API entry point

internal/
├── cli/                 # CLI Command logic
├── backend/             # API Handlers and Repositories
│   ├── data/            # Mock JSON data
│   ├── handlers/        # Gin handlers
│   └── repository/      # Interface and JSON implementation
└── domain/              # Shared entities (Conference)

tests/
├── contract/            # Pact tests (Consumer & Provider)
├── integration/         # E2E CLI tests
└── unit/                # Unit tests for CLI and Backend logic

build/
└── package/             # Dockerfiles
    └── backend.Dockerfile
```

**Structure Decision**: Expanded the existing structure to accommodate a multi-component layout (Backend + CLI). Shared business entities are moved to `internal/domain` to be accessible by both components.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Multi-binary layout | CLI needs a separate server to fetch data from as per spec. | Direct file reading in CLI would bypass the API requirement. |
| Contract Testing | Explicitly requested in spec. | Simple integration tests wouldn't provide the same contract decoupling. |