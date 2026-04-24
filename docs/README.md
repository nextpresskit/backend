# Documentation

How this folder is organised (aligned with [Diátaxis](https://diataxis.fr/): tutorials, how-to, reference, explanation).

| Type | Document | Purpose |
|------|----------|---------|
| **Tutorial** | [Root `README.md`](../README.md) | Fast path: clone, configure, run locally. |
| **How-to** | [Deployment (servers)](DEPLOYMENT.md) | Ubuntu, Nginx, systemd, `scripts/deploy`, branches. |
| **How-to** | [Local development](deployment/local.md) | Laptop setup, optional Nginx/systemd, ES/GraphQL, tests. |
| **How-to** | [Elasticsearch operations runbook](ELASTICSEARCH_OPERATIONS.md) | Index templates, upgrades/reindex, multi-cluster operations. |
| **How-to** | [Database seeding](SEEDING.md) | Run seeders, RBAC defaults, permission codes. |
| **How-to** | [Security and hardening](SECURITY.md) | CVE review, CORS policy, rate-limit tuning, JWT rotation guidance. |
| **Reference** | [`openapi.yaml`](openapi.yaml) | REST paths, request/response schemas. |
| **Reference** | [`internal/graphql/schema.graphqls`](../internal/graphql/schema.graphqls) | GraphQL schema (not in OpenAPI). |
| **Explanation** | [GraphQL vs REST split](../README.md#graphql-vs-rest-split) | Contract boundary: REST-first, GraphQL optional/read-focused. |
| **Reference** | [Root `CHANGELOG.md`](../CHANGELOG.md) | Release notes process and unreleased entries. |
| **Explanation** | [ADR folder](adr/README.md) | Architecture decision records process and conventions. |
| **Reference** | [`.env.example`](../.env.example) | All environment variables. |
| **Explanation** | [Roadmap](ROADMAP.md) | What is shipped, current themes, future direction. |
| **Task list** | [TODO](TODO.md) | Full shipped/`[ ]` open checklist; keep in sync with code. |

**Contributors:** [Contributing guide](../CONTRIBUTING.md).

## How these docs connect

```text
../README.md  ─────────────►  quick start, stack, links into docs/
       │
       ▼
docs/README.md  (this page) ─►  map of every doc
       │
       ├── ROADMAP.md  ───────►  short “why / shipped / themes”
       ├── TODO.md  ──────────►  full [x] / [ ] checklist (source of truth for scope)
       ├── openapi.yaml  ─────►  REST reference
       ├── ELASTICSEARCH_OPERATIONS.md  ─►  ES templates, upgrades, multi-cluster runbook
       ├── DEPLOYMENT.md + deployment/local.md  ─►  run on server vs laptop
       ├── SEEDING.md  ───────►  RBAC seed + permission table
       └── CONTRIBUTING.md  ──►  when to update the above on a PR
```

- **Start here** if you are new: [root `README.md`](../README.md) → then come back to this index.
- **Track work:** edit [`TODO.md`](TODO.md) with the code; refresh [`ROADMAP.md`](ROADMAP.md) when themes change ([`CONTRIBUTING.md`](../CONTRIBUTING.md)).

## Machine-readable API

Import [`openapi.yaml`](openapi.yaml) into Postman, Stoplight, or your gateway. Regenerate GraphQL code after schema edits: `make graphql`.

## Config templates

- `deploy/systemd/nextpress-backend@.service`
- `deploy/nginx/*.conf`

Instructions live in [DEPLOYMENT.md](DEPLOYMENT.md).
