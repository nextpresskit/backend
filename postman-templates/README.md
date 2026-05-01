# Postman - NextPressKit API

[Documentation index](../docs/README.md) · [Command reference](../docs/COMMANDS.md)

Canonical collection and environment JSON lives in this **`postman-templates/`** folder (tracked in git). The **`postman/`** directory at the repo root is **gitignored**: run **`postman-sync`** to copy any missing files from here into `postman/`, then apply values from `.env.example` / `.env`.

If you only need a quick path: run `./scripts/nextpresskit postman-sync`, then import files from `postman/`.

## Route groups (Public/Auth and Admin)

The API is split into two major groups:

| Type | Base URL | Auth | Use case |
|------|----------|------|----------|
| **Public/Auth** | `{{base_url}}` | Public routes: none. Auth routes: none. | Health/readiness checks, auth (`/auth/*`), and public content routes (`/posts/*`, `/pages/*`). |
| **Admin** | `{{base_url}}` | JWT via **cookie jar** (default) or **`Authorization: Bearer`** (see `jwt_auth_source`) | Management routes under `/admin/*`: posts, pages, taxonomy, media, RBAC, plugin management, and bootstrap/admin checks. |

### `jwt_auth_source` (environment variable)

Matches server behavior controlled by `JWT_AUTH_SOURCE` in `.env`:

| Value | Meaning |
|-------|---------|
| `cookie` (default) | After `POST /auth/login`, Postman stores HttpOnly cookies for `{{base_url}}`. Protected requests **do not** send `Authorization`; the collection pre-request script removes that header so the cookie jar is used. |
| `header` | Login/refresh responses include `tokens` in JSON. The collection scripts set `Authorization: Bearer …` from `access_token` (Public) or `admin_access_token` (Admin). |

Set this on each imported environment (`NextPressKit-*.postman_environment.json`).

## Collections

| Collection | File | Contents |
|------------|------|----------|
| **NextPressKit Public API** | `NextPressKit-Public-API.postman_collection.json` | Root/health/ready endpoints plus `/auth/*` and public content APIs. |
| **NextPressKit Admin API** | `NextPressKit-Admin-API.postman_collection.json` | All `/admin/*` endpoints requiring admin token and permissions. |

## Environments

Use one environment per target. Both collections rely on `{{base_url}}`. **`POST /auth/login`** in the Public collection uses `{{superadmin_email}}` and `{{superadmin_password}}`, which default to the seed superadmin (`SEED_SUPERADMIN_EMAIL` / `SEED_SUPERADMIN_PASSWORD` in `.env.example`). Override per environment (required for real staging/production accounts).

| Environment | File | Use case | `base_url` |
|-------------|------|----------|------------|
| **NextPressKit - Local** | `NextPressKit-Local.postman_environment.json` | Local Nginx + TLS (`make deploy`, `nextpresskit.local` in `/etc/hosts`) | `https://nextpresskit.local` |
| **NextPressKit - Dev** | `NextPressKit-Dev.postman_environment.json` | Dev deployment | `https://api-dev.example.com` |
| **NextPressKit - Staging** | `NextPressKit-Staging.postman_environment.json` | Staging deployment | `https://api-staging.example.com` |
| **NextPressKit - Production** | `NextPressKit-Production.postman_environment.json` | Production deployment | `https://api.example.com` |

> Replace the dev/staging/production `base_url` values with your actual domains. For **direct** `go run` / `make run` without Nginx, set local `base_url` to `http://127.0.0.1:9090` (or your `APP_PORT`).

### Sync from repo env files

Refresh **`postman/*.postman_environment.json`** from `.env.example` and `.env` (and optional shell overrides). On a fresh clone, this also creates **`postman/`** from these templates when needed:

```bash
./scripts/nextpresskit postman-sync
# or
make postman-sync
```

- Preview: `./scripts/nextpresskit postman-sync --dry-run`
- Windows: `.\scripts\nextpresskit.ps1 postman-sync`
- Tier URLs: `POSTMAN_LOCAL_BASE_URL`, `POSTMAN_DEV_BASE_URL`, `POSTMAN_STAGING_BASE_URL`, `POSTMAN_PRODUCTION_BASE_URL`, or set `NEXTPRESS_PUBLIC_HOST` for local `https://<host>`
- Clear token placeholders in the JSON: `POSTMAN_CLEAR_TOKENS=1 ./scripts/nextpresskit postman-sync`

Collections are not rewritten (requests use `{{base_url}}` only).

### Setup

1. Run **`postman-sync`** once so **`postman/`** contains the JSON (then import from that folder).
2. Import the two collections and the four environment files into Postman.
3. Select one environment. (Browser apps: set **`CORS_ORIGINS`** on the API to your frontend origin and use `credentials: 'include'`. Postman itself ignores CORS but still stores response cookies per host.)
4. Run **`POST /auth/login`** from the Public collection.
   - **`jwt_auth_source=cookie`:** cookies are stored automatically; response body is `{ "user": … }` only. Then run Admin requests against the same `base_url`.
   - **`jwt_auth_source=header`:** the login tests store `access_token`, `refresh_token`, and **`admin_access_token`** (copy of access) for the Admin collection script.
5. For **cookie mode**, no manual copy step is required. For **header mode**, use `admin_access_token` (already synced after login from the updated collection tests).

### Notes

- `POST /admin/bootstrap/claim-admin` is only available when `RBAC_BOOTSTRAP_ENABLED=true`.
- `GET /posts/search` and `POST /admin/posts/search/reindex` require Elasticsearch to be enabled.
- GraphQL (`/graphql`) is optional and controlled by `GRAPHQL_ENABLED`; it is not part of these REST collections. GraphQL `login` / `refresh` use the same cookie behavior when `JWT_AUTH_SOURCE=cookie`.
