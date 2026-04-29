# Next Press Kit

Next Press Kit is a starter kit for building modern backend APIs using Go, Gin, and PostgreSQL.

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

Install dependencies and run the app locally:

```bash
cp .env.example .env
make deps
make migrate-up
make seed
make run
```

The development server runs on `http://localhost:9090`.

To run without `Ctrl+C` terminal interrupt status noise, use background mode:

```bash
make start
make stop
```

## Scripts

Common development commands:

```text
make run              Start the API in the foreground with go run.
make start            Start the API in the background and write logs to .tmp/nextpress-api.log.
make stop             Stop the background API started with make start.
make build            Build the API binary into bin/server.
make test             Run the full test suite with verbose output.
make test-coverage    Run tests with coverage summary.
make migrate-up       Apply all pending database migrations.
make migrate-down     Roll back one database migration.
make migrate-version  Print the current migration version and dirty flag.
make seed             Run database seeders.
make graphql          Regenerate gqlgen code from the GraphQL schema.
make security-check   Run dependency vulnerability scanning with govulncheck.
make db-fresh         Drop and recreate the database schema by rerunning migrations.
```

Run `make help` for the complete command reference.

## Frontend Integration

Next Press Kit backend is designed to work with the frontend web project:

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

Next Press Kit backend starter for Go APIs.
