# Epic 4: Social Features & Roommates

**Goal**: Add attendee visibility, roommate requests, and booking management — delivering `requests`, `attendees` commands, and social discovery features that differentiate UNCONF from simple booking tools.

## Story 4.1: Attendee List API & Command

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

## Story 4.2: Roommate Request Data Model & API

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

## Story 4.3: Send Roommate Request

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

## Story 4.4: Manage Roommate Requests

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

## Story 4.5: Booking Cancellation

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

## Story 4.6: Roommate Departure Handling

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
