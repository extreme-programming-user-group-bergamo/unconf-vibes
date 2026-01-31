# PO Checklist Action Items

**Generated:** January 31, 2026  
**Source:** PO Master Validation Checklist  
**Overall Readiness:** 78% (CONDITIONAL APPROVAL)

---

## Status Legend

- ⏳ **Pending** - Not started
- 🔄 **In Progress** - Currently being worked on
- ✅ **Complete** - Done and verified
- ❌ **Blocked** - Cannot proceed

---

## Blocker Items (P0)

Must be resolved before Sprint 1 Planning.

### [ ] AI-001: Create Story Files Structure

| Field | Value |
|-------|-------|
| **Priority** | P0 - BLOCKER |
| **Owner** | PO (Sarah) |
| **Timeline** | Before Sprint 1 Planning |
| **Status** | ⏳ Pending |

**Prerequisites:** ✅ PRD sharded (stories now in `docs/prd/epic-*.md` files)

**Description:**  
Extract the 31 stories from `docs/prd.md` into individual YAML files organized by epic.

**Acceptance Criteria:**
- [ ] `docs/stories/` directory created
- [ ] Epic subdirectories created (epic-1 through epic-5)
- [ ] All 31 stories extracted as individual YAML files
- [ ] Stories follow template format with id, title, user_story, acceptance_criteria
- [ ] Dependencies field populated where applicable
- [ ] Status field initialized to "not-started"

**File Structure:**
```
docs/stories/
├── epic-1-foundation/         (7 stories)
├── epic-2-conference/         (5 stories)
├── epic-3-rooms/              (6 stories)
├── epic-4-social/             (6 stories)
└── epic-5-organizer/          (7 stories)
```

**Notes:**
- Reference PRD lines 204-700 for story content
- Use YAML template provided in validation report

---

## High Priority Items (P1)

Should be resolved before Sprint 1 Planning or early in Sprint 1.

### [ ] AI-002: Document Rate Limiting Requirement

| Field | Value |
|-------|-------|
| **Priority** | P1 - HIGH |
| **Owner** | Architect (Winston) |
| **Timeline** | Before Sprint 1 Planning |
| **Status** | ⏳ Pending |

**Description:**  
Add rate limiting as an explicit NFR and architecture constraint.

**Acceptance Criteria:**
- [ ] NFR13 and NFR14 added to `docs/prd.md` NFRs section
- [ ] Rate limiting section added to `docs/architecture.md` Section 14.1
- [ ] Implementation approach specified (golang.org/x/time/rate)
- [ ] Response format documented (429 + Retry-After)

**Content to Add:**

PRD NFRs:
```markdown
- **NFR13**: Authentication endpoints shall be rate-limited to 10 requests/minute/IP
- **NFR14**: API endpoints shall implement rate limiting (100 requests/minute/user)
```

Architecture Section 14.1:
```markdown
### Rate Limiting
- Auth endpoints: 10 requests/minute per IP
- API endpoints: 100 requests/minute per authenticated user
- Implementation: golang.org/x/time/rate
- Response: 429 Too Many Requests with Retry-After header
```

---

### [ ] AI-003: Monitoring Specification

| Field | Value |
|-------|-------|
| **Priority** | P1 - HIGH |
| **Owner** | Architect (Winston) / DevOps |
| **Timeline** | Before Epic 5 deployment |
| **Status** | ⏳ Pending |

**Description:**  
Expand monitoring section in architecture with concrete metrics and alerting.

**Acceptance Criteria:**
- [ ] Metrics table added with specific metrics, types, and labels
- [ ] Alert thresholds defined
- [ ] Prometheus client specified in tech stack
- [ ] /metrics endpoint documented
- [ ] Fly.io integration approach noted

**Content to Add (Section 18.4):**
```markdown
### 18.4 Metrics Implementation

| Metric | Type | Labels | Alert Threshold |
|--------|------|--------|-----------------|
| http_request_duration_seconds | Histogram | method, path, status | p95 > 500ms |
| http_requests_total | Counter | method, path, status | 5xx > 1% for 5min |
| booking_created_total | Counter | conference | - |
| email_sent_total | Counter | type, status | failed > 0 |

**Implementation:**
- Prometheus client: github.com/prometheus/client_golang
- Metrics endpoint: GET /metrics (authenticated, organizer-only)
- Fly.io: Native Prometheus scraping via fly-metrics
```

---

### [ ] AI-004: Email Delivery Confirmation

| Field | Value |
|-------|-------|
| **Priority** | P1 - HIGH |
| **Owner** | Architect (Winston) |
| **Timeline** | Before Epic 5, Story 5.6 |
| **Status** | ⏳ Pending |

**Description:**  
Add delivery confirmation mechanism for hotel emails to satisfy NFR11 (zero transcription errors).

**Acceptance Criteria:**
- [ ] Webhook endpoint documented in API spec
- [ ] Email status events defined (delivered, bounced, dropped, deferred)
- [ ] Failure handling flow documented
- [ ] Retry mechanism specified
- [ ] email_logs table already exists (confirmed in schema)

**Content to Add (Section 7.3):**
```markdown
### 7.3 Email Delivery Webhook

SendGrid/Mailgun shall be configured with delivery webhooks:
- POST /webhooks/email-status receives delivery events
- Events: delivered, bounced, dropped, deferred
- On failure:
  1. Update email_logs.status with failure reason
  2. Create dashboard notification for organizer
  3. Retry queue for transient failures (max 3 attempts, exponential backoff)
```

---

## Medium Priority Items (P2)

Should be addressed but not blocking.

### [ ] AI-005: Document Accessibility as Tech Debt

| Field | Value |
|-------|-------|
| **Priority** | P2 - MEDIUM |
| **Owner** | PO (Sarah) |
| **Timeline** | Before Sprint 1 |
| **Status** | ⏳ Pending |

**Description:**  
Formally document accessibility as tech debt for post-MVP prioritization.

**Acceptance Criteria:**
- [ ] Tech debt item created in project backlog
- [ ] Scope defined (screen reader support, WCAG 2.1 AA)
- [ ] Estimated effort noted (research + implementation)
- [ ] Linked to PRD accessibility section

---

### [ ] AI-006: Consider TUI Wireframes

| Field | Value |
|-------|-------|
| **Priority** | P2 - MEDIUM |
| **Owner** | UX Expert |
| **Timeline** | Optional, before Epic 3 |
| **Status** | ⏳ Pending |

**Description:**  
Create ASCII/text wireframes for key TUI screens to improve developer clarity.

**Screens to Consider:**
- Room Explorer (Story 3.3)
- Booking Wizard (Story 3.4)
- Organizer Dashboard (Story 5.4)

---

## Progress Tracking

| Date | Action | Item | Notes |
|------|--------|------|-------|
| 2026-01-31 | Created | All | Generated from PO Master Checklist |
| 2026-01-31 | Complete | PRD Sharding | 13 files created in docs/prd/ |
| 2026-01-31 | Complete | Architecture Sharding | 23 files created in docs/architecture/ |

---

## Approval Gate

**Criteria for Development Start:**
- [x] PRD complete and validated
- [x] Architecture document complete
- [ ] Story files extracted (AI-001)
- [ ] Rate limiting documented (AI-002)

**Approval Status:** ⏳ PENDING

---

*Generated by PO Master Validation Checklist*  
*Sarah (Product Owner)*
