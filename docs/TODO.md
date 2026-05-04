# TODO (full checklist)

**`[x]`** = shipped · **`[ ]`** = still open. [Documentation index](README.md) · [`ROADMAP.md`](ROADMAP.md) · [`CONTRIBUTING.md`](../CONTRIBUTING.md) · [`openapi.yaml`](openapi.yaml)

Use this as the source-of-truth checklist for work tracking.

---

## Platform & tooling

- [x] Go API entrypoint (`cmd/api`)
- [x] Database schema (`cmd/migrate`, AutoMigrate in `internal/platform/dbmigrate`, models in `internal/modules/*/persistence`)
- [x] Seed runner (`cmd/seed`)
- [x] Makefile targets (`build`, `run`, `migrate-*`, `seed`, `graphql`, …)
- [x] Environment loading (`.env` / `.env.example`)
- [x] Structured logging (zap)
- [x] Request ID middleware
- [x] In-memory rate limiting (public / auth / admin groups)
- [x] Interactive deploy wizard (`scripts/deploy`, `scripts/deploy.ps1`)
- [x] Nginx + systemd templates (`deploy/`)
- [x] Local run helper (`scripts/run_local.sh`)

---

## Auth & users

- [x] `POST /auth/register`
- [x] `POST /auth/login`
- [x] `POST /auth/refresh`
- [x] Bcrypt password hashing
- [x] JWT access + refresh tokens
- [x] User persistence (GORM)

---

## RBAC

- [x] Roles, permissions, user-role, role-permission schema
- [x] Permission middleware (`RequirePermission`)
- [x] `GET /admin/ping` (`admin:ping`)
- [x] `GET/POST /admin/roles`, permissions, grant to role, assign role to user (`rbac:manage`)
- [x] Seeded defaults (`make seed`, `docs/SEEDING.md`)
- [x] Optional `POST /admin/bootstrap/claim-admin` (`RBAC_BOOTSTRAP_ENABLED`)

---

## Posts - core

- [x] Admin list/create/get/update/delete posts (`posts:read` / `posts:write`)
- [x] `PUT` categories / tags on post
- [x] `GET/PUT` post metrics
- [x] `PUT` primary category
- [x] `GET/PUT` post SEO
- [x] `PUT` featured image fields
- [x] `PUT` series link on post
- [x] `PUT` coauthors
- [x] Gallery items CRUD on post
- [x] Changelog entries on post
- [x] Per-post syndication + global syndication admin routes
- [x] Post translations CRUD
- [x] Admin series CRUD (`/admin/series`)
- [x] Translation groups admin (`/admin/translation-groups`)
- [x] Derived fields / hooks in application layer (e.g. derived fields hook)
- [x] Scheduled publish fields on model + **background ticker** promoting due posts (`cmd/api`)
- [x] Public `GET /posts`, `GET /posts/{slug}` (published)
- [x] Public `GET /posts/search` when Elasticsearch enabled

---

## Pages

- [x] Admin pages CRUD (`pages:read` / `pages:write`)
- [x] Public `GET /pages/{slug}`

---

## Taxonomy

- [x] Categories CRUD (`categories:*`)
- [x] Tags CRUD (`tags:*`)

---

## Media

- [x] Upload + list + get (`media:*`)
- [x] Local filesystem storage (`Storage` interface, local implementation)
- [x] Public URL via `MEDIA_PUBLIC_BASE_URL` + Nginx/static guidance

---

## Plugins

- [x] `plugins` table + GORM repository
- [x] `GET/POST /admin/plugins`, `PUT /admin/plugins/{id}` (`plugins:manage`)
- [x] `PostSave` hook chain registration at startup
- [x] One **noop** hook slot per enabled plugin (chain runs; no real plugin logic)
- [x] **Real** `BeforePostSave` / `AfterPostSave` implementations (dispatch by plugin `slug` / `config`)
- [x] Document and implement **error / transaction policy** (fail post save vs log-and-continue)

---

## Elasticsearch (optional)

- [x] Client config + env vars (`.env.example`)
- [x] Posts index + public search + admin reindex when enabled
- [x] Indexing on save / scheduled publish path (when ES on)
- [x] Operational runbook: index templates, upgrades, multi-cluster (beyond current docs)

---

## GraphQL (optional)

- [x] gqlgen wiring; `post`, `posts`, `page(slug)` queries
- [x] Playground when `APP_ENV` local/dev and flag on
- [x] Public read parity slice: `categories`, `tags`, `searchPosts(q)` queries
- [x] Auth mutation parity slice: `register`, `login`, `refresh`
- [x] **Parity (defined scope)** with REST for public/auth slices: categories, tags, search, `register/login/refresh`, taxonomy mutations (`create/update/delete` category/tag); admin GraphQL types remain out-of-scope (REST-first)
- [x] Document intended GraphQL vs REST split in [`README.md`](README.md) (docs index) or an ADR

---

## Testing & CI

- [x] Middleware tests (auth, authorization, rate limit)
- [x] Posts public routes tests (with stubs)
- [x] Plugins hooks unit test
- [x] Posts derived-fields hook test
- [x] Config tests (GraphQL / ES-related)
- [x] Auth application service unit tests (register/login/refresh paths)
- [x] RBAC application service unit tests (roles/permissions/assignments)
- [x] Pages application service unit tests (create/update validation paths)
- [x] Taxonomy application service unit tests (category/tag create/update paths)
- [x] Plugins application service unit tests (register/update validation paths)
- [x] Media application service unit tests (upload/get/list validation paths)
- [ ] **Unit tests** for: `auth`, `rbac`, `user`, `pages`, `taxonomy`, `media`, `plugins` transport/application (most modules)
- [x] **Integration tests** with real Postgres (docker or CI service)
- [x] **CI workflow** (GitHub Actions or other): `go test`, `go vet`, optional `golangci-lint`
- [x] **OpenAPI** validation / drift check vs router (optional tooling)

---

## Operations & scale

- [x] Single-instance rate limits
- [x] **Shared** rate-limit store (e.g. **Redis**) for horizontal scale
- [x] **Metrics** (Prometheus/OpenTelemetry) beyond logs
- [x] **Health** variants (deep checks: DB, ES) - refine if needed
- [x] **Version** string in logs/binary aligned with tags/releases (`cmd/api` currently embeds a static label)

---

## Media & storage (future)

- [ ] Optional **object storage** backend (S3-compatible) implementing `Storage` interface
- [ ] Virus scan / MIME hardening hooks (if product requires)

---

## Product / admin (future)

- [ ] Dashboard / analytics APIs
- [ ] Settings / site config APIs
- [ ] Richer plugin management UI contracts (if separate from current admin JSON)
- [ ] **Example domain plugin** (e.g. ecommerce) shipped in-repo or as sibling module

---

## Security & hardening (ongoing)

- [x] Periodic dependency / CVE review (`go mod`, govulncheck)
- [x] Security headers, CORS policy documented per deployment
- [x] Rate-limit tuning and abuse-case tests
- [x] JWT key rotation story (if beyond single `JWT_SECRET`)

---

## Documentation

- [x] Root `README`, `docs/README`, `CONTRIBUTING`, `DEPLOYMENT`, `SEEDING`, `local.md`, `ROADMAP`, this file
- [x] **CHANGELOG** or release notes process (optional)
- [x] **ADR** folder for big decisions (optional)

