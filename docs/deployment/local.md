# Local development

**Tutorial / how-to** - run the API on your machine without Nginx or systemd.

## Prerequisites

- Go (`go.mod`)
- PostgreSQL
- Git

## Setup

```bash
git clone <repo-url> nextpress-backend && cd nextpress-backend
go mod download
cp .env.example .env
```

Configure `DB_*`, `JWT_SECRET`, `APP_ENV=local`. Optional flags: Elasticsearch, GraphQL - [`.env.example`](../../.env.example).

## Migrate, seed, run

```bash
make migrate-up
make seed
./scripts/run_local.sh   # or: go run ./cmd/api
```

| | |
|--|--|
| Base URL | `http://localhost:<APP_PORT>` (default 9090) |
| Checks | `GET /health`, `GET /ready` |
| REST | [`docs/openapi.yaml`](../openapi.yaml) |

## Optional: Elasticsearch

`ELASTICSEARCH_ENABLED=true`, `ELASTICSEARCH_URLS`. With `APP_ENV` `local` or `dev`, index auto-create defaults on. `GET /v1/posts/search` → 501 if disabled. Reindex: `POST /v1/admin/posts/search/reindex` (`posts:write`).

## Optional: GraphQL

`GRAPHQL_ENABLED=true`; playground only for `APP_ENV` `local` or `dev`. After editing `internal/graphql/schema.graphqls`: `make graphql`.

## Tests

```bash
go test ./...
go vet ./...
```

---

**See also:** [Documentation index](../README.md) · [Server deployment](../DEPLOYMENT.md) · [Seeding](../SEEDING.md) · [Roadmap](../ROADMAP.md) · [TODO](../TODO.md) (optional ES / GraphQL / tests **`[x]`** / **`[ ]`**)
