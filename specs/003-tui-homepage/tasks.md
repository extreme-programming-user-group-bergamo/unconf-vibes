# Tasks: "Charming" TUI Homepage

**Input**: Design documents from `specs/003-tui-homepage/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create TUI component directory structure in `internal/cli/ui/`
- [ ] T002 Refactor conference fetching in `internal/cli/list.go` to be reusable by TUI models

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure for the modular TUI engine and shared UI components

**⚠️ CRITICAL**: This phase must be completed before any user story implementation

- [ ] T003 Define `TUIState` and `DashboardWidget` types in `internal/cli/tui_types.go`
- [ ] T004 [P] Implement Status Bar component in `internal/cli/ui/status.go`
- [ ] T005 [P] Implement Command Bar component with `/` prompt in `internal/cli/ui/command.go`
- [ ] T006 Implement the main TUI engine (model swapper) in `internal/cli/tui_root.go`

**Checkpoint**: Foundation ready - homepage and navigation implementation can now begin

---

## Phase 3: User Story 1 - Enter the Dashboard (Priority: P1) 🎯 MVP

**Goal**: Launch `unconf` and see the "Charming" dashboard with current status and conference context.

**Independent Test**: Run `./unconf` and verify the full-screen dashboard appears with Header, Main Content Area, and Status Bar.

### Tests for User Story 1 (MANDATORY per Constitution Principle V)

- [ ] T007 [US1] Write failing TUI tests for Dashboard initialization and widget rendering in `internal/cli/tui_home_test.go`
- [ ] T008 [US1] Write failing test for "Proactive Discovery" state when no context exists in `internal/cli/tui_home_test.go`

### Implementation for User Story 1

- [ ] T009 [US1] Implement multi-column dashboard view with **Active Context widget (FR-004)** in `internal/cli/tui_home.go`
- [ ] T010 [P] [US1] Create dashboard widget components using Lip Gloss in `internal/cli/ui/widgets.go`
- [ ] T011 [US1] Integrate dashboard view as the default model in `internal/cli/tui_root.go`
- [ ] T012 [US1] Update `internal/cli/root.go` to launch the homepage TUI when no arguments are provided
- [ ] T013 [US1] Implement "Proactive Discovery" view logic for the homepage in `internal/cli/tui_home.go`

**Checkpoint**: User Story 1 is fully functional as the application's interactive entry point.

---

## Phase 4: User Story 2 - Navigate via Command Bar (Priority: P2)

**Goal**: Navigate between different views (Home, List, Rooms) using `/` commands.

**Independent Test**: From the homepage, type `/list` and verify a seamless transition to the conference list.

### Tests for User Story 2 (MANDATORY per Constitution Principle V)

- [ ] T014 [US2] Write failing tests for `/` command parsing and view transition logic in `internal/cli/ui/command_test.go`
- [ ] T015 [US2] Write failing test for "Go Back" (`Esc` / `/home`) navigation in `internal/cli/tui_root_test.go`

### Implementation for User Story 2

- [ ] T016 [US2] Refactor `internal/cli/tui_list.go` to implement the sub-model interface for navigation
- [ ] T017 [P] [US2] Implement slash command parsing for `/list`, `/rooms`, `/config`, and `/quit` in `internal/cli/ui/command.go`
- [ ] T018 [US2] Implement view transition and model swapping logic in `internal/cli/tui_root.go`
- [ ] T019 [US2] Implement "Go Back" functionality to return to the dashboard in `internal/cli/tui_root.go`

**Checkpoint**: Navigation between views is seamless and controlled via the command bar.

---

## Phase 5: User Story 3 - Real-time Status Monitoring (Priority: P3)

**Goal**: Persistent status bar reflecting system health and network connectivity.

**Independent Test**: Run the app and verify the status bar updates its "Connected/Offline" indicator based on backend availability.

### Tests for User Story 3 (MANDATORY per Constitution Principle V)

- [ ] T020 [US3] Write failing tests for background health check ticker and state updates in `internal/cli/tui_root_test.go`

### Implementation for User Story 3

- [ ] T021 [US3] Implement non-blocking background health check using `tea.Tick` in `internal/cli/tui_root.go`
- [ ] T022 [US3] Connect health check state to the Status Bar UI in `internal/cli/ui/status.go`
- [ ] T023 [US3] Add "Last Sync" timestamp to the Status Bar in `internal/cli/ui/status.go`

**Checkpoint**: Connectivity status is monitored in real-time without blocking the UI.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Responsive design, error handling, and performance optimization.

- [ ] T024 [P] Handle `tea.WindowSizeMsg` across all models for responsive layout in `internal/cli/tui_root.go`
- [ ] T025 [P] Implement adaptive colors for dark/light terminal backgrounds in `internal/cli/ui/` components
- [ ] T026 Implement graceful error display for corrupted local configuration in `internal/cli/tui_home.go`
- [ ] T027 Verify performance against SC-001 (init < 300ms), SC-003 (transition < 100ms), and **SC-004 (Status bar sync < 2s)**
- [ ] T028 Run final validation against `quickstart.md` scenarios
