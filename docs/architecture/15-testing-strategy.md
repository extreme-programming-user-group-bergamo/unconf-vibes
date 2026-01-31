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
| Room booking | TUI → Wizard | Booking created |
| Cancellation | `unconf cancel` | Booking cancelled |

---
