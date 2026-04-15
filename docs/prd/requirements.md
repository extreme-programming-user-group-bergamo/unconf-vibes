# Requirements

## Functional Requirements

### Authentication & Identity
- **FR1**: The system shall authenticate users via GitHub OAuth (device flow for CLI)
- **FR2**: The system shall allow users to configure their profile (display name, email, privacy preference)

### Conference Discovery & Context
- **FR3**: The system shall display a browsable list of available conferences with search/filter
- **FR4**: The system shall allow users to set an active conference context (`checkout`)
- **FR5**: The system shall display conference details (dates, location, pricing, capacity)

### Room Exploration (TUI)
- **FR6**: The system shall provide an interactive TUI for room exploration
- **FR7**: The TUI shall display room types, prices, and real-time availability
- **FR8**: The TUI shall show current occupants of rooms (respecting privacy settings)
- **FR9**: The TUI shall support filtering rooms by type (single/double/triple)

### Booking
- **FR10**: The system shall support guided wizard mode for first-time bookings
- **FR11**: The system shall support direct CLI booking (`book <room_number>`) for power users
- **FR12**: The system shall allow users to set privacy flag on booking (`--private`)
- **FR13**: The system shall allow users to add notes for hotel (dietary requests, accessibility needs)
- **FR14**: The system shall implement a roommate request system (request → accept/decline flow)

### Booking Management
- **FR15**: The system shall display current booking status
- **FR16**: The system shall allow users to view and manage incoming roommate requests
- **FR17**: The system shall allow users to cancel their booking (with confirmation)
- **FR18**: When a roommate cancels, the remaining person retains the room and can find a replacement, stay solo, or cancel

### Organizer Features
- **FR19**: The system shall allow organizers to create and configure conferences (dates, rooms, capacity, pricing)
- **FR20**: The system shall provide organizers a dashboard to view registrations with attendee preferences
- **FR21**: The system shall automatically send templated booking emails to hotels
- **FR22**: The system shall allow organizers to export registration data as CSV

### Privacy
- **FR23**: The system shall support two visibility tiers: Public (name visible) and Private (anonymous to other attendees)
- **FR24**: Organizers shall always be able to see all attendee information regardless of privacy setting

## Implementation Status Snapshot (2026-04-15)

| Area | Status | Notes |
|------|--------|-------|
| Authentication & Identity | Complete | Stories 1.1 through 1.9 are implemented, including auth middleware, user profile endpoints, current-session revoke, and automatic access-token refresh in the authenticated CLI client |
| Conference Discovery & Context | Complete | Stories 2.1 through 2.5 are implemented across API, CLI, context storage, and profile configuration |
| Room Exploration & Booking | Partially wired end-to-end | Room explorer TUI, booking wizard UI, direct CLI booking, room listing, booking service logic, and booking client support all exist, but the live server does not currently register `POST /bookings`, which blocks new bookings against a running API |
| Social Features & Roommates | Complete | Attendee listing, roommate requests, booking cancellation, and roommate-departure handling are implemented |
| Organizer Features & Hotel Automation | Complete | Conference create/edit, room management, organizer dashboard, CSV export, organizer ownership flows, and hotel email automation are implemented |

Follow-on implementation stories added after the initial PRD draft: **1.8**, **1.9**, **5.8**, and **5.9**.

## Non-Functional Requirements

### Performance
- **NFR1**: API responses shall complete in under 200ms
- **NFR2**: TUI shall render at 60fps for smooth interactions

### Platform Compatibility
- **NFR3**: The CLI shall be cross-platform (macOS, Linux, Windows)
- **NFR4**: The TUI shall work on standard terminals (iTerm2, Terminal.app, Windows Terminal, Linux TTYs)

### Usability
- **NFR5**: Registration flow shall be completable in under 3 minutes
- **NFR6**: The system shall use friendly, welcoming language accessible to non-power-users

### Budget & Infrastructure
- **NFR7**: The system shall operate within free-tier or minimal hosting costs (community/volunteer project)
- **NFR8**: Email delivery shall use SMTP standard (SendGrid/Mailgun for production)

### Security
- **NFR9**: All API communication shall use HTTPS
- **NFR10**: Authentication tokens shall be securely stored on the client

### Reliability
- **NFR11**: Hotel communication emails shall have zero transcription errors from automated templates
- **NFR12**: The system shall clearly indicate data staleness if offline/cached

---
