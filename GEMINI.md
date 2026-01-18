# GEMINI.md

This file provides guidance to Gemini when working with code in this repository.

## Project Overview

**UNCONF** - A CLI application with TUI for unconference registration with integrated room selection and automated hotel communication. Built for developer conferences like SoCraTes Italia and Polenta e Deploy.

The project follows a spec-driven development workflow. The core commands to interact with this workflow are backed by scripts in `.specify/scripts`.

## Tech Stack

- **CLI App**: Go 1.21+ with Cobra (commands), Bubble Tea/Bubbles/Lip Gloss (TUI)
- **TUI Aesthetic**: High-quality, interactive, and visually polished TUIs (referencing the "Crush" style). Static tables should be the fallback for scripting, while interactive modes must use the Charm stack.
- **Backend**: Go with Gin (REST API), containerized with Docker
- **Database**: SQLite (designed for PostgreSQL migration via repository pattern)
- **Config**: Viper for YAML/env configuration (`~/.unconf.yaml`)
- **Testing**: Go testing + testify, contract tests for API
- **Architecture**: Repository Pattern (Non-negotiable), Hexagonal/Clean Architecture principles

## Key Constitutional Principles (NON-NEGOTIABLE)

Defined in `.specify/memory/constitution.md`:

1.  **Repository Pattern**: All database access MUST go through repository interfaces. No direct SQL in business logic.
2.  **Privacy by Design**: Privacy defaults to private (opt-in). Name visibility is user-controlled.
3.  **Real-time State**: Room data syncs within 2 seconds. No stale data displayed.
4.  **Email Automation**: Structured data sent to hotel via configurable templates.
5.  **Test-First Development (TDD)**: Tests MUST be written and approved before implementation. Red-Green-Refactor cycle strictly enforced.
6.  **CLI-First Design**: Unix-like conventions, context system (`checkout`), `--json` output support.

## Current Focus: Feature `001-cli-version`

**Goal**: Implement the basic version of the CLI app with the `--version` option.
**Spec**: `specs/001-cli-version/spec.md`

**Requirements**:
-   Implement the root command and version command/flag.
-   Support `--version`, `-v`, and `version` subcommand.
-   Support `--help` and `-h`.
-   Exit code 0 for success, non-zero for errors.
-   **NO BACKEND** for this feature.
-   **STRICT TDD**: Write tests for version output -> Fail -> Implement -> Pass.

## Development Workflow

### Spec-Driven Development (SDD)
-   **Specs**: Located in `specs/`.
-   **Scripts**: Use scripts in `.specify/scripts/bash/` for feature workflow actions (e.g., `create-new-feature.sh`).

### Standard Commands
-   **Build**: `go build -o unconf ./cmd/unconf`
-   **Run**: `./unconf <command>`
-   **Test**: `go test ./...`
-   **Test (Integration)**: `go test -v ./tests/integration/...`

## File Structure Targets

```text
cmd/
├── unconf/              # CLI entry point (main.go)
└── ...                  # Future commands

internal/
├── tui/                 # Bubble Tea views
├── config/              # Viper config
└── ...

tests/
├── integration/         # E2E CLI tests
└── unit/                # Unit tests
```

## Reference Documents
-   `CLAUDE.md`: Runtime guidance and current project status.
-   `MVP.md`: Product vision and command specifications.
-   `.specify/memory/constitution.md`: Architectural rules and governance.

## Recent Changes
- 003-tui-homepage: Added [if applicable, e.g., PostgreSQL, CoreData, files or N/A]
- 002-backend-conference-list: Added Go 1.21+
- 002-backend-conference-list: Added [if applicable, e.g., PostgreSQL, CoreData, files or N/A]

## Future Tasks / Reminders

- **Distribution**: Set up [GoReleaser](https://goreleaser.com/) with GitHub Actions for automated cross-platform builds and releases. (Refer to MVP.md Section 12).
- **Documentation**: Update README.md and specs with installation instructions once GoReleaser is active (e.g., `brew install ...`).

## Active Technologies
- Static JSON file (`internal/backend/data/conferences.json`) (002-backend-conference-list)
