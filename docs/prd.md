# UNCONF CLI Product Requirements Document (PRD)

## Goals and Background Context

### Goals

- **Streamline conference registration** to under 3 minutes from start to confirmation
- **Enable social discovery** — let attendees see who's attending and choose roommates
- **Automate hotel communication** — eliminate manual email coordination
- **Reduce organizer overhead** by 90% through self-service tools and dashboards
- **Respect attendee privacy** — provide visibility controls (Public/Private) by design
- **Deliver a developer-delightful experience** — CLI + TUI hybrid matching community values

### Background Context

Developer unconferences like SoCraTes Italia and Polenta & Deploy follow the Open Space format where hotel room booking is inherently social — attendees want to know who else is attending and choose their roommates based on friendships or networking goals. The current process relies on email chains, shared spreadsheets, and manual hotel communication, creating 15-30 minute registration times and 5-10 hours of administrative work per event for organizers.

UNCONF CLI addresses this by providing a command-line tool with a text-based user interface (TUI) that embodies the collaborative spirit of unconferences. Following the conventions of developer tools like `kubectl` and `git`, it uses context-based workflows (`unconf checkout socrates-26`, `unconf rooms`, `unconf book 204`) to make registration fast and social discovery intuitive. The MVP targets deployment at SoCraTes Italia 2026 (Q2 2026).

### Change Log

| Date | Version | Description | Author |
|------|---------|-------------|--------|
| Jan 31, 2026 | 0.1 | Initial PRD draft | John (PM) |

---

## Requirements

### Functional Requirements

#### Authentication & Identity
- **FR1**: The system shall authenticate users via GitHub OAuth (device flow for CLI)
- **FR2**: The system shall allow users to configure their profile (display name, email, privacy preference)

#### Conference Discovery & Context
- **FR3**: The system shall display a browsable list of available conferences with search/filter
- **FR4**: The system shall allow users to set an active conference context (`checkout`)
- **FR5**: The system shall display conference details (dates, location, pricing, capacity)

#### Room Exploration (TUI)
- **FR6**: The system shall provide an interactive TUI for room exploration
- **FR7**: The TUI shall display room types, prices, and real-time availability
- **FR8**: The TUI shall show current occupants of rooms (respecting privacy settings)
- **FR9**: The TUI shall support filtering rooms by type (single/double/triple)

#### Booking
- **FR10**: The system shall support guided wizard mode for first-time bookings
- **FR11**: The system shall support direct CLI booking (`book <room_number>`) for power users
- **FR12**: The system shall allow users to set privacy flag on booking (`--private`)
- **FR13**: The system shall allow users to add notes for hotel (dietary requests, accessibility needs)
- **FR14**: The system shall implement a roommate request system (request → accept/decline flow)

#### Booking Management
- **FR15**: The system shall display current booking status
- **FR16**: The system shall allow users to view and manage incoming roommate requests
- **FR17**: The system shall allow users to cancel their booking (with confirmation)
- **FR18**: When a roommate cancels, the remaining person retains the room and can find a replacement, stay solo, or cancel

#### Organizer Features
- **FR19**: The system shall allow organizers to create and configure conferences (dates, rooms, capacity, pricing)
- **FR20**: The system shall provide organizers a dashboard to view registrations with attendee preferences
- **FR21**: The system shall automatically send templated booking emails to hotels
- **FR22**: The system shall allow organizers to export registration data as CSV

#### Privacy
- **FR23**: The system shall support two visibility tiers: Public (name visible) and Private (anonymous to other attendees)
- **FR24**: Organizers shall always be able to see all attendee information regardless of privacy setting

### Non-Functional Requirements

#### Performance
- **NFR1**: API responses shall complete in under 200ms
- **NFR2**: TUI shall render at 60fps for smooth interactions

#### Platform Compatibility
- **NFR3**: The CLI shall be cross-platform (macOS, Linux, Windows)
- **NFR4**: The TUI shall work on standard terminals (iTerm2, Terminal.app, Windows Terminal, Linux TTYs)

#### Usability
- **NFR5**: Registration flow shall be completable in under 3 minutes
- **NFR6**: The system shall use friendly, welcoming language accessible to non-power-users

#### Budget & Infrastructure
- **NFR7**: The system shall operate within free-tier or minimal hosting costs (community/volunteer project)
- **NFR8**: Email delivery shall use SMTP standard (SendGrid/Mailgun for production)

#### Security
- **NFR9**: All API communication shall use HTTPS
- **NFR10**: Authentication tokens shall be securely stored on the client

#### Reliability
- **NFR11**: Hotel communication emails shall have zero transcription errors from automated templates
- **NFR12**: The system shall clearly indicate data staleness if offline/cached

---

## User Interface Design Goals

### Overall UX Vision

UNCONF delivers a **"functional delight"** experience — satisfying to use without being flashy. The interface follows developer tool conventions (like `kubectl` and `git`), feeling immediately familiar to the target audience. First-timers are guided through a welcoming wizard experience, while power users get fast, scriptable CLI commands. The guiding principle: **the fun happens at the conference, not in the CLI**.

### Key Interaction Paradigms

1. **TUI for Discovery, CLI for Action** — Interactive TUI (Bubble Tea) for exploration (rooms, attendees, conference info) with direct CLI commands for rapid execution (`unconf book 204`)

2. **Context-Based Workflow** — Users `checkout` a conference context, and subsequent commands operate within that context (similar to git branches or kubectl contexts)

3. **Progressive Disclosure** — Information revealed as needed; defaults for common cases with advanced options available but not overwhelming

4. **Request-Based Social Interactions** — Roommate selection uses an explicit request/accept flow, giving control to both parties

### Core Screens and Views

| Screen | Purpose |
|--------|---------|
| **Conference List** | Browse/search available unconferences, see attendee counts |
| **Conference Info** | View dates, location, pricing, capacity, current attendees |
| **Room Explorer (TUI)** | Interactive room selection with occupancy, pricing, filtering |
| **Booking Wizard** | Step-by-step guided booking for first-timers |
| **Status View** | Current booking details, pending requests |
| **Requests View** | Manage incoming/outgoing roommate requests |
| **Organizer Dashboard** | Registration overview, attendee preferences, capacity |

### Accessibility

**None** (MVP) — Standard terminal accessibility. Screen reader support and WCAG compliance identified as areas needing further research.

### Branding

- **Aesthetic**: Clean, minimal terminal UI with thoughtful use of color
- **Tone**: Warm, welcoming, direct language — not intimidating
- **Visual Framework**: Lip Gloss (Charm) for styling
- **Constraint**: Must work on terminals with limited color support

### Target Devices and Platforms

**Cross-Platform CLI**:
- macOS (iTerm2, Terminal.app)
- Linux (standard TTYs)
- Windows (Windows Terminal)

No mobile or web interface planned — CLI-first tool for developers.

---

## Technical Assumptions

### Repository Structure: Monorepo

A single repository with Go standard layout:
```
/cmd         # CLI entry points
/internal    # Private application code
/pkg         # Shared libraries (if needed)
```

**Rationale**: Small team (1-3 developers), single codebase for CLI and backend simplifies development and deployment.

### Service Architecture

**CLI as API Client + Stateless Backend**

- **CLI/TUI**: Go application using Cobra + Viper for commands, Bubble Tea + Lip Gloss for TUI
- **Backend API**: Go + Gin (REST API), containerized with Docker
- **Separation**: CLI makes HTTP calls to backend; backend handles all data persistence and business logic

**Rationale**: Clean separation allows CLI updates independent of backend; stateless backend enables horizontal scaling if needed.

### Testing Requirements

**Unit + Integration Testing**

- `go test` for unit tests
- `golangci-lint` on every PR
- Integration tests for API endpoints
- MailHog for email testing in development

Manual testing convenience methods for TUI interactions (TUI testing is notoriously difficult to automate).

**Rationale**: Balance thoroughness with development velocity; TUI testing automation deferred for pragmatic reasons.

### Additional Technical Assumptions and Requests

#### Languages & Frameworks
- **Go 1.21+** — Primary language for both CLI and backend
- **Cobra** — CLI framework (industry standard)
- **Viper** — Configuration management
- **Bubble Tea** — TUI framework (Charm ecosystem)
- **Lip Gloss** — TUI styling
- **Gin** — HTTP framework for REST API

#### Database
- **SQLite** initially with **Repository pattern** — Enables future PostgreSQL migration without code changes
- Database file stored server-side (not in CLI)

#### Authentication
- **GitHub OAuth** — Device flow for CLI authentication
- **Token-based API auth** — JWT or similar for authenticated API requests

#### Email
- **SMTP standard** — For hotel communication
- **SendGrid/Mailgun** — Production email delivery
- **MailHog** — Local development/testing

#### CI/CD
- **GoReleaser** — Cross-platform builds
- **GitHub Actions** — CI pipeline
- **Git tags** — Trigger releases
- **GitHub Releases** — Binary distribution
- **Optional Homebrew tap** — For macOS users

#### Deployment
- **Docker** — Containerized backend
- **Free-tier hosting** — Community budget constraint (Fly.io, Railway, or similar)

#### Security
- **HTTPS only** — All API communication
- **Secure token storage** — CLI stores auth tokens in OS keychain or secure file

---

## Epic List

| Epic | Title | Goal |
|------|-------|------|
| **Epic 1** | Foundation & Authentication | Establish project infrastructure (CLI skeleton, backend API, database, CI/CD) and implement GitHub OAuth authentication — delivering a working `unconf login` command |
| **Epic 2** | Conference Discovery & Context | Enable users to browse conferences, view details, and set context — delivering `list`, `info`, and `checkout` commands |
| **Epic 3** | Room Exploration & Booking | Implement the core TUI room explorer and booking flow — delivering `rooms` TUI and `book` command with wizard mode |
| **Epic 4** | Social Features & Roommates | Add attendee visibility, roommate requests, and booking management — delivering `requests`, `status`, and social discovery features |
| **Epic 5** | Organizer Tools & Hotel Automation | Provide organizer dashboard, conference management, and automated hotel email — delivering organizer commands and email integration |

---

## Epic 1: Foundation & Authentication

**Goal**: Establish project infrastructure (CLI skeleton, backend API, database, CI/CD) and implement GitHub OAuth authentication — delivering a working `unconf login` command that proves the full stack works end-to-end.

### Story 1.1: Project Scaffolding

**As a** developer,  
**I want** the project repository initialized with Go module, folder structure, and basic tooling,  
**so that** I have a clean foundation to build upon.

**Acceptance Criteria:**
1. Go module initialized (`go.mod`) with appropriate module path
2. Standard Go project layout created (`/cmd`, `/internal`, `/pkg`)
3. `.gitignore` configured for Go projects
4. `README.md` with project overview and setup instructions
5. Makefile or task runner with common commands (`build`, `test`, `lint`)
6. `golangci-lint` configuration file present

### Story 1.2: CLI Framework Setup

**As a** developer,  
**I want** the Cobra CLI framework initialized with a root command,  
**so that** I can add subcommands for UNCONF features.

**Acceptance Criteria:**
1. Cobra initialized with `unconf` as root command
2. Viper integrated for configuration management
3. `--help` flag works and shows welcoming help text
4. `--version` flag displays version information
5. Configuration file support (`.unconf.yaml` or similar)
6. Basic error handling and exit codes established

### Story 1.3: Backend API Skeleton

**As a** developer,  
**I want** a basic Go + Gin backend API running,  
**so that** the CLI has an API to communicate with.

**Acceptance Criteria:**
1. Gin router initialized with health check endpoint (`GET /health`)
2. Structured logging configured (using `log/slog`)
3. Graceful shutdown handling implemented
4. CORS configured for development
5. Dockerfile created for containerized deployment
6. API responds with JSON; health endpoint returns `{"status": "ok"}`

### Story 1.4: Database Setup with Repository Pattern

**As a** developer,  
**I want** SQLite database initialized with repository pattern,  
**so that** data persistence is abstracted and migration-ready.

**Acceptance Criteria:**
1. SQLite database file created on first run
2. Database connection pool/manager implemented
3. Repository interface defined for data access abstraction
4. Initial schema migration system in place
5. `users` table created with basic fields (id, github_id, email, display_name, created_at)
6. Repository pattern allows swapping SQLite for PostgreSQL without business logic changes

### Story 1.5: CI/CD Pipeline

**As a** developer,  
**I want** GitHub Actions CI pipeline configured,  
**so that** code quality is enforced on every PR.

**Acceptance Criteria:**
1. GitHub Actions workflow triggers on PR and push to main
2. Pipeline runs `go test ./...` for all tests
3. Pipeline runs `golangci-lint` for static analysis
4. Pipeline fails if tests fail or linting errors exist
5. GoReleaser configuration file present (for future releases)
6. Build succeeds for linux/amd64, darwin/amd64, darwin/arm64, windows/amd64

### Story 1.6: GitHub OAuth Integration (Backend)

**As a** backend developer,  
**I want** GitHub OAuth device flow endpoints implemented,  
**so that** CLI users can authenticate without browser redirects.

**Acceptance Criteria:**
1. `POST /auth/device` endpoint initiates device flow, returns device_code and user_code
2. `POST /auth/token` endpoint polls for access token completion
3. GitHub OAuth app credentials configurable via environment variables
4. On successful auth, user record created/updated in database
5. JWT token issued for subsequent API authentication
6. Token expiry and refresh mechanism defined

### Story 1.7: CLI Login Command

**As an** attendee,  
**I want** to run `unconf login` to authenticate with my GitHub account,  
**so that** I can access UNCONF features.

**Acceptance Criteria:**
1. `unconf login` command initiates GitHub device flow
2. User shown device code and URL to enter it (https://github.com/login/device)
3. CLI polls backend for authentication completion
4. On success, auth token stored securely (OS keychain or config file)
5. Success message displayed with user's GitHub username
6. Subsequent commands can detect authenticated state
7. `unconf logout` command removes stored credentials

---

## Epic 2: Conference Discovery & Context

**Goal**: Enable users to browse conferences, view details, and set context — delivering `list`, `info`, and `checkout` commands that form the foundation for all subsequent conference-specific operations.

### Story 2.1: Conference Data Model & API

**As a** developer,  
**I want** the conference data model and CRUD API endpoints implemented,  
**so that** conference information can be stored and retrieved.

**Acceptance Criteria:**
1. `conferences` table created with fields: id, slug, name, description, location, start_date, end_date, capacity, created_at
2. Repository interface for conference data access
3. `GET /conferences` endpoint returns list of conferences (public)
4. `GET /conferences/{slug}` endpoint returns conference details
5. Conference status derived from dates (upcoming, active, past)
6. API responses include attendee count (respecting privacy)

### Story 2.2: CLI List Command

**As an** attendee,  
**I want** to run `unconf list` to see available conferences,  
**so that** I can discover unconferences I might attend.

**Acceptance Criteria:**
1. `unconf list` displays conferences in a formatted table
2. Table shows: name, dates, location, attendee count, status
3. Results sortable by date (default: upcoming first)
4. `--all` flag shows past conferences as well
5. Empty state message if no conferences available
6. Works without authentication (public data)

### Story 2.3: CLI Info Command

**As an** attendee,  
**I want** to run `unconf info` to see details about a conference,  
**so that** I can learn more before committing.

**Acceptance Criteria:**
1. `unconf info <conference-slug>` displays detailed conference information
2. Shows: name, description, dates, location, capacity, current attendee count
3. Shows room types available with price ranges
4. If context is set, `unconf info` (no argument) shows current context conference
5. Error message if conference slug not found
6. Works without authentication (public data)

### Story 2.4: Context Management (Checkout)

**As an** attendee,  
**I want** to run `unconf checkout <conference>` to set my active context,  
**so that** subsequent commands operate on that conference.

**Acceptance Criteria:**
1. `unconf checkout <slug>` sets active conference context
2. Context stored in local config file (persists across sessions)
3. `unconf checkout` (no argument) shows current context
4. Success message confirms context switch: "Switched to socrates-26"
5. Error if conference slug doesn't exist
6. Context indicator available for TUI header (story dependency for Epic 3)

### Story 2.5: User Profile Configuration

**As an** attendee,  
**I want** to run `unconf config` to set my display name and privacy preferences,  
**so that** my profile is ready for booking.

**Acceptance Criteria:**
1. `unconf config` opens interactive prompts for profile setup
2. Collects: display name, email (pre-filled from GitHub), privacy preference
3. Privacy options: Public (name visible) or Private (anonymous)
4. `unconf config --show` displays current profile settings
5. Profile data sent to backend and stored in user record
6. Can update individual fields: `unconf config --privacy private`

---

## Epic 3: Room Exploration & Booking

**Goal**: Implement the core TUI room explorer and booking flow — delivering `rooms` TUI and `book` command with both wizard mode and direct CLI booking. This is the heart of the MVP value proposition.

### Story 3.1: Room Data Model & API

**As a** developer,  
**I want** the room data model and API endpoints implemented,  
**so that** room information can be stored and retrieved.

**Acceptance Criteria:**
1. `rooms` table created with fields: id, conference_id, room_number, room_type (single/double/triple), price_per_night, capacity, created_at
2. `bookings` table created with fields: id, room_id, user_id, conference_id, status, notes, privacy_setting, created_at
3. Repository interfaces for rooms and bookings
4. `GET /conferences/{slug}/rooms` returns all rooms with availability
5. Room availability computed from bookings (spots taken vs capacity)
6. API returns occupant names only for Public bookings

### Story 3.2: TUI Framework Setup

**As a** developer,  
**I want** Bubble Tea TUI framework integrated into the CLI,  
**so that** interactive terminal interfaces can be built.

**Acceptance Criteria:**
1. Bubble Tea dependency added and initialized
2. Lip Gloss integrated for styling
3. Base TUI model structure established (Model, Update, View pattern)
4. Common components created: header with context indicator, footer with help
5. Color scheme defined that works on limited-color terminals
6. Graceful fallback if terminal doesn't support TUI features

### Story 3.3: Room Explorer TUI

**As an** attendee,  
**I want** to run `unconf rooms` to explore available rooms interactively,  
**so that** I can see options and make an informed choice.

**Acceptance Criteria:**
1. `unconf rooms` launches interactive TUI (requires conference context)
2. Rooms displayed in scrollable list with: room number, type, price, availability (e.g., "2/3 spots")
3. Room occupants shown (names for Public, "Private attendee" for Private)
4. Filter by room type (single/double/triple) via keyboard shortcut
5. Sort by price or availability
6. Press Enter on room to initiate booking flow
7. Press 'q' to exit TUI
8. Context indicator shown in header

### Story 3.4: Booking Wizard (TUI Flow)

**As a** first-time attendee,  
**I want** a guided booking wizard after selecting a room,  
**so that** I understand the process and don't miss important options.

**Acceptance Criteria:**
1. After selecting room in TUI, wizard flow begins
2. Step 1: Confirm room selection (show price, type, current occupants)
3. Step 2: Privacy setting (Public/Private) with clear explanation
4. Step 3: Optional notes for hotel (dietary, accessibility)
5. Step 4: Review and confirm booking
6. On confirm, booking created via API
7. Success screen with booking confirmation details
8. Error handling for room-full race conditions

### Story 3.5: Direct CLI Booking

**As a** power user,  
**I want** to run `unconf book <room>` to book directly without TUI,  
**so that** I can book quickly when I know what I want.

**Acceptance Criteria:**
1. `unconf book <room_number>` creates booking directly
2. `--private` flag sets privacy to Private (default: use profile setting)
3. `--notes "message"` adds hotel notes
4. Confirmation prompt before booking (can skip with `--yes`)
5. Success message with booking details
6. Error if room is full or doesn't exist
7. Error if user already has a booking for this conference

### Story 3.6: Booking Status Command

**As an** attendee,  
**I want** to run `unconf status` to see my current booking,  
**so that** I can verify my reservation details.

**Acceptance Criteria:**
1. `unconf status` shows current booking for active conference
2. Displays: room number, type, price, check-in/out dates, privacy setting
3. Shows roommates (if any) respecting their privacy settings
4. Shows pending roommate requests (incoming/outgoing) — placeholder for Epic 4
5. "No booking" message if not booked
6. `--all` flag shows bookings across all conferences

---

## Epic 4: Social Features & Roommates

**Goal**: Add attendee visibility, roommate requests, and booking management — delivering `requests`, `attendees` commands, and social discovery features that differentiate UNCONF from simple booking tools.

### Story 4.1: Attendee List API & Command

**As an** attendee,  
**I want** to see who else is attending the conference,  
**so that** I can connect with friends and find potential roommates.

**Acceptance Criteria:**
1. `GET /conferences/{slug}/attendees` returns list of public attendees
2. Response includes: display name, GitHub username (if public), room info (if booked)
3. Private attendees shown as count only ("+ 5 private attendees")
4. `unconf attendees` command displays attendee list
5. Formatted table: name, room (or "not booked yet")
6. Requires authentication to view (attendees are opt-in public)

### Story 4.2: Roommate Request Data Model & API

**As a** developer,  
**I want** the roommate request data model and API endpoints,  
**so that** users can request and manage roommates.

**Acceptance Criteria:**
1. `roommate_requests` table: id, requester_id, target_id, room_id, status (pending/accepted/declined), created_at
2. `POST /requests` creates a roommate request
3. `GET /requests` returns user's incoming and outgoing requests
4. `PUT /requests/{id}/accept` accepts request
5. `PUT /requests/{id}/decline` declines request
6. When accepted, target user added to requester's room booking
7. Validation: can't request if room is full, can't request self

### Story 4.3: Send Roommate Request

**As an** attendee with a booking,  
**I want** to invite someone to be my roommate,  
**so that** I can share a room with a friend.

**Acceptance Criteria:**
1. `unconf invite <username>` sends roommate request
2. Can also invite from TUI room view (when viewing own room)
3. Target user must exist and not have a booking for this conference
4. Request created with pending status
5. Success message: "Request sent to @username"
6. Error if target already has a booking or request pending
7. Can only invite if room has available spots

### Story 4.4: Manage Roommate Requests

**As an** attendee,  
**I want** to view and respond to roommate requests,  
**so that** I can accept or decline invitations.

**Acceptance Criteria:**
1. `unconf requests` shows incoming and outgoing requests
2. Incoming requests show: requester name, room details, status
3. Outgoing requests show: target name, status
4. `unconf requests accept <request_id>` accepts a request
5. `unconf requests decline <request_id>` declines a request
6. On accept, user added to room and booking created
7. Notification concept: accepted/declined shown in status command

### Story 4.5: Booking Cancellation

**As an** attendee,  
**I want** to cancel my booking,  
**so that** I can free up the room if my plans change.

**Acceptance Criteria:**
1. `unconf cancel` cancels current booking (with confirmation)
2. Confirmation shows what will happen: "Cancel your booking in Room 204?"
3. If user has roommates, they are notified (conceptually) and retain the room
4. Cancelled booking removed from room occupancy
5. Pending outgoing roommate requests cancelled
6. Success message confirms cancellation
7. `--yes` flag skips confirmation

### Story 4.6: Roommate Departure Handling

**As an** attendee whose roommate cancelled,  
**I want** to retain my room and have options,  
**so that** I'm not automatically reassigned.

**Acceptance Criteria:**
1. When a roommate cancels, remaining occupants keep the room
2. Remaining user can: find new roommate (send requests), stay solo, or cancel
3. Room availability updates to show open spot
4. Status command shows: "Room 204 (1/2 spots) - open spot available"
5. No automatic stranger assignment (per brief's resolved decisions)

---

## Epic 5: Organizer Tools & Hotel Automation

**Goal**: Provide organizer dashboard, conference management, and automated hotel email — delivering organizer commands and email integration that eliminate manual coordination overhead.

### Story 5.1: Organizer Role & Permissions

**As a** developer,  
**I want** organizer role and permissions implemented,  
**so that** organizers have access to management features.

**Acceptance Criteria:**
1. `conference_organizers` junction table: conference_id, user_id, role (owner/admin)
2. Organizer middleware validates permission for protected endpoints
3. Organizers can view all attendee data regardless of privacy settings
4. Conference creators automatically become owners
5. Owners can add/remove other organizers
6. API returns 403 for non-organizers on protected endpoints

### Story 5.2: Conference Creation & Setup

**As an** organizer,  
**I want** to create and configure a conference,  
**so that** attendees can discover and book rooms.

**Acceptance Criteria:**
1. `unconf create` launches interactive conference setup
2. Collects: name, slug, description, location, dates, capacity
3. `POST /conferences` creates conference (organizer only)
4. Conference created with organizer as owner
5. Success message with conference slug for sharing
6. `unconf edit` allows updating conference details

### Story 5.3: Room Configuration

**As an** organizer,  
**I want** to configure rooms for my conference,  
**so that** attendees can see and book available rooms.

**Acceptance Criteria:**
1. `unconf rooms add` adds a room (organizer command)
2. Collects: room number, type (single/double/triple), price, capacity
3. Bulk import option: `unconf rooms import <csv>`
4. `POST /conferences/{slug}/rooms` creates room (organizer only)
5. `unconf rooms edit <number>` updates room details
6. `unconf rooms remove <number>` removes room (fails if bookings exist)

### Story 5.4: Organizer Dashboard TUI

**As an** organizer,  
**I want** a dashboard to view registrations and attendee details,  
**so that** I can manage the event effectively.

**Acceptance Criteria:**
1. `unconf dashboard` launches organizer TUI (organizer only)
2. Shows: total registrations, capacity usage, room fill rates
3. Attendee list with: name, email, room, dietary/accessibility notes, privacy setting
4. Filter by: room type, booking status, has special requests
5. Search attendees by name
6. All attendee data visible (organizer bypasses privacy for logistics)

### Story 5.5: Email Template System

**As a** developer,  
**I want** a templated email system,  
**so that** automated hotel communications are structured and accurate.

**Acceptance Criteria:**
1. Email templates defined for: new booking, cancellation, modification
2. Templates include: guest name, room, dates, special requests
3. Template variables populated from booking data
4. Email rendering tested with MailHog in development
5. SMTP configuration via environment variables
6. SendGrid/Mailgun integration for production

### Story 5.6: Automated Hotel Email on Booking

**As an** organizer,  
**I want** automated emails sent to the hotel when bookings occur,  
**so that** I don't have to manually forward booking details.

**Acceptance Criteria:**
1. Hotel email address configurable per conference
2. On new booking: email sent to hotel with guest details and room
3. On cancellation: email sent to hotel with cancellation notice
4. Email includes: guest name, email, room number, dates, special notes
5. Organizer receives BCC copy of all hotel emails
6. Email delivery logged for audit trail
7. Retry mechanism for failed email delivery

### Story 5.7: CSV Export

**As an** organizer,  
**I want** to export registration data as CSV,  
**so that** I can use it in spreadsheets or share with the hotel.

**Acceptance Criteria:**
1. `unconf export` generates CSV of all bookings
2. CSV includes: name, email, room, dates, dietary/special notes
3. Output to file: `unconf export --output bookings.csv`
4. Output to stdout for piping: `unconf export | pbcopy`
5. Option to include cancelled bookings: `--include-cancelled`
6. Organizer-only command

---

## Checklist Results Report

### Executive Summary

| Metric | Result |
|--------|--------|
| **Overall PRD Completeness** | 92% |
| **MVP Scope Appropriateness** | Just Right |
| **Readiness for Architecture Phase** | Ready |
| **Critical Gaps** | None blocking |

### Category Analysis

| Category | Status | Notes |
|----------|--------|-------|
| 1. Problem Definition & Context | ✅ PASS | Clear problem statement, target users, success metrics from Brief |
| 2. MVP Scope Definition | ✅ PASS | Essential vs nice-to-have clear; Out of scope documented in Brief |
| 3. User Experience Requirements | ✅ PASS | User flows, TUI/CLI paradigm, accessibility noted as future research |
| 4. Functional Requirements | ✅ PASS | 24 FRs cover all MVP features; testable criteria |
| 5. Non-Functional Requirements | ✅ PASS | 12 NFRs cover performance, security, platform compatibility |
| 6. Epic & Story Structure | ✅ PASS | 5 Epics, 31 Stories; properly sequenced with dependencies |
| 7. Technical Guidance | ✅ PASS | Stack defined; architecture direction clear |
| 8. Cross-Functional Requirements | ⚠️ PARTIAL | Data model implicit in stories; migration strategy noted |
| 9. Clarity & Communication | ✅ PASS | Consistent language; user-focused throughout |

### Top Issues by Priority

**BLOCKERS**: None

**HIGH**:
- Data entity relationship diagram not explicitly documented (implicit in stories)
- GDPR/data retention policy not specified (noted as research area in Brief)

**MEDIUM**:
- Accessibility requirements deferred — should be revisited before launch
- No explicit load/capacity testing requirements (single conference MVP)

**LOW**:
- Could add sequence diagrams for complex flows (OAuth, roommate requests)
- Version numbering scheme for API not specified

### MVP Scope Assessment

**Scope is appropriate**:
- 31 stories across 5 epics is achievable for Q2 2026 target
- Each epic delivers clear, testable value
- No feature creep — V2 items clearly deferred

**Potential cuts if timeline pressure**:
- Story 5.4 (Organizer Dashboard TUI) could be simplified to CLI-only
- Story 5.7 (CSV Export) could be deferred if time-constrained

### Final Decision

**✅ READY FOR ARCHITECT** — The PRD is comprehensive, properly structured, and ready for architectural design.

---

## Next Steps

### Architect Prompt

> As the Architect, please review this PRD and the Project Brief (`docs/brief.md`) to create the Architecture Document. Focus on:
> 
> 1. **System architecture** — CLI client + backend API design
> 2. **Data model** — Entity relationships for users, conferences, rooms, bookings, requests
> 3. **API design** — REST endpoints following the functional requirements
> 4. **Authentication flow** — GitHub OAuth device flow implementation
> 5. **Email integration** — Hotel communication automation
> 6. **Deployment architecture** — Docker, CI/CD, free-tier hosting
> 
> The technical stack is defined: Go 1.21+, Cobra, Bubble Tea, Gin, SQLite (with Repository pattern for future PostgreSQL migration).

### UX Expert Prompt

> As the UX Expert, please review this PRD to create wireframes and detailed UI specifications for:
> 
> 1. **Room Explorer TUI** — The primary discovery interface
> 2. **Booking Wizard** — Step-by-step guided flow
> 3. **Organizer Dashboard** — Registration overview and management
> 
> Focus on progressive disclosure, first-timer friendliness, and the "functional delight" aesthetic described in the UI Design Goals section.

---

*Document created by John (Product Manager) using BMAD Method*  
*Source: Project Brief (docs/brief.md) + Brainstorming Session Results*
