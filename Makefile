GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: build test lint run-cli run-server clean migrate-up migrate-down docker-build

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

run-server: build
	./bin/unconf-server

clean:
	rm -rf bin
	rm -f *.out *.test

migrate-up:
	@echo "TODO: add migration tooling"

migrate-down:
	@echo "TODO: add migration tooling"

docker-build:
	@echo "TODO: add docker build"