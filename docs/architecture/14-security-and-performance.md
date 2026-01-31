# 14. Security and Performance

## 14.1 Security Requirements

- **Token Storage:** OS keychain (macOS Keychain, Linux Secret Service, Windows Credential Manager)
- **HTTPS:** All production traffic over TLS
- **JWT:** 24-hour expiry
- **Input Validation:** All inputs validated via `go-playground/validator`
- **SQL Injection:** Parameterized queries only

## 14.2 Performance Targets

- **API Response Time:** < 200ms (NFR1)
- **TUI Render:** 60 FPS (NFR2)
- **Database:** SQLite with WAL mode

## 14.3 Scalability

**Current Limits (SQLite):**
- Concurrent users: ~100
- Database size: ~10GB

**PostgreSQL Migration Triggers:**
- Multiple concurrent conferences
- More than 500 active users

---
