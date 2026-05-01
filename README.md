# NextPressKit

NextPressKit is a starter kit for building modern backend APIs using Go, Gin, and PostgreSQL.

The goal of this project is to give developers a strong starting point they can clone and build on, with common product needs already in place: authentication handling, content creation flows, and an administration area.

* Website: [nextpresskit.com](https://nextpresskit.com)
* Frontend repository: [nextpresskit/web](https://github.com/nextpresskit/web)

## Where to go next

Use this table if you're not sure which doc to open first.

| Goal | Start here |
|------|------------|
| Run the API locally in a few minutes | [Getting started](#getting-started) |
| Understand what each command does | [docs/COMMANDS.md](docs/COMMANDS.md) |
| See every doc and how it fits together | [docs/README.md](docs/README.md) |
| Deploy to Ubuntu (nginx, systemd, HTTPS) | [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) |
| HTTPS + nginx on your laptop (mkcert) | [docs/deployment/local.md](docs/deployment/local.md) · [macOS](docs/deployment/macos.md) |
| REST request/response shapes | [docs/openapi.yaml](docs/openapi.yaml) |
| Try endpoints in Postman | [postman-templates/README.md](postman-templates/README.md) (`postman-sync`) |
| JWT cookies, CORS, hardening | [docs/SECURITY.md](docs/SECURITY.md) |
| Seed data and RBAC defaults | [docs/SEEDING.md](docs/SEEDING.md) |
| Contribute or run checks before a PR | [CONTRIBUTING.md](CONTRIBUTING.md) |

## Project Concepts

* **Starter-first architecture**: designed for rapid project bootstrapping and customization.
* **Auth-ready foundations**: includes backend auth integration patterns and services.
* **Content-oriented workflows**: supports content creation and publishing APIs.
* **Admin capabilities**: includes administration routes and structures to manage app data.
* **Modern API stack**: Go + Gin + GORM for fast, consistent backend development.

## Tech Stack

* Go
* Gin
* PostgreSQL
* GORM
* JWT
* gqlgen
* Prometheus

## Getting started

You need **Go** (see `go.mod`) and **PostgreSQL** reachable with the credentials in `.env`. Copy [`.env.example`](.env.example) to `.env` if you do not have one yet (`setup` can create it).

Quick path:

1. `./scripts/nextpresskit setup` (or `make setup`)
2. `./scripts/nextpresskit run` (or `make run`)
3. Open `http://localhost:9090/health` (or your `APP_PORT`)

### Linux / macOS / Git Bash (same commands)

One-shot local bootstrap (modules, `.env` if missing, binaries, migrate, seed), then run the API:

```bash
./scripts/nextpresskit setup
./scripts/nextpresskit run
```

Or with **Make** (thin wrappers around the same scripts):

```bash
make setup
make run
```

From an **interactive terminal**, `setup` also runs **`scripts/setup-local-https.sh`**: tries to **install mkcert** via apt, dnf, pacman, zypper, or Homebrew when missing, then uses **mkcert** to write `~/.local/share/nextpresskit-ssl/cert.pem` and `key.pem` for **nextpresskit.local**, **localhost**, **127.0.0.1**, and **::1** (avoids hostname/SAN mismatches), prints a **`/etc/hosts`** hint when needed, and on **Linux** with **nginx** runs **`deploy apply-nginx --no-tls-menu`**. Set **`SKIP_SETUP_LOCAL_HTTPS=1`** to skip this step (CI and headless installs).

The API listens on **`APP_PORT`** (default **9090**). Foreground `run` frees the port first if a previous **same-repo** `bin/server` or `go run ./cmd/api` is still listening; **systemd** units named `nextpresskit-backend@*` are detected and not killed (stop them with `systemctl`).

### Windows (PowerShell)

From the repo root:

```powershell
.\scripts\nextpresskit.ps1 setup
.\scripts\nextpresskit.ps1 run
```

Interactive Nginx snippet wizard: `make deploy-ps` or `.\scripts\nextpresskit.ps1 deploy`. Full Linux server flows are still in `make deploy` / `bash scripts/deploy`.

### HTTPS / Nginx locally

For **HTTPS** (cookie auth in the browser) and reverse-proxy setup, see [docs/deployment/local.md](docs/deployment/local.md) and [docs/deployment/macos.md](docs/deployment/macos.md). Use **`make deploy`** (bash) or **`make deploy-ps`** (PowerShell) to generate configs under `deploy/generated/`.

Background mode (Unix): `make start` / `make stop` or `./scripts/nextpresskit start` / `stop`.

## Commands (summary)

Need command-by-command explanations? Open [docs/COMMANDS.md](docs/COMMANDS.md).

| Area | Unix CLI | Make | Windows PowerShell |
|------|----------|------|----------------------|
| Bootstrap | `./scripts/nextpresskit setup` | `make setup` | `.\scripts\nextpresskit.ps1 setup` |
| Modules + `.env` | `./scripts/nextpresskit install` | `make install` | `.\scripts\nextpresskit.ps1 install` |
| Build API only | `./scripts/nextpresskit build` | `make build` | `.\scripts\nextpresskit.ps1 build` |
| Build API + migrate + seed tools | `./scripts/nextpresskit build-all` | `make build-all` | `.\scripts\nextpresskit.ps1 build-all` |
| Migrate / seed | `./scripts/nextpresskit migrate-up` / `seed` | `make migrate-up` / `make seed` | same subcommands on `nextpresskit.ps1` |
| Run API | `./scripts/nextpresskit run` | `make run` | `.\scripts\nextpresskit.ps1 run` |
| CI-style checks | `./scripts/nextpresskit checks` | `make checks` | `.\scripts\nextpresskit.ps1 checks` |
| Deploy wizard | `./scripts/nextpresskit deploy` | `make deploy` | `make deploy-ps` or `nextpresskit.ps1 deploy` |
| Postman env files | `./scripts/nextpresskit postman-sync` | `make postman-sync` | `.\scripts\nextpresskit.ps1 postman-sync` |

Run **`./scripts/nextpresskit help`** or **`make help`** for the full list.

Postman templates are tracked under [`postman-templates/`](postman-templates/). Run **`./scripts/nextpresskit postman-sync`** to create a gitignored [`postman/`](postman/) workspace (copy missing JSON from templates, then apply `.env.example` / `.env`). Options: `--dry-run`, tier overrides like `POSTMAN_DEV_BASE_URL=…`. Details: [`postman-templates/README.md`](postman-templates/README.md).

## Frontend Integration

The NextPressKit backend is designed to work with the frontend web project:

* Frontend repo: [nextpresskit/web](https://github.com/nextpresskit/web)
* Backend repo: [nextpresskit/backend](https://github.com/nextpresskit/backend)
* API responsibilities include authentication, content APIs, and admin-related backend operations.
* This project can also be used separately with a different frontend or mock/local API consumers.

## API contract

This project includes an OpenAPI-first API contract and optional GraphQL support.

* OpenAPI spec: [docs/openapi.yaml](docs/openapi.yaml).
* REST endpoints cover auth, public content, and admin operations.
* API base path is configurable with `API_BASE_PATH` (see [.env.example](.env.example)).

### GraphQL vs REST

**REST** (OpenAPI) is the primary contract for writes and most product flows. **GraphQL** is optional and intended for read-focused use when you enable it in configuration. Regenerate GraphQL code after schema changes: `make graphql`. Details: [docs/README.md](docs/README.md).

### Authentication (JWT)

* **`JWT_AUTH_SOURCE=cookie` (default):** access and refresh tokens are issued as **HttpOnly** cookies (`JWT_ACCESS_COOKIE_NAME`, `JWT_REFRESH_COOKIE_NAME`). Login and refresh return **`user`** only in JSON. Protected routes accept the access cookie or, if you switch the server to header mode, a Bearer token.
* **`JWT_AUTH_SOURCE=header`:** tokens are returned in JSON (`tokens` + `user` on login/refresh); send **`Authorization: Bearer <access_jwt>`** to protected routes.

Cross-site browser apps must set **`CORS_ORIGINS`** to the real frontend origin and use **`credentials: 'include'`**. See [docs/SECURITY.md](docs/SECURITY.md) and [`.env.example`](.env.example).

## Documentation

The **single map of all docs** is [docs/README.md](docs/README.md). Skim that page whenever you feel lost.

**Common links:** [API versioning](docs/API_VERSIONING.md) · [Seeding](docs/SEEDING.md) · [Elasticsearch runbook](docs/ELASTICSEARCH_OPERATIONS.md) · [Roadmap](docs/ROADMAP.md) · [Task checklist](docs/TODO.md) · [Changelog](CHANGELOG.md)
