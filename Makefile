GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: build test lint run-cli run-server build-server clean migrate-up migrate-down docker-build

build:
	mkdir -p bin
	$(GO) build -o bin/unconf ./cmd/unconf
	$(GO) build -o bin/unconf-server ./cmd/server

test:
	$(GO) test -race ./...

lint:
	$(GOLANGCI_LINT) run ./...

run-cli: build
	./bin/unconf

run-server:
	$(GO) run ./cmd/server

build-server:
	mkdir -p bin
	$(GO) build -o bin/unconf-server ./cmd/server

clean:
	rm -rf bin
	rm -f *.out *.test

migrate-up:
	@echo "TODO: add migration tooling"

migrate-down:
	@echo "TODO: add migration tooling"

docker-build:
	docker build -t unconf:latest .