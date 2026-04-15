# UNCONF CLI

UNCONF is a Go-based CLI + TUI application for unconference registration, room selection, and booking workflows.

## Project Overview

- CLI-first experience with a REST backend
- Tech stack centered on Go 1.24+, Gin, Cobra, Bubble Tea, and SQLite
- Architecture and development standards are documented under `docs/`

## Current Status

- Core attendee flows are implemented: `login`, `logout`, `config`, `list`, `info`, `checkout`, `rooms`, `book`, `status`, `cancel`, `attendees`, `invite`, and `requests`.
- Organizer flows are also implemented: `create`, `edit`, `rooms add|edit|remove|import`, `dashboard`, `export`, and organizer-assignment API endpoints.
- Follow-on hardening stories were completed for auth/session lifecycle and status/cancel consistency.
- Known release-readiness gap: the CLI, client, and TUI all submit bookings through `POST /bookings`, but the server currently registers only `GET /bookings` and `DELETE /bookings/{id}`. Existing bookings can still be listed, cancelled, exported, and used in attendee/request flows.

## Quick Start

```bash
make build
./bin/unconf
```

Implemented command examples:

```bash
unconf login
unconf checkout socrates-26
unconf rooms
unconf status
unconf dashboard
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

MailHog development verification (optional, local only):

```bash
UNCONF_MAILHOG_TEST=1 \
UNCONF_SMTP_HOST=localhost \
UNCONF_SMTP_PORT=1025 \
UNCONF_MAILHOG_API_URL=http://localhost:8025 \
go test ./internal/email -run TestMailHog_NewBookingRenderAndSend -count=1
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

## Database

- Default SQLite file: `.unconf.db`
- Override with environment variable: `UNCONF_DB_PATH`
- Database initialization + migrations run automatically on CLI and server startup

### Database Migration

Run migrations locally:

```bash
make migrate-up
```

Rollback migrations:

```bash
make migrate-down
```

Recreate schema from scratch:

```bash
make migrate-fresh
```

Migration files live in `migrations/` and follow `{version}_{name}.up.sql` / `{version}_{name}.down.sql`.

golang-migrate docs: https://github.com/golang-migrate/migrate

## Docker

Build image:

```bash
docker build -t unconf:latest .
```

Run container:

```bash
docker volume create unconf-data
docker run --rm -p 8080:8080 \
	-e UNCONF_DB_PATH=/data/unconf.db \
	-v unconf-data:/data \
	unconf:latest
```

This keeps SQLite data across container restarts.

## Linting

```bash
make lint
```

## CI/CD

GitHub Actions workflows are defined in `.github/workflows/`:

- `ci.yaml` runs on `pull_request` and on push to `main`
- `release.yaml` runs on push of tags matching `v*`, builds all release targets, and publishes a GitHub Release
- `deploy.yaml` runs on push of tags matching `v*` (placeholder)

Run the same quality gates locally:

```bash
make test
make lint
make build
```

Create a release tag to trigger the release workflow:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Local release smoke test:

```bash
make release-local
```

Note: SQLite uses `github.com/mattn/go-sqlite3` (CGO), so `make release-local` runs a single-target GoReleaser build to avoid cross-compilation toolchain issues.

## Project Structure

Project structure reference: `docs/architecture/11-unified-project-structure.md`

## Documentation

- Architecture docs: `docs/architecture/`
- Product docs: `docs/prd/`

## Contributing

Follow coding standards in `docs/architecture/coding-standards.md`.
