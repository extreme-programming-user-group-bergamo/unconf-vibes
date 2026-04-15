# UNCONF CLI Tech Stack

> Definitive technology choices for the project as implemented in the repository. See [architecture.md](../architecture.md) for full context.

## Core Technologies

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **Language** | Go | 1.24.0 | All application code |
| **CLI Framework** | Cobra | v1.10.2 | Command structure |
| **CLI Config** | Viper | v1.21.0 | Configuration management |
| **TUI Framework** | Bubble Tea | v1.3.10 | Interactive terminal UI |
| **TUI Styling** | Lip Gloss | v1.1.0 | Terminal styling |
| **HTTP Framework** | Gin | v1.11.0 | REST API |
| **Database** | SQLite | 3.40+ | Data persistence |
| **DB Driver** | mattn/go-sqlite3 | v1.14.34 | SQLite bindings |
| **Migrations** | golang-migrate | v4.19.1 | Schema migrations |

## Authentication & Security

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **PASETO** | aidanwoods.dev/go-paseto | v1.4.0 | Access-token issue and validation |
| **GitHub Device Flow** | GitHub REST API via `net/http` | 2022-11-28 API headers | CLI-friendly OAuth without browser redirects |
| **Token Storage** | zalando/go-keyring | v0.2.6 | Secure CLI storage for access and refresh tokens |

## HTTP & Networking

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **HTTP Client** | go-resty/resty/v2 | v2.17.2 | CLI API calls |
| **Validation** | Gin binding + validator/v10 | indirect via Gin | Request validation in handlers |

## Logging & Monitoring

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **Logging** | log/slog | stdlib | Structured logging |

## Email

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **SMTP** | `net/smtp` | stdlib | Baseline SMTP delivery |
| **SendGrid** | HTTP API via internal sender | API v3 | Production email delivery option |
| **Mailgun** | HTTP API via internal sender | v3 endpoints | Production email delivery option |
| **Dev Testing** | MailHog | local service | Local email capture and rendering verification |

## Testing & Quality

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **Testing** | go test | stdlib | Unit and integration tests |
| **Assertions** | stretchr/testify | v1.11.1 | Test utilities |
| **Linting** | golangci-lint | v2.x | Static analysis |

## Build & Delivery

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **Task Runner** | Make | - | Build automation |
| **Release** | GoReleaser | workflow-managed | Local snapshot and tagged release builds |
| **CI/CD** | GitHub Actions | - | Continuous integration and packaging |
| **Container** | Docker | 24+ | Server packaging |
| **Deployment** | Docker-based target | TBD | Runtime deployment config is not committed in the repo |

---

## Quick Install (Development)

```bash
# Required
brew install go          # Go 1.24+
brew install sqlite      # SQLite

# Recommended
brew install golangci-lint
brew install golang-migrate
brew install goreleaser

# Optional
brew install docker
```

## Version Constraints

| Constraint | Reason |
|------------|--------|
| Go 1.24+ | Matches the module baseline and current standard-library feature set |
| SQLite 3.40+ | Required for the migration and query patterns used in the repo |
| Bubble Tea 1.3+ | Stable 1.x API and improved terminal behavior |

## Migration Path

**SQLite → PostgreSQL:**
- Repository interfaces enable swap without service changes.
- Trigger when concurrency or event scale outgrows single-file SQLite.
- No PostgreSQL implementation package exists in the current repository yet.

**Current deployment posture:**
- Docker packaging and GitHub workflows exist.
- Runtime deployment target remains intentionally open; no committed Fly.io or equivalent configuration is present.

---

## Version Pinning

```go
// direct dependencies from go.mod
require (
    aidanwoods.dev/go-paseto v1.4.0
    github.com/charmbracelet/bubbletea v1.3.10
    github.com/charmbracelet/lipgloss v1.1.0
    github.com/gin-gonic/gin v1.11.0
    github.com/go-resty/resty/v2 v2.17.2
    github.com/golang-migrate/migrate/v4 v4.19.1
    github.com/mattn/go-sqlite3 v1.14.34
    github.com/spf13/cobra v1.10.2
    github.com/spf13/viper v1.21.0
    github.com/stretchr/testify v1.11.1
    github.com/zalando/go-keyring v0.2.6
)
```

---

## Not Using (and Why)

| Technology | Reason Not Chosen |
|------------|-------------------|
| **PostgreSQL** | SQLite is still sufficient for the current repo scope and simplifies MVP operations |
| **Redis** | No cache tier is needed for the current CLI/API workload |
| **gRPC** | The existing CLI client and JSON REST API cover current requirements well |
| **GraphQL** | The implemented workflow is command-oriented and maps cleanly to REST |
| **GORM** | Repository pattern plus explicit SQL keeps control over schema and queries |
| **External OAuth SDK** | Direct GitHub device-flow HTTP integration keeps the auth path explicit and dependency-light |
