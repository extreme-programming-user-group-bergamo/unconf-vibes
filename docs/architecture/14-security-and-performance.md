# 14. Security and Performance

## 14.1 Security Requirements

- **Token Storage:** OS keychain (macOS Keychain, Linux Secret Service, Windows Credential Manager)
- **HTTPS:** All production traffic over TLS
- **Token Format:** PASETO `v4.local` (symmetric authenticated encryption)
- **Token TTL:** 24-hour expiry
- **Key Rotation:** `kid` footer support; rotate signing/encryption keys every 90 days
- **Replay Protection:** `jti` per token with denylist checks for revoked tokens
- **Input Validation:** All inputs validated via `go-playground/validator`
- **SQL Injection:** Parameterized queries only

## 14.2 Token Claim Validation Rules

- **Issuer (`iss`):** must equal `unconf-api`
- **Audience (`aud`):** must include `unconf-cli`
- **Not-Before (`nbf`):** token not accepted before effective time
- **Expiry (`exp`):** hard fail on expired token
- **Issued-At (`iat`):** reject tokens older than 24h
- **Clock Skew:** tolerate at most ±60 seconds

## 14.3 Session Governance Security

- **Server Session Store:** maintain active sessions keyed by `session_id`
- **Per-Session Revocation:** immediate revoke for a single device/session
- **Revoke Others:** support account-level cleanup except current session
- **Idle Session Policy:** optionally expire inactive sessions after 30 days
- **Audit Trail:** log session create/refresh/revoke events with user and session identifiers

## 14.4 Edge and Abuse Controls

- **Device Polling Limits:** rate-limit `POST /auth/token` by client/device code
- **Auth Endpoint Limits:** stricter limits on `/auth/*` than read-only endpoints
- **CORS Policy:** explicit trusted origins only; deny wildcard in production
- **Brute-Force Protection:** temporary lockout/backoff on repeated invalid auth attempts

## 14.5 Performance Targets

- **API Response Time:** < 200ms (NFR1)
- **TUI Render:** 60 FPS (NFR2)
- **Database:** SQLite with WAL mode

## 14.6 Scalability

**Current Limits (SQLite):**
- Concurrent users: ~100
- Database size: ~10GB

**PostgreSQL Migration Triggers:**
- Multiple concurrent conferences
- More than 500 active users

---
