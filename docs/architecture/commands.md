# Commands

> Make targets for building, testing, and managing the UNCONF project. Run from the repository root.

## Development

| Command | Description |
|---------|-------------|
| `make build` | Build both CLI and server binaries to `bin/` |
| `make run-cli` | Build and run the CLI (loads `.env`) |
| `make run-server` | Run the API server (loads `.env`) |
| `make build-server` | Build the server binary only |

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
