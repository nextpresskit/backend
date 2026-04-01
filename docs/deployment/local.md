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

## 4. Tests

```bash
go test ./...
go vet ./...
```

---

[← Menu](../DEPLOYMENT.md) · [Seeding](../SEEDING.md) · [Phases](../PHASES.md)

