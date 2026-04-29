# Postman - NextPressKit API

Postman collections and environments for the NextPressKit backend.

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

Set this on each imported environment (`NextPress-*.postman_environment.json`).

## Collections

| Collection | File | Contents |
|------------|------|----------|
| **NextPressKit Public API** | `NextPress-Public-API.postman_collection.json` | Root/health/ready endpoints plus `/auth/*` and public content APIs. |
| **NextPressKit Admin API** | `NextPress-Admin-API.postman_collection.json` | All `/admin/*` endpoints requiring admin token and permissions. |

## Environments

Use one environment per target. Both collections rely on `{{base_url}}`.

| Environment | File | Use case | `base_url` |
|-------------|------|----------|------------|
| **NextPressKit - Local** | `NextPress-Local.postman_environment.json` | Local development (`APP_PORT=9090` by default) | `http://localhost:9090` |
| **NextPressKit - Dev** | `NextPress-Dev.postman_environment.json` | Dev deployment | `https://api-dev.example.com` |
| **NextPressKit - Staging** | `NextPress-Staging.postman_environment.json` | Staging deployment | `https://api-staging.example.com` |
| **NextPressKit - Production** | `NextPress-Production.postman_environment.json` | Production deployment | `https://api.example.com` |

> Replace the dev/staging/production `base_url` values with your actual domains.

### Setup

1. Import the two collections and the four environment files into Postman.
2. Select one environment. (Browser apps: set **`CORS_ORIGINS`** on the API to your frontend origin and use `credentials: 'include'`. Postman itself ignores CORS but still stores response cookies per host.)
3. Run **`POST /auth/login`** from the Public collection.
   - **`jwt_auth_source=cookie`:** cookies are stored automatically; response body is `{ "user": … }` only. Then run Admin requests against the same `base_url`.
   - **`jwt_auth_source=header`:** the login tests store `access_token`, `refresh_token`, and **`admin_access_token`** (copy of access) for the Admin collection script.
4. For **cookie mode**, no manual copy step is required. For **header mode**, use `admin_access_token` (already synced after login from the updated collection tests).

### Notes

- `POST /admin/bootstrap/claim-admin` is only available when `RBAC_BOOTSTRAP_ENABLED=true`.
- `GET /posts/search` and `POST /admin/posts/search/reindex` require Elasticsearch to be enabled.
- GraphQL (`/graphql`) is optional and controlled by `GRAPHQL_ENABLED`; it is not part of these REST collections. GraphQL `login` / `refresh` use the same cookie behavior when `JWT_AUTH_SOURCE=cookie`.
