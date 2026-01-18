# Research: "Charming" TUI Homepage

## Overview
This document consolidates research and technical decisions for the UNCONF homepage TUI, focusing on the "Crush" aesthetic and the Charm Bracelet stack.

## Decision 1: Layout Management with Lip Gloss
- **Decision**: Use `lipgloss.JoinHorizontal` and `lipgloss.JoinVertical` with structured `Place` and `Width/Height` constraints.
- **Rationale**: Lip Gloss is the standard for styling in the Charm ecosystem. To achieve the "Crush" aesthetic (multi-column widgets), we need precise control over padding, borders, and alignment.
- **Alternatives Considered**: 
    - `table` component from Bubbles: Too rigid for a dashboard overview.
    - Raw ANSI strings: Unmaintainable and lacks adaptive color support.

## Decision 2: Single-Session Navigation (Model Swapping)
- **Decision**: Implement a root `mainModel` that holds a `currentView` interface. The `Update` function will delegate to the active model.
- **Rationale**: Ensures "visually seamless" transitions (SC-003) without terminal flickering or losing app state.
- **Alternatives Considered**: 
    - `os.Exec` to sub-commands: Causes screen flickering and loses the "integrated app" feel.

## Decision 3: Slash Command System
- **Decision**: Use a dedicated `textinput` component from Bubbles for the Command Bar, triggered by `/`.
- **Rationale**: Familiar to users of modern apps (Crush, Slack, Discord). Allows for extensible command parsing without cluttering the UI with buttons.
- **Implementation**: The Command Bar will be a persistent component at the bottom, just above the status bar.

## Decision 4: Non-blocking Backend Health Checks
- **Decision**: Use `tea.Tick` and `tea.Cmd` to perform background pings to the backend.
- **Rationale**: Required by FR-008 to ensure the Status Bar (SC-004) stays updated without blocking user interaction.
- **Alternatives Considered**: 
    - Synchronous checks on every `Update`: Would cause visible lag in the TUI (violating SC-003).

## Decision 5: Adaptive Rendering for Resize
- **Decision**: Handle `tea.WindowSizeMsg` in all sub-models and use Lip Gloss to re-calculate layout dimensions dynamically.
- **Rationale**: Mandatory for FR-007 and SC-005. Ensures the TUI remains "Charming" on any terminal size.
