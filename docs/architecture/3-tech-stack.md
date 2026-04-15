# 3. Tech Stack

This is the **definitive technology selection** for UNCONF CLI. All development must use these exact technologies and versions.

| Category | Technology | Version | Purpose | Rationale |
|----------|------------|---------|---------|-----------|
| **Core Language** | Go | 1.24.0 | CLI, API, all backend | Performance, single binary distribution, strong concurrency |
| **CLI Framework** | Cobra | v1.10.2 | Command structure, flags, help | Industry standard (kubectl, gh, docker) |
| **CLI Config** | Viper | v1.21.0 | Configuration management | Seamless Cobra integration, multi-source config |
| **TUI Framework** | Bubble Tea | v1.3.10 | Interactive terminal UI | Best-in-class Go TUI, Charm ecosystem |
| **TUI Styling** | Lip Gloss | v1.1.0 | Terminal styling, colors | Native Bubble Tea companion |
| **HTTP Framework** | Gin | v1.11.0 | REST API routing | Fast, minimal, well-documented |
| **Database** | SQLite | 3.40+ | Data persistence | Simple, file-based, migration-ready |
| **DB Driver** | mattn/go-sqlite3 | v1.14.34 | SQLite Go bindings | Most mature SQLite driver |
| **Migrations** | golang-migrate | v4.19.1 | Schema migrations | Database-agnostic, CLI and library |
| **Authentication** | PASETO | aidanwoods.dev/go-paseto v1.4.0 | API token authentication | Safer by default than JWT defaults and matches the current implementation |
| **GitHub Device Flow** | GitHub REST API via `net/http` | 2022-11-28 API headers | GitHub OAuth integration | Keeps the auth path explicit without an extra OAuth SDK |
| **HTTP Client** | Resty + `net/http` | v2.17.2 | CLI API calls and provider integrations | Resty for fluent API, stdlib where simpler |
| **Logging** | log/slog | stdlib | Structured logging | Standard library, no extra dependency |
| **Email** | `net/smtp` plus internal SendGrid/Mailgun senders | stdlib + HTTP APIs | SMTP and provider delivery | Covers local SMTP plus hosted provider delivery without a heavyweight mail client dependency |
| **Validation** | Gin binding + validator/v10 | indirect via Gin | Input validation | Reuses Gin's validation integration |
| **Testing** | go test + testify | v1.11.1 | Unit & integration tests | Standard + assertions/mocks |
| **Linting** | golangci-lint | v2.x (pin latest stable) | Static analysis | Multi-linter aggregator |
| **Build Tool** | Make | - | Task runner | Universal, simple |
| **Release** | GoReleaser | workflow-managed | Cross-platform builds | Automated releases, checksums |
| **CI/CD** | GitHub Actions | - | Continuous integration | Native GitHub integration |
| **Containerization** | Docker | 24+ | Backend deployment | Standard containerization |
| **Hosting** | Deployment target TBD | - | Production deployment | Docker packaging exists, but runtime hosting config is not committed |
| **Dev Email** | MailHog | v1.0+ | Local email testing | Captures SMTP in development |
| **Prod Email** | SendGrid or Mailgun | API v3 endpoints | Production email delivery | Both providers are supported by the current internal sender abstraction |

---
