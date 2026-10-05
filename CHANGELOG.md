# Changelog

All notable changes to this project should be documented in this file.

Format:
- Keep a top `## Unreleased` section for ongoing work.
- Move items into a dated release section when tagging.
- Group entries by type: `Added`, `Changed`, `Fixed`, `Security`.

## Unreleased

### Added
- Admin users API: `GET /admin/users` (`limit`/`offset`/`q`, `{users,total,limit,offset}`, role names per user) and `GET /admin/users/{id}` (numeric id or uuid), guarded by the new `users:read` permission, which `make seed` grants to the `admin` role. OpenAPI, Postman template, unit tests and a Postgres integration test included.
- CI: gofmt check, `go test -race`, govulncheck; actions pinned by commit SHA with read-only token permissions.
- Modular kit: `internal/kit` (`Module`, `Deps`), `internal/app.Run`, `internal/appregistry` default registry, `MODULES` env to filter modules; `cmd/migrate` and `cmd/seed` use the same list. Docs: `docs/MODULES.md`, ADR `docs/adr/0001-module-composition.md`.
- `pkg/seed/helpers` for shared demo seed helpers.
- Postman: `jwt_auth_source` on all environments; collection pre-request scripts for cookie vs Bearer JWT; Auth folder script for cookie-mode refresh/logout bodies.
- OpenAPI: `cookieAuth` security scheme, auth response and cookie documentation, optional refresh/logout bodies.

### Changed
- **Breaking:** Removed the in-tree `plugins` module (`/admin/plugins`, `plugins:manage`, DB-driven post hooks). Post-save hooks are derived fields + optional Elasticsearch only.
- RBAC default seed now inserts only permission codes declared by enabled modules (`Permissions()`), via `SeedRBACDefaults(db, codes)`.
- Removed the optional GraphQL API (gqlgen), `GRAPHQL_*` settings, and `make graphql` / CLI `graphql` generation. The HTTP surface is REST per OpenAPI only.
- Documentation: root `README`, `docs/README`, `docs/SECURITY`, `docs/DEPLOYMENT`, `docs/deployment/local`, and `postman-templates/README` updated for HttpOnly JWT cookies and `JWT_AUTH_SOURCE`.
- Documentation: added `docs/COMMANDS.md` and cross-links across docs for faster navigation and plain-language command explanations.

### Fixed
- (fill in during development)

### Security
- (fill in during development)

