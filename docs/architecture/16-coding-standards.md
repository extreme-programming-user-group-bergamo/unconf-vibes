# 16. Coding Standards

## 16.1 Critical Rules

1. **Repository Pattern:** All DB access through repository interfaces
2. **Error Wrapping:** `fmt.Errorf("context: %w", err)`
3. **Context Propagation:** Pass `context.Context` to all I/O functions
4. **Logging:** Use `zerolog`, never `fmt.Println`
5. **Config:** Use Viper wrapper, never `os.Getenv` directly
6. **No Globals:** Dependency injection only
7. **Defer Cleanup:** Always `defer rows.Close()`

## 16.2 Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Packages | lowercase | `service`, `repository` |
| Interfaces | PascalCase | `BookingRepository` |
| Files | snake_case | `booking_service.go` |
| API routes | kebab-case | `/roommate-requests` |
| DB tables | snake_case | `roommate_requests` |

## 16.3 Git Commits

```
type(scope): description

feat(booking): add roommate request flow
fix(tui): prevent crash on empty room list
```

---
