# Next Press Kit

Next Press Kit is a starter kit for building modern backend APIs using Go, Gin, and PostgreSQL.

The goal of this project is to give developers a strong starting point they can clone and build on, with common product needs already in place: authentication handling, content creation flows, and an administration area.

* Website: nextpresskit.com
* Frontend repository: nextpresskit/web

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

## Scripts

```bash
make run
make build
make test
make test-coverage
make migrate-up
make migrate-down
make migrate-version
make seed
make graphql
make security-check
make db-fresh
```

## Frontend Integration

Next Press Kit backend is designed to work with the frontend web project:

* Frontend repo: <https://github.com/nextpresskit/web>
* API responsibilities include authentication, content APIs, and admin-related backend operations.
* This project can also be used separately with a different frontend or mock/local API consumers.

## API Contract

This project includes an OpenAPI-first API contract and optional GraphQL support.

* OpenAPI spec lives in `docs/openapi.yaml`.
* REST endpoints cover auth, public content, and admin operations.
* GraphQL is optional and can be enabled via environment configuration.
* API base path is configurable using `API_BASE_PATH`.

## Documentation

This project includes documentation for setup, architecture, and operations.

* Docs index: `docs/README.md`
* API versioning: `docs/API_VERSIONING.md`
* Seeding guide: `docs/SEEDING.md`
* Local deployment: `docs/deployment/local.md`
* Full deployment guide: `docs/DEPLOYMENT.md`
* Roadmap: `docs/ROADMAP.md`
* TODO checklist: `docs/TODO.md`

## About

Next Press Kit backend starter for Go APIs.
