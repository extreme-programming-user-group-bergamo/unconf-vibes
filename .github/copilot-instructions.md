# Copilot IDE Instructions

These instructions govern how the AI assistant interacts with the user inside the IDE.
For project architecture, coding rules, and conventions see `AGENTS.md`.

## Rules

1. When the assistant needs questions, menus, or clarification from the user, the assistant shall use the `user-input` MCP tools (`user_elicitation` or `user_input`).
2. Before finalizing any output, the assistant shall ask for confirmation via `user_elicitation` and shall continue iterating until the user confirms.
3. If an MCP tool call fails, the assistant shall retry it exactly once before proceeding without it.

## Agent Behavior Rules

4. The system shall match existing code style, conventions, and types found in the target package.
5. The system shall not create custom scripts to edit files, nor add markdown files unless explicitly instructed.
6. If summarizing context, the system shall include instruction to re-read the agent persona file and `.bmad-core/core-config.yaml`.

## VERIFY before responding

1. Did I use `user-input` MCP for every question or clarification?
2. Did I confirm completion with the user via `user_elicitation` before finalizing?
3. If an MCP call failed, did I retry exactly once?
