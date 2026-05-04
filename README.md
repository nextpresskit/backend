# NextPressKit

NextPressKit is a starter kit for building modern backend APIs using Go, Gin, and PostgreSQL.

The goal of this project is to give developers a strong starting point they can clone and build on, with common product needs already in place: authentication handling, content creation flows, and an administration area.

* Website: [nextpresskit.com](https://nextpresskit.com)
* Frontend repository: [nextpresskit/web](https://github.com/nextpresskit/web)

## Where to go next

Use this table if you're not sure which doc to open first.

| Goal | Start here |
|------|------------|
| Understand what each command does | [docs/COMMANDS.md](./docs/COMMANDS.md) |
| See every doc and how it fits together | [docs/README.md](./docs/README.md) |
| Deploy to Ubuntu (nginx, systemd, HTTPS) | [docs/DEPLOYMENT.md](./docs/DEPLOYMENT.md) |
| HTTPS + nginx on your laptop (mkcert) | [docs/deployment/local.md](./docs/deployment/local.md) · [macOS](./docs/deployment/macos.md) |
| REST request/response shapes | [docs/openapi.yaml](./docs/openapi.yaml) |
| Try endpoints in Postman | [postman-templates/README.md](./postman-templates/README.md) (`postman-sync`) |
| JWT cookies, CORS, hardening | [docs/SECURITY.md](./docs/SECURITY.md) |
| Database (migrate, seed, fresh start) | [docs/COMMANDS.md](./docs/COMMANDS.md#database-and-seed-data) · [docs/SEEDING.md](./docs/SEEDING.md) |
| Contribute or run checks before a PR | [CONTRIBUTING.md](./CONTRIBUTING.md) |

## Project Concepts

* Starter-first architecture for bootstrapping and customization.
* Auth-ready foundations (patterns and services for sign-in and tokens).
* Content-oriented workflows (creation and publishing APIs).
* Admin routes and structures for managing app data.
* Modern API stack: Go, Gin, GORM.

## Tech Stack

* Go
* Gin
* PostgreSQL
* GORM
* JWT
* gqlgen
* Prometheus

## Quick start

You need Go (see `go.mod`) and PostgreSQL. Start the server, then create an empty database (and user if needed) that match `DB_*` in [.env.example](./.env.example), for example database `nextpresskit` and user `postgres` on `localhost:5432`.

Running `./scripts/nextpresskit setup` (or `make setup`) first runs `install`, which copies `.env.example` to `.env` if `.env` is missing. Edit at least `JWT_SECRET` and double-check `DB_*` before the migrate step runs.

Quick path:

1. `./scripts/nextpresskit setup` (or `make setup`) — modules, `.env` if needed, build, migrate, seed
2. `./scripts/nextpresskit run` (or `make run`)
3. Open `http://localhost:9090/health` (replace `9090` with `APP_PORT` from `.env` if you changed it). Use `/ready` if you want to confirm PostgreSQL is wired up.

More database commands (`migrate-up`, `seed`, `db-fresh`): [docs/COMMANDS.md](./docs/COMMANDS.md#database-and-seed-data) and [docs/SEEDING.md](./docs/SEEDING.md).

### Linux / macOS / Git Bash (same commands)

Copy-paste:

```bash
./scripts/nextpresskit setup
./scripts/nextpresskit run
```

Or with Make (thin wrappers around the same scripts):

```bash
make setup
make run
```

From an interactive terminal, `setup` may also run `scripts/setup-local-https.sh`: it tries to install mkcert (apt, dnf, pacman, zypper, or Homebrew), generates certs under `~/.local/share/nextpresskit-ssl/`, can print a `/etc/hosts` hint, and on Linux with nginx may run `deploy apply-nginx --no-tls-menu`. Set `SKIP_SETUP_LOCAL_HTTPS=1` to skip (CI and headless).

The API listens on `APP_PORT` (default 9090). Foreground `run` frees the port if another same-repo `bin/server` or `go run ./cmd/api` is still listening; systemd units named `nextpresskit-backend@*` are left alone (stop with `systemctl`).

### Windows (PowerShell)

From the repo root:

```powershell
.\scripts\nextpresskit.ps1 setup
.\scripts\nextpresskit.ps1 run
```

Interactive Nginx snippet wizard: `make deploy-ps` or `.\scripts\nextpresskit.ps1 deploy`. Full Linux server flows are still in `make deploy` / `bash scripts/deploy`.

### HTTPS / Nginx locally

For HTTPS (browser cookie auth) and reverse-proxy setup, see [docs/deployment/local.md](./docs/deployment/local.md) and [docs/deployment/macos.md](./docs/deployment/macos.md). Use `make deploy` (bash) or `make deploy-ps` (PowerShell) to generate configs under `deploy/generated/`.

Background mode (Unix): `make start` / `make stop` or `./scripts/nextpresskit start` / `stop`.

## Commands (summary)

Need command-by-command explanations? Open [docs/COMMANDS.md](./docs/COMMANDS.md).

Most common confusion: `setup` is for local bootstrap; `deploy` is for deployment/config and release flows.

| Area | Unix CLI | Make | Windows PowerShell |
|------|----------|------|----------------------|
| Bootstrap | `./scripts/nextpresskit setup` | `make setup` | `.\scripts\nextpresskit.ps1 setup` |
| Modules + `.env` | `./scripts/nextpresskit install` | `make install` | `.\scripts\nextpresskit.ps1 install` |
| Build API only | `./scripts/nextpresskit build` | `make build` | `.\scripts\nextpresskit.ps1 build` |
| Build API + migrate + seed tools | `./scripts/nextpresskit build-all` | `make build-all` | `.\scripts\nextpresskit.ps1 build-all` |
| Database | `./scripts/nextpresskit migrate-up`, `seed`, `db-fresh` | `make migrate-up`, `make seed`, `make db-fresh` | same on `nextpresskit.ps1` (after `db-fresh`, run `seed` for demo data) |
| Run API | `./scripts/nextpresskit run` | `make run` | `.\scripts\nextpresskit.ps1 run` |
| CI-style checks | `./scripts/nextpresskit checks` | `make checks` | `.\scripts\nextpresskit.ps1 checks` |
| Deploy wizard | `./scripts/nextpresskit deploy` | `make deploy` | `make deploy-ps` or `nextpresskit.ps1 deploy` |
| Postman env files | `./scripts/nextpresskit postman-sync` | `make postman-sync` | `.\scripts\nextpresskit.ps1 postman-sync` |

Run `./scripts/nextpresskit help` or `make help` for the full list.

Postman templates live under [postman-templates/](./postman-templates/). Run `./scripts/nextpresskit postman-sync` to create a gitignored [postman/](./postman/) workspace. Options: `--dry-run`, tier URLs such as `POSTMAN_DEV_BASE_URL`. Details: [postman-templates/README.md](./postman-templates/README.md).

## Frontend Integration

The NextPressKit backend is designed to work with the frontend web project:

* Frontend repo: [nextpresskit/web](https://github.com/nextpresskit/web)
* Backend repo: [nextpresskit/backend](https://github.com/nextpresskit/backend)
* API responsibilities include authentication, content APIs, and admin-related backend operations.
* This project can also be used separately with a different frontend or mock/local API consumers.

## API contract

This project includes an OpenAPI-first API contract and optional GraphQL support.

* OpenAPI spec: [docs/openapi.yaml](./docs/openapi.yaml).
* REST endpoints cover auth, public content, and admin operations.
* API base path is configurable with `API_BASE_PATH` (see [.env.example](./.env.example)).

### GraphQL vs REST

**REST** (OpenAPI) is the primary contract for writes and most product flows. **GraphQL** is optional and intended for read-focused use when you enable it in configuration. Regenerate GraphQL code after schema changes: `make graphql`. Details: [docs/README.md](./docs/README.md).

### Authentication (JWT)

* **`JWT_AUTH_SOURCE=cookie` (default):** access and refresh tokens are issued as **HttpOnly** cookies (`JWT_ACCESS_COOKIE_NAME`, `JWT_REFRESH_COOKIE_NAME`). Login and refresh return **`user`** only in JSON. Protected routes accept the access cookie or, if you switch the server to header mode, a Bearer token.
* **`JWT_AUTH_SOURCE=header`:** tokens are returned in JSON (`tokens` + `user` on login/refresh); send **`Authorization: Bearer <access_jwt>`** to protected routes.

Cross-site browser apps must set **`CORS_ORIGINS`** to the real frontend origin and use **`credentials: 'include'`**. See [docs/SECURITY.md](./docs/SECURITY.md) and [.env.example](./.env.example).

## Documentation

The single map of all docs is [docs/README.md](./docs/README.md). Skim that page whenever you feel lost.

Common links: [API versioning](./docs/API_VERSIONING.md) · [Seeding](./docs/SEEDING.md) · [Elasticsearch runbook](./docs/ELASTICSEARCH_OPERATIONS.md) · [Roadmap](./docs/ROADMAP.md) · [Task checklist](./docs/TODO.md) · [Changelog](./CHANGELOG.md)
