# Roadmap

**Explanation** — product scope and direction. **Every checkbox** (shipped vs open): [`TODO.md`](TODO.md) — use **`[ ]`** lines as the backlog; this page stays short.

**Related:** [Documentation index](README.md) · [Contributing](../CONTRIBUTING.md) · [Commands](COMMANDS.md) · **REST contract** [`openapi.yaml`](openapi.yaml)

Keep this file short: shipped capabilities, what you are actively improving, and rough future themes. Owners and threading: your issue tracker + [`TODO.md`](TODO.md).

Use this page for direction and use [`TODO.md`](TODO.md) for execution details.

---

## Shipped

- **Platform:** Gin, GORM (AutoMigrate + `cmd/migrate` / `cmd/seed`), config, logging—see [`COMMANDS.md`](COMMANDS.md#database-and-seed-data) and [`DEPLOYMENT.md`](DEPLOYMENT.md).
- **Auth:** Register/login/refresh, JWT access + refresh, bcrypt.
- **RBAC:** Roles, permissions, middleware, admin APIs, seeded defaults, optional bootstrap.
- **Content & admin APIs:** Posts, pages, taxonomy, media; public + admin HTTP APIs; rate limits, request ID, OpenAPI.
- **Plugins (baseline):** `plugins` table, admin CRUD, `PostSave` hook chain (handlers still to be implemented — see **[Plugins](TODO.md#plugins)** in [`TODO.md`](TODO.md)).

---

## In progress

Typical focus: **plugin hook implementations**, **test/CI coverage**, **GraphQL parity** (if desired). Details: unchecked lines in [`TODO.md`](TODO.md).

---

## Later

See unchecked items under **Product / admin**, **Operations & scale**, **Security**, and **Documentation** in [`TODO.md`](TODO.md).

---

## Historical note

Numbered **phases** (1-5) were internal planning labels during early development; they are retired in favor of the sections above.
