# Security and hardening

[← Documentation index](README.md) · [`.env.example`](../.env.example) (JWT / CORS vars) · [Command reference](COMMANDS.md)

Practical baseline guidance for secure operation of NextPressKit.

For a local-first setup flow, start with [`deployment/local.md`](deployment/local.md), then return here to harden settings.

## Dependency and CVE review

Run regularly (for example weekly and before release):

```bash
make security-check
go list -m -u all
```

- `make security-check` runs `govulncheck` against project packages.
- Review actionable findings and upgrade affected modules.

## CORS policy by deployment

`CORS_ORIGINS` controls allowed origins. Keep it explicit in production:

- **Production:** set exact origins (comma-separated), e.g. `https://app.example.com,https://admin.example.com`.
- **Local/dev:** empty `CORS_ORIGINS` allows all origins for faster iteration.

Operational notes:
- When `CORS_ORIGINS` is set, credentials are allowed and response headers are constrained.
- When unset, wildcard behavior is enabled and browser credentials are incompatible by design.
- Prefer terminating public traffic through Nginx and applying TLS before the app.

## JWT delivery: cookies vs Authorization header

The API reads the access JWT from either HttpOnly cookies or the `Authorization: Bearer` header, depending on **`JWT_AUTH_SOURCE`** (see [`.env.example`](../.env.example)):

| Mode | Access token | Refresh token | Login/refresh JSON |
|------|----------------|---------------|---------------------|
| `cookie` (default) | `JWT_ACCESS_COOKIE_NAME` (default `access_token`) | `JWT_REFRESH_COOKIE_NAME` (default `refresh_token`) | `user` only; tokens are not returned in the body |
| `header` | Client sends `Authorization: Bearer <jwt>` | Client stores refresh from JSON and sends it on `POST /auth/refresh` | `tokens` + `user` |

Cross-site browser flows (SPA on a different origin than the API) require:

- `CORS_ORIGINS` set to the exact frontend origin(s), and the client must use **`credentials: 'include'`** (or equivalent).
- Cookie attributes aligned with that deployment: defaults use **`SameSite=None`** and **`Secure`** (see `JWT_COOKIE_*` in `.env.example`).

**HTTPS at the browser:** those defaults mean the browser must reach the API over **`https://`** (Nginx + Let’s Encrypt on servers, or **mkcert** locally) for cookies to work in cross-site flows. See [deployment/local.md](deployment/local.md) (local HTTPS, HTTP-only dev options) and [DEPLOYMENT.md § TLS](DEPLOYMENT.md#4-tls-https).

Postman and other non-browser clients can use **cookie mode** (cookie jar after login) or **header mode**; see [`postman-templates/README.md`](../postman-templates/README.md).

## Rate-limit tuning and abuse tests

- Start with conservative `RATE_LIMIT_*` values from `.env.example`.
- Keep separate limits for public, auth, and admin scopes.
- Verify behavior with automated tests in `internal/platform/middleware/rate_limit_test.go`.
- Re-tune limits based on real traffic and error rate telemetry.

## JWT key rotation story

Current implementation uses one signing secret (`JWT_SECRET`) for access and refresh tokens.

Recommended rotation process:
1. Announce maintenance window for token invalidation.
2. Set new `JWT_SECRET` and redeploy API instances.
3. Invalidate old refresh tokens (users re-authenticate).
4. Verify new token issuance and refresh flows.

Future enhancement path:
- Add support for key IDs (`kid`) and multi-key verification window for zero-downtime rotation.

