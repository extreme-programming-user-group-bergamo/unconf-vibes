GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: build build-cli build-server test lint run-cli run-server clean migrate-up migrate-down migrate-fresh docker-build release-local

build: build-cli build-server

build-cli:
	mkdir -p bin
	CGO_ENABLED=0 $(GO) build -o bin/unconf ./cmd/unconf

build-server:
	mkdir -p bin
	CGO_ENABLED=1 $(GO) build -o bin/unconf-server ./cmd/server

test:
	$(GO) test -race ./...

lint:
	$(GOLANGCI_LINT) run ./...

run-cli: build-cli
	set -a && . ./.env && set +a && ./bin/unconf

run-server: build-server
	set -a && . ./.env && set +a && ./bin/unconf-server

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
