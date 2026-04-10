# TODO (full checklist)

**`[x]`** = shipped · **`[ ]`** = still open. [`ROADMAP.md`](ROADMAP.md) · [`CONTRIBUTING.md`](../CONTRIBUTING.md) · [`openapi.yaml`](openapi.yaml)

---

## Platform & tooling

- [x] Go API entrypoint (`cmd/api`)
- [x] SQL migrations runner (`cmd/migrate`, `pkg/migrate`)
- [x] Seed runner (`cmd/seed`)
- [x] Makefile targets (`build`, `run`, `migrate-*`, `seed`, `graphql`, …)
- [x] Environment loading (`.env` / `.env.example`)
- [x] Structured logging (zap)
- [x] Request ID middleware
- [x] In-memory rate limiting (public / auth / admin groups)
- [x] Deployment script (`scripts/deploy`)
- [x] Nginx + systemd templates (`deploy/`)
- [x] Local run helper (`scripts/run_local.sh`)

---

## Auth & users

- [x] `POST /v1/auth/register`
- [x] `POST /v1/auth/login`
- [x] `POST /v1/auth/refresh`
- [x] Bcrypt password hashing
- [x] JWT access + refresh tokens
- [x] User persistence (GORM)

---

## RBAC

- [x] Roles, permissions, user-role, role-permission schema
- [x] Permission middleware (`RequirePermission`)
- [x] `GET /v1/admin/ping` (`admin:ping`)
- [x] `GET/POST /v1/admin/roles`, permissions, grant to role, assign role to user (`rbac:manage`)
- [x] Seeded defaults (`make seed`, `docs/SEEDING.md`)
- [x] Optional `POST /v1/admin/bootstrap/claim-admin` (`RBAC_BOOTSTRAP_ENABLED`)

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
- [x] Admin series CRUD (`/v1/admin/series`)
- [x] Translation groups admin (`/v1/admin/translation-groups`)
- [x] Derived fields / hooks in application layer (e.g. derived fields hook)
- [x] Scheduled publish fields on model + **background ticker** promoting due posts (`cmd/api`)
- [x] Public `GET /v1/posts`, `GET /v1/posts/{slug}` (published)
- [x] Public `GET /v1/posts/search` when Elasticsearch enabled

---

## Pages

- [x] Admin pages CRUD (`pages:read` / `pages:write`)
- [x] Public `GET /v1/pages/{slug}`

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

## Menus

- [x] Menus CRUD + items (`menus:*`)
- [x] Public `GET /v1/menus/{slug}`

---

## Plugins

- [x] `plugins` table + GORM repository
- [x] `GET/POST /v1/admin/plugins`, `PUT /v1/admin/plugins/{id}` (`plugins:manage`)
- [x] `PostSave` hook chain registration at startup
- [x] One **noop** hook slot per enabled plugin (chain runs; no real plugin logic)
- [ ] **Real** `BeforePostSave` / `AfterPostSave` implementations (dispatch by plugin `slug` / `config`)
- [ ] Document and implement **error / transaction policy** (fail post save vs log-and-continue)

---

## Elasticsearch (optional)

- [x] Client config + env vars (`.env.example`)
- [x] Posts index + public search + admin reindex when enabled
- [x] Indexing on save / scheduled publish path (when ES on)
- [ ] Operational runbook: index templates, upgrades, multi-cluster (beyond current docs)

---

## GraphQL (optional)

- [x] gqlgen wiring; `post`, `posts`, `page(slug)` queries
- [x] Playground when `APP_ENV` local/dev and flag on
- [ ] **Parity** with REST: categories, tags, menus, search, mutations, auth, admin types - as you define scope
- [ ] Document intended GraphQL vs REST split in [`README.md`](README.md) (docs index) or an ADR

---

## Testing & CI

- [x] Middleware tests (auth, authorization, rate limit)
- [x] Posts public routes tests (with stubs)
- [x] Plugins hooks unit test
- [x] Posts derived-fields hook test
- [x] Config tests (GraphQL / ES-related)
- [ ] **Unit tests** for: `auth`, `rbac`, `user`, `pages`, `taxonomy`, `media`, `menus`, `plugins` transport/application (most modules)
- [ ] **Integration tests** with real Postgres (docker or CI service)
- [ ] **CI workflow** (GitHub Actions or other): `go test`, `go vet`, optional `golangci-lint`
- [ ] **OpenAPI** validation / drift check vs router (optional tooling)

---

## Operations & scale

- [x] Single-instance rate limits
- [ ] **Shared** rate-limit store (e.g. **Redis**) for horizontal scale
- [ ] **Metrics** (Prometheus/OpenTelemetry) beyond logs
- [ ] **Health** variants (deep checks: DB, ES) - refine if needed
- [ ] **Version** string in logs/binary aligned with tags/releases (`cmd/api` currently embeds a static label)

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

- [ ] Periodic dependency / CVE review (`go mod`, govulncheck)
- [ ] Security headers, CORS policy documented per deployment
- [ ] Rate-limit tuning and abuse-case tests
- [ ] JWT key rotation story (if beyond single `JWT_SECRET`)

---

## Documentation

- [x] Root `README`, `docs/README`, `CONTRIBUTING`, `DEPLOYMENT`, `SEEDING`, `local.md`, `ROADMAP`, this file
- [ ] **CHANGELOG** or release notes process (optional)
- [ ] **ADR** folder for big decisions (optional)

