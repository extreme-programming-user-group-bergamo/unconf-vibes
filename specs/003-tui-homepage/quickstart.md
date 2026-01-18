# Quickstart: "Charming" TUI Homepage

## Launching the App
To enter the interactive TUI mode:

```bash
make build
./unconf
```

## Dashboard Overview
The homepage consists of three main areas:
1. **Header**: Displays "UNCONF" and the current version.
2. **Main Dashboard**: Multi-column widgets showing your active conference and current status.
3. **Command & Status Bar**: 
   - **Command Bar**: Type `/` followed by a command.
   - **Status Bar**: Shows your email, privacy status, and connectivity.

## Essential Commands
Type these into the command bar:
- `/list` (or `/ls`): Switch to the interactive conference list.
- `/rooms` (or `/map`): Open the room selection view (requires active context).
- `/config`: Open user settings.
- `/quit` (or `ctrl+c`): Exit the application.

## Development Hotkeys
- `r`: Force refresh dashboard data.
- `esc`: Close the command bar if active.
- `tab`: Cycle through dashboard widgets (future implementation).
