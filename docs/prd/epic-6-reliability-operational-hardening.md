# Epic 6: Reliability & Operational Hardening

**Goal**: Eliminate high-impact booking consistency and configuration reliability gaps so the MVP behaves predictably under concurrency, uses safe network defaults, and exposes one coherent operator/developer command surface.

**Command boundary note:** Attendee and organizer product commands remain on `unconf`; server process runtime configuration and diagnostics remain on `unconf-server`. This epic clarifies and enforces that split for config semantics and docs.

## Story 6.1: Atomic Room Capacity Enforcement

**As an** attendee,  
**I want** room-capacity checks and booking creation to be atomic,  
**so that** concurrent booking attempts cannot overbook the same room.

**Acceptance Criteria:**
1. Booking creation executes capacity validation and insert/update in one transaction boundary
2. Concurrent booking attempts for the final available spot yield exactly one success and deterministic conflict responses for the rest
3. API conflict response remains stable (`409` + domain error such as `room_full`) for rejected attempts
4. Integration coverage includes a concurrent-create scenario that proves no persisted over-capacity state

## Story 6.2: Atomic Roommate Accept Flow

**As an** attendee,  
**I want** roommate-request acceptance to be atomic,  
**so that** partial side effects do not leave requests, bookings, and occupancy out of sync.

**Acceptance Criteria:**
1. Request status transition, booking mutation/creation, and occupancy updates are committed in one transaction
2. If any step fails, no partial state is persisted (request status and booking state both roll back)
3. Accepting a request re-checks room and conference capacity within the same transaction before commit
4. Error responses distinguish business conflicts (already booked/full/not allowed) from internal failures
5. Integration tests cover successful accept, conflict during accept, and rollback on injected failure

## Story 6.3: Deterministic Config Flag API Client Wiring

**As a** CLI user,  
**I want** `--config` to deterministically control which API base URL and auth settings are loaded,  
**so that** each command runs against the environment I explicitly selected.

**Acceptance Criteria:**
1. All API-using commands resolve configuration from one shared loader honoring precedence (`--config` > env > default file)
2. `--config` on the root command affects every subcommand that initializes an API client
3. Startup logs/debug output expose the effective config source and API base URL when verbose mode is enabled
4. Regression tests verify two different config files drive different API endpoints in the same test suite

## Story 6.4: Conference-Level Capacity Gate on Booking Create

**As an** organizer,  
**I want** booking creation to enforce conference-level capacity in addition to room capacity,  
**so that** total registrations cannot exceed conference limits.

**Acceptance Criteria:**
1. Booking create path checks active booking count against conference capacity before commit
2. If conference capacity is reached, booking creation returns a stable conflict response (`409` + `conference_full`)
3. Capacity check is performed in the same transaction as booking persistence to avoid race-induced overflow
4. `unconf status`, organizer dashboard totals, and attendee counts remain consistent with enforced conference capacity

## Story 6.5: Safe HTTP Client Defaults for Timeout and Retry

**As a** maintainer,  
**I want** shared HTTP client defaults for timeout and bounded retries,  
**so that** transient network faults do not hang commands or cause brittle failures.

**Acceptance Criteria:**
1. CLI API client uses a non-zero default request timeout suitable for interactive commands
2. Retry policy is defined for retryable failures (transport errors and selected 5xx), with bounded attempts and backoff
3. Non-retryable responses (4xx business errors) return immediately without retry loops
4. Timeout and retry settings are configurable but have safe defaults when unset
5. Client tests cover timeout behavior, retry on transient failure, and no-retry behavior on 4xx

## Story 6.6: Clear Server vs CLI Config Semantics

**As a** developer/operator,  
**I want** separate and unambiguous config keys for server listen address and CLI API endpoint,  
**so that** local/dev/prod setups do not misroute traffic or bind incorrectly.

**Acceptance Criteria:**
1. `unconf-server` runtime bind/listen setting uses server-specific keys/flags only (for example `server.listen_addr`)
2. `unconf` API target uses client-specific keys/flags only (for example `client.api_base_url`)
3. Deprecated ambiguous keys are either removed or mapped with explicit deprecation warnings and migration guidance
4. Help text and examples for both binaries reflect the split and avoid cross-usage ambiguity
5. Config parsing tests verify each binary ignores irrelevant keys from the other surface

## Story 6.7: Docs-Behavior Parity for Shipped Surface

**As a** user and contributor,  
**I want** docs to match the shipped CLI/API behavior,  
**so that** onboarding, operation, and development decisions are based on accurate guidance.

**Acceptance Criteria:**
1. PRD and architecture sections that describe booking consistency, config semantics, and command surface are updated to match implemented behavior
2. README and command examples are updated for current `unconf` and `unconf-server` config usage
3. Removed or out-of-scope commands/endpoints are explicitly deleted or marked as non-shipping
4. A lightweight release checklist item verifies doc parity for any command/config contract changes in this epic

---
