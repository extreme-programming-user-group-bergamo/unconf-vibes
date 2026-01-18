<!--
Sync Impact Report - Constitution Update
Version: 1.0.0 → 1.1.0 (CLI architecture update)
Ratification Date: 2025-01-17
Last Amended: 2025-01-17

Changes:
- Updated tech stack from React/Node.js to Go CLI + Go Backend
- Added CLI-First Principle (VI) for Unix-like command design
- Updated Security section for CLI-specific session handling
- Updated Performance Standards for TUI responsiveness
- Aligned with MVP.md command specifications

Templates Status:
✅ .specify/templates/plan-template.md - Reviewed, constitution check section compatible
✅ .specify/templates/spec-template.md - Reviewed, requirement sections align with principles
✅ .specify/templates/tasks-template.md - Reviewed, task ordering supports TDD principle

Follow-up Items: None
-->

# UNCONF Platform Constitution

## Core Principles

### I. Repository Pattern (NON-NEGOTIABLE)

The database layer MUST be abstracted through repository interfaces to enable seamless
migration from SQLite to PostgreSQL without requiring changes to business logic.

**Rules**:

- All database access MUST go through repository interfaces
- No direct SQL queries in business logic or service layers
- Repository implementations MUST be swappable via dependency injection
- Data access patterns MUST be database-agnostic

**Rationale**: The project starts with SQLite for simplicity but must support PostgreSQL
in production. Repository pattern ensures business logic remains unchanged during
migration, reducing risk and enabling incremental transition.

### II. Privacy by Design (NON-NEGOTIABLE)

All features involving participant visibility MUST respect individual privacy controls.
Every participant MUST have explicit control over their name visibility in shared contexts.

**Rules**:

- Privacy settings MUST default to private (opt-in for visibility)
- Room occupancy displays MUST honor visibility preferences
- No participant data MUST be exposed without explicit consent
- Privacy controls MUST be granular and per-context (e.g., room selection vs. participant list)
- CLI `--private` flag MUST override default visibility settings

**Rationale**: Conference attendees have varying comfort levels with social visibility.
Privacy-first design respects individual preferences while enabling optional social
features, aligning with GDPR principles and community values.

### III. Real-time State Consistency

Room availability and occupancy MUST reflect the current system state without manual refresh.

**Rules**:

- TUI MUST update automatically when room state changes
- No stale data MUST be displayed to users making decisions
- Concurrent booking conflicts MUST be prevented or detected
- State synchronization MUST occur within 2 seconds of change
- CLI commands MUST fetch fresh data on each invocation

**Rationale**: Multiple users selecting rooms simultaneously require real-time updates
to prevent double-bookings and ensure informed decisions about roommate selection.

### IV. Email Automation with Structured Data

Hotel notifications MUST be automated with complete, structured booking data sent via
configurable email templates.

**Rules**:

- All booking confirmations MUST trigger automated hotel notification
- Email templates MUST be configurable by organizers
- Hotel emails MUST contain all required booking fields (dates, room type, participant
  details, special requests)
- Email delivery failures MUST be logged and retried

**Rationale**: Manual coordination creates errors and delays. Automated structured emails
reduce organizer workload, improve hotel accuracy, and provide audit trail for bookings.

### V. Test-First Development (NON-NEGOTIABLE)

All features MUST follow Test-Driven Development (TDD). Tests MUST be written and
approved before implementation begins.

**Rules**:

- Write contract/integration tests first
- User approval of tests required before implementation
- Tests MUST fail initially (red phase)
- Implement until tests pass (green phase)
- Refactor with passing tests as safety net
- Red-Green-Refactor cycle strictly enforced

**Rationale**: TDD ensures requirements are testable, drives better design through
test-first thinking, and provides regression safety. User-approved tests validate
shared understanding before implementation investment.

### VI. CLI-First Design

The application MUST follow Unix-like CLI conventions for intuitive developer experience.

**Rules**:

- Command syntax MUST follow `unconf <command> [subcommand] [options]` pattern
- Context system MUST reduce repetition (like `kubectl` or `git`)
- Commands MUST support both interactive (TUI) and scripting modes
- **All interactive elements and TUI views MUST be built using the Charm Bracelet stack (Bubble Tea, Bubbles, Lip Gloss) and follow the high-contrast, polished "Crush" aesthetic.**
- Output MUST support `--json` flag for machine-readable format
- Exit codes MUST follow Unix conventions (0 = success, non-zero = error)

**Rationale**: Target users are developers who expect familiar CLI patterns. Unix-like
design reduces learning curve and enables integration with shell scripts and automation.

## Security & Privacy Requirements

### GDPR Compliance

- User data MUST be stored with consent tracking
- Data retention periods MUST be configurable (default: 2 years after last participation)
- Users MUST be able to export their data (right to portability)
- Users MUST be able to request deletion (right to erasure)
- Privacy policy MUST be displayed before registration

### Authentication & Session Security

- Magic links MUST use 256-bit cryptographically secure tokens
- Magic link tokens MUST expire after 15 minutes and be single-use
- Session tokens MUST expire after 30 minutes of inactivity
- Session credentials MUST be stored securely in `~/.unconf.yaml` (file permissions 600)
- Rate limiting MUST protect authentication endpoints (3 requests/email/hour)

## Development Workflow

### Test Gates

- Contract tests MUST pass before API implementation
- Integration tests MUST pass before feature completion
- No code review without passing tests
- No deployment without full test suite green

### Code Review Requirements

- All PRs MUST be reviewed before merge
- Repository pattern violations MUST be rejected
- Privacy control bypasses MUST be rejected
- Test coverage MUST not decrease

### Performance Standards

- Room availability queries MUST complete <200ms (p95)
- TUI MUST render at 60fps for smooth navigation
- Email sending MUST not block user confirmation flow (async)
- Database migrations MUST support zero-downtime deployment

## Governance

### Amendment Process

1. Proposed changes MUST be documented with rationale
2. Constitution amendments require explicit approval
3. Breaking changes require migration plan
4. All dependent templates MUST be updated to reflect changes

### Versioning Policy

- **MAJOR**: Principle removal, backward-incompatible changes
- **MINOR**: New principles added, material expansions
- **PATCH**: Clarifications, wording improvements, typo fixes

### Compliance Review

- All PRs MUST verify compliance with constitution
- Complexity deviations MUST be justified in plan documentation
- Use CLAUDE.md for runtime development guidance aligned with these principles

**Version**: 1.1.0 | **Ratified**: 2025-01-17 | **Last Amended**: 2025-01-17
