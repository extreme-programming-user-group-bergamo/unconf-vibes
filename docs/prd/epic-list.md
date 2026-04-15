# Epic List

| Epic | Title | Goal | Implementation Status (2026-04-15) | Notes |
|------|-------|------|------------------------------------|-------|
| **Epic 1** | Foundation & Authentication | Establish project infrastructure (CLI skeleton, backend API, database, CI/CD) and implement GitHub OAuth authentication — delivering a working `unconf login` command | Complete | Delivery extends through Stories 1.8 and 1.9; proposed next-step story: 1.10 Active Session Management |
| **Epic 2** | Conference Discovery & Context | Enable users to browse conferences, view details, and set context — delivering `list`, `info`, and `checkout` commands | Complete | CLI, API, and local context storage all match the shipped codebase |
| **Epic 3** | Room Exploration & Booking | Implement the core TUI room explorer and booking flow — delivering `rooms` TUI and `book` command with wizard mode | Partially wired end-to-end | Explorer, wizard UI, direct-booking CLI path, booking/status services, and tests are present, but `POST /bookings` is not currently registered in the live API router; proposed next-step stories: 3.7 and 3.8 |
| **Epic 4** | Social Features & Roommates | Add attendee visibility, roommate requests, and booking management — delivering `requests`, `status`, and social discovery features | Complete | Request lifecycle, attendee listing, cancellation, and roommate-departure handling are all implemented |
| **Epic 5** | Organizer Tools & Hotel Automation | Provide organizer dashboard, conference management, and automated hotel email — delivering organizer commands and email integration | Complete | Delivery extends through Stories 5.8 and 5.9; proposed next-step story: 5.10 Organizer Booking Confirmation Decision |

---
