# Documentation

This page lists all documentation in one place. If you opened one extra file after the root readme, use [Where to go next](../README.md#where-to-go-next) as the short menu; it points here when you need more detail.

---

## Pick your path

| You are… | Read first | Then |
|----------|------------|------|
| New to the project | [../README.md](../README.md) (install, setup, run) | [Local HTTPS / nginx](deployment/local.md) if you need browser cookies over HTTPS |
| Which command do I run? | [COMMANDS.md](COMMANDS.md) | [Database and seed](COMMANDS.md#database-and-seed-data); [command cheat sheet](../README.md#commands-summary) |
| Deploying to production | [DEPLOYMENT.md](DEPLOYMENT.md) | [SECURITY.md](SECURITY.md), [production.ssl-snippet.conf.example](../deploy/nginx/production.ssl-snippet.conf.example) |
| Developing locally (advanced) | [deployment/local.md](deployment/local.md) | [macOS notes](deployment/macos.md) if applicable |
| Calling the HTTP API | [openapi.yaml](openapi.yaml) | [postman-templates/README.md](../postman-templates/README.md), postman-sync |
| Elasticsearch in production | [ELASTICSEARCH_OPERATIONS.md](ELASTICSEARCH_OPERATIONS.md) | — |
| Contributing code | [../CONTRIBUTING.md](../CONTRIBUTING.md) | [TODO.md](TODO.md), update [openapi.yaml](openapi.yaml) when REST changes |

---

## Full doc list (by type)

Docs follow [Diátaxis](https://diataxis.fr/) ideas: tutorials, how-tos, reference, explanation.

| Type | Document | Purpose |
|------|----------|---------|
| Tutorial | [Root README](../README.md) | Fast path: `./scripts/nextpresskit setup`, `make setup`, or PowerShell `nextpresskit.ps1 setup`. |
| How-to | [Deployment (servers)](DEPLOYMENT.md) | Ubuntu, Nginx, systemd, HTTPS (Certbot), interactive scripts/deploy. |
| Reference | [Command reference](COMMANDS.md) | What each command does and when to use it. |
| How-to | [Local development](deployment/local.md) | Laptop setup, HTTPS with mkcert, optional Nginx/systemd, ES/GraphQL, tests. |
| How-to | [Local development (macOS)](deployment/macos.md) | Homebrew, paths, mkcert, Nginx on Mac. |
| How-to | [Elasticsearch operations runbook](ELASTICSEARCH_OPERATIONS.md) | Index templates, upgrades/reindex, multi-cluster operations. |
| Explanation | [API versioning strategy](API_VERSIONING.md) | Current decision (`API_BASE_PATH`) and migration paths (URL/header versioning). |
| How-to | [Database seeding](SEEDING.md) | What demo data looks like; superadmin env vars. Commands: [COMMANDS.md](COMMANDS.md#database-and-seed-data). |
| How-to | [Security and hardening](SECURITY.md) | CVE review, CORS policy, rate-limit tuning, JWT rotation guidance. |
| Reference | [openapi.yaml](openapi.yaml) | REST paths, request/response schemas. |
| Reference | [internal/graphql/schema.graphqls](../internal/graphql/schema.graphqls) | GraphQL schema (not in OpenAPI). |
| Explanation | [GraphQL vs REST](../README.md#graphql-vs-rest) | Contract boundary: REST-first, GraphQL optional/read-focused. |
| Reference | [CHANGELOG](../CHANGELOG.md) | Release notes process and unreleased entries. |
| Explanation | [ADR folder](adr/README.md) | Architecture decision records process and conventions. |
| Reference | [.env.example](../.env.example) | All environment variables. |
| Explanation | [Roadmap](ROADMAP.md) | What is shipped, current themes, future direction. |
| Task list | [TODO](TODO.md) | Full checklist; keep in sync with code. |

Contributors: [Contributing guide](../CONTRIBUTING.md).

## How these docs connect (simple flow)

1. [README](../README.md): run the app, command cheat sheet, JWT overview.
2. This index: pick the right how-to or reference.
3. [COMMANDS.md](COMMANDS.md): all commands, including [database and seed](COMMANDS.md#database-and-seed-data).
4. Deploy: laptop [local.md](deployment/local.md) and [macos.md](deployment/macos.md); server [DEPLOYMENT.md](DEPLOYMENT.md).
5. Planning: [ROADMAP.md](ROADMAP.md) and [TODO.md](TODO.md).

Track work in [TODO.md](TODO.md) with the code; refresh [ROADMAP.md](ROADMAP.md) when themes change ([CONTRIBUTING.md](../CONTRIBUTING.md)).

## Machine-readable API

Import [`openapi.yaml`](openapi.yaml) into Postman, Stoplight, or your gateway. It documents Bearer and cookie JWT security for protected routes. Regenerate GraphQL code after schema edits: `make graphql`.

Ready-made collections: [`postman-templates/`](../postman-templates) (see [`postman-templates/README.md`](../postman-templates/README.md) for jwt_auth_source and cookie jar vs header mode). Local gitignored `postman/`: `./scripts/nextpresskit postman-sync` or `make postman-sync` (optional `--dry-run`; tier URLs such as `POSTMAN_*_BASE_URL`).

## Config templates

- `deploy/systemd/nextpresskit-backend@.service`
- `deploy/nginx/*.conf`
- `deploy/nginx/production.ssl-snippet.conf.example` (optional manual TLS)
- `scripts/deploy` (bash) and `scripts/deploy.ps1`: interactive wizard writes `deploy/generated/` (see [DEPLOYMENT.md](DEPLOYMENT.md#interactive-deploy-scriptsdeploy))

Instructions live in [DEPLOYMENT.md](DEPLOYMENT.md) and [deployment/local.md](deployment/local.md).
