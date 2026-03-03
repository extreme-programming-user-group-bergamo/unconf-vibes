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

Manual API check:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

## API Server

Start API server:

```bash
make run-server
```

- Default port: `8080`
- Health endpoint: `http://localhost:8080/health`
- API endpoint reference: `docs/architecture/5-api-specification.md`

## Docker

Build image:

```bash
docker build -t unconf:latest .
```

Run container:

```bash
docker run --rm -p 8080:8080 unconf:latest
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
