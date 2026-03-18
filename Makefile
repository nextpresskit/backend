# Makefile for nextpress-backend

# Default environment (used in future phases for migrations/seeds if needed)
APP_ENV ?= local

# Build variables
BINARY_NAME=server
MIGRATE_BINARY=migrate
SEED_BINARY=seed

.PHONY: all build run clean test migrate-up migrate-down migrate-steps migrate-drop migrate-version seed seed-build db-fresh help tidy deps

## help: Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## all: Build the server binary
all: build

## build: Build the API server binary (bin/server)
build:
	@echo "Building nextpress-backend server..."
	mkdir -p bin
	go build -o bin/$(BINARY_NAME) ./cmd/api
	@echo "Done."

## run: Run the API server directly with go run
run:
	go run ./cmd/api

## clean: Clean build artifacts
clean:
	rm -rf bin/
	go clean

# =============================================================================
# Database Seeding
# =============================================================================

## seed: Run all seeders
seed:
	go run ./cmd/seed

## seed-build: Build the seed binary (bin/seed)
seed-build:
	@echo "Building seed binary..."
	mkdir -p bin
	go build -o bin/$(SEED_BINARY) ./cmd/seed
	@echo "Done."

# =============================================================================
# Database Migrations (placeholders until migrations are added)
# =============================================================================

## migrate-up: Run all pending migrations
migrate-up:
	go run ./cmd/migrate -command=up

## migrate-down: Rollback the last migration
migrate-down:
	go run ./cmd/migrate -command=down

## migrate-steps: Run a specific number of migration steps
migrate-steps:
	@echo "Usage: make migrate-steps STEPS=n"
	@true

## migrate-version: Show current migration version
migrate-version:
	go run ./cmd/migrate -command=version

## migrate-drop: Drop all tables (dangerous)
migrate-drop:
	@echo "WARNING: This will drop all tables in the database!"
	@read -p "Are you sure? [y/N] " confirm && [ $${confirm:-N} = y ]
	go run ./cmd/migrate -command=drop

## db-fresh: Drop all tables then run all migrations
db-fresh:
	$(MAKE) migrate-drop
	$(MAKE) migrate-up

## test: Run tests
test:
	go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	go test -cover ./...

## tidy: Tidy up dependencies
tidy:
	go mod tidy

## deps: Download dependencies
deps:
	go mod download
