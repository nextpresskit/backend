# Kit modules

[← Docs index](README.md)

NextPressKit is modular.
Each feature is a module under `internal/modules/<name>`.

## Default registry

The compile-time default list (order matters for migrations and HTTP wiring) lives in [`internal/appregistry/registry.go`](../internal/appregistry/registry.go):

| Order | Module id | Responsibility |
|-------|-----------|------------------|
| 1 | `user` | Users table + demo user seed; admin `GET /admin/users` (paginated, `q` search, roles) and `GET /admin/users/{id}` behind `users:read` |
| 2 | `rbac` | Roles, permissions, RBAC admin APIs, `/admin/ping`, optional bootstrap |
| 3 | `auth` | Register/login/refresh/logout/me |
| 4 | `taxonomy` | Categories and tags |
| 5 | `media` | Uploads + local static serving when `MEDIA_PUBLIC_BASE_URL` is a path |
| 6 | `posts` | Posts + public routes + Elasticsearch hooks + scheduled publish loop |
| 7 | `pages` | Pages + public routes |

## Choose modules at runtime

Set `MODULES` in `.env` as comma-separated module ids.
If empty, all default modules are enabled.

Some dependencies are auto-added:

- `auth` → also enables `user`, `rbac`
- Any of `posts`, `pages`, `taxonomy`, `media` → also enables `user`, `rbac`, `auth`
- `posts` → also enables `taxonomy`, `media` (FKs / featured media)

Unknown module ids are ignored.

Examples:

```bash
# Minimal API (auth + RBAC only)
MODULES=auth,user,rbac

# Full CMS (same as empty MODULES)
MODULES=user,rbac,auth,taxonomy,media,posts,pages
```

## Change default registry

Edit `internal/appregistry/registry.go`.

## What commands use modules

- `cmd/api` for HTTP routes/services
- `cmd/migrate` for schema migration
- `cmd/seed` for demo/RBAC seed data

## Adding a new module

1. Create `internal/modules/<name>/` with your usual `domain`, `application`, `persistence`, `transport` layout.
2. Add `internal/modules/<name>/module/module.go` implementing [`kit.Module`](../internal/kit/module.go).
3. Append the module to `ModuleRegistry()` in [`internal/appregistry/registry.go`](../internal/appregistry/registry.go) in the correct position for FKs.
4. Return permission codes from `Permissions()` if the module introduces RBAC codes seeded for the admin role.
5. Update [`docs/openapi.yaml`](openapi.yaml) and Postman templates when you add HTTP surface.

## Plugin note

Dynamic plugin loading is not part of this core kit.
