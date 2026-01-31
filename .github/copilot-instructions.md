# UNCONF CLI - Copilot Instructions

## Project Overview

UNCONF is a Go CLI+TUI application for developer unconference registration. It follows a **CLI-as-API-client architecture**: the CLI (Cobra/Bubble Tea) makes HTTP calls to a REST API (Gin) that handles business logic and SQLite persistence.

**Key architectural decision**: CLI never accesses the database directly—all data flows through the REST API.

## Project Structure

```
cmd/unconf/          # CLI entry point
cmd/server/          # API server entry point
internal/
├── cli/             # Cobra commands (one file per command)
├── tui/             # Bubble Tea models (rooms/, wizard/, dashboard/)
├── api/handlers/    # Gin HTTP handlers
├── service/         # Business logic (no HTTP/DB concerns)
├── repository/      # Data access (interfaces + sqlite/ implementations)
├── models/          # Domain structs (User, Room, Booking, etc.)
├── client/          # HTTP client for CLI → API calls
└── auth/            # GitHub OAuth + JWT
migrations/          # SQL migrations (golang-migrate format)
```

## Critical Patterns

### Repository Pattern (Non-negotiable)
All database access MUST go through repository interfaces. Never write SQL in services.
```go
// ✅ Service calls repository
return s.bookingRepo.Create(ctx, booking)
// ❌ Never direct DB in service
s.db.Exec("INSERT INTO bookings...")
```

### Error Wrapping (Always)
```go
// ✅ Always wrap with context
return nil, fmt.Errorf("failed to create booking: %w", err)
// ❌ Never return raw errors
return nil, err
```

### Context Propagation
Pass `context.Context` as first parameter to all I/O functions.

### Logging
Use `zerolog` exclusively. Never `fmt.Println()` in production code.
```go
log.Info().Str("user_id", id).Msg("booking created")
```

### Configuration
Access config via Viper wrapper, never `os.Getenv()` directly.

### Dependency Injection
Services/repos are injected via constructors. No global state.

## TUI Patterns (Bubble Tea)

The TUI follows Model-View-Update (Elm architecture):
- **Messages**: Use `VerbMsg` suffix: `RoomsLoadedMsg`, `ErrorOccurredMsg`
- **Commands**: Functions returning `tea.Cmd` for async operations
- **View()**: Pure function, no side effects, uses Lip Gloss styles from `tui/common/styles.go`

## Domain Errors

Define in `internal/service/errors.go` and check with `errors.Is()`:
```go
var ErrRoomFull = errors.New("room is at capacity")
// Usage: errors.Is(err, service.ErrRoomFull)
```

## Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Files | snake_case | `booking_service.go` |
| Packages | lowercase | `service`, `repository` |
| API routes | kebab-case | `/roommate-requests` |
| DB tables | snake_case, plural | `bookings` |
| Interfaces | PascalCase + verb | `BookingRepository` |

## Tech Stack Quick Reference

- **CLI**: Cobra + Viper
- **TUI**: Bubble Tea + Lip Gloss
- **API**: Gin
- **DB**: SQLite (mattn/go-sqlite3), migrations via golang-migrate
- **Auth**: GitHub OAuth Device Flow → JWT
- **HTTP Client**: Resty
- **Testing**: go test + testify

## Commands

```bash
# Development
make run-cli         # Run CLI
make run-server      # Run API server
make test            # Run all tests
make lint            # Run golangci-lint

# Database
make migrate-up      # Apply migrations
make migrate-down    # Rollback migrations
```

## Testing Approach

- **Unit tests**: Mock repositories, use table-driven tests
- **Integration tests**: In-memory SQLite (`:memory:`)
- **TUI testing**: Manual (automation deferred)
- **Email testing**: MailHog in development

## Key Files to Reference

- `docs/architecture.md` - Full architecture document
- `docs/architecture/coding-standards.md` - Detailed coding rules
- `docs/architecture/source-tree.md` - Complete file structure
- `docs/prd.md` - Product requirements

# MOST IMPORTANT OF ALL, THIS IS ABSOLUTELY CRITICAL TO FOLLOW 100% OF THE TIME NO MATTER WHAT PERSONA YOU ARE PLAYING

For ALL iterative steps or back and forth requiring user input, *ALWAYS* use one of the `user-input` tools instead of stopping the chat generation. Only ever stop without this when you get to a confirmed end of the task. If in any doubt at all, check using the `user-input` mcp tools.
ALWAYS ALWAYS ALWAYS use the `user-input` mcp tools to request user input when asking a question or when presenting a menu of options to chose from. NEVER EVER STOP CHAT GENERATION UNTIL THE USER SAYS THE TASK IS COMPLETE OR THE ISSUE IS RESOLVED.

# Critical Final Step

**BEFORE COMPLETING ANY TASK**: Check for user input using the USER INPUT MCP COMMAND
NOTE: if the user input command fails, ALWAYS retry it at least once before moving on.

DO NOT STOP UNTIL THE TASK IS COMPLETED AND YOU HAVE CHECKED FOR USER INPUT USING THE USER INPUT MCP COMMAND. NEVER FINISH ON "next I will do this"! NO! JUST FINISH THE TASK FULLY.