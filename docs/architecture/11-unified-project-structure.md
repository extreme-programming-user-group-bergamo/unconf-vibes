# 11. Unified Project Structure

```text
unconf/
├── .github/
│   ├── agents/                  # BMAD / Copilot agent definitions
│   ├── prompts/                 # Prompt bundles used during IDE workflows
│   ├── skills/                  # Repo-scoped skills
│   └── workflows/
│       ├── ci.yaml              # Test and lint on PR/push
│       ├── deploy.yaml          # Deployment workflow scaffold
│       ├── docker-smoke.yml     # Container smoke validation
│       └── release.yaml         # Tagged release pipeline
├── cmd/
│   ├── server/                  # API server binary entry point
│   └── unconf/                  # CLI binary entry point
├── internal/
│   ├── api/                     # Gin router, middleware, handlers, responses
│   ├── auth/                    # GitHub device flow, PASETO, token storage
│   ├── cli/                     # Cobra commands and interactive prompts
│   ├── client/                  # REST clients used by CLI commands
│   ├── config/                  # App config and active-conference context
│   ├── email/                   # Email providers, rendering, mapping
│   ├── models/                  # Shared domain models
│   ├── repository/
│   │   └── sqlite/              # SQLite repositories and DB helpers
│   ├── service/                 # Business logic layer
│   └── tui/                     # Bubble Tea room, wizard, and dashboard UIs
├── migrations/                  # Embedded schema migrations
├── assets/                      # Logos and static image assets
├── bin/                         # Built binaries
├── configs/                     # Reserved for config files (currently empty)
├── docs/                        # PRD, architecture, QA, story docs
├── Dockerfile
├── Makefile
├── MVP.md
├── README.md
├── go.mod
└── AGENTS.md
```

- The current repository uses `/internal` rather than a split `/internal` plus `/pkg` layout.
- Runtime deployment files such as `docker-compose.yml` or `fly.toml` are not currently committed.
- Built artifacts live in `bin/`, but source-of-truth command and server entry points remain under `cmd/`.

---
