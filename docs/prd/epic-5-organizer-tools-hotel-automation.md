# Epic 5: Organizer Tools & Hotel Automation

**Goal**: Provide organizer dashboard, conference management, and automated hotel email — delivering organizer commands and email integration that eliminate manual coordination overhead.

## Story 5.1: Organizer Role & Permissions

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

## Story 5.2: Conference Creation & Setup

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

## Story 5.3: Room Configuration

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

## Story 5.4: Organizer Dashboard TUI

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

## Story 5.5: Email Template System

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

## Story 5.6: Automated Hotel Email on Booking

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

## Story 5.7: CSV Export

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

## Story 5.8: Cancel Command

**As an** attendee,
**I want** a safe and explicit `unconf cancel` flow,
**so that** I can cancel my current booking without ambiguity and immediately free the room.

**Acceptance Criteria:**
1. `unconf cancel` routes through the backend API and requires explicit confirmation by default
2. `--yes` skips the confirmation prompt for scripted or power-user flows
3. Successful cancellation removes the booking from active occupancy projections and keeps roommate-preservation behavior intact

## Story 5.9: Status Command Lifecycle Consistency

**As an** attendee,
**I want** `unconf status` to reflect my live booking lifecycle immediately after cancellations,
**so that** I can trust what the CLI shows me after booking-state transitions.

**Acceptance Criteria:**
1. `unconf status` excludes cancelled bookings from active status projections
2. `unconf cancel` followed by `unconf status` yields a deterministic no-active-booking outcome in the same conference context
3. Status output remains privacy-aware for roommate visibility and consistent with existing cancellation behavior

## Story 5.10: Organizer Booking Confirmation Decision

**As an** organizer,
**I want** booking-confirmation behavior to be explicit in the product surface,
**so that** attendee flows, hotel automation, and the docs do not disagree about whether a manual confirmation step exists.

**Acceptance Criteria:**
1. Product and engineering decide whether organizer confirmation remains in scope or is removed from the design
2. If retained, the API, CLI, and architecture docs all expose and describe the same confirmation behavior
3. If removed, legacy references to organizer confirmation are deleted from the PRD and architecture set

**Decision (2026-04-16):** `removed` path selected. Manual organizer booking confirmation is out of scope for the current product surface; there is no `unconf confirm` command and no `PUT /bookings/{id}/confirm` API contract.

---
