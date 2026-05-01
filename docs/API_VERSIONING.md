# API Versioning Strategy

[← Documentation index](README.md) · [OpenAPI contract](openapi.yaml)

This document records the current API versioning decision for NextPressKit and how to evolve it safely.

## One-line summary

The API is unversioned by default, and you can switch to path versioning later by setting `API_BASE_PATH=/v1` without refactoring handlers.

## Current Decision

- Default strategy: **unversioned paths**.
- Optional path prefix: **`API_BASE_PATH`**.
- Default value: empty (`""`) so routes are unversioned.

Examples:

- `API_BASE_PATH=""` -> `POST /auth/login`, `GET /posts`, `GET /admin/ping`
- `API_BASE_PATH="/v1"` -> `POST /v1/auth/login`, `GET /v1/posts`, `GET /v1/admin/ping`

This keeps the API simple now while making URL-path versioning a no-refactor config change later.

## Why This Approach

- Keeps early development friction low (no forced version policy yet).
- Avoids hardcoding `/v1` across handlers and docs.
- Gives a reversible, low-risk bridge to path versioning when needed.
- Fits current architecture where routes are centralized in `cmd/api`.

## Runtime Behavior

- All REST groups are mounted under `API_BASE_PATH`.
- GraphQL default path is derived from the same base:
  - `<API_BASE_PATH>/graphql`
  - So defaults become:
    - `/graphql` when base path is empty
    - `/v1/graphql` when `API_BASE_PATH=/v1`

You can still override GraphQL explicitly with `GRAPHQL_PATH`.

## Configuration

In `.env`:

```env
# Leave empty for unversioned endpoints.
API_BASE_PATH=
```

Rules:

- Empty or `/` means no prefix.
- Missing leading slash is normalized (`v1` -> `/v1`).
- Trailing slash is trimmed (`/v1/` -> `/v1`).

## If We Choose URL Path Versioning Later

1. Set `API_BASE_PATH=/v1`.
2. Update external clients to call `/v1/*`.
3. Keep unversioned compatibility only if needed (temporary rewrite/proxy rule).
4. Announce migration window and cutoff date.

No handler-level route refactor is required.

## If We Choose Header Versioning Later

Recommended transition:

1. Keep canonical server routes under one internal base path.
2. Add middleware that maps version header (for example `X-API-Version: 1`) to the chosen route group or contract behavior.
3. Maintain explicit OpenAPI docs for supported header versions.
4. Add version usage metrics before deprecating any behavior.

Notes:

- Header versioning keeps URLs clean but is harder to test and cache correctly.
- Path versioning is usually easier for CDN/proxy visibility and debugging.

## Compatibility Policy (Baseline)

Until formal versioning is introduced:

- Additive changes only by default (new optional fields/endpoints).
- Avoid breaking field renames/removals without a migration plan.
- Document contract-affecting changes in `CHANGELOG.md`.

## Deprecation Policy Template

When versioning starts, apply this baseline:

- Announce deprecation with timeline.
- Provide migration guide and examples.
- Track old/new version usage.
- Remove old behavior only after the published sunset date.
