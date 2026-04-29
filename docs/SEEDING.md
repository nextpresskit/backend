# Database seeding

`make seed` runs the full seeding pipeline and is safe to run repeatedly.

## What `make seed` does

1. Seeds RBAC defaults (`pkg/seed/rbac_defaults.go`) - baseline `admin` role and core permissions.
2. Seeds a deterministic full dataset (`pkg/seed/full_dataset.go`) with **100 records per table** for local/dev use.
3. Seeds a `superadmin` account and links it to both `superadmin` and `admin` roles.

All seeders are idempotent (`ON CONFLICT ...`) so reruns update/keep existing deterministic rows instead of duplicating them.

## Prerequisites

1. Apply migrations: `make migrate-up`
2. Ensure `.env` has valid `DB_*` values

## Run seeders

```bash
make seed
# or:
go run ./cmd/seed
# or:
make seed-build && ./bin/seed
```

## Superadmin credentials

Configure in `.env` (defaults shown in `.env.example`):

```bash
SEED_SUPERADMIN_EMAIL=superadmin@nextpress.local
SEED_SUPERADMIN_PASSWORD=SuperAdmin123!
```

The seeded superadmin user is deterministic and updated on reruns (same identity, latest configured credentials).

## Tables seeded with 100 rows

- `users` (includes `superadmin` as one of the 100)
- `roles`
- `permissions` (100 total with RBAC defaults + generated permissions)
- `role_permissions`
- `user_roles`
- `posts`, `pages`
- `categories`, `tags`
- `media`
- `plugins`
- `post_categories`, `post_tags`
- `post_seo`, `post_metrics`
- `series`, `post_series`
- `post_coauthors`, `post_gallery_items`, `post_changelog`, `post_syndication`
- `translation_groups`, `post_translations`

## Deploy

Set `RUN_SEED_ON_DEPLOY=true` in `.env` to run `./bin/seed` automatically when you choose the **release** steps in the interactive `./scripts/deploy` wizard.

---

**See also:** [Documentation index](README.md) · [Deployment](DEPLOYMENT.md) · [TODO](TODO.md)
