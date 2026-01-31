# User Interface Design Goals

## Overall UX Vision

UNCONF delivers a **"functional delight"** experience — satisfying to use without being flashy. The interface follows developer tool conventions (like `kubectl` and `git`), feeling immediately familiar to the target audience. First-timers are guided through a welcoming wizard experience, while power users get fast, scriptable CLI commands. The guiding principle: **the fun happens at the conference, not in the CLI**.

## Key Interaction Paradigms

1. **TUI for Discovery, CLI for Action** — Interactive TUI (Bubble Tea) for exploration (rooms, attendees, conference info) with direct CLI commands for rapid execution (`unconf book 204`)

2. **Context-Based Workflow** — Users `checkout` a conference context, and subsequent commands operate within that context (similar to git branches or kubectl contexts)

3. **Progressive Disclosure** — Information revealed as needed; defaults for common cases with advanced options available but not overwhelming

4. **Request-Based Social Interactions** — Roommate selection uses an explicit request/accept flow, giving control to both parties

## Core Screens and Views

| Screen | Purpose |
|--------|---------|
| **Conference List** | Browse/search available unconferences, see attendee counts |
| **Conference Info** | View dates, location, pricing, capacity, current attendees |
| **Room Explorer (TUI)** | Interactive room selection with occupancy, pricing, filtering |
| **Booking Wizard** | Step-by-step guided booking for first-timers |
| **Status View** | Current booking details, pending requests |
| **Requests View** | Manage incoming/outgoing roommate requests |
| **Organizer Dashboard** | Registration overview, attendee preferences, capacity |

## Accessibility

**None** (MVP) — Standard terminal accessibility. Screen reader support and WCAG compliance identified as areas needing further research.

## Branding

- **Aesthetic**: Clean, minimal terminal UI with thoughtful use of color
- **Tone**: Warm, welcoming, direct language — not intimidating
- **Visual Framework**: Lip Gloss (Charm) for styling
- **Constraint**: Must work on terminals with limited color support

## Target Devices and Platforms

**Cross-Platform CLI**:
- macOS (iTerm2, Terminal.app)
- Linux (standard TTYs)
- Windows (Windows Terminal)

No mobile or web interface planned — CLI-first tool for developers.

---
