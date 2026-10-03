# NextPressKit

NextPressKit is a [PN Scripts](https://pnscripts.com) product ([product page](https://pnscripts.com/products/nextpresskit)): a starter kit for building modern backend APIs using Go, Gin, and PostgreSQL. This repository is its backend API.

- Website: [nextpresskit.com](https://nextpresskit.com)
- Frontend: [nextpresskit/web](https://github.com/nextpresskit/web)

## Quick Start

Requirements:
- Go version from `go.mod`
- PostgreSQL running locally or remotely

1. Copy env and update database + JWT settings:

```bash
cp .env.example .env
```

2. Run setup:

```bash
make setup
```

3. Run API:

```bash
make run
```

4. Check:
- `http://localhost:9090/health`
- `http://localhost:9090/ready`

If you changed `APP_PORT`, use that port instead of `9090`.

## Most Used Commands

- `make setup`: first-time setup (interactive menu on a terminal)
- `make run`: run API in foreground
- `make postman-sync`: create/update local `postman/` files
- `./scripts/nextpresskit help`: full command list

## Auth Modes

- `JWT_AUTH_SOURCE=cookie` (default): browser-friendly HttpOnly cookies
- `JWT_AUTH_SOURCE=header`: Bearer token mode for scripts/clients

For cross-site browser auth, set `CORS_ORIGINS` and use HTTPS.

## Where To Read Next

- All docs index: [`docs/README.md`](./docs/README.md)
- Commands reference: [`docs/COMMANDS.md`](./docs/COMMANDS.md)
- Local setup + HTTPS: [`docs/deployment/local.md`](./docs/deployment/local.md)
- Server deployment: [`docs/DEPLOYMENT.md`](./docs/DEPLOYMENT.md)
- Troubleshooting: [`docs/TROUBLESHOOTING.md`](./docs/TROUBLESHOOTING.md)
- Security baseline: [`docs/SECURITY.md`](./docs/SECURITY.md)
- API contract: [`docs/openapi.yaml`](./docs/openapi.yaml)
- Postman templates: [`postman-templates/README.md`](./postman-templates/README.md)
