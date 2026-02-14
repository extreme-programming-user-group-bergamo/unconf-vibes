# UNCONF CLI

UNCONF is a Go-based CLI + TUI application for unconference registration, room selection, and booking workflows.

## Project Overview

- CLI-first experience with a REST backend
- Tech stack centered on Go 1.24+, Gin, Cobra, Bubble Tea, and SQLite
- Architecture and development standards are documented under `docs/`

## Quick Start

```bash
make build
./bin/unconf
```

Planned command examples:

```bash
unconf login
unconf help
```

## Development Setup

Prerequisites:

- Go 1.24+
- SQLite 3.40+
- Make
- golangci-lint v2.x

## Building

```bash
make build
```

This produces:

- `bin/unconf`
- `bin/unconf-server`

## Testing

```bash
make test
```

## Linting

```bash
make lint
```

## Project Structure

Project structure reference: `docs/architecture/11-unified-project-structure.md`

## Documentation

- Architecture docs: `docs/architecture/`
- Product docs: `docs/prd/`

## Contributing

Follow coding standards in `docs/architecture/coding-standards.md`.
