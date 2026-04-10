# Local

Run nextpress-backend on your machine for development or testing. No systemd or Nginx; you run the binary from the project folder.

| Item   | Value                    |
|--------|--------------------------|
| Folder | project root (any path)  |
| Branch | your branch              |

---

## Prerequisites

| Requirement   | Details |
|---------------|---------|
| **Go**        | Version in `go.mod`. |
| **PostgreSQL**| Required for a working API (auth, CMS, RBAC). Configure `DB_*` in `.env`. |
| **Git**       | To clone the repository. |

---

## 1. Clone and setup

```bash
git clone <repo-url> nextpress-backend
cd nextpress-backend
go mod download
cp .env.example .env
```

Edit `.env`:

- **`DB_*`** — point at your local Postgres database.
- **`APP_ENV=local`** — recommended local environment name (matches `.env.example` and `scripts/run_local.sh`).
- **`JWT_SECRET`** — set a strong, non-default value for any environment beyond short-lived local testing.
- Optional: **`RATE_LIMIT_*`**, **`MEDIA_*`**, **`RBAC_BOOTSTRAP_ENABLED`** — see `.env.example`.

---

## 2. Migrations and seed

```bash
make migrate-up   # apply all SQL migrations
make seed         # RBAC defaults (admin role + permissions)
```

---

## 3. Run the API

```bash
./scripts/run_local.sh
# or
go run ./cmd/api
```

Listen URL: `http://localhost:<APP_PORT>` (default **9090**).

**Smoke checks**

- `GET /health` — process up  
- `GET /ready` — database reachable  
- API reference: `docs/openapi.yaml`

---

## 4. Optional: Elasticsearch and GraphQL

These features are **off by default**. Enable them in `.env` when you need full-text search or GraphQL reads.

### Elasticsearch

- Set `ELASTICSEARCH_ENABLED=true` and `ELASTICSEARCH_URLS` (e.g. `http://localhost:9200`).
- With `APP_ENV=local` or `dev`, `ELASTICSEARCH_AUTO_CREATE_INDEX` defaults to **true** so `{ELASTICSEARCH_INDEX_PREFIX}_posts` is created on startup if missing. For production, prefer managing indices yourself and set `ELASTICSEARCH_AUTO_CREATE_INDEX=false`.
- **Search**: `GET /v1/posts/search?q=...` (public, same rate limit as other public routes). If Elasticsearch is disabled, the API returns **501** with `{"error":"search_disabled"}`.
- **Scheduled posts**: when a post is promoted to published by the background scheduler, it is synced to Elasticsearch the same way as saves (so search stays consistent).
- **Reindex**: `POST /v1/admin/posts/search/reindex` (JWT + `posts:write`) walks all published posts and upserts them into the search index. Returns **501** if Elasticsearch is disabled.
- Run Elasticsearch locally, for example:

```bash
docker run -d --name es -p 9200:9200 -e "discovery.type=single-node" -e "xpack.security.enabled=false" docker.elastic.co/elasticsearch/elasticsearch:8.12.2
```

### GraphQL

- Set `GRAPHQL_ENABLED=true`. Default path: `GRAPHQL_PATH=/v1/graphql` (`POST` and `GET` for queries).
- Optional UI: `GRAPHQL_PLAYGROUND_ENABLED=true` only takes effect when `APP_ENV` is **local** or **dev** (ignored on staging/production even if set).
- After changing `internal/graphql/schema.graphqls`, run **`make graphql`** to regenerate `internal/graphql/generated/`.

---

## 5. Tests

```bash
go test ./...
go vet ./...
```

---

[← Menu](../DEPLOYMENT.md) · [Seeding](../SEEDING.md) · [Phases](../PHASES.md)

