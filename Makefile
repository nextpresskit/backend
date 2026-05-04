# NextPressKit backend — developer tasks
#
# Cross-platform CLI: ./scripts/nextpresskit (Linux/macOS/Git Bash) or .\scripts\nextpresskit.ps1 (Windows).
# Requires: Go (see go.mod), PostgreSQL for migrate/seed/run.
# Config: copy .env.example to .env (or: make install / nextpresskit install).

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
	security-check deploy deploy-nginx deploy-ps checks postman-sync

## help: List targets and short descriptions
help:
	@echo "Usage: make [target]   — or: ./scripts/nextpresskit <command>"
	@echo ""
	@echo "Common:"
	@echo "  make install   Go modules + .env from .env.example if missing"
	@echo "  make setup     install + build-all + migrate-up + seed"
	@echo "  make migrate-up / seed   Sync DB schema, then demo data"
	@echo "  make db-fresh  Dev reset: empty public + migrate-up (add make seed after if you want data)"
	@echo "  make run       Foreground API"
	@echo "  make deploy       Interactive Nginx/TLS wizard (Linux/macOS bash)"
	@echo "  make deploy-nginx Non-interactive Nginx install (scripts/deploy apply-nginx)"
	@echo "  make deploy-ps Interactive deploy (Windows PowerShell)"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'

## install: go mod download; create .env from .env.example if missing
install:
	@bash scripts/nextpresskit install

## setup: Bootstrap + local HTTPS helper when TTY (mkcert + nginx on Linux); SKIP_SETUP_LOCAL_HTTPS=1 to skip
setup: install build-all migrate-up seed
	@if [ -t 0 ] && [ "$$SKIP_SETUP_LOCAL_HTTPS" != "1" ]; then bash scripts/setup-local-https.sh || true; fi
	@echo "Setup done."

## all: Alias of build
all: build

## build: Produce bin/server from cmd/api
build:
	@bash scripts/nextpresskit build

## build-all: Build bin/server, bin/migrate, bin/seed
build-all:
	@bash scripts/nextpresskit build-all

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
	@bash scripts/nextpresskit clean

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
	@bash scripts/nextpresskit seed

## seed-build: Build bin/seed only
seed-build:
	@echo "Building $(SEED_BINARY)..."
	@mkdir -p bin
	go build -o bin/$(SEED_BINARY) ./cmd/seed
	@echo "Done."

## migrate-up: GORM AutoMigrate via internal/platform/dbmigrate (module persistence)
migrate-up:
	@bash scripts/nextpresskit migrate-up

## migrate-down: Removed (fails with guidance); use db-fresh for dev reset
migrate-down:
	@bash scripts/nextpresskit migrate-down

## migrate-steps: Removed (fails with guidance)
migrate-steps:
	@test -n "$(STEPS)" || (echo >&2 "Usage: make migrate-steps STEPS=n (removed; see docs/COMMANDS.md)"; exit 1)
	@bash scripts/nextpresskit migrate-steps "$(STEPS)"

## migrate-version: Prints notice (no version table with AutoMigrate)
migrate-version:
	@bash scripts/nextpresskit migrate-version

## migrate-drop: Drop all public tables (interactive; sets ALLOW_SCHEMA_DROP)
migrate-drop:
	@bash scripts/nextpresskit migrate-drop

## db-fresh: migrate-drop then migrate-up (destructive)
db-fresh:
	@bash scripts/nextpresskit db-fresh

## test: Run all tests (verbose)
test:
	@bash scripts/nextpresskit test

## test-coverage: Run tests with coverage summary
test-coverage:
	@bash scripts/nextpresskit test-coverage

## test-integration: Integration tests (Postgres; set DB_* or skipped)
test-integration:
	@bash scripts/nextpresskit test-integration

## tidy: go mod tidy
tidy:
	@bash scripts/nextpresskit tidy

## deps: go mod download (modules only; no .env)
deps:
	@command -v go >/dev/null 2>&1 || { echo "Go not found" >&2; exit 1; }
	go mod download

## security-check: govulncheck
security-check:
	@bash scripts/nextpresskit security-check

## checks: CI-style suite (see scripts/nextpresskit checks)
checks:
	@bash scripts/nextpresskit checks

## graphql: Regenerate gqlgen code
graphql:
	@bash scripts/nextpresskit graphql

## postman-sync: Seed postman/ from postman-templates/; refresh env JSON from .env.example + .env (optional POSTMAN_* URLs)
postman-sync:
	@bash scripts/nextpresskit postman-sync
