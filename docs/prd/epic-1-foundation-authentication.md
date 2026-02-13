# Epic 1: Foundation & Authentication

**Goal**: Establish project infrastructure (CLI skeleton, backend API, database, CI/CD) and implement GitHub OAuth authentication — delivering a working `unconf login` command that proves the full stack works end-to-end.

## Story 1.1: Project Scaffolding

**As a** developer,  
**I want** the project repository initialized with Go module, folder structure, and basic tooling,  
**so that** I have a clean foundation to build upon.

**Acceptance Criteria:**
1. Go module initialized (`go.mod`) with appropriate module path
2. Standard Go project layout created (`/cmd`, `/internal`, `/pkg`)
3. `.gitignore` configured for Go projects
4. `README.md` with project overview and setup instructions
5. Makefile or task runner with common commands (`build`, `test`, `lint`)
6. `golangci-lint` configuration file present

## Story 1.2: CLI Framework Setup

**As a** developer,  
**I want** the Cobra CLI framework initialized with a root command,  
**so that** I can add subcommands for UNCONF features.

**Acceptance Criteria:**
1. Cobra initialized with `unconf` as root command
2. Viper integrated for configuration management
3. `--help` flag works and shows welcoming help text
4. `--version` flag displays version information
5. Configuration file support (`.unconf.yaml` or similar)
6. Basic error handling and exit codes established

## Story 1.3: Backend API Skeleton

**As a** developer,  
**I want** a basic Go + Gin backend API running,  
**so that** the CLI has an API to communicate with.

**Acceptance Criteria:**
1. Gin router initialized with health check endpoint (`GET /health`)
2. Structured logging configured (using `log/slog`)
3. Graceful shutdown handling implemented
4. CORS configured for development
5. Dockerfile created for containerized deployment
6. API responds with JSON; health endpoint returns `{"status": "ok"}`

## Story 1.4: Database Setup with Repository Pattern

**As a** developer,  
**I want** SQLite database initialized with repository pattern,  
**so that** data persistence is abstracted and migration-ready.

**Acceptance Criteria:**
1. SQLite database file created on first run
2. Database connection pool/manager implemented
3. Repository interface defined for data access abstraction
4. Initial schema migration system in place
5. `users` table created with basic fields (id, github_id, email, display_name, created_at)
6. Repository pattern allows swapping SQLite for PostgreSQL without business logic changes

## Story 1.5: CI/CD Pipeline

**As a** developer,  
**I want** GitHub Actions CI pipeline configured,  
**so that** code quality is enforced on every PR.

**Acceptance Criteria:**
1. GitHub Actions workflow triggers on PR and push to main
2. Pipeline runs `go test ./...` for all tests
3. Pipeline runs `golangci-lint` for static analysis
4. Pipeline fails if tests fail or linting errors exist
5. GoReleaser configuration file present (for future releases)
6. Build succeeds for linux/amd64, darwin/amd64, darwin/arm64, windows/amd64

## Story 1.6: GitHub OAuth Integration (Backend)

**As a** backend developer,  
**I want** GitHub OAuth device flow endpoints implemented,  
**so that** CLI users can authenticate without browser redirects.

**Acceptance Criteria:**
1. `POST /auth/device` endpoint initiates device flow, returns device_code and user_code
2. `POST /auth/token` endpoint polls for access token completion
3. GitHub OAuth app credentials configurable via environment variables
4. On successful auth, user record created/updated in database
5. PASETO token issued for subsequent API authentication
6. Token expiry and refresh mechanism defined

## Story 1.7: CLI Login Command

**As an** attendee,  
**I want** to run `unconf login` to authenticate with my GitHub account,  
**so that** I can access UNCONF features.

**Acceptance Criteria:**
1. `unconf login` command initiates GitHub device flow
2. User shown device code and URL to enter it (https://github.com/login/device)
3. CLI polls backend for authentication completion
4. On success, auth token stored securely in OS keychain
5. Success message displayed with user's GitHub username
6. Subsequent commands can detect authenticated state
7. `unconf logout` command removes stored credentials

---
