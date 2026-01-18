# Tasks: Backend Initialization & Conference Listing

**Input**: Design documents from `specs/002-backend-conference-list/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: TDD is explicitly requested ("Strict TDD", "Contract Testing"). Contract tests MUST be written before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Install dependencies `github.com/gin-gonic/gin`, `github.com/go-resty/resty/v2`, `github.com/pact-foundation/pact-go/v2`, `go.uber.org/zap`, and `github.com/jedib0t/go-pretty/v6`
- [ ] T002 [P] Create backend directories: `cmd/backend`, `internal/backend/handlers`, `internal/backend/repository`, `internal/backend/data`, `build/package`
- [ ] T003 [P] Create testing directories: `tests/contract`, `internal/domain`
- [ ] T004 Create `build/package/backend.Dockerfile` for containerized deployment

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure and data modeling required for all stories

- [ ] T005 [P] Define `Conference` entity in `internal/domain/conference.go`
- [ ] T006 Setup structured JSON logger using Zap in `internal/backend/logger.go`
- [ ] T007 Define `ConferenceRepository` interface in `internal/backend/repository/interface.go`
- [ ] T008 [P] Create mock data file `internal/backend/data/conferences.json` per data-model.md
- [ ] T009 Implement JSON-based `ConferenceRepository` in `internal/backend/repository/json_repo.go`

**Checkpoint**: Foundation ready - backend components and domain entities are established.

---

## Phase 3: User Story 1 - Start Backend Service (Priority: P1) 🎯 MVP

**Goal**: Launch a containerized backend service with a working health check endpoint.

**Independent Test**: Build and run the Docker container, then `curl http://localhost:8080/health` and expect 200 OK.

### Implementation for User Story 1

- [ ] T010 [US1] Implement health check handler in `internal/backend/handlers/health.go`
- [ ] T011 [US1] Create backend entry point in `cmd/backend/main.go` with Gin and structured logging
- [ ] T012 [US1] Verify containerized startup and health endpoint per `quickstart.md`

**Checkpoint**: User Story 1 is functional. The backend can be deployed and monitored.

---

## Phase 4: User Story 2 - Retrieve Conference List (Priority: P1)

**Goal**: CLI command `unconf list` fetches and displays conferences from the backend API.

**Independent Test**: Run `unconf list` and verify the table output matches `conferences.json`.

### Tests for User Story 2 (TDD Required) ⚠️

- [ ] T013 [US2] Create Pact consumer test for the CLI in `tests/contract/cli_consumer_test.go`
- [ ] T014 [US2] Generate Pact contract file (JSON) from consumer test

### Implementation for User Story 2

- [ ] T015 [US2] Implement conference list handler in `internal/backend/handlers/conference.go`
- [ ] T016 [US2] Register `/conferences` route in `cmd/backend/main.go` with repository injection
- [ ] T017 [US2] Create Pact provider verification test in `tests/contract/backend_provider_test.go`
- [ ] T018 [US2] Implement `list` command logic in `internal/cli/list.go` using `resty` and `go-pretty`
- [ ] T019 [US2] Add `--json` flag support to `list` command in `internal/cli/list.go`
- [ ] T020 [US2] Wire up `list` command to root in `internal/cli/root.go`

**Checkpoint**: User Story 2 is functional. Full client-server communication is established and verified via contracts.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Finalizing error handling, documentation, and validation.

- [ ] T021 [US2] Enhance CLI connection error handling with troubleshooting tips in `internal/cli/list.go`
- [ ] T022 [P] Verify implementation against OpenAPI spec in `specs/002-backend-conference-list/contracts/openapi.yaml`
- [ ] T023 Run full verification suite per `quickstart.md`
- [ ] T024 [US2] Verify CLI and Backend behavior when conference list is empty in `internal/backend/data/conferences.json`

## Phase 6: TUI Refinement (Charm Stack)
**Purpose**: Align interactive commands with the "Crush" aesthetic using the Charm Bracelet stack.

- [ ] T025 [US2] Install Charm Bracelet dependencies: `bubbletea`, `lipgloss`, `bubbles`
- [ ] T026 [US2] Implement interactive list model in `internal/cli/tui_list.go`
- [ ] T027 [US2] Update `list` command in `internal/cli/list.go` to launch interactive TUI by default when no flags are provided
- [ ] T028 [US2] Verify TUI behavior and responsiveness

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion.
- **User Story 1 (Phase 3)**: Depends on Foundational completion.
- **User Story 2 (Phase 4)**: Depends on Foundational completion. Can run in parallel with US1 implementation but needs backend entry point (T011) for full integration.
- **Polish (Phase 5)**: Depends on all user stories completion.

### Parallel Opportunities

- T002, T003 (Directories) can run in parallel.
- T005, T006, T008 (Base components) can run in parallel.
- T013 (Consumer Test) can be started as soon as domain entities (T005) are ready.
- T022 (OpenAPI verification) can run in parallel with polish tasks.

---

## Implementation Strategy

### MVP First (US1 + US2)

1. Complete Setup and Foundational phases.
2. Implement US1 (Health Check + Docker) to ensure the service runs.
3. Implement US2 (List Command + Contract) to deliver the core business value.
4. Validate both stories independently using the specified criteria.

---

## Notes

- Use `go mod tidy` after T001.
- Ensure Pact Mock Server is running for T013.
- Use `ldflags` for versioning if needed, but focus on functional requirements first.