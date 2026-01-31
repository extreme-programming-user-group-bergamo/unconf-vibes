# 17. Error Handling Strategy

## 17.1 Error Response Format

```json
{
    "error": {
        "code": "room_full",
        "message": "Room 204 has no available spots",
        "timestamp": "2026-01-31T10:30:00Z"
    }
}
```

## 17.2 Domain Errors

```go
var (
    ErrRoomFull       = errors.New("room is at capacity")
    ErrAlreadyBooked  = errors.New("user already has a booking")
    ErrBookingNotFound = errors.New("booking not found")
    ErrUnauthorized   = errors.New("unauthorized")
    ErrForbidden      = errors.New("forbidden")
)
```

## 17.3 CLI Error Display

```go
func HandleError(err error) {
    if apiErr, ok := err.(*client.APIError); ok {
        fmt.Fprintf(os.Stderr, "Error: %s\n", apiErr.Message)
        // Show helpful tip based on error code
    }
    os.Exit(1)
}
```

---
