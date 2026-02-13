<agent-instructions project="UNCONF">

<CRITICAL>
ALWAYS: use `user-input` MCP for questions/menus/clarification
NEVER: stop until user confirms complete
ALWAYS: ask for confirmation using `user_elicitation` tool from `user-input` MCP before finalizing any output with the question: "Is the task complete and ready to be finalized?" and only finalize if the user confirms, otherwise keep iterating by asking for clarification or additional input until the user confirms completion via `user_elicitation` tool with the same question.
RETRY: `user-input` failures once before proceeding
</CRITICAL>

<context>
Go CLI+TUI for unconference registration
CLI (Cobra/BubbleTea) → REST API (Gin) → SQLite
CLI NEVER accesses DB directly

cmd/unconf/        CLI entry
cmd/server/        API entry  
internal/cli/      Cobra commands
internal/tui/      BubbleTea models
internal/api/      Gin handlers
internal/service/  Business logic
internal/repository/ Data access
internal/models/   Domain structs
internal/client/   HTTP client
internal/auth/     GitHub OAuth+JWT
migrations/        golang-migrate
</context>

<rules priority="1">
Repository: all DB via interfaces, no SQL in services
  ✓ s.bookingRepo.Create(ctx, booking)
  ✗ s.db.Exec("INSERT...")
Errors: always wrap with context
  ✓ fmt.Errorf("failed to X: %w", err)
  ✗ return err
Context: first param to all I/O functions
</rules>

<rules priority="2">
Logging: log/slog only, no fmt.Println
Config: Viper wrapper, no os.Getenv
DI: constructor injection, no globals
</rules>

<rules priority="3">
Domain errors: internal/service/errors.go + errors.Is()
</rules>

<tui>
Messages: VerbMsg suffix (RoomsLoadedMsg)
Commands: tea.Cmd for async
View: pure, Lip Gloss from tui/common/styles.go
</tui>

<naming>
files: snake_case | packages: lowercase
routes: kebab-case | tables: snake_case plural
interfaces: PascalCase+verb
</naming>

<commands>
make run-cli | make run-server | make test | make lint
make migrate-up | make migrate-down
</commands>

<testing>
unit: mock repos, table-driven
integration: :memory: SQLite
tui: manual | email: MailHog
</testing>

<docs>
docs/architecture.md | docs/prd.md
docs/architecture/coding-standards.md
docs/architecture/source-tree.md
</docs>

</agent-instructions>