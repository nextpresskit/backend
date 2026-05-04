# Commands Guide

What to run, in plain language. Use either `./scripts/nextpresskit` or `make`; they run the same workflows.

[← Documentation index](README.md) · [Quick start](../README.md#quick-start) · [Contributing checks](../CONTRIBUTING.md#before-you-open-a-pr)

## Command families

- `./scripts/nextpresskit ...` (Unix): main command runner with grouped subcommands.
- `make ...`: thin wrappers around the same script workflows.
- `.\scripts\nextpresskit.ps1 ...` (Windows PowerShell): Windows equivalent.

Use whichever family matches your environment. The behavior is intentionally aligned.

## Choose by task

| I want to… | Use this |
|------------|----------|
| Set up a fresh clone | `./scripts/nextpresskit setup` |
| Run the API now | `./scripts/nextpresskit run` |
| Run checks before PR | `./scripts/nextpresskit checks` |
| Create or update database tables | `./scripts/nextpresskit migrate-up` |
| Load demo data (run after migrate) | `./scripts/nextpresskit seed` |
| Start over on a dev database | `./scripts/nextpresskit db-fresh`, then `seed` (see [Database and seed data](#database-and-seed-data)) |
| Generate deploy config | `./scripts/nextpresskit deploy` or `make deploy` |
| Sync Postman env files | `./scripts/nextpresskit postman-sync` |

## Commonly confused commands

| Pair | Difference |
|------|------------|
| `setup` vs `deploy` | `setup` bootstraps local development (deps, env, migrate, seed). `deploy` runs deployment/config wizard flows (nginx/systemd/TLS/release steps). |
| `setup` vs `install` | `setup` is fuller (includes migration/seed/build helpers). `install` is lighter (dependency and prerequisite initialization). |
| `run` vs `start` | `run` keeps the API in the foreground. `start` runs it in the background (Unix) so your shell stays free. |
| `checks` vs `test` | `checks` is CI-style multi-check flow. `make test` runs Go tests only. |

## Most-used commands

| Command | What it does | When to use it |
|---------|---------------|----------------|
| `./scripts/nextpresskit setup` | Full local bootstrap: modules, `.env` (if missing), build helpers, migrate, seed. | First run on a new clone. |
| `./scripts/nextpresskit run` | Runs API in foreground (dev mode). | Normal local development. |
| `./scripts/nextpresskit start` | Starts API in background (Unix). | Keep server running while you use the shell. |
| `./scripts/nextpresskit stop` | Stops background API process (Unix). | Clean shutdown after `start`. |
| `./scripts/nextpresskit checks` | CI-style local checks. | Before pushing or opening a PR. |
| `./scripts/nextpresskit postman-sync` | Creates/updates gitignored `postman/` from templates + env values. | Before importing Postman collections/environments. |
| `./scripts/nextpresskit deploy` | Interactive deployment/config generation wizard. | Generate nginx/systemd snippets or run release steps. |

`setup` is for local development bootstrap. `deploy` is for deployment and release workflows.

`make` equivalents:

- `make setup`, `make run`, `make start`, `make stop`, `make checks`, `make postman-sync`, `make deploy`

Windows equivalents:

- `.\scripts\nextpresskit.ps1 setup|run|checks|postman-sync|deploy`

## Setup, build, run

| Command | Description |
|---------|-------------|
| `./scripts/nextpresskit install` | Downloads modules and initializes local prerequisites (`.env` support included). |
| `./scripts/nextpresskit build` | Builds API binary only. |
| `./scripts/nextpresskit build-all` | Builds API + migrate + seed binaries. |
| `./scripts/nextpresskit run` | Runs API in foreground. |
| `./scripts/nextpresskit start` | Runs API in background (Unix). |
| `./scripts/nextpresskit stop` | Stops background API (Unix). |

<a id="database-and-seed-data"></a>

## Database and seed data

Tables come from the Go models via GORM AutoMigrate, wired up in [`internal/platform/dbmigrate/migrate.go`](../internal/platform/dbmigrate/migrate.go). There is no separate SQL migration tree: change the models, then run migrate-up.

- migrate-up: sync the database schema with the code.
- seed: add repeatable demo data (roles, sample posts, a superadmin, and more). Run this after migrate-up.
- db-fresh: local development only. Drops every table in the `public` schema (you confirm first), runs migrate-up again, and stops. It does not run seed; run seed yourself if you want the demo dataset back.

PostgreSQL must be running, with `DB_*` set in [`.env`](../.env.example) so migrate and seed can connect.

| Situation | Commands |
|-----------|----------|
| New clone, full setup | `make setup` (build, migrate, seed), or `make migrate-up` then `make seed` |
| Clean slate on your machine | `make db-fresh`, then `make seed` |
| Pulled code with model changes | `make migrate-up` |
| Refresh demo data only | `make seed` (safe to repeat) |

Same names work as `./scripts/nextpresskit …` or `make …`. On Windows: `.\scripts\nextpresskit.ps1`. After `make build-all`, `./bin/migrate -command=up` and `./bin/seed` match migrate-up and seed.

| Subcommand | Meaning |
|------------|---------|
| migrate-up | AutoMigrate from module [`persistence`](../internal/modules) models (FK-safe order). |
| migrate-down | Removed (no versioned SQL downs). Locally use db-fresh or migrate-drop plus migrate-up. |
| migrate-version | Prints a note: there is no `schema_migrations` table with AutoMigrate. |
| migrate-drop | Drops all `public` tables after confirmation (`ALLOW_SCHEMA_DROP` is set for you). |
| db-fresh | migrate-drop plus migrate-up only; no seed. |
| seed | Upserts RBAC defaults and the full demo dataset ([SEEDING.md](SEEDING.md)). |

On servers, releases usually run `bin/migrate -command=up` and sometimes `bin/seed` ([DEPLOYMENT.md](DEPLOYMENT.md)). Do not use db-fresh in production.

## API contract and quality checks

| Command | Description |
|---------|-------------|
| `make graphql` | Regenerates GraphQL code after schema changes. |
| `./scripts/nextpresskit checks` | Runs project checks used in CI-style validation. |
| `make test` | Runs Go test suites. |
| `go vet ./...` | Static analysis for suspicious constructs. |
| `make security-check` | Runs vulnerability checks (`govulncheck`). |

## Deployment workflow commands

| Command | Description |
|---------|-------------|
| `./scripts/deploy` | Interactive deployment wizard (Linux/macOS/Git Bash). |
| `.\scripts\deploy.ps1` | Interactive deployment wizard (PowerShell). |
| `make deploy` | Wrapper for Unix deploy wizard. |
| `make deploy-ps` | Convenience target for PowerShell deploy flow. |

For full production guidance, branch promotion model, and TLS options, see `DEPLOYMENT.md`.

## Postman and environments

| Command | Description |
|---------|-------------|
| `./scripts/nextpresskit postman-sync` | Syncs environment JSON into gitignored `postman/`. |
| `./scripts/nextpresskit postman-sync --dry-run` | Shows what would change, without writing files. |
| `POSTMAN_CLEAR_TOKENS=1 ./scripts/nextpresskit postman-sync` | Clears token placeholders in generated Postman env files. |

## Help and discovery

| Command | Description |
|---------|-------------|
| `./scripts/nextpresskit help` | Full command list and usage. |
| `make help` | Available make targets and short descriptions. |

If a command is unfamiliar, run the `help` variants first and then use the equivalent command family for your OS.
