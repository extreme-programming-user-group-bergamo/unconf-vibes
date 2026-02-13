# 15. Testing Strategy

## 15.1 Testing Pyramid

- **60% Unit Tests** — Service layer with mocked dependencies
- **30% Integration Tests** — Repository with in-memory SQLite
- **10% E2E Tests** — Manual (TUI difficult to automate)

## 15.2 Test Examples

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

## 15.3 E2E Testing (Manual)

| Scenario | Steps | Expected |
|----------|-------|----------|
| Login flow | `unconf login` | Token stored |
| Claim validation (issuer/audience) | Send token with wrong `iss` or missing `aud` | Request denied (401) |
| Claim validation (time-based) | Send expired token or future `nbf` | Request denied (401) |
| Token refresh | `POST /auth/refresh` with valid token | New token issued, old token invalidated |
| Token revocation | `POST /auth/revoke` then call protected endpoint | Request denied (401) |
| Session listing | `GET /auth/sessions` after multi-device login | All active sessions returned with `current` flag |
| Revoke other sessions | `POST /auth/revoke-others` then test old device token | Old device request denied (401) |
| Authorization matrix | Attendee calls organizer-only endpoints | Request denied (403) |
| Room booking | TUI → Wizard | Booking created |
| Cancellation | `unconf cancel` | Booking cancelled |

---
