# UNCONF CLI Source Tree

> Quick reference for the current repository structure. See [architecture.md](../architecture.md) for the full architecture.

## Repository Structure

```text
unconf/
├── .github/
│   ├── agents/                    # BMAD / Copilot agent definitions
│   ├── prompts/                   # Prompt bundles
│   ├── skills/                    # Repo-scoped skills
│   └── workflows/
│       ├── ci.yaml               # Test and lint workflow
│       ├── deploy.yaml           # Deployment workflow scaffold
│       ├── docker-smoke.yml      # Container smoke validation
│       └── release.yaml          # Tagged release builds
├── cmd/
│   ├── server/
│   │   ├── main.go               # `unconf-server` entry point (`serve` wiring + startup lifecycle)
│   │   └── main_integration_test.go
│   └── unconf/
│       └── main.go               # `unconf` product CLI entry point
├── internal/
│   ├── api/
│   │   ├── routes.go             # Gin route registration
│   │   ├── middleware/           # Auth, CORS, logging, organizer checks
│   │   ├── handlers/             # HTTP handlers
│   │   └── responses/            # Error envelope helpers
│   ├── auth/
│   │   ├── github_provider.go    # GitHub device flow client
│   │   ├── token_service.go      # PASETO issue/validate helpers
│   │   └── store.go              # Keyring-backed token storage
│   ├── cli/
│   │   ├── root.go               # `unconf` root command and product workflow wiring
│   │   ├── login.go / logout.go / config.go
│   │   ├── list.go / info.go / checkout.go
│   │   ├── rooms.go / rooms_manage.go
│   │   ├── book.go / status.go / cancel.go
│   │   ├── attendees.go / invite.go / requests.go
│   │   ├── create.go / edit.go / dashboard.go / export.go
│   │   └── conference_forms.go
│   ├── servercli/
│   │   ├── root.go               # `unconf-server` root command and lifecycle wiring
│   │   ├── serve.go              # API startup + graceful shutdown orchestration
│   │   └── db.go                 # Backend/admin DB utilities (`db status`)
│   ├── client/
│   │   ├── client.go / auth_client.go
│   │   ├── bookings.go / rooms.go / requests.go
│   │   ├── attendees.go / dashboard.go / export.go
│   │   └── conferences.go
│   ├── config/
│   │   ├── config.go             # Viper-backed configuration
│   │   └── context.go            # Active conference context manager
│   ├── email/
│   │   ├── provider.go           # Provider selection
│   │   ├── renderer.go / mapper.go / service.go
│   │   ├── smtp_sender.go / sendgrid_sender.go / mailgun_sender.go
│   │   └── templates/
│   │       ├── new_booking.tmpl
│   │       ├── cancellation.tmpl
│   │       └── modification.tmpl
│   ├── models/
│   │   ├── user.go / conference.go / room.go / booking.go / request.go
│   │   ├── conference_organizer.go
│   │   ├── refresh_session.go
│   │   └── email_log.go
│   ├── repository/
│   │   ├── interfaces.go
│   │   ├── errors.go
│   │   ├── mock_repository.go
│   │   └── sqlite/
│   │       ├── db.go
│   │       ├── user_repository.go
│   │       ├── conference_repository.go
│   │       ├── room_repository.go
│   │       ├── booking_repository.go
│   │       ├── request_repository.go
│   │       ├── organizer_repository.go
│   │       ├── refresh_session_repository.go
│   │       └── email_log_repository.go
│   ├── service/
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── conference_service.go
│   │   ├── room_service.go
│   │   ├── booking_service.go
│   │   ├── request_service.go
│   │   ├── attendee_service.go
│   │   ├── organizer_service.go
│   │   ├── hotel_email_service.go
│   │   └── errors.go
│   └── tui/
│       ├── common/               # Shared styles, header, footer
│       ├── rooms/                # Room explorer TUI
│       ├── wizard/               # Booking wizard TUI
│       └── dashboard/            # Organizer dashboard TUI
├── migrations/
│   ├── 1_init.*.sql through 9_email_logs.*.sql
│   └── embed.go                  # Embedded migration registry
├── assets/                       # Logos and image assets
├── bin/                          # Built binaries (`unconf`, `unconf-server`)
├── configs/                      # Present but currently empty
├── docs/                         # PRD, architecture, story, and QA docs
├── Dockerfile
├── Makefile
├── MVP.md
├── README.md
├── go.mod
└── AGENTS.md
```

## Current Layout Notes

- `configs/` exists in the repository but does not currently contain committed configuration files.
- `bin/` contains built artifacts and is not source-of-truth code.
- CLI boundary for this split: `unconf` owns attendee/organizer product workflows; `unconf-server` owns backend/admin operations such as `serve` and `db status`.
- Server runtime + backend DB admin command ownership is in `cmd/server/main.go` + `internal/servercli/`.
- The router wires the attendee booking lifecycle endpoints: `POST /bookings`, `GET /bookings`, and `DELETE /bookings/{id}`.
- Older placeholders such as `/pkg`, `/scripts`, `docker-compose.yml`, and `fly.toml` are not present in the current repository.

## Key Directories

| Directory | Purpose | When to Modify |
|-----------|---------|----------------|
| `.github/workflows/` | CI, release, and container smoke automation | Workflow or release changes |
| `cmd/` | Binary entry points (`unconf`, `unconf-server`) | Bootstrap or process lifecycle changes |
| `internal/cli/` | `unconf` product CLI commands and prompts | Attendee/organizer UX command changes |
| `internal/servercli/` | `unconf-server` backend/admin CLI commands | Server lifecycle and backend utility changes |
| `internal/tui/` | Bubble Tea UIs | Interactive room, wizard, or dashboard changes |
| `internal/api/` | Router, middleware, handlers | Endpoint and auth-surface changes |
| `internal/service/` | Business rules | Domain behavior changes |
| `internal/repository/sqlite/` | Persistence | Query, transaction, or schema-adapter changes |
| `internal/models/` | Shared data models | Schema or API-contract changes |
| `migrations/` | Database schema history | Schema evolution |
| `docs/` | Product, architecture, story, and QA artifacts | Documentation maintenance |

## File Naming

- **Go files:** `snake_case.go`
- **Test files:** `*_test.go` next to source
- **Migrations:** `{version}_{name}.up.sql` / `.down.sql`
- **Email templates:** `.tmpl` files under `internal/email/templates/`

## Import Order

```go
import (
	// 1. Standard library
	"context"
	"fmt"
	"log/slog"

	// 2. Third-party
	"github.com/gin-gonic/gin"

	// 3. Internal packages
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/service"
)
```
