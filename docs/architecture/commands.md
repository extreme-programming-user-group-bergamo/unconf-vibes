# Commands

> Make targets for building, testing, and managing the UNCONF project. Run from the repository root.

## CLI Surfaces

| Binary | Ownership | Examples |
|--------|-----------|----------|
| `unconf` | Attendee + organizer product UX commands | `unconf login`, `unconf checkout socrates-26`, `unconf dashboard` |
| `unconf-server` | Backend/admin operations | `unconf-server serve`, `unconf-server db status` |

## Development

| Command | Description |
|---------|-------------|
| `make build` | Build both CLI and server binaries to `bin/` |
| `make build-cli` | Build only `unconf` with `CGO_ENABLED=0` |
| `make build-server` | Build only `unconf-server` with `CGO_ENABLED=1` |
| `make run-cli` | Build and run `unconf` (loads `.env`) |
| `make run-server` | Run `unconf-server serve` (loads `.env`) |

## Quality

| Command | Description |
|---------|-------------|
| `make test` | Run all tests with race detection |
| `make lint` | Run golangci-lint |

## Database

| Command | Description |
|---------|-------------|
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Roll back all migrations |
| `make migrate-fresh` | Roll back and re-apply all migrations |

## Packaging

| Command | Description |
|---------|-------------|
| `make docker-build` | Build Docker image (`unconf:latest`) |
| `make release-local` | Local GoReleaser build (snapshot) |
| `make clean` | Remove build artifacts |
