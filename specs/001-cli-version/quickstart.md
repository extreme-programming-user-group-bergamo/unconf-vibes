# Phase 1: Quickstart Guide

**Feature**: 001-cli-version

## Prerequisites

- Go 1.21 or higher installed
- `make` (optional, for build scripts)

## Building the CLI

To build the CLI with version information:

```bash
# Basic build
go build -o unconf ./cmd/unconf

# Build with version injection (simulated for dev)
go build -ldflags "-X 'github.com/xpugbg/unconf-vibes/cmd.Version=0.1.0'" -o unconf ./cmd/unconf
```

## Running the Feature

### Check Version
```bash
./unconf --version
# Output: unconf version 0.1.0
```

### Check Help
```bash
./unconf --help
# Output: Usage information...
```

## Running Tests

```bash
# Unit tests
go test ./...

# Integration tests (verifying binary output)
go test -v ./tests/integration/...
```
