# 10. CLI & Backend Architecture

## 10.1 CLI Command Organization

```
internal/cli/
├── root.go           # Root command, global flags
├── login.go          # unconf login
├── logout.go         # unconf logout
├── list.go           # unconf list
├── info.go           # unconf info <slug>
├── checkout.go       # unconf checkout <slug>
├── config.go         # unconf config
├── rooms.go          # unconf rooms (TUI)
├── book.go           # unconf book <room>
├── status.go         # unconf status
├── cancel.go         # unconf cancel
├── invite.go         # unconf invite <username>
├── requests.go       # unconf requests
├── attendees.go      # unconf attendees
├── dashboard.go      # unconf dashboard (organizer)
├── create.go         # unconf create (organizer)
├── confirm.go        # unconf confirm (organizer)
└── export.go         # unconf export (organizer)
```

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

```
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
│   ├── conferences.go
│   ├── rooms.go
│   ├── bookings.go
│   ├── requests.go
│   ├── users.go
│   └── organizer.go
└── responses/
    ├── success.go
    └── errors.go
```

---
