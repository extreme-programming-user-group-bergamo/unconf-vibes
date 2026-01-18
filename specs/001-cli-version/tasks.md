# Tasks: CLI App Foundation with Version Command

**Feature**: 001-cli-version
**Spec**: [spec.md](spec.md)
**Plan**: [plan.md](plan.md)
**Status**: Pending

## Phase 1: Setup
*Goal: Initialize project structure and dependencies*

- [ ] T001 Initialize Go module `github.com/xpugbg/unconf-vibes`
- [ ] T002 Install dependencies `github.com/spf13/cobra` and `github.com/stretchr/testify`
- [ ] T003 Create project directories: `cmd/unconf`, `internal/cli`, `tests/integration`, `tests/unit`

## Phase 2: Foundational
*Goal: Establish entry point and root command structure*

- [ ] T004 Create skeleton `cmd/unconf/main.go` entry point
- [ ] T005 Create basic `internal/cli/root.go` with Cobra root command definition

## Phase 3: User Story 1 - Check Application Version (P1)
*Goal: Implement version command and flags with TDD*
*Test Criteria: `unconf --version` prints "unconf version X.Y.Z"*

- [ ] T006 [US1] Create unit test for version logic in `tests/unit/version_test.go` (Expect Fail)
- [ ] T007 [US1] Create integration test for `--version` binary execution in `tests/integration/cli_test.go` (Expect Fail)
- [ ] T008 [US1] Implement `version` variable and injection point in `internal/cli/root.go`
- [ ] T009 [US1] Implement `version` subcommand logic in `internal/cli/version.go`
- [ ] T010 [US1] Add `--version` and `-v` flags to root command in `internal/cli/root.go`
- [ ] T011 [US1] Wire up version logic and command execution in `cmd/unconf/main.go`
- [ ] T012 [US1] Verify all version tests pass

## Phase 4: User Story 2 - Get Help Information (P2)
*Goal: Ensure help and usage information is displayed correctly*
*Test Criteria: `unconf --help` shows usage info, `unconf` (no args) shows brief usage*

- [ ] T013 [US2] Create integration test for `--help` and no-args in `tests/integration/cli_test.go` (Expect Fail)
- [ ] T014 [US2] Configure root command usage and help strings in `internal/cli/root.go`
- [ ] T015 [US2] Verify help tests pass

## Phase 5: Polish & Cross-Cutting
*Goal: Handle edge cases and finalize build*

- [ ] T016 Create integration tests for unknown commands/flags in `tests/integration/cli_test.go`
- [ ] T017 Run full test suite (`go test ./...`) and verify `go build` produces working binary

## Dependencies

1. **Setup** (T001-T003) -> **Foundational** (T004-T005)
2. **Foundational** -> **US1** (T006-T012)
3. **US1** -> **US2** (T013-T015)
4. **US2** -> **Polish** (T016-T017)

## Parallel Execution Opportunities

- T006 (Unit Test) and T007 (Integration Test) can be written in parallel.
- T013 (Help Test) can be written while T008-T011 (Version Implementation) is in progress, though running it requires the binary build.

## Implementation Strategy

1. **Strict TDD**: Write the test first, verify it fails, then implement the code.
2. **Integration Tests**: These rely on building the binary. The test harness should compile `cmd/unconf` to a temp location for execution.
3. **Version Injection**: Use a linker flag `-ldflags "-X ..."` during the build process to set the version.
