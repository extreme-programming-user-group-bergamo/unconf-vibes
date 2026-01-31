# 20. Glossary

## Domain Terms

| Term | Definition |
|------|------------|
| **Unconference** | A participant-driven meeting format (also called Open Space) where the agenda is created by attendees on the day. Examples: SoCraTes Italia, Polenta & Deploy |
| **Attendee** | A user who has a booking (active or requested) for a conference. Implicit role — no separate entity |
| **Organizer** | A user with owner or admin role for a conference. Can view all attendee data and confirm bookings |
| **Roommate** | A user sharing a hotel room with another attendee. Created via roommate request flow |
| **Privacy Setting** | User's visibility preference: **Public** (name shown to other attendees) or **Private** (appears as "Private attendee") |
| **Context** | The currently selected conference for CLI commands. Set via `unconf checkout <slug>` |
| **Booking Status** | Lifecycle state of a reservation: **requested** → **confirmed** → **cancelled** |

## Technical Terms

| Term | Definition |
|------|------------|
| **Device Flow** | OAuth 2.0 authentication method for CLI applications. User enters a code at a URL rather than being redirected |
| **TUI** | Text User Interface — interactive terminal application (vs GUI). Built with Bubble Tea |
| **Repository Pattern** | Design pattern that abstracts data access behind interfaces, enabling database swaps without changing business logic |
| **Service Layer** | Layer containing business logic, sitting between handlers (HTTP) and repositories (data) |
| **WAL Mode** | Write-Ahead Logging — SQLite configuration enabling concurrent reads while writing |
| **JWT** | JSON Web Token — stateless authentication token containing encoded claims |
| **Monorepo** | Single repository containing multiple related projects/packages (CLI + server in this case) |

## UNCONF CLI Commands

| Command | Purpose |
|---------|---------|
| `unconf login` | Authenticate with GitHub |
| `unconf logout` | Clear stored credentials |
| `unconf list` | Show available conferences |
| `unconf info <slug>` | Show conference details |
| `unconf checkout <slug>` | Set active conference context |
| `unconf rooms` | Launch room explorer TUI |
| `unconf book <room>` | Book a room directly |
| `unconf status` | Show current booking |
| `unconf cancel` | Cancel booking |
| `unconf invite <user>` | Send roommate request |
| `unconf requests` | Manage roommate requests |
| `unconf attendees` | List conference attendees |
| `unconf config` | Update profile settings |
| `unconf dashboard` | Organizer dashboard TUI |
| `unconf confirm <id>` | Confirm booking (organizer) |
| `unconf export` | Export CSV (organizer) |

## API Status Codes

| Code | Meaning | When Used |
|------|---------|-----------|
| 200 | OK | Successful GET, PUT, DELETE |
| 201 | Created | Successful POST (resource created) |
| 202 | Accepted | Auth polling (pending authorization) |
| 400 | Bad Request | Validation error, business rule violation |
| 401 | Unauthorized | Missing or invalid token |
| 403 | Forbidden | Valid token but insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 500 | Internal Server Error | Unexpected server error |

## Acronyms

| Acronym | Expansion |
|---------|-----------|
| **API** | Application Programming Interface |
| **CLI** | Command-Line Interface |
| **CORS** | Cross-Origin Resource Sharing |
| **CRUD** | Create, Read, Update, Delete |
| **DX** | Developer Experience |
| **MVP** | Minimum Viable Product |
| **NFR** | Non-Functional Requirement |
| **OAuth** | Open Authorization |
| **REST** | Representational State Transfer |
| **SMTP** | Simple Mail Transfer Protocol |
| **SQL** | Structured Query Language |
| **TLS** | Transport Layer Security (HTTPS) |
| **TUI** | Text User Interface |
| **UX** | User Experience |

---

*Document created by Winston (Architect) using BMAD Method*  
*Based on: docs/prd.md, docs/brief.md*
