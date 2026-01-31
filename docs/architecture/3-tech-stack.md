# 3. Tech Stack

This is the **definitive technology selection** for UNCONF CLI. All development must use these exact technologies and versions.

| Category | Technology | Version | Purpose | Rationale |
|----------|------------|---------|---------|-----------|
| **Core Language** | Go | 1.21+ | CLI, API, all backend | Performance, single binary distribution, strong concurrency |
| **CLI Framework** | Cobra | v1.8+ | Command structure, flags, help | Industry standard (kubectl, gh, docker) |
| **CLI Config** | Viper | v1.18+ | Configuration management | Seamless Cobra integration, multi-source config |
| **TUI Framework** | Bubble Tea | v0.25+ | Interactive terminal UI | Best-in-class Go TUI, Charm ecosystem |
| **TUI Styling** | Lip Gloss | v0.10+ | Terminal styling, colors | Native Bubble Tea companion |
| **HTTP Framework** | Gin | v1.9+ | REST API routing | Fast, minimal, well-documented |
| **Database** | SQLite | 3.40+ | Data persistence | Simple, file-based, migration-ready |
| **DB Driver** | mattn/go-sqlite3 | v1.14+ | SQLite Go bindings | Most mature SQLite driver |
| **Migrations** | golang-migrate | v4.16+ | Schema migrations | Database-agnostic, CLI and library |
| **Authentication** | JWT | dgrijalva/jwt-go v5 | API token authentication | Stateless, standard |
| **OAuth Client** | golang.org/x/oauth2 | latest | GitHub OAuth integration | Official Go OAuth library |
| **HTTP Client** | net/http + resty | v2.11+ | CLI API calls | Resty for fluent API, retries |
| **Logging** | zerolog | v1.32+ | Structured logging | Fast, zero-allocation, JSON output |
| **Email** | gomail | v2.0+ | SMTP email sending | Simple, reliable SMTP client |
| **Validation** | go-playground/validator | v10+ | Input validation | Struct tags, comprehensive rules |
| **Testing** | go test + testify | v1.8+ | Unit & integration tests | Standard + assertions/mocks |
| **Linting** | golangci-lint | v1.55+ | Static analysis | Multi-linter aggregator |
| **Build Tool** | Make | - | Task runner | Universal, simple |
| **Release** | GoReleaser | v1.24+ | Cross-platform builds | Automated releases, checksums |
| **CI/CD** | GitHub Actions | - | Continuous integration | Native GitHub integration |
| **Containerization** | Docker | 24+ | Backend deployment | Standard containerization |
| **Hosting** | Fly.io | - | Production deployment | Free tier, Docker support |
| **Dev Email** | MailHog | v1.0+ | Local email testing | Captures SMTP in development |
| **Prod Email** | SendGrid | API v3 | Production email delivery | Reliable, free tier available |

---
