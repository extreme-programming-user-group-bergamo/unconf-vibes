# Epic 2: Conference Discovery & Context

**Goal**: Enable users to browse conferences, view details, and set context — delivering `list`, `info`, and `checkout` commands that form the foundation for all subsequent conference-specific operations.

## Story 2.1: Conference Data Model & API

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

## Story 2.2: CLI List Command

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

## Story 2.3: CLI Info Command

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

## Story 2.4: Context Management (Checkout)

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

## Story 2.5: User Profile Configuration

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
