# Phase 0: Research & Technical Decisions

**Feature**: 002-backend-conference-list
**Date**: 2026-01-17

## Technical Context Resolution

### 1. Backend Web Framework
**Decision**: `github.com/gin-gonic/gin`
**Rationale**: Explicitly mentioned in `MVP.md`. It's lightweight, fast, and has excellent middleware support for logging and recovery.

### 2. Contract Testing
**Decision**: `github.com/pact-foundation/pact-go/v2`
**Rationale**: Requested "Contract Testing". Pact is the industry standard for Consumer-Driven Contract (CDC) testing. It allows the CLI (Consumer) to define its expectations and the Backend (Provider) to verify it meets them.

### 3. Structured Logging
**Decision**: `go.uber.org/zap`
**Rationale**: High-performance structured logging. Supports JSON output natively, which fits the container observability requirement.

### 4. API Client for CLI
**Decision**: `github.com/go-resty/resty/v2`
**Rationale**: More ergonomic than standard `net/http` for building REST clients. Built-in support for JSON marshaling/unmarshaling and easy error handling.

### 5. Table Rendering for CLI
**Decision**: `github.com/jedib0t/go-pretty/v6/table`
**Rationale**: Standard Go library for rendering beautiful tables in the terminal.

## Constitution Check Analysis
- **TDD (Principle V)**: **CRITICAL**. Contract tests will be the first artifacts implemented. The CLI will define the Pact, and the backend will verify it.
- **CLI-First (Principle VI)**: **CRITICAL**. The `list` command must support `--json`.
- **Repository Pattern (Principle I)**: Although we use mock data, we will design the backend with a `ConferenceRepository` interface to be constitution-compliant from the start.

## Unknowns Resolved
- **API Protocol**: REST over HTTP/1.1.
- **Serialization**: JSON.
- **Port Mapping**: Backend will default to port 8080.
