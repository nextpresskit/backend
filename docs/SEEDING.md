# Database seeding

[← Documentation index](README.md) · [Commands: migrate and seed](COMMANDS.md#database-and-seed-data)

Seeding adds predictable demo content so you can try the API without entering everything by hand. Tables must already exist: run migrate-up first, or use setup, which does migrate and seed. All command options (including db-fresh) are in [COMMANDS.md](COMMANDS.md#database-and-seed-data).

```bash
make migrate-up && make seed          # usual path
make db-fresh && make seed            # wipe public schema, recreate tables, then seed (dev only)
make seed                             # run again anytime; upserts, not duplicates
```

You can also run `go run ./cmd/seed`, `./bin/seed` after `make seed-build`, or the nextpresskit script; behavior is the same.

## What gets seeded

1. RBAC baseline (`pkg/seed/rbac_defaults.go`): admin role and core permission codes.
2. Full demo dataset (`pkg/seed/full_dataset.go`): about 100 rows per table for local use.
3. Superadmin: one privileged user tied to both superadmin and admin roles.

Reruns are safe: seeders use upserts so you do not pile up duplicate keys.

## Superadmin credentials

Configure in `.env` (defaults shown in `.env.example`):

```bash
SEED_SUPERADMIN_EMAIL=superadmin@nextpresskit.local
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

If you want the deploy wizard release step to run `./bin/seed` automatically, set `RUN_SEED_ON_DEPLOY=true` in `.env` ([DEPLOYMENT.md](DEPLOYMENT.md)).

---

See also: [Documentation index](README.md) · [Deployment](DEPLOYMENT.md) · [TODO](TODO.md)
