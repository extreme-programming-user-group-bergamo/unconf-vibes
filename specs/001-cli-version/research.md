# Phase 0: Research & Technical Decisions

**Feature**: 001-cli-version
**Date**: 2026-01-17

## Technical Context Resolution

### 1. Language & Dependencies
**Decision**: Go 1.21+ with Cobra framework.
**Rationale**: Explicitly defined in `MVP.md` and `CLAUDE.md`. Cobra is the de-facto standard for Go CLIs (used by kubectl, docker, etc.), enabling the required "Unix-like" subcommand structure (`unconf <command>`).
**Alternatives**: `urfave/cli` (popular but less standard for complex apps), `flag` stdlib (too simple for the full MVP scope).

### 2. Project Structure
**Decision**: Standard Go Layout (`cmd/unconf/main.go`).
**Rationale**: Aligns with `CLAUDE.md` target structure and Go community standards.
- `cmd/unconf/main.go`: Entry point.
- `cmd/root.go`: Root command definition (where global flags like `--version` often live or are tied to).
- `cmd/version.go`: Explicit version subcommand.
**Reference**: [Standard Go Project Layout](https://github.com/golang-standards/project-layout).

### 3. Version Injection
**Decision**: Use `ldflags` at build time.
**Rationale**: Standard Go practice to avoid hardcoding versions in source code.
**Implementation**:
```bash
go build -ldflags "-X main.version=0.1.0" -o unconf ./cmd/unconf
```
The application will have a `version` variable in the root package (or a `cmd` package) that defaults to "dev" but is overwritten by the linker.

### 4. Testing Strategy
**Decision**: `testify` for assertions, `os/exec` for integration tests of the binary.
**Rationale**: `spec.md` requires strict TDD.
- **Unit Tests**: Test the Cobra command execution and output capture directly in Go.
- **Integration Tests**: Build the binary and run it to verify exit codes and stdout (ensures `main` package wiring is correct).

## Constitution Check Analysis
- **Repository Pattern**: N/A (no database).
- **Privacy**: N/A.
- **Real-time State**: N/A.
- **Email**: N/A.
- **TDD**: **CRITICAL**. Must write tests before implementation.
- **CLI-First**: **CRITICAL**. Must support `--json` (future) and standard flags (`--help`, `--version`).

## Unknowns Resolved
- **Platform**: Cross-platform Go binary (macOS/Linux primary targets per "Unix-like" requirement).
- **Performance**: <100ms startup is standard for Go binaries.
