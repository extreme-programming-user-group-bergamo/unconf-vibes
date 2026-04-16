# 15. Testing Strategy

## 15.1 Testing Pyramid

- **60% Unit Tests** — Service layer with mocked dependencies
- **30% Integration Tests** — Repository with in-memory SQLite
- **10% Command/TUI Smoke Tests** — Manual or focused integration validation for full user flows

## 15.2 Current Coverage Notes (2026-04-16)

- Automated tests exist across middleware, handlers, services, repositories, CLI commands, HTTP clients, and TUI packages.
- `cmd/server/main_integration_test.go` exercises server bootstrap and integration behavior.
- `internal/servercli/root_test.go` covers `unconf-server` command surface and `db status` behavior.
- Room explorer, booking wizard state, organizer dashboard behavior, cancellation, attendee flows, and CSV export all have automated coverage.
- End-to-end CLI lifecycle smoke coverage now exists for `login -> checkout -> rooms/book -> status -> cancel` against the real in-memory router composition (`internal/cli/booking_lifecycle_smoke_test.go`).

## 15.3 Representative Test Examples

```go
// Unit test
func TestBookingService_CreateBooking_RoomFull(t *testing.T) {
    mockRepo := new(MockBookingRepo)
    // ... assertions
}

// Integration test
func TestBookingRepository_Create(t *testing.T) {
    db := setupTestDB(t) // in-memory SQLite
    // ... assertions
}
```

## 15.4 Manual Smoke Tests

| Scenario | Steps | Expected |
|----------|-------|----------|
| Login flow | `unconf login` | Token stored |
| Claim validation (issuer/audience) | Send token with wrong `iss` or wrong `aud` | Request denied (401) |
| Claim validation (time-based) | Send expired token or future `nbf` | Request denied (401) |
| Token refresh | `POST /auth/refresh` with valid refresh token | New access and refresh tokens issued |
| Token revocation | `POST /auth/revoke` then call protected endpoint | Request denied (401) |
| Authorization matrix | Attendee calls organizer-only endpoints | Request denied (403) |
| Room explorer | `unconf rooms` | TUI loads available rooms for active context |
| Room booking submission | `unconf rooms` fallback/list + `unconf book <room>` | Booking is created and visible via `unconf status` |
| Cancellation | `unconf cancel` | Booking cancelled |
| Organizer dashboard | `unconf dashboard` | Organizer metrics and attendee detail visible |
| CSV export | `unconf export --output bookings.csv` | CSV written with conference bookings |
| Server lifecycle | `unconf-server serve` then `curl /health` | API process starts and health returns `{"status":"ok"}` |
| Backend DB diagnostics | `unconf-server db status` | Prints `db_path`, `migration_version`, and `dirty` |

---
