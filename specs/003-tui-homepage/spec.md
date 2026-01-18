# Feature Specification: "Charming" TUI Homepage

**Feature Branch**: `003-tui-homepage`
**Created**: 2026-01-17
**Status**: Draft
**Input**: User description: "implement the 'homepage' of the 'unconf' app, so that I run 'unconf' and I enter a TUI that implements command, status bar, etc. Take 'Crush' app as a reference"

## Clarifications

### Session 2026-01-17

- Q: When navigating to sub-views, should we use a single TUI session or execute sub-processes? → A: Single-TUI Session (Swap internal models).
- Q: How should the user interact with the Command Bar? → A: Command Prompt using / (slash) commands (e.g., /list).
- Q: How should the main area be partitioned when a conference is active? → A: Dashboard Widgets (Multi-column layout with boxes/panels).
- Q: What should be the primary visual content when no conference context is active? → A: Proactive Discovery (Display a list of upcoming/open conferences in the main widget area).
- Q: Should the command bar be persistently visible? → A: Persistent (Always visible at the bottom of the screen, above the status bar).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Enter the Dashboard (Priority: P1)

As a developer, I want to launch `unconf` and immediately see my current status and active conference context in a polished dashboard so that I can quickly decide my next action.

**Why this priority**: This is the entry point of the application and defines the primary user experience.

**Independent Test**: Can be tested by running `unconf` without arguments and verifying that a full-screen TUI appears with the current user config and active context.

**Acceptance Scenarios**:

1. **Given** the user has a valid configuration and an active context, **When** I run `unconf`, **Then** I see a "Charming" homepage showing the conference name, my name, and my booking status.
2. **Given** the user has NO active context, **When** I run `unconf`, **Then** the homepage displays a proactive discovery view in the main widget area showing available/upcoming conferences with a call to action to select one.

---

### User Story 2 - Navigate via Command Bar (Priority: P2)

As a user, I want to navigate between different sections (Conferences, Rooms, Config) using a command bar or keyboard shortcuts so that I can perform tasks without typing full CLI commands.

**Why this priority**: Improves usability and provides the "Crush-like" integrated experience.

**Independent Test**: Can be tested by interacting with the TUI menu/command bar and verifying that it triggers the correct sub-views or actions.

**Acceptance Scenarios**:

1. **Given** I am on the homepage, **When** I select "List Conferences" from the command bar, **Then** I am transitioned to the conference list view.
2. **Given** I am on the homepage, **When** I press a dedicated shortcut (e.g., 'r'), **Then** the app switches to the room selection view (if a context is active).

---

### User Story 3 - Real-time Status Monitoring (Priority: P3)

As a user, I want to see a persistent status bar that reflects system health and network connectivity so that I am aware of the backend status.

**Why this priority**: Enhances the "pro" feel and provides useful feedback for troubleshooting.

**Independent Test**: Can be tested by observing the status bar update when the backend goes offline or online.

**Acceptance Scenarios**:

1. **Given** the backend is healthy, **When** I view the status bar, **Then** it shows a green "Connected" indicator.
2. **Given** the backend is unreachable, **When** I view the status bar, **Then** it shows a red "Offline" indicator with the last sync time.

### Edge Cases

- What happens when the terminal is resized to a very small dimension? (TUI must adapt or show a "Terminal too small" warning).
- How does the homepage handle long conference names or long usernames? (Text should be truncated gracefully or wrapped according to "Crush" style).
- What happens if the local configuration file is corrupted? (App should show an error state in the TUI with an option to re-configure).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST launch a full-screen interactive TUI (Bubble Tea) when `unconf` is executed without subcommands.
- **FR-002**: System MUST implement a "Crush-inspired" layout featuring a Header (App name/Version), a Main Content Area with a widget-based multi-column layout (boxes/panels), and a persistent Status Bar at the bottom.
- **FR-003**: System MUST provide a persistent "Command Bar" positioned above the status bar, featuring a `/` (slash) command prompt (e.g., `/list`, `/rooms`) to navigate and trigger actions via internal model swapping.
- **FR-004**: The Homepage MUST display the current "active context" (active conference ID and name) prominently.
- **FR-005**: The Status Bar MUST display the current user's email and their privacy setting (Public/Private).
- **FR-006**: The TUI MUST follow the **"Crush" aesthetic**: high-contrast colors (Lip Gloss), structured borders, and clear typography.
- **FR-007**: System MUST detect terminal resize events and re-render the TUI to fit the new dimensions.
- **FR-008**: System MUST perform a non-blocking health check to the backend and update the status bar indicator.

### Key Entities

- **TUI Context**: Represents the current state of the interface (active view, selection, connectivity status).
- **User Profile**: Local user data (name, email, privacy) displayed in the status bar.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The homepage TUI initializes and renders the first frame in under 300ms on a standard terminal.
- **SC-002**: 100% of the UI layout is built using the Charm Bracelet stack (Bubble Tea, Lip Gloss).
- **SC-003**: Navigation between the homepage and any sub-view (like `list`) is visually seamless and occurs in under 100ms.
- **SC-004**: The status bar accurately reflects backend connectivity within 2 seconds of a state change.
- **SC-005**: All text elements are legible on both dark and light terminal backgrounds (proper use of Lip Gloss adaptive colors).