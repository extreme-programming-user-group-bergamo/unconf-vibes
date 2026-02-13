# 3. Tech Stack

This is the **definitive technology selection** for UNCONF CLI. All development must use these exact technologies and versions.

| Category | Technology | Version | Purpose | Rationale |
|----------|------------|---------|---------|-----------|
| **Core Language** | Go | 1.24+ | CLI, API, all backend | Performance, single binary distribution, strong concurrency |
| **CLI Framework** | Cobra | v1.10+ | Command structure, flags, help | Industry standard (kubectl, gh, docker) |
| **CLI Config** | Viper | v1.21+ | Configuration management | Seamless Cobra integration, multi-source config |
| **TUI Framework** | Bubble Tea | v1.3+ | Interactive terminal UI | Best-in-class Go TUI, Charm ecosystem |
| **TUI Styling** | Lip Gloss | v1.1+ | Terminal styling, colors | Native Bubble Tea companion |
| **HTTP Framework** | Gin | v1.11+ | REST API routing | Fast, minimal, well-documented |
| **Database** | SQLite | 3.40+ | Data persistence | Simple, file-based, migration-ready |
| **DB Driver** | mattn/go-sqlite3 | v1.14+ | SQLite Go bindings | Most mature SQLite driver |
| **Migrations** | golang-migrate | v4.19+ | Schema migrations | Database-agnostic, CLI and library |
| **Authentication** | JWT | golang-jwt/jwt/v5 v5.3+ | API token authentication | Maintained community fork |
| **OAuth Client** | golang.org/x/oauth2 | latest | GitHub OAuth integration | Official Go OAuth library |
| **HTTP Client** | net/http + resty | v2.17+ | CLI API calls | Resty for fluent API, retries |
| **Logging** | log/slog | stdlib (Go 1.24+) | Structured logging | Standard library, no extra dependency |
| **Email** | gomail | v2.0+ | SMTP email sending | Simple, reliable SMTP client |
| **Validation** | go-playground/validator | v10.30+ | Input validation | Struct tags, comprehensive rules |
| **Testing** | go test + testify | v1.11+ | Unit & integration tests | Standard + assertions/mocks |
| **Linting** | golangci-lint | v2.x (pin latest stable) | Static analysis | Multi-linter aggregator |
| **Build Tool** | Make | - | Task runner | Universal, simple |
| **Release** | GoReleaser | v1.24+ | Cross-platform builds | Automated releases, checksums |
| **CI/CD** | GitHub Actions | - | Continuous integration | Native GitHub integration |
| **Containerization** | Docker | 24+ | Backend deployment | Standard containerization |
| **Hosting** | Fly.io | - | Production deployment | Free tier, Docker support |
| **Dev Email** | MailHog | v1.0+ | Local email testing | Captures SMTP in development |
| **Prod Email** | SendGrid | API v3 | Production email delivery | Reliable, free tier available |

---
