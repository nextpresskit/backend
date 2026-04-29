# NextPressKit backend - developer tasks
#
# Requires: Go (see go.mod), PostgreSQL for migrate/seed/run.
# Config: copy .env.example to .env (DB_*, JWT_*, etc.).

# Binaries written under bin/
BINARY_NAME   := server
MIGRATE_BINARY := migrate
SEED_BINARY   := seed
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
API_PID_FILE ?= .tmp/nextpress-api.pid
API_LOG_FILE ?= .tmp/nextpress-api.log

# migrate-steps: direction (up = apply N, down = roll back N)
MIGRATE_CMD ?= up

.PHONY: help all build run start stop clean \
	test test-coverage test-integration tidy deps graphql \
	seed seed-build \
	migrate-up migrate-down migrate-steps migrate-drop migrate-version db-fresh \
	security-check deploy

## deploy: Interactive deploy wizard (Nginx, TLS, systemd snippets, optional git/build/migrate); Windows: scripts/deploy.ps1
deploy:
	@bash scripts/deploy

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
	go build -ldflags "-X main.version=$(VERSION)" -o bin/$(BINARY_NAME) ./cmd/api
	@echo "Done."

## run: Start the API with go run (loads .env from cwd if present)
run:
	@port="$$(awk -F= '/^APP_PORT=/{print $$2; exit}' .env 2>/dev/null | tr -d '[:space:]')"; \
	if [ -z "$$port" ]; then port=9090; fi; \
	if ss -ltn "( sport = :$$port )" | awk 'NR>1 { found=1 } END { exit found ? 0 : 1 }'; then \
		echo "Port $$port is already in use. Stop the running process or change APP_PORT in .env."; \
		exit 0; \
	fi; \
	interrupted=0; trap 'interrupted=1' INT; \
	go run ./cmd/api; status=$$?; \
	if [ $$interrupted -eq 1 ] || [ $$status -eq 130 ]; then exit 0; fi; \
	if [ $$status -ne 0 ]; then \
		echo "run finished with exit code $$status"; \
	fi; \
	exit 0

## start: Start the API in background (recommended for local dev)
start:
	@mkdir -p "$$(dirname "$(API_PID_FILE)")"; \
	if [ -f "$(API_PID_FILE)" ]; then \
		pid="$$(cat "$(API_PID_FILE)")"; \
		if kill -0 "$$pid" 2>/dev/null; then \
			echo "API already running (pid=$$pid)."; \
			exit 0; \
		fi; \
		rm -f "$(API_PID_FILE)"; \
	fi; \
	port="$$(awk -F= '/^APP_PORT=/{print $$2; exit}' .env 2>/dev/null | tr -d '[:space:]')"; \
	if [ -z "$$port" ]; then port=9090; fi; \
	if ss -ltn "( sport = :$$port )" | awk 'NR>1 { found=1 } END { exit found ? 0 : 1 }'; then \
		echo "Port $$port is already in use. Stop the running process or change APP_PORT in .env."; \
		exit 0; \
	fi; \
	nohup go run ./cmd/api > "$(API_LOG_FILE)" 2>&1 & echo $$! > "$(API_PID_FILE)"; \
	sleep 1; \
	pid="$$(cat "$(API_PID_FILE)")"; \
	if kill -0 "$$pid" 2>/dev/null; then \
		echo "API started in background (pid=$$pid)."; \
		echo "Logs: $(API_LOG_FILE)"; \
	else \
		echo "API failed to start. See $(API_LOG_FILE)"; \
		rm -f "$(API_PID_FILE)"; \
	fi

## stop: Stop the background API started with start
stop:
	@if [ ! -f "$(API_PID_FILE)" ]; then \
		echo "API is not running (no pid file)."; \
		exit 0; \
	fi; \
	pid="$$(cat "$(API_PID_FILE)")"; \
	if ! kill -0 "$$pid" 2>/dev/null; then \
		echo "API is not running (stale pid file)."; \
		rm -f "$(API_PID_FILE)"; \
		exit 0; \
	fi; \
	kill -TERM "$$pid"; \
	for _ in 1 2 3 4 5 6 7 8 9 10; do \
		if ! kill -0 "$$pid" 2>/dev/null; then \
			rm -f "$(API_PID_FILE)"; \
			echo "API stopped."; \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "API did not stop in time; sending SIGKILL."; \
	kill -KILL "$$pid" 2>/dev/null || true; \
	rm -f "$(API_PID_FILE)"; \
	echo "API stopped."

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

## test-integration: Run integration tests that require real services (Postgres)
test-integration:
	go test -tags=integration -v ./internal/platform/database

## tidy: go mod tidy
tidy:
	go mod tidy

## deps: go mod download
deps:
	go mod download

## security-check: Run dependency vulnerability scan (govulncheck)
security-check:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## graphql: Regenerate gqlgen code from internal/graphql/schema.graphqls
graphql:
	go run github.com/99designs/gqlgen generate
