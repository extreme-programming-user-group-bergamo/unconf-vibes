# Phase 1: Quickstart Guide

**Feature**: 002-backend-conference-list

## Prerequisites

- Go 1.21+
- Docker installed
- Pact Go v2 prerequisites (Pact FFI library installed, see [Pact Go docs](https://github.com/pact-foundation/pact-go#prerequisites))
- Pact CLI tools (optional, for manual verification)

## Running the Backend (Docker)

```bash
# Build the backend image
docker build -t unconf-backend -f build/package/backend.Dockerfile .

# Run the container
docker run -p 8080:8080 unconf-backend
```

## Running the CLI

```bash
# Build the CLI
go build -o unconf ./cmd/unconf

# List conferences (requires backend running)
./unconf list

# List conferences in JSON format
./unconf list --json
```

## Running Contract Tests

```bash
# Run all tests including Pact verification
go test ./...
```
