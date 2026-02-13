# UNCONF CLI Source Tree

> Quick reference for project structure. See [architecture.md](../architecture.md) for full documentation.

## Repository Structure

```
unconf/
├── .github/
│   └── workflows/
│       ├── ci.yaml              # Test, lint on PR/push
│       ├── release.yaml         # GoReleaser on version tag
│       └── deploy.yaml          # Fly.io deployment on tag
│
├── cmd/                         # 📦 Binary entry points
│   ├── unconf/                  # CLI binary
│   │   └── main.go              # CLI entry: initialize Cobra, run root command
│   └── server/                  # API server binary
│       └── main.go              # Server entry: setup Gin, start HTTP
│
├── internal/                    # 🔒 Private application code
│   │
│   ├── api/                     # HTTP API layer (Gin)
│   │   ├── routes.go            # Route definitions, middleware chain
│   │   ├── middleware/
│   │   │   ├── auth.go          # JWT validation, user extraction
│   │   │   ├── cors.go          # CORS headers
│   │   │   ├── logger.go        # Request logging (slog)
│   │   │   ├── organizer.go     # Organizer permission check
│   │   │   └── recovery.go      # Panic recovery
│   │   ├── handlers/
│   │   │   ├── health.go        # GET /health
│   │   │   ├── auth.go          # POST /auth/device, /auth/token
│   │   │   ├── conferences.go   # Conference CRUD
│   │   │   ├── rooms.go         # Room listing, creation
│   │   │   ├── bookings.go      # Booking lifecycle
│   │   │   ├── requests.go      # Roommate requests
│   │   │   ├── users.go         # User profile
│   │   │   └── organizer.go     # Organizer-only endpoints
│   │   └── responses/
│   │       ├── success.go       # Standard success wrapper
│   │       └── errors.go        # Error response, mapping
│   │
│   ├── cli/                     # CLI commands (Cobra)
│   │   ├── root.go              # Root command, global flags (--json, --verbose)
│   │   ├── login.go             # unconf login
│   │   ├── logout.go            # unconf logout
│   │   ├── list.go              # unconf list
│   │   ├── info.go              # unconf info <slug>
│   │   ├── checkout.go          # unconf checkout <slug>
│   │   ├── config.go            # unconf config
│   │   ├── rooms.go             # unconf rooms → launches TUI
│   │   ├── book.go              # unconf book <room>
│   │   ├── status.go            # unconf status
│   │   ├── cancel.go            # unconf cancel
│   │   ├── invite.go            # unconf invite <username>
│   │   ├── requests.go          # unconf requests [accept|decline]
│   │   ├── attendees.go         # unconf attendees
│   │   ├── dashboard.go         # unconf dashboard (organizer TUI)
│   │   ├── create.go            # unconf create (organizer)
│   │   ├── confirm.go           # unconf confirm <id> (organizer)
│   │   └── export.go            # unconf export (organizer)
│   │
│   ├── tui/                     # TUI components (Bubble Tea)
│   │   ├── rooms/
│   │   │   ├── model.go         # RoomExplorerModel (state)
│   │   │   ├── view.go          # View() rendering
│   │   │   └── messages.go      # Message types (RoomsLoadedMsg, etc.)
│   │   ├── wizard/
│   │   │   ├── model.go         # BookingWizardModel
│   │   │   ├── steps.go         # Step definitions (confirm, privacy, notes)
│   │   │   └── view.go          # Step rendering
│   │   ├── dashboard/
│   │   │   ├── model.go         # OrganizerDashboardModel
│   │   │   └── view.go          # Dashboard rendering
│   │   └── common/
│   │       ├── styles.go        # Lip Gloss style definitions
│   │       ├── header.go        # Shared header component
│   │       └── footer.go        # Shared footer with help
│   │
│   ├── service/                 # Business logic layer
│   │   ├── auth.go              # GitHub OAuth, JWT generation
│   │   ├── conference.go        # Conference operations
│   │   ├── room.go              # Room operations, availability
│   │   ├── booking.go           # Booking lifecycle, validation
│   │   ├── request.go           # Roommate request logic
│   │   ├── user.go              # User profile management
│   │   ├── email.go             # Email orchestration
│   │   └── errors.go            # Domain errors (ErrRoomFull, etc.)
│   │
│   ├── repository/              # Data access layer
│   │   ├── interfaces.go        # Repository interfaces
│   │   ├── sqlite/              # SQLite implementations
│   │   │   ├── user.go
│   │   │   ├── conference.go
│   │   │   ├── room.go
│   │   │   ├── booking.go
│   │   │   ├── request.go
│   │   │   └── organizer.go
│   │   └── postgres/            # Future: PostgreSQL implementations
│   │       └── .gitkeep
│   │
│   ├── models/                  # Domain models (shared)
│   │   ├── user.go              # User struct
│   │   ├── conference.go        # Conference struct
│   │   ├── room.go              # Room struct
│   │   ├── booking.go           # Booking struct, BookingStatus
│   │   └── request.go           # RoommateRequest struct
│   │
│   ├── auth/                    # Authentication
│   │   ├── github.go            # GitHub OAuth client
│   │   ├── jwt.go               # JWT generation/validation
│   │   └── store.go             # Token storage (keyring)
│   │
│   ├── email/                   # Email service
│   │   ├── service.go           # Email orchestration
│   │   ├── smtp.go              # SMTP sender (dev/production)
│   │   ├── sendgrid.go          # SendGrid API sender
│   │   └── templates/
│   │       ├── booking.html     # New booking notification
│   │       ├── cancellation.html# Cancellation notice
│   │       └── confirmation.html# Booking confirmation
│   │
│   ├── config/                  # Configuration
│   │   ├── config.go            # Viper wrapper
│   │   └── defaults.go          # Default values
│   │
│   └── client/                  # HTTP client for CLI
│       ├── client.go            # Base client, auth injection
│       ├── conferences.go       # Conference API calls
│       ├── rooms.go             # Room API calls
│       ├── bookings.go          # Booking API calls
│       └── requests.go          # Roommate request API calls
│
├── migrations/                  # Database migrations (golang-migrate)
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   ├── 000002_create_conferences.up.sql
│   ├── 000002_create_conferences.down.sql
│   ├── 000003_create_rooms.up.sql
│   ├── 000003_create_rooms.down.sql
│   ├── 000004_create_bookings.up.sql
│   ├── 000004_create_bookings.down.sql
│   ├── 000005_create_requests.up.sql
│   ├── 000005_create_requests.down.sql
│   ├── 000006_create_email_logs.up.sql
│   └── 000006_create_email_logs.down.sql
│
├── configs/                     # Configuration files
│   ├── config.example.yaml      # Example config for developers
│   └── config.dev.yaml          # Development defaults
│
├── scripts/                     # Build/deploy scripts
│   ├── build.sh                 # Build both binaries
│   ├── migrate.sh               # Run migrations
│   └── seed.sh                  # Seed dev data
│
├── docs/                        # Documentation
│   ├── prd.md                   # Product Requirements
│   ├── brief.md                 # Project Brief
│   ├── architecture.md          # This architecture document
│   ├── api.yaml                 # OpenAPI specification
│   └── architecture/            # Sharded architecture docs
│       ├── source-tree.md       # This file
│       ├── tech-stack.md        # Technology choices
│       └── coding-standards.md  # Development guidelines
│
├── testdata/                    # Test fixtures
│   └── conferences.json         # Sample conference data
│
├── Dockerfile                   # Server container build
├── docker-compose.yml           # Local dev (server + MailHog)
├── fly.toml                     # Fly.io deployment config
├── Makefile                     # Task runner
├── .goreleaser.yaml             # CLI release config
├── .golangci.yaml               # Linter config
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Key Directories

| Directory | Purpose | When to Modify |
|-----------|---------|----------------|
| `cmd/` | Binary entry points | Rarely (initialization only) |
| `internal/cli/` | CLI commands | Adding new commands |
| `internal/tui/` | TUI components | UI changes |
| `internal/api/handlers/` | HTTP handlers | Adding/changing endpoints |
| `internal/service/` | Business logic | Business rule changes |
| `internal/repository/` | Data access | Query changes, new DB support |
| `internal/models/` | Domain models | Schema changes |
| `migrations/` | Database schema | Schema changes |

## File Naming

- **Go files:** `snake_case.go` (e.g., `booking_service.go`)
- **Test files:** `*_test.go` next to source (e.g., `booking_test.go`)
- **Migrations:** `NNNNNN_description.up.sql` / `.down.sql`
- **Templates:** `purpose.html` (e.g., `booking.html`)

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
    "unconf/internal/models"
    "unconf/internal/service"
)
```
