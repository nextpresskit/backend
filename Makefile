# NextPress Backend - developer tasks
#
# Requires: Go (see go.mod), PostgreSQL for migrate/seed/run.
# Config: copy .env.example to .env (DB_*, JWT_*, etc.).

# Binaries written under bin/
BINARY_NAME   := server
MIGRATE_BINARY := migrate
SEED_BINARY   := seed

# migrate-steps: direction (up = apply N, down = roll back N)
MIGRATE_CMD ?= up

.PHONY: help all build run clean \
	test test-coverage tidy deps graphql \
	seed seed-build \
	migrate-up migrate-down migrate-steps migrate-drop migrate-version db-fresh

## help: List targets and short descriptions
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'

## all: Same as build
all: build

## build: Produce bin/server from cmd/api
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	go build -o bin/$(BINARY_NAME) ./cmd/api
	@echo "Done."

## run: Start the API with go run (loads .env from cwd if present)
run:
	go run ./cmd/api

## clean: Remove bin/ and go clean
clean:
	rm -rf bin/
	go clean

# --- Database: seed -----------------------------------------------------------

## seed: Run seeders (go run ./cmd/seed)
seed:
	go run ./cmd/seed

## seed-build: Build bin/seed
seed-build:
	@echo "Building $(SEED_BINARY)..."
	@mkdir -p bin
	go build -o bin/$(SEED_BINARY) ./cmd/seed
	@echo "Done."

# --- Database: migrations (migrations/, cmd/migrate) ---------------------------

## migrate-up: Apply all pending migrations
migrate-up:
	go run ./cmd/migrate -command=up

## migrate-down: Roll back one migration (-steps=1; omitting steps would roll back all)
migrate-down:
	go run ./cmd/migrate -command=down -steps=1

## migrate-steps: Apply or roll back STEPS migrations (MIGRATE_CMD=up|down)
migrate-steps:
	@test -n "$(STEPS)" || (echo >&2 "Usage: make migrate-steps STEPS=n [MIGRATE_CMD=up|down]"; exit 1)
	go run ./cmd/migrate -command=$(MIGRATE_CMD) -steps=$(STEPS)

## migrate-version: Print current schema version / dirty flag
migrate-version:
	go run ./cmd/migrate -command=version

## migrate-drop: Drop all tables (interactive confirm)
migrate-drop:
	@echo "WARNING: This drops all tables in the configured database."
	@read -p "Type y to continue: " confirm && [ "$${confirm:-N}" = y ]
	go run ./cmd/migrate -command=drop

## db-fresh: migrate-drop then migrate-up (destructive)
db-fresh: migrate-drop migrate-up

# --- Tests and modules ---------------------------------------------------------

## test: Run all tests with verbose output
test:
	go test -v ./...

## test-coverage: Run tests with coverage summary
test-coverage:
	go test -cover ./...

## tidy: go mod tidy
tidy:
	go mod tidy

## deps: go mod download
deps:
	go mod download

## graphql: Regenerate gqlgen code from internal/graphql/schema.graphqls
graphql:
	go run github.com/99designs/gqlgen generate
