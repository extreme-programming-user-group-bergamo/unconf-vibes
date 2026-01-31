# 12. Development Workflow

## 12.1 Prerequisites

```bash
go version   # 1.21 or higher
sqlite3      # For local database
docker       # For MailHog

# Recommended
brew install golangci-lint
brew install golang-migrate
```

## 12.2 Initial Setup

```bash
git clone https://github.com/unconf/unconf.git
cd unconf
go mod download
cp configs/config.example.yaml ~/.unconf/config.yaml
make migrate-up
make test
```

## 12.3 Development Commands

```bash
# Start server + MailHog
docker-compose up -d

# Run server
make run-server

# Run CLI
make run-cli -- login
make run-cli -- rooms

# Tests
make test
make test-coverage

# Lint
make lint

# Migrations
make migrate-up
make migrate-down
```

## 12.4 Environment Variables

```bash
# Server (.env)
PORT=8080
DATABASE_URL=sqlite3://./unconf.db
JWT_SECRET=your-dev-secret-min-32-chars
GITHUB_CLIENT_ID=xxx
GITHUB_CLIENT_SECRET=xxx
SMTP_HOST=localhost
SMTP_PORT=1025

# CLI (~/.unconf/config.yaml)
api_endpoint: http://localhost:8080/v1
```

---
