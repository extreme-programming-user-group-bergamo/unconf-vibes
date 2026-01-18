# Data Model: "Charming" TUI Homepage

## Overview
The Homepage TUI relies on an internal state machine to manage navigation, connectivity, and dashboard content.

## Key Entities

### TUIState
Represents the global state of the CLI application during an interactive session.

- **activeView**: Enum { `HOME`, `LIST`, `ROOMS`, `CONFIG` }
- **context**: The currently active conference ID (if any).
- **connectivity**: Enum { `CONNECTED`, `OFFLINE`, `SYNCING` }
- **userProfile**: Cached local user configuration (name, email, privacy).
- **lastSync**: Timestamp of the last successful backend health check.
- **commandInput**: String (current contents of the `/` prompt).

### DashboardWidget
Represents a visual panel on the homepage dashboard.

- **title**: String (e.g., "Active Conference", "My Status").
- **content**: List of strings or formatted text.
- **type**: Enum { `INFO`, `ACTION_LIST`, `STATUS` }.
- **style**: Lip Gloss style definition.

## State Transitions

1. **App Launch**: 
    - `Initialize` -> Load Config -> Check Backend -> Set `activeView` to `HOME`.
2. **Navigation**: 
    - User types `/list` -> `activeView` changes to `LIST` -> Model swapping occurs.
3. **Health Check**:
    - Background Tick -> HTTP Ping -> Update `connectivity` and `lastSync`.
4. **Resize**:
    - `WindowSizeMsg` -> Recalculate Widget widths/heights.
