GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: build test lint run-cli run-server build-server clean migrate-up migrate-down migrate-fresh docker-build release-local

build:
	mkdir -p bin
	$(GO) build -o bin/unconf ./cmd/unconf
	$(GO) build -o bin/unconf-server ./cmd/server

test:
	$(GO) test -race ./...

lint:
	$(GOLANGCI_LINT) run ./...

run-cli: build
	set -a && . ./.env && set +a && ./bin/unconf

run-server:
	set -a && . ./.env && set +a && $(GO) run ./cmd/server

build-server:
	mkdir -p bin
	$(GO) build -o bin/unconf-server ./cmd/server

clean:
	rm -rf bin
	rm -f *.out *.test

migrate-up:
	$(GO) run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.0 -path migrations -database "sqlite3://.unconf.db" up

migrate-down:
	$(GO) run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.0 -path migrations -database "sqlite3://.unconf.db" down

migrate-fresh:
	$(GO) run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.0 -path migrations -database "sqlite3://.unconf.db" down
	$(GO) run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.0 -path migrations -database "sqlite3://.unconf.db" up

docker-build:
	docker build -t unconf:latest .

release-local:
	goreleaser check
	goreleaser build --snapshot --clean --single-target