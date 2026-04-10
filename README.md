# NextPress Backend

Headless-style **CMS HTTP API** in Go: **REST** ([OpenAPI](docs/openapi.yaml)), optional **GraphQL**, **PostgreSQL**, **JWT** auth, **RBAC** on admin routes.

**Docs:** [`docs/README.md`](docs/README.md) (index) · **Checklist:** [`docs/TODO.md`](docs/TODO.md) · **Direction:** [`docs/ROADMAP.md`](docs/ROADMAP.md) · **Contributing:** [`CONTRIBUTING.md`](CONTRIBUTING.md)

## Stack

| | |
|--|--|
| Layout | Modular monolith - `internal/modules/*` |
| HTTP | Gin |
| Persistence | GORM + SQL migrations (`migrations/`, `cmd/migrate`) |
| Plugins | Registry + post-save hooks - status: [roadmap](docs/ROADMAP.md), tasks: [TODO](docs/TODO.md#plugins) |

## Requirements

- Go ≥ 1.26 ([`go.mod`](go.mod))
- PostgreSQL

## Quick start

```bash
cp .env.example .env   # set DB_* and JWT_SECRET

make deps
make migrate-up
make seed
make run
```

| | |
|--|--|
| API | `http://localhost:9090` (`APP_PORT`) |
| Health | `GET /health` · `GET /ready` |
| REST spec | [`docs/openapi.yaml`](docs/openapi.yaml) |

## Makefile

```bash
make help
make build run test tidy deps
make migrate-up migrate-down migrate-version
make db-fresh          # destructive
make seed
make graphql           # after editing internal/graphql/schema.graphqls
```

Configuration: [`.env.example`](.env.example). Optional Elasticsearch / GraphQL notes: [`docs/deployment/local.md`](docs/deployment/local.md).

## API surface (summary)

- **Auth:** `POST /v1/auth/register`, `/login`, `/refresh`
- **Public:** posts, pages, menus (and search when Elasticsearch is enabled)
- **GraphQL:** if enabled - `GRAPHQL_PATH` (default `/v1/graphql`)
- **Admin:** `/v1/admin/*` - JWT + permissions

Details: OpenAPI and source.

## RBAC

[`make seed`](docs/SEEDING.md) loads default roles and permissions. Assign `admin` via RBAC APIs or optional bootstrap (`RBAC_BOOTSTRAP_ENABLED`) - see [roadmap](docs/ROADMAP.md).

## Repository layout

```text
cmd/          api, migrate, seed
internal/     config, platform, graphql, server, modules
migrations/   SQL
pkg/          shared libraries
deploy/       nginx, systemd templates
docs/         guides, OpenAPI (see docs/README.md)
scripts/      deploy, local run
```

## Deployment

Servers and Git flow: [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md). Local deep-dive: [`docs/deployment/local.md`](docs/deployment/local.md).
