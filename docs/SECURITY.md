# Security and hardening

Practical baseline guidance for secure operation of NextPress.

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

