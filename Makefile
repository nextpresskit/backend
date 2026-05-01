# NextPressKit backend — developer tasks
#
# Cross-platform CLI: ./scripts/nextpress (Linux/macOS/Git Bash) or .\scripts\nextpress.ps1 (Windows).
# Requires: Go (see go.mod), PostgreSQL for migrate/seed/run.
# Config: copy .env.example to .env (or: make install / nextpress install).

BINARY_NAME   := server
MIGRATE_BINARY := migrate
SEED_BINARY   := seed
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
# Default matches APP_DEV_RUNTIME_BASENAME in .env (nextpresskit); override if you change only Makefile.
DEV_RUNTIME_BASENAME ?= nextpresskit
API_PID_FILE ?= .tmp/$(DEV_RUNTIME_BASENAME)-api.pid
API_LOG_FILE ?= .tmp/$(DEV_RUNTIME_BASENAME)-api.log

MIGRATE_CMD ?= up

.PHONY: help all build build-all install setup run start stop clean \
	test test-coverage test-integration tidy deps graphql \
	seed seed-build \
	migrate-up migrate-down migrate-steps migrate-drop migrate-version db-fresh \
	security-check deploy deploy-nginx deploy-ps checks

## help: List targets and short descriptions
help:
	@echo "Usage: make [target]   — or: ./scripts/nextpress <command>"
	@echo ""
	@echo "Common:"
	@echo "  make install   Go modules + .env from .env.example if missing"
	@echo "  make setup     install + build-all + migrate-up + seed"
	@echo "  make run       Foreground API"
	@echo "  make deploy       Interactive Nginx/TLS wizard (Linux/macOS bash)"
	@echo "  make deploy-nginx Non-interactive Nginx install (scripts/deploy apply-nginx)"
	@echo "  make deploy-ps Interactive deploy (Windows PowerShell)"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'

## install: go mod download; create .env from .env.example if missing
install:
	@bash scripts/nextpress install

## setup: Bootstrap + local HTTPS helper when TTY (mkcert + nginx on Linux); SKIP_SETUP_LOCAL_HTTPS=1 to skip
setup: install build-all migrate-up seed
	@if [ -t 0 ] && [ "$$SKIP_SETUP_LOCAL_HTTPS" != "1" ]; then bash scripts/setup-local-https.sh || true; fi
	@echo "Setup done."

## all: Alias of build
all: build

## build: Produce bin/server from cmd/api
build:
	@bash scripts/nextpress build

## build-all: Build bin/server, bin/migrate, bin/seed
build-all:
	@bash scripts/nextpress build-all

## run: Start the API in the foreground (go run)
run:
	@bash scripts/dev-run.sh

## start: Start the API in the background (logs under .tmp/ from DEV_RUNTIME_BASENAME)
start:
	@bash scripts/dev-start.sh

## stop: Stop background API + soft-clear same-repo listeners on APP_PORT
stop:
	@bash scripts/dev-stop.sh

## clean: Remove bin/ and go clean
clean:
	@bash scripts/nextpress clean

## deploy: Interactive deploy wizard (Nginx, TLS, systemd); Windows: make deploy-ps
deploy:
	@bash scripts/deploy

## deploy-nginx: Non-interactive Nginx write + Linux install (sudo); ./scripts/deploy apply-nginx --help
deploy-nginx:
	@bash scripts/deploy apply-nginx

## deploy-ps: Interactive deploy wizard (PowerShell; run from repo root on Windows)
deploy-ps:
	@if command -v pwsh >/dev/null 2>&1; then \
		pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/deploy.ps1; \
	elif command -v powershell.exe >/dev/null 2>&1; then \
		powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/deploy.ps1; \
	else \
		echo "Install PowerShell 7+ (pwsh) or use Git Bash: make deploy" >&2; \
		exit 1; \
	fi

## seed: Run seeders
seed:
	@bash scripts/nextpress seed

## seed-build: Build bin/seed only
seed-build:
	@echo "Building $(SEED_BINARY)..."
	@mkdir -p bin
	go build -o bin/$(SEED_BINARY) ./cmd/seed
	@echo "Done."

## migrate-up: Apply all pending migrations
migrate-up:
	@bash scripts/nextpress migrate-up

## migrate-down: Roll back one migration
migrate-down:
	@bash scripts/nextpress migrate-down

## migrate-steps: Apply or roll back STEPS migrations (MIGRATE_CMD=up|down)
migrate-steps:
	@test -n "$(STEPS)" || (echo >&2 "Usage: make migrate-steps STEPS=n [MIGRATE_CMD=up|down]"; exit 1)
	@bash scripts/nextpress migrate-steps "$(STEPS)"

## migrate-version: Print current migration version
migrate-version:
	@bash scripts/nextpress migrate-version

## migrate-drop: Drop all tables (interactive confirm)
migrate-drop:
	@bash scripts/nextpress migrate-drop

## db-fresh: migrate-drop then migrate-up (destructive)
db-fresh:
	@bash scripts/nextpress db-fresh

## test: Run all tests (verbose)
test:
	@bash scripts/nextpress test

## test-coverage: Run tests with coverage summary
test-coverage:
	@bash scripts/nextpress test-coverage

## test-integration: Integration tests (Postgres; set DB_* or skipped)
test-integration:
	@bash scripts/nextpress test-integration

## tidy: go mod tidy
tidy:
	@bash scripts/nextpress tidy

## deps: go mod download (modules only; no .env)
deps:
	@command -v go >/dev/null 2>&1 || { echo "Go not found" >&2; exit 1; }
	go mod download

## security-check: govulncheck
security-check:
	@bash scripts/nextpress security-check

## checks: CI-style suite (see scripts/nextpress checks)
checks:
	@bash scripts/nextpress checks

## graphql: Regenerate gqlgen code
graphql:
	@bash scripts/nextpress graphql
