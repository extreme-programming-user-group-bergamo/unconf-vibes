# UNCONF CLI Coding Standards

> Development guidelines for AI agents and human developers. See [architecture.md](../architecture.md) for full context.

## Critical Rules

These rules prevent common bugs and ensure consistency. **Violations should block PR merge.**

### 1. Repository Pattern

All database access MUST go through repository interfaces.

```go
// ✅ CORRECT
func (s *BookingService) CreateBooking(ctx context.Context, input Input) (*Booking, error) {
    return s.bookingRepo.Create(ctx, booking)
}

// ❌ WRONG - Direct DB access in service
func (s *BookingService) CreateBooking(ctx context.Context, input Input) (*Booking, error) {
    return s.db.Exec("INSERT INTO bookings...")
}
```

### 2. Error Wrapping

Always wrap errors with context. Never return raw errors.

```go
// ✅ CORRECT
if err != nil {
    return nil, fmt.Errorf("failed to create booking: %w", err)
}

// ❌ WRONG - No context
if err != nil {
    return nil, err
}
```

### 3. Context Propagation

Pass `context.Context` as first parameter to all functions that do I/O.

```go
// ✅ CORRECT
func (r *bookingRepo) Create(ctx context.Context, b *Booking) error {
    _, err := r.db.ExecContext(ctx, query, args...)
    return err
}

// ❌ WRONG - Missing context
func (r *bookingRepo) Create(b *Booking) error {
    _, err := r.db.Exec(query, args...)
    return err
}
```

### 4. Logging

Use `log/slog` for all output. Never use `fmt.Println()` in production code.

```go
// ✅ CORRECT
slog.Info("booking created",
    "user_id", userID,
    "room", roomNumber,
)

// ❌ WRONG
fmt.Printf("Booking created for user %s\n", userID)
```

### 5. Configuration

Never read environment variables directly. Use the Config wrapper.

```go
// ✅ CORRECT
apiEndpoint := cfg.GetAPIEndpoint()

// ❌ WRONG
apiEndpoint := os.Getenv("API_ENDPOINT")
```

### 6. No Globals

Services and repositories are passed via dependency injection.

```go
// ✅ CORRECT
type BookingHandler struct {
    svc *service.BookingService
}

func NewBookingHandler(svc *service.BookingService) *BookingHandler {
    return &BookingHandler{svc: svc}
}

// ❌ WRONG - Global service
var bookingService *service.BookingService
```

### 7. Defer for Cleanup

Always clean up resources with defer.

```go
// ✅ CORRECT
rows, err := db.QueryContext(ctx, query)
if err != nil {
    return err
}
defer rows.Close()

// ❌ WRONG - Rows never closed
rows, err := db.QueryContext(ctx, query)
// process rows...
```

---

## Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| **Packages** | lowercase, single word | `service`, `repository` |
| **Interfaces** | PascalCase, often verb+er | `BookingRepository`, `EmailSender` |
| **Structs** | PascalCase | `BookingService`, `RoomExplorerModel` |
| **Public methods** | PascalCase | `CreateBooking`, `GetByID` |
| **Private methods** | camelCase | `handleKeyPress`, `parseResponse` |
| **Variables** | camelCase | `bookingID`, `roomCount` |
| **Constants** | PascalCase or UPPER_CASE | `BookingStatusRequested`, `MAX_RETRIES` |
| **Files** | snake_case | `booking_service.go`, `booking_test.go` |
| **Test files** | `*_test.go` | `booking_test.go` |
| **API routes** | kebab-case | `/roommate-requests`, `/auth/device` |
| **DB tables** | snake_case, plural | `bookings`, `roommate_requests` |
| **DB columns** | snake_case | `created_at`, `github_id` |

---

## File Organization

### Package Structure

```go
// internal/service/booking.go

package service

import (
    // 1. Standard library
    "context"
    "fmt"
    
    // 2. Third-party
    "log/slog"
    
    // 3. Internal packages
    "unconf/internal/models"
    "unconf/internal/repository"
)

// Interface (if needed in same package)
type BookingCreator interface {
    CreateBooking(ctx context.Context, input CreateBookingInput) (*models.Booking, error)
}

// Implementation struct
type bookingService struct {
    bookingRepo repository.BookingRepository
    roomRepo    repository.RoomRepository
    emailSvc    EmailSender
}

// Constructor
func NewBookingService(
    bookingRepo repository.BookingRepository,
    roomRepo repository.RoomRepository,
    emailSvc EmailSender,
) *bookingService {
    return &bookingService{
        bookingRepo: bookingRepo,
        roomRepo:    roomRepo,
        emailSvc:    emailSvc,
    }
}

// Methods
func (s *bookingService) CreateBooking(ctx context.Context, input CreateBookingInput) (*models.Booking, error) {
    // Implementation
}
```

---

## Error Handling

### Domain Errors

Define domain-specific errors in the service package:

```go
// internal/service/errors.go
package service

import "errors"

var (
    ErrRoomFull       = errors.New("room is at capacity")
    ErrAlreadyBooked  = errors.New("user already has a booking for this conference")
    ErrBookingNotFound = errors.New("booking not found")
    ErrUnauthorized   = errors.New("unauthorized")
    ErrForbidden      = errors.New("forbidden: insufficient permissions")
)
```

### Error Checking

Use `errors.Is()` for comparison:

```go
if errors.Is(err, service.ErrRoomFull) {
    // Handle room full case
}
```

### Handler Error Mapping

Map domain errors to HTTP status codes in handlers:

```go
switch {
case errors.Is(err, service.ErrRoomFull):
    c.JSON(http.StatusBadRequest, errorResponse("room_full", "Room has no available spots"))
case errors.Is(err, service.ErrNotFound):
    c.JSON(http.StatusNotFound, errorResponse("not_found", "Resource not found"))
default:
    log.Error().Err(err).Msg("unexpected error")
    c.JSON(http.StatusInternalServerError, errorResponse("internal", "An unexpected error occurred"))
}
```

---

## TUI Patterns (Bubble Tea)

### Message Naming

```go
// Messages use Verb + "Msg" suffix
type RoomsLoadedMsg []Room
type ErrorOccurredMsg error
type BookingCreatedMsg Booking
```

### Model Structure

```go
type RoomExplorerModel struct {
    // State fields are lowercase (private)
    rooms   []Room
    cursor  int
    loading bool
    err     error
    
    // Dependencies
    client  *api.Client
    styles  Styles
}
```

### Commands

```go
// Commands are functions returning tea.Cmd
func loadRooms(client *api.Client, slug string) tea.Cmd {
    return func() tea.Msg {
        rooms, err := client.GetRooms(context.Background(), slug)
        if err != nil {
            return ErrorOccurredMsg(err)
        }
        return RoomsLoadedMsg(rooms)
    }
}
```

### View Rules

- `View()` is a pure function — no side effects
- Same input always produces same output
- Use styles from the Styles struct, never hardcode

---

## Git Commit Format

```
type(scope): description

Types:
- feat:     New feature
- fix:      Bug fix
- docs:     Documentation only
- style:    Formatting (no code change)
- refactor: Code restructuring
- test:     Adding/fixing tests
- chore:    Maintenance tasks

Examples:
feat(booking): add roommate request flow
fix(tui): prevent crash on empty room list
docs(readme): add installation instructions
test(service): add booking cancellation tests
refactor(repository): extract common query logic
chore(deps): update Bubble Tea to v0.25.0
```

---

## Testing Guidelines

### Unit Tests

- Mock external dependencies
- Test one thing per test
- Use table-driven tests for variations

```go
func TestBookingService_CreateBooking(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateBookingInput
        setup   func(*MockBookingRepo)
        wantErr error
    }{
        {
            name:  "success",
            input: CreateBookingInput{RoomNumber: "204"},
            setup: func(m *MockBookingRepo) {
                m.On("Create", mock.Anything, mock.Anything).Return(nil)
            },
            wantErr: nil,
        },
        {
            name:  "room full",
            input: CreateBookingInput{RoomNumber: "204"},
            setup: func(m *MockBookingRepo) {
                // Setup for full room
            },
            wantErr: ErrRoomFull,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Tests

- Use in-memory SQLite
- Clean up after each test
- Test actual SQL queries

```go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)
    runMigrations(db)
    t.Cleanup(func() { db.Close() })
    return db
}
```

---

## Code Review Checklist

Before approving a PR, verify:

- [ ] Repository pattern used for all DB access
- [ ] Errors wrapped with context
- [ ] Context passed to I/O functions
- [ ] slog used for logging
- [ ] No direct `os.Getenv()` calls
- [ ] No global mutable state
- [ ] Resources cleaned up with defer
- [ ] Tests cover happy path and error cases
- [ ] Commit messages follow format
