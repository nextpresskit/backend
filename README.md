## nextpress-backend

nextpress-backend is a **production-ready CMS backend** written in Go, designed as a **modular monolith** with a **global PostgreSQL database instance**, a **clean service layer**, and a future-proof **plugin system** (implemented in later phases).

The current codebase includes the **foundation phase**:

- Strict, opinionated folder structure
- Application bootstrap with Gin
- Environment and configuration loading
- Deployment-oriented scripts and configuration skeletons (Makefile, systemd, Nginx)

### High-level architecture

- **Modular Monolith**: Features are grouped by domain under `internal/modules`, sharing a single process and database, while keeping clear boundaries for maintainability.
- **Global DB instance**: A single Postgres connection (via GORM) is initialized once and injected where needed to avoid duplicated connection logic.
- **Service layer**: Thin HTTP handlers, reusable application services, and clear separation of concerns without over-engineered DDD layers.

### Folder structure

```text
.
├── cmd/
│   └── api/
│       └── main.go          # HTTP API entrypoint
├── internal/
│   ├── config/              # Configuration loading and environment helpers
│   ├── http/                # HTTP server setup (Gin engine, middlewares, routing)
│   ├── db/                  # Global DB initialization and lifecycle management
│   ├── modules/             # Feature modules (content, auth, media, etc.)
│   ├── logging/             # Centralized logger setup
│   └── shared/              # Shared helpers (errors, responses, etc.)
├── deploy/
│   ├── nginx/                        # Nginx configs per environment
│   └── systemd/                      # Systemd template unit
├── docs/
│   └── deployment/                   # Deployment guides (local/dev/staging/production)
├── scripts/
│   ├── run_local.sh                  # Local run helper
│   └── deploy                        # Ubuntu deploy script (systemd + Nginx)
└── Makefile                          # Common developer tasks (build, run, test, docker)
```

> Later phases will populate `internal/modules` with concrete CMS features (auth, content, media, plugins, etc.).

### Getting started

#### Prerequisites

- Go 1.26+
- PostgreSQL (for DB-backed phases)

#### Local development

```bash
cp .env.example .env
go mod download
make build
make run    # or ./scripts/run_local.sh
```

#### Configuration

- `.env` (optional) – for local overrides, loaded via `godotenv`.
- `deployments/configs/app.<env>.yaml` – environment-specific defaults.

Key environment variables (Phase 1) – see `.env.example`:

- `APP_NAME` – application name (default: `NextPress`)
- `APP_ENV` – e.g. `development`, `staging`, `production` (default: `development`)
- `APP_PORT` – port for the HTTP server (default: `9090`)

Database, JWT, and module-level configuration will be introduced in later phases.

For full deployment instructions (Ubuntu + systemd + Nginx), see:

- `docs/DEPLOYMENT.md`
- `docs/deployment/production.md`
- `docs/deployment/staging.md`
- `docs/deployment/dev.md`
- `docs/deployment/local.md`

API Reference:
- OpenAPI spec: `docs/openapi.yaml`

# NextPress Backend

NextPress is a modular CMS backend written in Go.

## Stack

- Go 1.26
- Gin HTTP Framework
- PostgreSQL
- GORM
- Zap Logger
- JWT Authentication

## Architecture

The project follows:

- Clean Architecture
- Modular Monolith
- Domain Driven Design

## Project Structure

cmd/api → application entry point

internal/config → configuration system  
internal/platform → shared infrastructure  
internal/modules → domain modules

pkg → reusable utilities

## Git workflow

- Long-lived branches:
  - `dev` – active development and integration testing.
  - `staging` – pre-production testing.
  - `main` – production.
- Promotion flow:
  - New work is done on feature branches off `dev`, then merged into `dev`.
  - To promote: merge `dev` into `staging`, push; merge `staging` into `main`, push.
  - To keep branches in sync: merge `main` back into `staging` and `dev`, push.
- Server mappings (recommended):
  - `/var/www/nextpress-backend-dev` → tracks `dev` → service `nextpress-backend@dev`.
  - `/var/www/nextpress-backend-staging` → tracks `staging` → service `nextpress-backend@staging`.
  - `/var/www/nextpress-backend-production` → tracks `main` → service `nextpress-backend@production`.

See `docs/DEPLOYMENT.md` for how this ties into the deploy script and systemd units.
