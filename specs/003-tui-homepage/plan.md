# Implementation Plan: "Charming" TUI Homepage

**Branch**: `003-tui-homepage` | **Date**: 2026-01-17 | **Spec**: [specs/003-tui-homepage/spec.md](specs/003-tui-homepage/spec.md)

## Summary
Implement a polished, "Crush-inspired" homepage dashboard using the Charm Bracelet stack (Bubble Tea, Lip Gloss). This will serve as the central hub for the UNCONF CLI, integrating navigation via slash commands and real-time status monitoring.

## Technical Context

**Language/Version**: Go 1.24  
**Primary Dependencies**: `bubbletea`, `lipgloss`, `bubbles`, `cobra`, `resty`  
**Storage**: N/A (Local config via Viper)  
**Testing**: `go test`, Bubble Tea test helpers  
**Target Platform**: Terminal (Cross-platform)
**Project Type**: Single CLI Application  
**Performance Goals**: <300ms init, 60fps rendering  
**Constraints**: <100ms navigation transition, persistent status bar  
**Scale/Scope**: Main entry point for the application.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **VI. CLI-First Design**: PASS. Strictly uses Charm stack and follows Unix conventions.
- **III. Real-time State Consistency**: PASS. Status bar reflects real-time backend health.
- **V. Test-First Development**: PASS. Plan includes unit tests for TUI state transitions.

## Project Structure

### Documentation (this feature)

```text
specs/003-tui-homepage/
├── plan.md              # This file
├── research.md          # TUI architecture and layout research
├── data-model.md        # TUI state and widget entities
├── quickstart.md        # TUI interaction guide
├── contracts/           # N/A (Internal TUI communication)
└── tasks.md             # Implementation tasks
```

### Source Code (repository root)

```text
internal/
├── cli/
│   ├── root.go          # Launches the homepage TUI
│   ├── list.go          # Refactored as a sub-model
│   ├── tui_list.go      # Sub-model for list view
│   ├── tui_home.go      # New: Dashboard view
│   ├── tui_root.go      # New: Main TUI engine (model swapper)
│   └── ui/              # Shared UI components (status bar, command bar)
│       ├── status.go
│       └── command.go
```

**Structure Decision**: Extending `internal/cli` to support a modular TUI architecture. Each view (Home, List, Rooms) will be a separate Bubble Tea model that the root model can swap between.

## Complexity Tracking

*No violations identified.*