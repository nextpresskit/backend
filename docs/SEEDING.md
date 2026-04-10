# Database seeding

**How-to** - run seeders and understand default RBAC data.

Seeders load **reference data** (especially RBAC). They are **idempotent**: natural keys use `ON CONFLICT DO NOTHING` (or equivalent) so repeat runs do not duplicate roles, permissions, etc.

## Prerequisites

1. Migrations applied: `make migrate-up`.
2. `.env` contains valid `DB_*` (same as `cmd/migrate` / `cmd/seed`).

## Run

```bash
make seed
# or: go run ./cmd/seed
# or: make seed-build && ./bin/seed
```

## RBAC defaults

Source: `pkg/seed/rbac_defaults.go`.

### Role

| Name | Notes |
|------|--------|
| `admin` | Fixed UUID `00000000-0000-0000-0000-000000000001` |

### Permissions (granted to `admin`)

| Code | Typical use |
|------|-------------|
| `admin:ping` | `GET /v1/admin/ping` |
| `rbac:manage` | Role/permission APIs, user-role assignment |
| `posts:read` / `posts:write` | Posts and post-taxonomy links |
| `pages:read` / `pages:write` | Pages |
| `categories:read` / `categories:write` | Categories |
| `tags:read` / `tags:write` | Tags |
| `media:read` / `media:write` | Media |
| `menus:read` / `menus:write` | Menus and items |
| `plugins:manage` | Plugin admin endpoints |

After seeding, assign `admin` to a user via RBAC APIs or optional bootstrap - see [`ROADMAP.md`](ROADMAP.md) (RBAC).

## Deploy

`RUN_SEED_ON_DEPLOY=true` in `.env` runs `./bin/seed` during `scripts/deploy`.

---

**See also:** [Documentation index](README.md) · [Deployment](DEPLOYMENT.md) · [TODO](TODO.md) (**Auth & users** / **RBAC** sections)
