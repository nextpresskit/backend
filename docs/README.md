# Documentation

This page is the **table of contents** for the repo. If you only read one extra file after the root readme, read **[../README.md § Where to go next](../README.md#where-to-go-next)** — it points you here with a goal-based menu.

---

## Pick your path

| You are… | Read first | Then |
|----------|------------|------|
| **New to the project** | [../README.md](../README.md) (install + `setup` / `run`) | [Local HTTPS / nginx](deployment/local.md) only if you need browser cookies against HTTPS |
| **Unsure which command to run** | [COMMANDS.md](COMMANDS.md) | [../README.md](../README.md#commands-summary) for the short matrix |
| **Deploying to production** | [DEPLOYMENT.md](DEPLOYMENT.md) | [SECURITY.md](SECURITY.md), [production.ssl-snippet.conf.example](../deploy/nginx/production.ssl-snippet.conf.example) |
| **Developing locally (advanced)** | [deployment/local.md](deployment/local.md) | [macOS notes](deployment/macos.md) if applicable |
| **Calling the HTTP API** | [openapi.yaml](openapi.yaml) | [postman-templates/README.md](../postman-templates/README.md), `postman-sync` |
| **Elasticsearch in production** | [ELASTICSEARCH_OPERATIONS.md](ELASTICSEARCH_OPERATIONS.md) | — |
| **Contributing code** | [../CONTRIBUTING.md](../CONTRIBUTING.md) | [TODO.md](TODO.md), update [openapi.yaml](openapi.yaml) when REST changes |

---

## Full doc list (by type)

Docs follow [Diátaxis](https://diataxis.fr/) ideas: tutorials, how-tos, reference, explanation.

| Type | Document | Purpose |
|------|----------|---------|
| **Tutorial** | [Root `README.md`](../README.md) | Fast path: `./scripts/nextpresskit setup`, `make setup`, or `.\scripts\nextpresskit.ps1 setup`. |
| **How-to** | [Deployment (servers)](DEPLOYMENT.md) | Ubuntu, Nginx, systemd, HTTPS (Certbot), interactive `scripts/deploy`. |
| **Reference** | [Command reference](COMMANDS.md) | What each command does and when to use it. |
| **How-to** | [Local development](deployment/local.md) | Laptop setup, HTTPS with mkcert, optional Nginx/systemd, ES/GraphQL, tests. |
| **How-to** | [Local development (macOS)](deployment/macos.md) | Homebrew, paths, mkcert, Nginx on Mac. |
| **How-to** | [Elasticsearch operations runbook](ELASTICSEARCH_OPERATIONS.md) | Index templates, upgrades/reindex, multi-cluster operations. |
| **Explanation** | [API versioning strategy](API_VERSIONING.md) | Current decision (`API_BASE_PATH`) and migration paths (URL/header versioning). |
| **How-to** | [Database seeding](SEEDING.md) | Run seeders, RBAC defaults, permission codes. |
| **How-to** | [Security and hardening](SECURITY.md) | CVE review, CORS policy, rate-limit tuning, JWT rotation guidance. |
| **Reference** | [`openapi.yaml`](openapi.yaml) | REST paths, request/response schemas. |
| **Reference** | [`internal/graphql/schema.graphqls`](../internal/graphql/schema.graphqls) | GraphQL schema (not in OpenAPI). |
| **Explanation** | [GraphQL vs REST](../README.md#graphql-vs-rest) | Contract boundary: REST-first, GraphQL optional/read-focused. |
| **Reference** | [Root `CHANGELOG.md`](../CHANGELOG.md) | Release notes process and unreleased entries. |
| **Explanation** | [ADR folder](adr/README.md) | Architecture decision records process and conventions. |
| **Reference** | [`.env.example`](../.env.example) | All environment variables. |
| **Explanation** | [Roadmap](ROADMAP.md) | What is shipped, current themes, future direction. |
| **Task list** | [TODO](TODO.md) | Full shipped/`[ ]` open checklist; keep in sync with code. |

**Contributors:** [Contributing guide](../CONTRIBUTING.md).

## How these docs connect (simple flow)

1. **[../README.md](../README.md)** — run the app; command cheat sheet; JWT overview.
2. **This index** — find the right how-to or reference.
3. **Command details** — [COMMANDS.md](COMMANDS.md) for plain-language command descriptions.
4. **Deploy paths** — laptop: [deployment/local.md](deployment/local.md) + [macOS](deployment/macos.md); server: [DEPLOYMENT.md](DEPLOYMENT.md).
5. **Scope / planning** — [ROADMAP.md](ROADMAP.md) (themes) and [TODO.md](TODO.md) (checklist).

- **Track work:** edit [`TODO.md`](TODO.md) with the code; refresh [`ROADMAP.md`](ROADMAP.md) when themes change ([`CONTRIBUTING.md`](../CONTRIBUTING.md)).

## Machine-readable API

Import [`openapi.yaml`](openapi.yaml) into Postman, Stoplight, or your gateway. It documents **Bearer** and **cookie** JWT security for protected routes. Regenerate GraphQL code after schema edits: `make graphql`.

Ready-made collections: [`postman-templates/`](../postman-templates) (see [`postman-templates/README.md`](../postman-templates/README.md) for `jwt_auth_source` / cookie jar vs header mode). Local gitignored `postman/`: `./scripts/nextpresskit postman-sync` or `make postman-sync` (optional `--dry-run`; tier URLs: `POSTMAN_*_BASE_URL`).

## Config templates

- `deploy/systemd/nextpresskit-backend@.service`
- `deploy/nginx/*.conf`
- `deploy/nginx/production.ssl-snippet.conf.example` (optional manual TLS)
- `scripts/deploy` (bash) / `scripts/deploy.ps1` — interactive wizard → `deploy/generated/` (see [DEPLOYMENT.md § Interactive deploy](DEPLOYMENT.md#interactive-deploy-scriptsdeploy))

Instructions live in [DEPLOYMENT.md](DEPLOYMENT.md) and [deployment/local.md](deployment/local.md).
