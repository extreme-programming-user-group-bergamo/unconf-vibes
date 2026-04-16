# Epic List

| Epic | Title | Goal | Implementation Status (2026-04-15) | Notes |
|------|-------|------|------------------------------------|-------|
| **Epic 1** | Foundation & Authentication | Establish project infrastructure (CLI skeleton, backend API, database, CI/CD) and implement GitHub OAuth authentication — delivering a working `unconf login` command | Complete | Delivery extends through Stories 1.8 and 1.9; proposed next-step story: 1.10 Active Session Management |
| **Epic 2** | Conference Discovery & Context | Enable users to browse conferences, view details, and set context — delivering `list`, `info`, and `checkout` commands | Complete | CLI, API, and local context storage all match the shipped codebase |
| **Epic 3** | Room Exploration & Booking | Implement the core TUI room explorer and booking flow — delivering `rooms` TUI and `book` command with wizard mode | Partially wired end-to-end | Explorer, wizard UI, direct-booking CLI path, booking/status services, and tests are present; `POST /bookings` is now registered in the live API router; proposed next-step story: 3.8 |
| **Epic 4** | Social Features & Roommates | Add attendee visibility, roommate requests, and booking management — delivering `requests`, `status`, and social discovery features | Complete | Request lifecycle, attendee listing, cancellation, and roommate-departure handling are all implemented |
| **Epic 5** | Organizer Tools & Hotel Automation | Provide organizer dashboard, conference management, and automated hotel email — delivering organizer commands and email integration | Complete | Delivery extends through Story 5.10 with the explicit `removed` decision for manual organizer booking confirmation (`unconf confirm` / `PUT /bookings/{id}/confirm`) |
| **Epic 6** | Reliability & Operational Hardening | Deliver the prioritized P0/P1 hardening pass for booking consistency, request atomicity, config correctness, networking resilience, and docs-behavior parity | Planned | Introduces Stories 6.1–6.7 as post-MVP high-impact reliability work |

---
