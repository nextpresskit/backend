# NextPressKit

NextPressKit is a starter kit for building modern backend APIs using Go, Gin, and PostgreSQL.

The goal of this project is to give developers a strong starting point they can clone and build on, with common product needs already in place: authentication handling, content creation flows, and an administration area.

* Website: [nextpresskit.com](https://nextpresskit.com)
* Frontend repository: [nextpresskit/web](https://github.com/nextpresskit/web)

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

## Getting Started

You need **Go** (see `go.mod`) and **PostgreSQL** reachable with the credentials in `.env`.

### Linux / macOS / Git Bash (same commands)

One-shot local bootstrap (modules, `.env` if missing, binaries, migrate, seed), then run the API:

```bash
./scripts/nextpress setup
./scripts/nextpress run
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
.\scripts\nextpress.ps1 setup
.\scripts\nextpress.ps1 run
```

Interactive Nginx snippet wizard: `make deploy-ps` or `.\scripts\nextpress.ps1 deploy`. Full Linux server flows remain in `make deploy` / `bash scripts/deploy`.

### HTTPS / Nginx locally

For **HTTPS** (cookie auth in the browser) and reverse-proxy setup, see [docs/deployment/local.md](docs/deployment/local.md) and [docs/deployment/macos.md](docs/deployment/macos.md). Use **`make deploy`** (bash) or **`make deploy-ps`** (PowerShell) to generate configs under `deploy/generated/`.

Background mode (Unix): `make start` / `make stop` or `./scripts/nextpress start` / `stop`.

## Commands (summary)

| Area | Unix CLI | Make | Windows PowerShell |
|------|----------|------|----------------------|
| Bootstrap | `./scripts/nextpress setup` | `make setup` | `.\scripts\nextpress.ps1 setup` |
| Modules + `.env` | `./scripts/nextpress install` | `make install` | `.\scripts\nextpress.ps1 install` |
| Build API only | `./scripts/nextpress build` | `make build` | `.\scripts\nextpress.ps1 build` |
| Build API + migrate + seed tools | `./scripts/nextpress build-all` | `make build-all` | `.\scripts\nextpress.ps1 build-all` |
| Migrate / seed | `./scripts/nextpress migrate-up` / `seed` | `make migrate-up` / `make seed` | same subcommands on `nextpress.ps1` |
| Run API | `./scripts/nextpress run` | `make run` | `.\scripts\nextpress.ps1 run` |
| CI-style checks | `./scripts/nextpress checks` | `make checks` | `.\scripts\nextpress.ps1 checks` |
| Deploy wizard | `./scripts/nextpress deploy` | `make deploy` | `make deploy-ps` or `nextpress.ps1 deploy` |

Run **`./scripts/nextpress help`** or **`make help`** for the full list.

## Frontend Integration

The NextPressKit backend is designed to work with the frontend web project:

* Frontend repo: [nextpresskit/web](https://github.com/nextpresskit/web)
* Backend repo: [nextpresskit/backend](https://github.com/nextpresskit/backend)
* API responsibilities include authentication, content APIs, and admin-related backend operations.
* This project can also be used separately with a different frontend or mock/local API consumers.

## API Contract

This project includes an OpenAPI-first API contract and optional GraphQL support.

* OpenAPI spec lives in [docs/openapi.yaml](https://github.com/nextpresskit/backend/blob/main/docs/openapi.yaml).
* REST endpoints cover auth, public content, and admin operations.
* GraphQL is optional and can be enabled via environment configuration.
* API base path is configurable using `API_BASE_PATH`.

### Authentication (JWT)

* **`JWT_AUTH_SOURCE=cookie` (default):** access and refresh tokens are issued as **HttpOnly** cookies (`JWT_ACCESS_COOKIE_NAME`, `JWT_REFRESH_COOKIE_NAME`). Login and refresh return **`user`** only in JSON. Protected routes accept the access cookie or, if you switch the server to header mode, a Bearer token.
* **`JWT_AUTH_SOURCE=header`:** tokens are returned in JSON (`tokens` + `user` on login/refresh); send **`Authorization: Bearer <access_jwt>`** to protected routes.

Cross-site browser apps must set **`CORS_ORIGINS`** to the real frontend origin and use **`credentials: 'include'`**. See [docs/SECURITY.md](docs/SECURITY.md) and [`.env.example`](.env.example).

## Documentation

This project includes documentation for setup, architecture, and operations.

* Docs index: [docs/README.md](https://github.com/nextpresskit/backend/blob/main/docs/README.md)
* API versioning: [docs/API_VERSIONING.md](https://github.com/nextpresskit/backend/blob/main/docs/API_VERSIONING.md)
* Seeding guide: [docs/SEEDING.md](https://github.com/nextpresskit/backend/blob/main/docs/SEEDING.md)
* Local deployment: [docs/deployment/local.md](https://github.com/nextpresskit/backend/blob/main/docs/deployment/local.md)
* Full deployment guide: [docs/DEPLOYMENT.md](https://github.com/nextpresskit/backend/blob/main/docs/DEPLOYMENT.md)
* Roadmap: [docs/ROADMAP.md](https://github.com/nextpresskit/backend/blob/main/docs/ROADMAP.md)
* TODO checklist: [docs/TODO.md](https://github.com/nextpresskit/backend/blob/main/docs/TODO.md)

## About

NextPressKit backend starter for Go APIs.
