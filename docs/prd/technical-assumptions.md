# Technical Assumptions

## Repository Structure: Monorepo

A single repository with Go standard layout:
```
/cmd         # CLI entry points
/internal    # Private application code
/pkg         # Shared libraries (if needed)
```

**Rationale**: Small team (1-3 developers), single codebase for CLI and backend simplifies development and deployment.

## Service Architecture

**CLI as API Client + Stateless Backend**

- **CLI/TUI**: Go application using Cobra + Viper for commands, Bubble Tea + Lip Gloss for TUI
- **Backend API**: Go + Gin (REST API), containerized with Docker
- **Separation**: CLI makes HTTP calls to backend; backend handles all data persistence and business logic

**Rationale**: Clean separation allows CLI updates independent of backend; stateless backend enables horizontal scaling if needed.

## Testing Requirements

**Unit + Integration Testing**

- `go test` for unit tests
- `golangci-lint` on every PR
- Integration tests for API endpoints
- MailHog for email testing in development

Manual testing convenience methods for TUI interactions (TUI testing is notoriously difficult to automate).

**Rationale**: Balance thoroughness with development velocity; TUI testing automation deferred for pragmatic reasons.

## Additional Technical Assumptions and Requests

### Languages & Frameworks
- **Go 1.24+** — Primary language for both CLI and backend
- **Cobra** — CLI framework (industry standard)
- **Viper** — Configuration management
- **Bubble Tea** — TUI framework (Charm ecosystem)
- **Lip Gloss** — TUI styling
- **Gin** — HTTP framework for REST API

### Database
- **SQLite** initially with **Repository pattern** — Enables future PostgreSQL migration without code changes
- Database file stored server-side (not in CLI)

### Authentication
- **GitHub OAuth** — Device flow for CLI authentication
- **Token-based API auth** — PASETO (`v4.local`) for authenticated API requests

### Email
- **SMTP standard** — For hotel communication
- **SendGrid/Mailgun** — Production email delivery
- **MailHog** — Local development/testing

### CI/CD
- **GoReleaser** — Cross-platform builds
- **GitHub Actions** — CI pipeline
- **Git tags** — Trigger releases
- **GitHub Releases** — Binary distribution
- **Optional Homebrew tap** — For macOS users

### Deployment
- **Docker** — Containerized backend
- **Free-tier hosting** — Community budget constraint (Fly.io, Railway, or similar)

### Security
- **HTTPS only** — All API communication
- **Secure token storage** — CLI stores auth tokens in OS keychain

---
