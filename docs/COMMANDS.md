# Commands Guide

Clear, practical command reference for day-to-day work in this repository.

[← Documentation index](README.md) · [Quick start](../README.md#getting-started) · [Contributing checks](../CONTRIBUTING.md#before-you-open-a-pr)

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
| Apply DB changes | `./scripts/nextpresskit migrate-up` |
| Seed local data | `./scripts/nextpresskit seed` |
| Generate deploy config | `./scripts/nextpresskit deploy` or `make deploy` |
| Sync Postman env files | `./scripts/nextpresskit postman-sync` |

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

## Database and seed data

| Command | Description |
|---------|-------------|
| `./scripts/nextpresskit migrate-up` | Applies pending SQL migrations. |
| `./scripts/nextpresskit migrate-down` | Rolls back the latest migration batch (use carefully). |
| `./scripts/nextpresskit seed` | Runs idempotent seeders (RBAC defaults + deterministic dataset). |
| `make migrate-up` / `make seed` | Same actions through Make wrappers. |

For seeding details and credentials, see `SEEDING.md`.

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
