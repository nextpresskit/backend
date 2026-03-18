# Database Seeding

Seeders populate reference data and optional development data. They are **idempotent**: running them multiple times will not duplicate data.

## Prerequisites

- Database migrations have been applied (`make migrate-up`).
- `.env` has valid `DB_*` settings so the seed command can connect.

## How to Run

### Run all seeders

```bash
make seed
# or
go run ./cmd/seed
# or build then run
make seed-build
./bin/seed
```

## What is seeded right now

### RBAC defaults

- **Tables:** `roles`, `permissions`, `role_permissions`
- **Data:**
  - role: `admin`
  - permissions: `admin:ping`, `rbac:manage`
  - grants: `admin` gets both permissions

These are used by:

- `GET /v1/admin/ping` (requires `admin:ping`)
- Admin management endpoints (requires `rbac:manage`)

