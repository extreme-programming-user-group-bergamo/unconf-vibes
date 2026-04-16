# 10. CLI & Backend Architecture

## 10.1 CLI Command Organization

```text
internal/cli/
├── root.go                # `unconf` root command and product workflow wiring
├── login.go               # unconf login
├── logout.go              # unconf logout
├── list.go                # unconf list
├── info.go                # unconf info <slug>
├── checkout.go            # unconf checkout <slug>
├── config.go              # unconf config
├── rooms.go               # unconf rooms attendee explorer
├── rooms_manage.go        # unconf rooms add|edit|remove|import
├── book.go                # unconf book <room>
├── status.go              # unconf status
├── cancel.go              # unconf cancel
├── attendees.go           # unconf attendees
├── invite.go              # unconf invite <username>
├── requests.go            # unconf requests
├── create.go              # unconf create
├── edit.go                # unconf edit [slug]
├── dashboard.go           # unconf dashboard
├── export.go              # unconf export
└── conference_forms.go    # Shared interactive create/edit prompts

internal/servercli/
├── root.go                # `unconf-server` root command and lifecycle wiring
├── serve.go               # unconf-server serve
└── db.go                  # unconf-server db status
```

- `unconf` is the user/organizer product CLI surface.
- `unconf-server` is the backend/admin CLI surface and defaults to `serve` when no subcommand is provided.
- `db status` lives under `unconf-server` as backend/admin functionality.
- `unconf rooms` is a composite surface: attendee exploration lives in `rooms.go`, while organizer room-management subcommands live in `rooms_manage.go`.
- `unconf edit` replaces the earlier planned organizer placeholder command set; there is no standalone `confirm.go` in the current codebase.
- The direct-booking CLI and booking wizard use the live `POST /bookings` route.
- Organizer manual booking confirmation is removed from the active surface; no `unconf confirm` command or `PUT /bookings/{id}/confirm` route is expected.

## 10.2 TUI Architecture (Bubble Tea)

```go
// Model-View-Update pattern
type RoomExplorerModel struct {
    rooms       []api.RoomWithAvailability
    cursor      int
    filter      string
    loading     bool
    err         error
    styles      Styles
}

func (m RoomExplorerModel) Init() tea.Cmd { return m.loadRooms }
func (m RoomExplorerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* handle input */ }
func (m RoomExplorerModel) View() string { /* render UI */ }
```

## 10.3 API Server Architecture

```text
internal/api/
├── routes.go         # Route definitions
├── middleware/
│   ├── auth.go       # PASETO authentication
│   ├── cors.go       # CORS configuration
│   ├── logger.go     # Request logging
│   └── organizer.go  # Organizer permission check
├── handlers/
│   ├── health.go
│   ├── auth.go
│   ├── attendees.go
│   ├── bookings.go
│   ├── conferences.go
│   ├── organizers.go
│   ├── requests.go
│   ├── rooms.go
│   └── users.go
└── responses/
    └── errors.go
```

- `routes.go` currently registers booking create/list/cancel routes.
- Organizer-only dashboard, export, room management, and organizer membership routes are all part of the current router surface.

---
