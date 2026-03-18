## Project Phases

This document describes the planned phases for nextpress-backend.

---

### Phase 1 – Core Infrastructure

**Goals:**

- Initialize project as a modular monolith in Go.
- Add HTTP server with Gin.
- Configure PostgreSQL connection via GORM with a global DB instance.
- Introduce configuration system and logging.
- Provide a migration system.
- Provide deployment tooling (Makefile, systemd, Nginx, docs, scripts).

---

### Phase 2 – Authentication

**Goals:**

- Implement authentication and basic user management.
- Use JWT-based access and refresh tokens.
- Use bcrypt for password hashing.
- Keep logic in DDD-style `user` and `auth` modules.
- Expose auth endpoints for register, login, and refresh.

---

### Phase 3 – RBAC (Roles and Permissions)

**Goals:**

- Implement Role-Based Access Control (RBAC).
- Support users, roles, permissions, and their relations.
- Add middleware to enforce permissions on routes.
- Provide admin APIs to manage roles and permissions.

**Current status:**

- RBAC database schema is present (`roles`, `permissions`, `user_roles`, `role_permissions`).
- Authorization middleware exists (`RequirePermission`) and is wired with a sample protected route:
  - `GET /v1/admin/ping` requires `admin:ping`.
- Admin RBAC APIs exist (guarded by `rbac:manage`):
  - `GET /v1/admin/roles`, `POST /v1/admin/roles`
  - `GET /v1/admin/permissions`, `POST /v1/admin/permissions`
  - `POST /v1/admin/roles/:role_id/permissions` (grant permission to role)
  - `POST /v1/admin/users/:user_id/roles` (assign role to user)
- RBAC defaults are seeded via `make seed` / `go run ./cmd/seed` (admin role + base permissions).
- Optional one-time bootstrap endpoint (guarded by auth + env flag):
  - `POST /v1/admin/bootstrap/claim-admin` (requires `RBAC_BOOTSTRAP_ENABLED=true`)

---

### Phase 4 – CMS Core

**Goals:**

- Implement core CMS entities:
  - posts, pages, media, categories, tags, menus.
- Provide CRUD APIs and relations for content.
- Enable filtering, searching, and listing for content entities.

**Current status:**

- Posts: schema + basic CRUD API exists (RBAC-protected):
  - `GET /v1/admin/posts` (requires `posts:read`; supports `status`, `author_id`, `q`, `limit`, `offset`)
  - `GET /v1/admin/posts/:id` (requires `posts:read`)
  - `POST /v1/admin/posts` (requires `posts:write`)
  - `PUT /v1/admin/posts/:id` (requires `posts:write`)
  - `DELETE /v1/admin/posts/:id` (requires `posts:write`)
  - `PUT /v1/admin/posts/:id/categories` (requires `posts:write`)
  - `PUT /v1/admin/posts/:id/tags` (requires `posts:write`)
- Pages: schema + basic CRUD API exists (RBAC-protected):
  - `GET /v1/admin/pages` (requires `pages:read`; supports `status`, `author_id`, `q`, `limit`, `offset`)
  - `GET /v1/admin/pages/:id` (requires `pages:read`)
  - `POST /v1/admin/pages` (requires `pages:write`)
  - `PUT /v1/admin/pages/:id` (requires `pages:write`)
  - `DELETE /v1/admin/pages/:id` (requires `pages:write`)
- Taxonomy: categories and tags basic CRUD APIs exist (RBAC-protected):
  - `GET /v1/admin/categories` (requires `categories:read`)
  - `POST /v1/admin/categories` (requires `categories:write`)
  - `PUT /v1/admin/categories/:id` (requires `categories:write`)
  - `DELETE /v1/admin/categories/:id` (requires `categories:write`)
  - `GET /v1/admin/tags` (requires `tags:read`)
  - `POST /v1/admin/tags` (requires `tags:write`)
  - `PUT /v1/admin/tags/:id` (requires `tags:write`)
  - `DELETE /v1/admin/tags/:id` (requires `tags:write`)
- Media: upload + list APIs exist (RBAC-protected):
  - `POST /v1/admin/media` (multipart form field `file`, requires `media:write`)
  - `GET /v1/admin/media` (requires `media:read`)
  - `GET /v1/admin/media/:id` (requires `media:read`)
  - uploads are served at `MEDIA_PUBLIC_BASE_URL` (default `/uploads`)
- Menus: schema + basic APIs exist (RBAC-protected):
  - `GET /v1/admin/menus` (requires `menus:read`)
  - `POST /v1/admin/menus` (requires `menus:write`)
  - `GET /v1/admin/menus/:id` (requires `menus:read`)
  - `PUT /v1/admin/menus/:id` (requires `menus:write`)
  - `DELETE /v1/admin/menus/:id` (requires `menus:write`)
  - `GET /v1/admin/menus/:id/items` (requires `menus:read`)
  - `PUT /v1/admin/menus/:id/items` (requires `menus:write`)
- Public APIs (no auth):
  - `GET /v1/posts` (published only; supports `q`, `category_id`, `tag_id`, `limit`, `offset`)
  - `GET /v1/posts/:slug` (published only)
  - `GET /v1/pages/:slug` (published only)
  - `GET /v1/menus/:slug` (items returned as a tree)

---

### Phase 5 – Plugin System

**Goals:**

- Provide a plugin mechanism for extending the CMS without modifying core.
- Start with a database-driven plugin model (no Go plugins initially).
- Allow plugins to register routes, permissions, migrations, and hooks.
- Implement a plugin loader that wires plugins at startup.

---

### Phase 6 – Admin API

**Goals:**

- Provide an admin-facing API for:
  - dashboard/analytics,
  - settings management,
  - plugin management.
- Expose secure endpoints for environment and configuration management.

---

### Phase 7 – Example Ecommerce Plugin

**Goals:**

- Demonstrate plugin capabilities with a non-trivial example:
  - products, orders, cart, payments.
- Implement ecommerce as a plugin, not part of the core.
- Provide example APIs for catalog, cart, and checkout flows.