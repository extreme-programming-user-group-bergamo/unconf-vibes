# 11. Unified Project Structure

```
unconf/
├── .github/
│   └── workflows/
│       ├── ci.yaml              # Test, lint on PR
│       ├── release.yaml         # GoReleaser on tag
│       └── deploy.yaml          # Deploy to Fly.io
├── cmd/
│   ├── unconf/                  # CLI binary
│   │   └── main.go
│   └── server/                  # API server binary
│       └── main.go
├── internal/
│   ├── api/                     # HTTP API (Gin)
│   ├── cli/                     # CLI commands (Cobra)
│   ├── tui/                     # TUI components (Bubble Tea)
│   ├── service/                 # Business logic
│   ├── repository/              # Data access
│   │   ├── interfaces.go
│   │   └── sqlite/
│   ├── models/                  # Domain models
│   ├── auth/                    # Authentication
│   ├── email/                   # Email service
│   ├── config/                  # Configuration
│   └── client/                  # HTTP client for CLI
├── migrations/                  # Database migrations
├── configs/                     # Configuration files
├── scripts/                     # Build/deploy scripts
├── docs/                        # Documentation
├── Dockerfile
├── docker-compose.yml
├── fly.toml
├── Makefile
├── .goreleaser.yaml
├── .golangci.yaml
├── go.mod
└── README.md
```

---
