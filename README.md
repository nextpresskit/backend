# nextpress-backend

Production-oriented **CMS API** in Go.

- **Architecture**: modular monolith (`internal/modules/*`)
- **HTTP**: Gin
- **DB**: PostgreSQL + GORM + SQL migrations (`migrations/`, `cmd/migrate`)
- **Auth**: bcrypt passwords, JWT access + refresh
- **Authorization**: RBAC permissions on `/v1/admin/*`
- **CMS core**: posts, pages, taxonomy (categories/tags), media, menus (+ public read APIs)
- **Hardening**: request IDs + structured logs, in-memory rate limiting
- **Extensibility (WIP)**: plugin registry + post-save hook chain

## Status / roadmap

This repo is developed in phases (infra → auth → RBAC → CMS core → plugins). The authoritative roadmap is `docs/PHASES.md`.

## Requirements

- **Go**: 1.26 (see `go.mod`)
- **PostgreSQL**: required for a working API (auth/CMS/RBAC)

## Quick start (local)

```bash
cp .env.example .env
# edit .env: set DB_* to your Postgres, change JWT_SECRET

make deps
make migrate-up
make seed
make run
```

- **Base URL**: `http://localhost:<APP_PORT>` (default `9090`)
- **Health**: `GET /health`
- **Readiness** (DB check): `GET /ready`
- **API spec**: `docs/openapi.yaml`

## Common commands

```bash
make help
make build
make test
make tidy

make migrate-up
make migrate-down
make migrate-version
make db-fresh   # dangerous: drops all tables

make seed
```

## Configuration (.env)

Primary reference is `.env.example` (all variables, with short notes). Highlights:

- **App**: `APP_NAME`, `APP_ENV`, `APP_PORT`
- **Database**: `DB_*` (host/port/name/user/password/sslmode)
- **JWT**: `JWT_SECRET`, `JWT_ACCESS_TTL`, `JWT_REFRESH_TTL`
- **RBAC bootstrap** (optional): `RBAC_BOOTSTRAP_ENABLED=true` enables `POST /v1/admin/bootstrap/claim-admin`
- **Media**: `MEDIA_STORAGE_DIR`, `MEDIA_PUBLIC_BASE_URL`, `MEDIA_MAX_UPLOAD_BYTES`
- **Rate limiting**: `RATE_LIMIT_ENABLED`, `RATE_LIMIT_*_MAX_PER_MINUTE`

## API overview

- **Auth**: `POST /v1/auth/register`, `/v1/auth/login`, `/v1/auth/refresh`
- **Public read** (no auth): `GET /v1/posts`, `/v1/posts/:slug`, `/v1/pages/:slug`, `/v1/menus/:slug`
- **Admin**: `/v1/admin/*` (JWT + permission checks)

Full list and schemas are in `docs/openapi.yaml`.

## RBAC: getting an admin user

RBAC defaults are seeded by `make seed` (see `docs/SEEDING.md`). After you have at least one registered user:

- **Recommended**: use the RBAC admin APIs (guarded by `rbac:manage`) to assign roles/permissions
- **Optional**: enable the one-time bootstrap endpoint with `RBAC_BOOTSTRAP_ENABLED=true` (see `cmd/api/main.go` and `docs/PHASES.md`)

## Repository layout

```text
cmd/          entrypoints (api, migrate, seed)
internal/     app code (config, platform, modules, server wiring)
migrations/   timestamped SQL migrations
pkg/          shared libraries used by entrypoints
deploy/       nginx + systemd templates
docs/         roadmap, deployment, seeding, OpenAPI
scripts/      deploy + local run helpers
Makefile      developer commands
```

## Git workflow and deployment

- **Branches**: `dev` → `staging` → `main` (promotion by merge + push)
- **Workflow**: `docs/GIT_FLOW.md`
- **Deployment hub**: `docs/DEPLOYMENT.md` and `docs/deployment/*.md`

## Documentation index

Start at `docs/README.md`.
