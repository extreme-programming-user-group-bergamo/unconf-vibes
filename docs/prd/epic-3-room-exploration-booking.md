# Epic 3: Room Exploration & Booking

**Goal**: Implement the core TUI room explorer and booking flow — delivering `rooms` TUI and `book` command with both wizard mode and direct CLI booking. This is the heart of the MVP value proposition.

## Story 3.1: Room Data Model & API

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

## Story 3.2: TUI Framework Setup

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

## Story 3.3: Room Explorer TUI

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

## Story 3.4: Booking Wizard (TUI Flow)

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

## Story 3.5: Direct CLI Booking

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

## Story 3.6: Booking Status Command

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
