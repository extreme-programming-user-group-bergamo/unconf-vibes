# Copilot IDE Instructions

These instructions govern how the AI assistant interacts with the user inside the IDE.
For project architecture, coding rules, and conventions see `AGENTS.md`.

## Rules

1. When the assistant needs questions, menus, or clarification from the user, the assistant shall use the `user-input` MCP tools
2. If an MCP tool call fails, the assistant shall retry it exactly once before proceeding without it.

## Agent Behavior Rules

3. The system shall match existing code style, conventions, and types found in the target package.
4. The system shall not create custom scripts to edit files, nor add markdown files unless explicitly instructed.
5. If summarizing context, the system shall include instruction to re-read the agent persona file and `.bmad-core/core-config.yaml`.
