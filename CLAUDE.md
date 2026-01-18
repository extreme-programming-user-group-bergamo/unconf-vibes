# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**UNCONF** - A CLI application with TUI for unconference registration with integrated room selection and automated hotel communication. Built for developer conferences like SoCraTes Italia and Polenta e Deploy.

See [MVP.md](MVP.md) for product vision and CLI command specifications. See [.specify/memory/constitution.md](.specify/memory/constitution.md) for architectural principles.

**Status**: Specification phase - no code implemented yet.

## Tech Stack

- **CLI App**: Go 1.21+ with Cobra (commands), **MANDATORY** Bubble Tea/Bubbles/Lip Gloss (TUI) following the "Crush" aesthetic for interactive views.
- **Backend**: Go with Gin (REST API), containerized with Docker
- **Database**: SQLite (designed for PostgreSQL migration via repository pattern)
- **Config**: Viper for YAML/env configuration (`~/.unconf.yaml`)
- **Testing**: Go testing + testify, contract tests for API
- **Email**: SendGrid/Mailgun/SMTP (MailHog for local testing)

## Current Structure

```text
.claude/commands/        # Claude Code slash commands (speckit workflow)
.gemini/commands/        # Gemini CLI commands
.specify/
├── memory/              # constitution.md - architectural principles
├── scripts/             # Bash scripts for feature workflow
└── templates/           # Templates for spec, plan, tasks
specs/
└── 001-*/               # Feature folders with spec.md, plan.md, tasks.md
```

## Target Structure (after implementation)

```text
cmd/
├── unconf/              # CLI entry point
│   └── main.go
├── config.go            # `unconf config` command
├── list.go              # `unconf list` command
├── checkout.go          # `unconf checkout` command
├── info.go              # `unconf info` command
├── rooms.go             # `unconf rooms` command (TUI)
├── book.go              # `unconf book` command
├── status.go            # `unconf status` command
└── cancel.go            # `unconf cancel` command

internal/
├── api/                 # HTTP client for backend communication
├── config/              # Viper configuration handling
├── models/              # Domain entities (Participant, Room, Booking)
├── tui/                 # Bubble Tea views and components
│   ├── rooms/           # Room selection TUI
│   └── common/          # Shared TUI components
└── utils/               # Helpers, formatters

server/
├── cmd/                 # Server entry point
├── api/                 # Gin routes and handlers
├── models/              # Domain entities
├── repositories/        # Data access interfaces + SQLite implementations
├── services/            # Business logic
└── middleware/          # Auth, validation, rate limiting

tests/
├── contract/            # API contract tests
├── integration/         # End-to-end CLI tests
└── unit/                # Unit tests
```

## CLI Commands

- `unconf config` - Configure user settings (name, email, privacy)
- `unconf list` (alias: `ls`) - List available conferences
- `unconf checkout <id>` - Set active conference context
- `unconf info` - Show conference details
- `unconf rooms` (alias: `map`) - Interactive TUI for room selection
- `unconf book [room_id]` - Book a room (wizard or direct)
- `unconf status` (alias: `whoami`) - Show current booking status
- `unconf cancel` - Cancel current booking

See [MVP.md](MVP.md) for complete command specifications.

## Constitutional Principles (NON-NEGOTIABLE)

These rules are defined in `.specify/memory/constitution.md`:

1. **Repository Pattern**: All database access through interfaces. No direct SQL in business logic.

2. **Privacy by Design**: Privacy defaults to private (opt-in). Name visibility is user-controlled.

3. **Real-time State**: Room data syncs within 2 seconds. No stale data displayed.

4. **Email Automation**: All bookings trigger automated hotel notification.

5. **Test-First Development (TDD)**: Tests written before implementation. Red-Green-Refactor.

6. **CLI-First Design**: Unix-like conventions, context system, `--json` output support.

## Authentication: Magic Links (Passwordless)

The system uses passwordless authentication via magic links:

1. User enters email to request login
2. System generates 32-byte cryptographically secure token (15-min expiry)
3. Email sent with one-time login link
4. User clicks link, token validated, JWT session created

**Key Details**:

- Token stored as SHA-256 hash (never plaintext)
- Single-use tokens (invalidated after login)
- Rate limiting: 3 requests per email per hour
- Sessions: 30-min JWT stored locally in `~/.unconf.yaml`

## Development Commands

```bash
# Build CLI
go build -o unconf ./cmd/unconf

# Run CLI
./unconf list
./unconf rooms --interactive

# Run backend server
go run ./server/cmd/main.go

# Run tests
go test ./...
go test -v ./tests/integration/...

# Run with race detector
go test -race ./...
```

## Key Entities

- **Participant**: Email, profile (name, dietary, accessibility), nameVisible toggle
- **Conference**: ID, name, dates, location, rooms configuration
- **Room**: ID, type (single/double/triple), price, features, occupants
- **Booking**: Participant + Room + Conference, status, notes

## Security Requirements

- Sessions: 30-minute timeout, stored securely in local config
- Magic links: 256-bit random tokens, 15-min expiry, single-use
- Rate limiting: 3 magic link requests/email/hour
- GDPR: Data export, deletion rights, consent tracking
