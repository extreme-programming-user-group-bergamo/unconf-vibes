# Risk Profile: UNCONF CLI Project

**Date:** January 31, 2026  
**Reviewer:** Quinn (Test Architect)  
**Scope:** Project-level risk assessment based on PRD and Architecture documents

---

## Executive Summary

| Metric | Value |
|--------|-------|
| **Total Risks Identified** | 18 |
| **Critical Risks** | 2 |
| **High Risks** | 5 |
| **Medium Risks** | 6 |
| **Low Risks** | 5 |
| **Overall Risk Score** | 68/100 (Moderate-High) |

The UNCONF CLI project has a **moderate-high risk profile** primarily driven by:
1. **Security concerns** around GitHub OAuth device flow and JWT token management
2. **Data integrity risks** in the roommate request/booking workflow
3. **Operational risks** related to automated hotel email communication

Immediate attention required on authentication security and booking workflow validation.

---

## Risk Matrix

| Risk ID | Description | Probability | Impact | Score | Priority |
|---------|-------------|-------------|--------|-------|----------|
| SEC-001 | JWT token theft from CLI storage | Medium (2) | High (3) | **6** | High |
| SEC-002 | GitHub OAuth device flow hijacking | Low (1) | High (3) | **3** | Low |
| SEC-003 | Insufficient input validation on API endpoints | Medium (2) | High (3) | **6** | High |
| SEC-004 | Authorization bypass (organizer-only endpoints) | Medium (2) | High (3) | **6** | High |
| PERF-001 | SQLite concurrent write contention | High (3) | Medium (2) | **6** | High |
| PERF-002 | TUI 60fps requirement not met on slow terminals | Medium (2) | Low (1) | **2** | Low |
| DATA-001 | Booking state corruption on concurrent requests | Medium (2) | High (3) | **6** | High |
| DATA-002 | Roommate request race conditions | Medium (2) | Medium (2) | **4** | Medium |
| DATA-003 | Privacy setting not respected in API responses | Medium (2) | Medium (2) | **4** | Medium |
| BUS-001 | Hotel email contains incorrect booking data | Low (1) | High (3) | **3** | Low |
| BUS-002 | User abandons registration (>3 min goal) | Medium (2) | Medium (2) | **4** | Medium |
| BUS-003 | Roommate workflow confuses users | Medium (2) | Medium (2) | **4** | Medium |
| OPS-001 | Fly.io free tier resource exhaustion | Medium (2) | Medium (2) | **4** | Medium |
| OPS-002 | SQLite database file corruption | Low (1) | High (3) | **3** | Low |
| OPS-003 | No monitoring/alerting in place | High (3) | High (3) | **9** | Critical |
| OPS-004 | Email delivery failures undetected | High (3) | High (3) | **9** | Critical |
| TECH-001 | SQLite → PostgreSQL migration complexity | Low (1) | Medium (2) | **2** | Low |
| TECH-002 | Cross-platform CLI binary issues | Medium (2) | Medium (2) | **4** | Medium |

---

## Detailed Risk Analysis

### Critical Risks (Score 9)

#### OPS-003: No Monitoring/Alerting in Place

**Category:** Operational  
**Description:** Architecture document mentions monitoring but no specific tools or implementation defined. Without observability, issues in production will be invisible until users report them.

**Affected Components:**
- API Server
- Database connections
- Email service
- Authentication flow

**Detection Method:** Architecture review revealed logging (zerolog) but no mention of metrics, dashboards, or alerting.

**Mitigation Strategy:**
| Strategy | Actions |
|----------|---------|
| **Preventive** | Add Prometheus metrics endpoint, Configure Fly.io metrics |
| **Detective** | Set up error alerting (email/Slack on 5xx spike) |
| **Corrective** | Create runbook for common failure scenarios |

**Testing Requirements:**
- Load test to establish baseline metrics
- Synthetic monitoring for health endpoint
- Alert testing (fire test alerts)

**Residual Risk:** Medium — Some edge cases may still go undetected  
**Owner:** DevOps/Architect  
**Timeline:** Before production deployment

---

#### OPS-004: Email Delivery Failures Undetected

**Category:** Operational  
**Description:** Hotel communication is automated via email (SendGrid/Mailgun), but there's no mechanism to detect failed deliveries. A failed email means a missed booking with the hotel.

**Affected Components:**
- `internal/email/` service
- SendGrid/Mailgun integration
- Booking confirmation workflow

**Detection Method:** PRD states "zero transcription errors" requirement (NFR11) but architecture lacks delivery confirmation mechanism.

**Mitigation Strategy:**
| Strategy | Actions |
|----------|---------|
| **Preventive** | Use SendGrid webhook for delivery events, Store email status in DB |
| **Detective** | Dashboard showing email delivery status per booking |
| **Corrective** | Manual retry mechanism for failed emails, Organizer notification |

**Testing Requirements:**
- Integration test with SendGrid sandbox
- Webhook handling tests
- Failure scenario simulation

**Residual Risk:** Low — With webhooks, failures will be detected  
**Owner:** Backend Developer  
**Timeline:** Epic 5 (Hotel Automation)

---

### High Risks (Score 6)

#### SEC-001: JWT Token Theft from CLI Storage

**Category:** Security  
**Description:** CLI stores JWT tokens locally. If stored insecurely, tokens could be stolen by malicious software or other users on shared machines.

**Affected Components:**
- `internal/auth/` (CLI token storage)
- OS keychain integration

**Mitigation Strategy:**
| Strategy | Actions |
|----------|---------|
| **Preventive** | Use OS keychain (macOS Keychain, Windows Credential Manager, Linux secret-service) |
| **Detective** | Log token usage with device fingerprint |
| **Corrective** | Token revocation endpoint, Short token expiry |

**Testing Requirements:**
- Verify keychain integration on all platforms
- Test fallback to config file with proper permissions (0600)
- Token refresh flow tests

**Residual Risk:** Low — Keychain provides OS-level protection  
**Owner:** CLI Developer  
**Timeline:** Epic 1 (Story 1.7)

---

#### SEC-003: Insufficient Input Validation on API Endpoints

**Category:** Security  
**Description:** API accepts user input (notes, display names, search queries). Without proper validation, XSS, SQL injection, or command injection attacks are possible.

**Affected Components:**
- All API handlers accepting user input
- Booking notes field
- User profile fields
- Conference search

**Mitigation Strategy:**
| Strategy | Actions |
|----------|---------|
| **Preventive** | Use go-playground/validator on all input structs, Parameterized queries only |
| **Detective** | Input validation error logging |
| **Corrective** | Sanitization before storage |

**Testing Requirements:**
- Fuzzing tests for all input endpoints
- SQL injection test suite
- XSS payload testing on text fields

**Residual Risk:** Low — With proper validation  
**Owner:** Backend Developer  
**Timeline:** All Epics

---

#### SEC-004: Authorization Bypass (Organizer-Only Endpoints)

**Category:** Security  
**Description:** Several endpoints are organizer-only (create conference, export CSV, confirm bookings). Authorization logic must be robust.

**Affected Components:**
- `POST /conferences`
- `GET /conferences/{slug}/export`
- `PUT /bookings/{id}/confirm`
- Organizer middleware

**Mitigation Strategy:**
| Strategy | Actions |
|----------|---------|
| **Preventive** | Centralized organizer middleware, Role check at service layer too |
| **Detective** | Log authorization failures with user context |
| **Corrective** | N/A (prevention is key) |

**Testing Requirements:**
- Negative authorization tests (non-organizer attempts)
- Role boundary testing
- Organizer revocation scenarios

**Residual Risk:** Low — With proper middleware  
**Owner:** Backend Developer  
**Timeline:** Epic 5

---

#### PERF-001: SQLite Concurrent Write Contention

**Category:** Performance  
**Description:** SQLite has limited concurrent write support. During high-registration periods (conference announcements), multiple booking attempts could cause lock contention.

**Affected Components:**
- `internal/repository/sqlite/`
- Booking creation
- Roommate requests

**Mitigation Strategy:**
| Strategy | Actions |
|----------|---------|
| **Preventive** | Enable WAL mode, Connection pooling with busy timeout, Optimistic locking on bookings |
| **Detective** | Monitor database lock wait times |
| **Corrective** | Graceful retry with backoff |

**Testing Requirements:**
- Concurrent booking load test (50+ simultaneous)
- Lock timeout behavior verification
- WAL mode confirmation

**Residual Risk:** Medium — May need PostgreSQL migration sooner than planned  
**Owner:** Backend Developer  
**Timeline:** Epic 1 (Story 1.4)

---

#### DATA-001: Booking State Corruption on Concurrent Requests

**Category:** Data  
**Description:** Room capacity validation and booking creation are not atomic. Two users could book the last spot in a room simultaneously.

**Affected Components:**
- Booking service
- Room availability check
- `POST /bookings` endpoint

**Mitigation Strategy:**
| Strategy | Actions |
|----------|---------|
| **Preventive** | Database transaction with SELECT FOR UPDATE (or SQLite equivalent), Check capacity inside transaction |
| **Detective** | Post-booking capacity validation alert |
| **Corrective** | Manual overbooking resolution by organizer |

**Testing Requirements:**
- Concurrent booking race condition tests
- Capacity enforcement verification
- Transaction isolation testing

**Residual Risk:** Low — With proper transactions  
**Owner:** Backend Developer  
**Timeline:** Epic 3

---

### Medium Risks (Score 4)

#### DATA-002: Roommate Request Race Conditions

**Description:** User A requests User B while User B requests User A, or user accepts two different roommate requests.

**Mitigation:** Transactional state machine for requests, conflict detection logic.

#### DATA-003: Privacy Setting Not Respected

**Description:** API might leak private user info in attendee lists or room occupancy views.

**Mitigation:** Privacy check at service layer, integration tests for privacy scenarios.

#### BUS-002: User Abandons Registration (>3 min goal)

**Description:** Complex flow or confusing TUI could cause users to abandon.

**Mitigation:** Wizard mode for first-timers, UX testing, analytics on drop-off points.

#### BUS-003: Roommate Workflow Confuses Users

**Description:** Request/accept/decline flow may not be intuitive to all users.

**Mitigation:** Clear TUI messaging, help text, status indicators.

#### OPS-001: Fly.io Free Tier Resource Exhaustion

**Description:** 256MB RAM, shared CPU may not handle peak loads.

**Mitigation:** Monitor resource usage, define upgrade criteria, graceful degradation.

#### TECH-002: Cross-Platform CLI Binary Issues

**Description:** GoReleaser builds for multiple platforms; edge cases in terminal handling.

**Mitigation:** CI matrix testing, beta testing on each platform.

---

### Low Risks (Score 2-3)

| Risk ID | Description | Mitigation |
|---------|-------------|------------|
| SEC-002 | OAuth device flow hijacking | User code expiry, HTTPS only |
| PERF-002 | TUI 60fps on slow terminals | Graceful degradation, test on constrained terminals |
| BUS-001 | Hotel email errors | Template validation, organizer preview |
| OPS-002 | SQLite file corruption | Regular backups, WAL mode |
| TECH-001 | PostgreSQL migration complexity | Repository pattern in place |

---

## Recommendations

### Must-Fix Before MVP

1. **Implement observability** (OPS-003) — Add Prometheus metrics, error alerting
2. **Email delivery tracking** (OPS-004) — SendGrid webhooks, delivery status in DB
3. **Secure token storage** (SEC-001) — OS keychain integration
4. **Input validation** (SEC-003) — go-playground/validator on all inputs
5. **Transactional booking** (DATA-001) — Atomic capacity check + booking

### Monitor in Production

1. SQLite write contention metrics
2. API response times vs 200ms SLA
3. Registration completion rate and timing
4. Email delivery success rate
5. Failed authorization attempts

### Future Improvements

1. Consider PostgreSQL earlier if concurrent users exceed 100
2. Add automated E2E tests when TUI stabilizes
3. Implement rate limiting on public endpoints
4. Add audit logging for compliance

---

## Risk Heatmap

```
                    IMPACT
              Low    Medium    High
         ┌─────────┬─────────┬─────────┐
    High │PERF-002 │OPS-001  │OPS-003* │
         │         │TECH-002 │OPS-004* │
P        ├─────────┼─────────┼─────────┤
R   Med  │         │DATA-002 │SEC-001  │
O        │         │DATA-003 │SEC-003  │
B        │         │BUS-002  │SEC-004  │
         │         │BUS-003  │PERF-001 │
         │         │         │DATA-001 │
         ├─────────┼─────────┼─────────┤
    Low  │TECH-001 │         │SEC-002  │
         │         │         │BUS-001  │
         │         │         │OPS-002  │
         └─────────┴─────────┴─────────┘

* = Critical (requires immediate attention)
```

---

## Appendix: Risk Score Calculation

**Risk Score = Probability × Impact**

| Level | Probability | Impact |
|-------|-------------|--------|
| High | 3 (>70% chance) | 3 (Severe: data breach, system down) |
| Medium | 2 (30-70% chance) | 2 (Moderate: degraded performance) |
| Low | 1 (<30% chance) | 1 (Minor: cosmetic issues) |

**Score Interpretation:**
- 9: Critical (Red) — Immediate action required
- 6: High (Orange) — Address before release
- 4: Medium (Yellow) — Monitor and plan mitigation
- 2-3: Low (Green) — Accept or defer
- 1: Minimal (Blue) — Document only
