# Project Brief: UNCONF CLI

> "Choose your room, choose your roommate, focus on the conference."

**Version:** 1.0  
**Date:** January 18, 2026  
**Status:** Draft  

---

## Executive Summary

**UNCONF** is a command-line application with a text-based user interface (TUI) designed to transform conference registration and hotel room booking into a fast, social, and developer-friendly experience.

**Problem:** Open Space conferences like SoCraTes Italia and Polenta & Deploy require complex coordination between attendees, organizers, and hotels. The current process involves email chains, shared spreadsheets, and manual hotel communication — creating friction for attendees and significant overhead for organizers.

**Solution:** A CLI tool that lets developers register for conferences and book hotel rooms in under 3 minutes, with social discovery features that let them see who's attending and choose their roommates — while maintaining privacy controls and automating hotel communication.

**Target Market:** Developer communities organizing unconferences and Open Space events (initially Italian tech community, expandable to global developer conferences).

**Value Proposition:** Replace bureaucratic registration workflows with a "nerd-delightful" experience that embodies the collaborative spirit of unconferences.

---

## Problem Statement

### Current State

Developer conferences following the Open Space or unconference format face a unique challenge: hotel room booking is part of the social experience. Attendees want to:
- Know who else is attending
- Choose roommates based on existing friendships or networking goals
- Have transparency on room availability and pricing

### Pain Points

**For Attendees:**
- Multiple email exchanges to understand room options
- No visibility into who's already registered or rooming with whom
- Manual tracking of booking status
- Lack of control over privacy (name visibility)

**For Organizers:**
- Managing a shared spreadsheet with room assignments
- Manually forwarding booking details to hotels
- Tracking dietary restrictions and accessibility needs separately
- Answering repetitive questions about availability

**For Hotels:**
- Receiving unstructured booking requests via email
- Risk of data transcription errors
- No standardized format for special requests

### Impact

- Registration process takes 15-30 minutes instead of <3 minutes
- Organizers spend 5-10 hours per event on administrative coordination
- Some attendees choose single rooms to avoid awkward roommate coordination
- Information silos reduce the community-building potential of the event

### Why Now

The Italian developer unconference community is growing (SoCraTes Italia, Polenta & Deploy, XP Days). A tool built by the community for the community can become the standard — and potentially expand to global Open Space events.

---

## Proposed Solution

### Core Concept

UNCONF is a **CLI-first application** with an optional **TUI (Terminal User Interface)** for exploration. It follows the developer tool conventions of `kubectl` and `git`, using a context-based workflow.

```
unconf checkout socrates-26     # Set context
unconf rooms                     # Explore (TUI)
unconf book 204                  # Book (CLI)
```

### Key Differentiators

1. **Social Room Selection** — See who's in which room, choose your roommate
2. **Privacy by Design** — Attendees control their visibility (Private/Public)
3. **Hotel Automation** — Structured email templates sent directly to hotels
4. **CLI + TUI Hybrid** — TUI for discovery, CLI for rapid action
5. **Open Space Spirit** — Reflects transparency and collaboration values

### Why This Will Succeed

- **Community alignment**: Built for developers by developers, using familiar tool patterns
- **Progressive complexity**: Simple for first-timers, powerful for power users
- **No lock-in**: Hotels still receive email; no complex integrations required
- **Scalable model**: Can expand to any conference that needs room coordination

---

## Target Users

### Primary User Segment: Conference Attendees

**Profile:**
- Software developers, ages 25-50
- Comfortable with terminal/CLI tools (but not necessarily power users)
- Attend 1-3 unconferences per year
- Value community connection and networking

**Behaviors:**
- Prefer self-service over email back-and-forth
- Check who's attending before committing
- Often attend with friends or colleagues

**Needs:**
- Fast registration with minimal friction
- Transparency on room availability and pricing
- Choice in roommate selection
- Control over personal information visibility

**Sub-segments:**
- **First-Timers**: Need guidance and reassurance about the process
- **Returning Attendees**: Want shortcuts and familiar faces
- **Privacy-Conscious**: Need granular control over visibility

### Secondary User Segment: Conference Organizers

**Profile:**
- Community volunteers or professional event organizers
- Manage 1-3 events per year
- Have limited time for administrative tasks

**Behaviors:**
- Currently use spreadsheets and email
- Need to coordinate with hotels
- Track dietary and accessibility requirements

**Needs:**
- Dashboard to monitor registrations
- Automated hotel communication
- Easy data export for logistics
- Capacity management tools

### Tertiary User Segment: Hotel Partners

**Profile:**
- Hotel reservation staff
- Receive group bookings via email
- Need structured data to process reservations

**Needs:**
- Clear, structured booking information
- Standard format for special requests
- Reliable guest data (no transcription errors)

---

## Goals & Success Metrics

### Business Objectives

- **Adoption**: 80% of target conference attendees use UNCONF for registration (within 6 months of launch at first event)
- **Efficiency**: Reduce organizer administrative time by 90%
- **Satisfaction**: Achieve >4.5/5 user satisfaction rating
- **Expansion**: Deploy at 3+ Italian unconferences within first year

### User Success Metrics

- **Registration time**: <3 minutes from start to confirmation
- **Social discovery usage**: 80% of users view attendee list before booking
- **Privacy adoption**: >50% of users customize their privacy settings
- **Repeat usage**: 70% retention at subsequent events

### Key Performance Indicators (KPIs)

- **Time to first booking**: Minutes from account creation to confirmed booking
- **Error rate**: Hotel communication errors requiring manual correction
- **Support requests**: Volume of questions organizers receive about registration
- **Feature adoption**: % of users using TUI vs CLI-only paths

---

## MVP Scope

### Core Features (Must Have)

#### Attendee Features
- **`login`**: GitHub OAuth authentication
- **`config`**: User profile setup (name, email, privacy preference)
- **`list`**: Browse available conferences with search/filter
- **`checkout`**: Set active conference context
- **`info`**: View conference details (dates, location, pricing)
- **`rooms`**: Interactive TUI for room exploration
  - View room types, prices, availability
  - See current occupants (where privacy permits)
  - Filter by type (single/double/triple)
- **`book`**: Reserve a room
  - Wizard mode for guided booking
  - Direct mode (`book 204`) for power users
  - Privacy flag (`--private`)
  - Notes for hotel (`--notes "dietary request"`)
  - Roommate request system (target person accepts/declines)
- **`status`**: View current booking status
- **`requests`**: View/manage incoming roommate requests
- **`cancel`**: Cancel existing booking

#### Organizer Features
- **Conference setup**: Create/configure conference (dates, rooms, capacity)
- **Attendee dashboard**: View registrations with preferences
- **Hotel email automation**: Templated emails on booking
- **CSV export**: Download registration data

#### Privacy & UX
- **Visibility tiers**: Public (name visible) / Private (anonymous)
- **Progressive disclosure**: Simple defaults, advanced options available
- **Friendly language**: Welcoming tone, not intimidating
- **Context indicator**: Always show current conference in TUI

#### Technical Infrastructure
- **Backend API**: Go + Gin, SQLite (Repository pattern)
- **CLI Framework**: Cobra + Viper
- **TUI**: Bubble Tea + Lip Gloss
- **Email**: SMTP with SendGrid/Mailgun option

### Out of Scope for MVP

- **Protected visibility tier** (middle tier between public/private) → V2
- **Calendar .ics file generation** → V2
- **Interest/Going states** (pre-registration signaling) → V2
- **Social proof indicators** (booking momentum) → V2
- **Fuzzy search for conferences** → V2
- **Invite link sharing** → V2
- **Terminal push notifications** → V2+
- **Offline mode** → V2+
- **Mobile app or web interface** → Not planned
- **Payment processing** (hotels handle payment directly) → Not planned

### MVP Success Criteria

The MVP is successful if:
1. A complete conference (SoCraTes Italia 2026) can be run using only UNCONF
2. Registration time averages <3 minutes
3. Zero hotel communication errors from the automated emails
4. Organizers report >80% reduction in email/spreadsheet work
5. >50% of attendees view the room map before booking

---

## Post-MVP Vision

### Phase 2 Features

- **Enhanced privacy**: Protected tier (visible to registered attendees only)
- **Calendar integration**: .ics file download on booking confirmation
- **Interest signals**: "Interested" / "Going" states before room booking opens
- **Social proof**: "5 people booked today", occupancy percentages
- **Expert CLI mode**: `unconf book socrates-26 --room 204 --roommate @marco`
- **Invite links**: Share codes for quick onboarding
- **Context safety**: Confirmation dialogs for destructive actions

### Long-term Vision (1-2 Years)

- **Multi-language support**: Italian, English, German, Spanish
- **Historical data platform**: Help organizers plan based on past events
- **Conference templates**: Clone settings from previous editions
- **Third-party integrations**: Eventbrite import, Slack notifications
- **Web dashboard**: For non-CLI users and analytics

### Expansion Opportunities

- **Global unconferences**: SoCraTes Germany, SoCraTes UK, etc.
- **Tech meetups**: Expand beyond conferences to recurring events
- **Corporate offsites**: Team retreats with room coordination needs
- **White-label version**: Customizable for specific communities

---

## Technical Considerations

### Platform Requirements

- **Target Platforms**: macOS, Linux, Windows (cross-compiled)
- **Terminal Support**: Works on standard terminals (iTerm2, Terminal.app, Windows Terminal, Linux TTYs)
- **Performance Requirements**: TUI renders at 60fps, API responses <200ms

### Technology Preferences

- **CLI/TUI**: Go 1.21+, Cobra, Viper, Bubble Tea, Lip Gloss
- **Backend**: Go + Gin (REST API), containerized with Docker
- **Database**: SQLite initially (Repository pattern for future PostgreSQL migration)
- **Email**: SMTP standard, SendGrid/Mailgun for production, MailHog for testing

### Architecture Considerations

- **Repository Structure**: Monorepo with `/cmd`, `/internal`, `/pkg` layout
- **Service Architecture**: CLI as API client, stateless backend
- **Authentication**: GitHub OAuth (device flow for CLI)
- **Integration Requirements**: GitHub OAuth API, Email SMTP
- **Security**: API authentication (token-based), HTTPS for all communication

### CI/CD

- **Build**: GoReleaser with GitHub Actions
- **Test**: `go test` + `golangci-lint` on every PR
- **Release**: Git tags trigger cross-platform builds
- **Distribution**: GitHub Releases, optional Homebrew tap

---

## Constraints & Assumptions

### Constraints

- **Budget**: Community/volunteer project (no paid services beyond minimal hosting)
- **Timeline**: MVP ready for SoCraTes Italia 2026 (target: Q2 2026)
- **Resources**: Small team (1-3 developers), asynchronous collaboration
- **Technical**: Must work on terminals with limited color support

### Key Assumptions

- Conference organizers will adopt a CLI tool (validated by community interest)
- Attendees are comfortable running terminal commands
- Hotels will accept structured email for bookings (no API integration required)
- The Italian unconference community will pilot the first deployment
- Privacy controls are sufficient with two tiers (public/private) for MVP

---

## Risks & Open Questions

### Key Risks

- **Adoption risk**: Non-technical attendees may prefer web forms
  - *Mitigation*: Provide excellent first-timer guidance; consider web UI in future
- **Hotel resistance**: Hotels may prefer their existing booking systems
  - *Mitigation*: Structure emails to match their expected format; keep them in control of confirmations
- **Scope creep**: Temptation to add "just one more feature"
  - *Mitigation*: Strict MVP definition; defer all V2 items
- **Community fragmentation**: Different conferences want different features
  - *Mitigation*: Configurable settings per conference; core features work for all

### Open Questions

*All open questions have been resolved — see Resolved Decisions below.*

### Resolved Decisions

| Question | Decision | Rationale |
|----------|----------|-----------|
| Roommate preference conflicts | **Request system** — Target person receives request and decides to accept/decline | Gives control to the person being requested; avoids race conditions |
| Waiting list when full | **No waiting list for MVP** — Users see "Sold Out" and check back manually | Keeps MVP simple; can add waiting list in V2 if demand exists |
| Email verification | **GitHub OAuth login** — Authenticate via GitHub | Developer-friendly, verified identity, no fake signups, on-brand for CLI tool |
| Roommate cancellation | **Room reserved by remaining person** — They can find replacement, stay solo, or cancel | Puts control in user's hands; no automatic stranger assignment |
| No-show billing | **Out of scope** — UNCONF facilitates booking; payment is between attendee and hotel | Keeps scope focused; avoids payment complexity |

### Areas Needing Further Research

- User testing with actual unconference attendees (both technical and less technical)
- Hotel partner interviews to understand their workflow
- Legal/GDPR considerations for storing attendee data
- Accessibility review of TUI interactions (screen readers, keyboard-only)

---

## Appendices

### A. Research Summary

**Inputs to this brief:**
- [README.md](../README.md) — Initial concept and architecture overview
- [MVP.md](../MVP.md) — Detailed functional specifications
- [Brainstorming Session](brainstorming-session-results.md) — UX exploration (34 ideas, 5 design principles)

**Key insights from brainstorming:**
1. TUI for discovery, CLI for action — hybrid approach
2. Progressive disclosure prevents overwhelming first-timers
3. Privacy controls must be upfront, not buried
4. Social proof builds excitement for the event
5. Organizer tools are essential, not optional

### B. Design Principles

From the brainstorming session, these principles guide all UX decisions:

1. **Progressive Disclosure** — Reveal info as needed
2. **TUI for Discovery, CLI for Action** — Best of both worlds
3. **Privacy by Design** — Visibility controls upfront
4. **Functional Delight** — Satisfying to use, not flashy
5. **Magnify Community, Minimize Friction** — Social + efficient

### C. References

- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Bubble Tea TUI](https://github.com/charmbracelet/bubbletea)
- [SoCraTes Conference Format](https://www.socrates-conference.de/)
- [Open Space Technology](https://openspaceworld.org/)

---

## Next Steps

### Immediate Actions

1. **Review this brief** with core team / community stakeholders
2. **Validate assumptions** with 2-3 unconference organizers
3. **Prioritize P1 vs P2 features** for MVP development
4. **Create technical architecture document** from this brief
5. **Set up project repository** with initial Go module structure

### PM Handoff

This Project Brief provides the full context for UNCONF CLI. Please start in 'PRD Generation Mode', review the brief thoroughly to work with the user to create the PRD section by section as the template indicates, asking for any necessary clarification or suggesting improvements.

---

*Document created by Mary (Business Analyst) using BMAD Method*  
*Source: Brainstorming session + MVP.md + README.md*
