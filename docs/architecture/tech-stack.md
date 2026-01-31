# UNCONF CLI Tech Stack

> Definitive technology choices for the project. See [architecture.md](../architecture.md) for full context.

## Core Technologies

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **Language** | Go | 1.21+ | All application code |
| **CLI Framework** | Cobra | v1.8+ | Command structure |
| **CLI Config** | Viper | v1.18+ | Configuration management |
| **TUI Framework** | Bubble Tea | v0.25+ | Interactive terminal UI |
| **TUI Styling** | Lip Gloss | v0.10+ | Terminal styling |
| **HTTP Framework** | Gin | v1.9+ | REST API |
| **Database** | SQLite | 3.40+ | Data persistence |
| **DB Driver** | mattn/go-sqlite3 | v1.14+ | SQLite bindings |
| **Migrations** | golang-migrate | v4.16+ | Schema migrations |

## Authentication & Security

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **JWT** | golang-jwt/jwt | v5+ | API authentication |
| **OAuth** | golang.org/x/oauth2 | latest | GitHub integration |
| **Token Storage** | zalando/go-keyring | v0.2+ | Secure CLI storage |

## HTTP & Networking

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **HTTP Client** | go-resty/resty | v2.11+ | CLI API calls |
| **Validation** | go-playground/validator | v10+ | Input validation |

## Logging & Monitoring

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **Logging** | rs/zerolog | v1.32+ | Structured logging |

## Email

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **SMTP Client** | go-gomail/gomail | v2.0+ | Email sending |
| **Dev Testing** | MailHog | v1.0+ | Local email capture |
| **Production** | SendGrid | API v3 | Email delivery |

## Testing & Quality

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **Testing** | go test | stdlib | Unit/integration tests |
| **Assertions** | stretchr/testify | v1.8+ | Test utilities |
| **Linting** | golangci-lint | v1.55+ | Static analysis |

## Build & Deploy

| Category | Technology | Version | Purpose |
|----------|------------|---------|---------|
| **Task Runner** | Make | - | Build automation |
| **Release** | GoReleaser | v1.24+ | Cross-platform builds |
| **CI/CD** | GitHub Actions | - | Continuous integration |
| **Container** | Docker | 24+ | Server packaging |
| **Hosting** | Fly.io | - | Production deployment |

---

## Quick Install (Development)

```bash
# Required
brew install go          # Go 1.21+
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
| Go 1.21+ | Required for `log/slog`, improved generics |
| SQLite 3.40+ | JSON functions, window functions |
| Bubble Tea 0.25+ | Stable API, performance improvements |

## Migration Path

**SQLite → PostgreSQL:**
- Repository interfaces enable swap without service changes
- Triggers: >500 users, multiple concurrent conferences
- Implementation: Add `internal/repository/postgres/` package

**Fly.io → Other:**
- Docker-based, portable to any container platform
- Volume mount pattern works on most PaaS

---

## Version Pinning

```go
// go.mod excerpt
require (
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.0
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.10.0
    github.com/gin-gonic/gin v1.9.1
    github.com/mattn/go-sqlite3 v1.14.22
    github.com/golang-migrate/migrate/v4 v4.17.0
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/go-resty/resty/v2 v2.11.0
    github.com/rs/zerolog v1.32.0
    github.com/go-playground/validator/v10 v10.18.0
    github.com/stretchr/testify v1.8.4
    github.com/zalando/go-keyring v0.2.3
    gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
)
```

---

## Not Using (and Why)

| Technology | Reason Not Chosen |
|------------|-------------------|
| **PostgreSQL** | MVP simplicity; SQLite sufficient for <500 users |
| **Redis** | No caching needed; room availability must be real-time |
| **gRPC** | REST simpler for CLI client; no streaming needed |
| **GraphQL** | PRD specifies REST; CRUD operations fit REST well |
| **GORM** | Repository pattern + raw SQL preferred for control |
| **Echo/Chi** | PRD specifies Gin; all three are good choices |
| **Zap** | zerolog faster, simpler API |
