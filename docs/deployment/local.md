# Local development

[← Documentation index](../README.md) · [Server deployment (Ubuntu)](../DEPLOYMENT.md) · [Command reference](../COMMANDS.md)

**Tutorial / how-to** - run the API on your machine. The default path is `go run` / `make run` with no reverse proxy. Optionally use **systemd** and **Nginx** on the same box to mirror how [server deployment](../DEPLOYMENT.md) runs the binary behind a proxy.

Platform-specific setup: **[macOS](macos.md)** (Homebrew, paths) · **Ubuntu laptop** (below, [Ubuntu on servers](../DEPLOYMENT.md)). Interactive Nginx/TLS/setup: [§ Interactive deploy](../DEPLOYMENT.md#interactive-deploy-scriptsdeploy) (`make deploy` or `scripts/deploy.ps1` on Windows).

## Fast local path (recommended)

If you just want the API running quickly:

```bash
./scripts/nextpresskit setup   # install deps, prepare .env, migrate, seed
./scripts/nextpresskit run     # run API in foreground
```

Make equivalents:

```bash
make setup
make run
```

## Already running the project?

Nothing in this guide **requires** you to change a working setup. If you already use `make run`, a local Nginx reverse proxy, or a server install from [DEPLOYMENT.md](../DEPLOYMENT.md), you can keep doing that.

**Optional upgrade:** add **HTTPS** locally (recommended for browser cookie auth that matches production defaults) or on the server (Let’s Encrypt). See [HTTPS locally (recommended)](#https-locally-recommended) and [DEPLOYMENT.md § TLS](../DEPLOYMENT.md#4-tls-https).

When you switch the public base URL to `https://`, update **clients** (Postman environments, frontend API URL) and ensure **`CORS_ORIGINS`** lists the exact `https://` frontend origins you use.

## Why HTTPS matters for cookie auth

Default [`.env.example`](../../.env.example) uses `JWT_COOKIE_SAME_SITE=none` and `JWT_COOKIE_SECURE=true`. With `SameSite=None`, the API sets **`Secure`** cookies ([`internal/platform/jwtcookie/jwtcookie.go`](../../internal/platform/jwtcookie/jwtcookie.go)); browsers will not store or send them on plain **`http://`** for cross-site flows.

- **Production-like browser tests** (SPA on another origin, `credentials: 'include'`): terminate **HTTPS** in front of the API (Nginx + mkcert locally, or Let’s Encrypt on a server).
- **Staying on HTTP:** use **`JWT_AUTH_SOURCE=header`** and Bearer tokens for API-only checks, or relax cookies **only** for same-site dev (see [If you stay on HTTP](#if-you-stay-on-http)).

## Prerequisites

- Go (`go.mod`)
- PostgreSQL
- Git

## Ubuntu laptop (local)

Typical packages:

```bash
sudo apt update
sudo apt install -y golang-go postgresql postgresql-contrib git nginx
```

Install a supported Go toolchain if the distro package is too old (see [Go downloads](https://go.dev/dl/)). Create a DB user/database for `DB_*` in `.env`. **mkcert** for local HTTPS: [install from FiloSottile/mkcert releases](https://github.com/FiloSottile/mkcert) or your package manager, then follow [HTTPS locally (recommended)](#https-locally-recommended). Nginx site files live under `/etc/nginx/sites-available/` (same patterns as [DEPLOYMENT.md](../DEPLOYMENT.md)).

## Setup

```bash
git clone <repo-url> nextpresskit-backend && cd nextpresskit-backend
go mod download
cp .env.example .env
```

Configure `DB_*`, `JWT_SECRET`, `APP_ENV=local`. For auth, set `JWT_AUTH_SOURCE` (`cookie` default vs `header`) and `JWT_COOKIE_*` / cookie names as needed; use `CORS_ORIGINS` when testing cookie auth from a separate frontend origin. Optional flags: Elasticsearch, GraphQL - [`.env.example`](../../.env.example).

## Migrate, seed, run

Create tables, load demo data, then start the API ([full command guide](../COMMANDS.md#database-and-seed-data)):

```bash
make migrate-up
make seed
./scripts/run_local.sh   # or: go run ./cmd/api
```

Reset everything in the `public` schema on your laptop: `make db-fresh && make seed`.

| | |
|--|--|
| Base URL | `http://localhost:<APP_PORT>` (default 9090) |
| Checks | `GET /health`, `GET /ready` |
| REST | [`docs/openapi.yaml`](../openapi.yaml) |

## If you stay on HTTP

Use one of these when you are not ready to terminate TLS locally:

| Goal | Suggestion |
|------|------------|
| API / Postman / scripts only | `JWT_AUTH_SOURCE=header`; send `Authorization: Bearer …` after login. |
| Browser, API and frontend **same site** (no cross-origin cookie need) | You may set `JWT_COOKIE_SAME_SITE=lax` and `JWT_COOKIE_SECURE=false` in `.env` for local only. **Do not** use this shape for cross-site SPAs; those still need HTTPS. |

Cross-site flows with default cookie settings require **HTTPS** at the browser (see below).

## HTTPS locally (recommended)

Use **[mkcert](https://github.com/FiloSottile/mkcert)** so the OS and browsers trust a local CA. Works on Linux and macOS (see [macOS notes](macos.md) for Homebrew paths).

1. **Install and install the local CA** (once per machine):

   ```bash
   mkcert -install
   ```

2. **Issue a certificate** for the hostnames you will use (examples):

   ```bash
   mkdir -p ~/.local/share/nextpresskit-ssl
   cd ~/.local/share/nextpresskit-ssl
   mkcert localhost 127.0.0.1 ::1 api.nextpresskit.local
   ```

   This creates `localhost+2.pem` and `localhost+2-key.pem` (names vary with the command). Add `api.nextpresskit.local` to `/etc/hosts` pointing at `127.0.0.1` if you use a fake hostname.

3. **Nginx** terminates TLS and proxies to the Go process (same idea as [Optional: Nginx in front](#optional-nginx-in-front-local)). Add a `server` block that listens on **443** and references the mkcert files:

   ```nginx
   server {
       listen 443 ssl;
       server_name localhost api.nextpresskit.local;

       ssl_certificate     /home/YOU/.local/share/nextpresskit-ssl/localhost+2.pem;
       ssl_certificate_key /home/YOU/.local/share/nextpresskit-ssl/localhost+2-key.pem;

       location /uploads/ {
           alias /absolute/path/to/your/clone/storage/uploads/;
           expires 30d;
           add_header Cache-Control "public, max-age=2592000";
       }

       location / {
           proxy_pass         http://127.0.0.1:9090;
           proxy_set_header   Host $host;
           proxy_set_header   X-Real-IP $remote_addr;
           proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header   X-Forwarded-Proto $scheme;
       }
   }
   ```

   Adjust `alias`, certificate paths, `proxy_pass` port (`APP_PORT` in `.env`), and `server_name` to match your clone.

4. **Test and reload Nginx**

   ```bash
   sudo nginx -t && sudo systemctl reload nginx
   ```

   Call the API at `https://localhost/...` or `https://api.nextpresskit.local/...`. Set `CORS_ORIGINS` to your frontend’s `https://` origin if it is not same-site.

**Without Nginx:** you can terminate TLS with another tool (for example **Caddy** with automatic local HTTPS, or a dev proxy). The app still listens on HTTP internally; only the edge needs TLS for browser clients.

## Optional: systemd (same machine)

Use this when you want the API as a managed service (restart on failure, start on boot) instead of a foreground terminal.

1. **Build a release-style binary** from the repo root (same layout as servers):

   ```bash
   make build
   ```

2. **Choose how paths map to the unit file**

   The checked-in template [`deploy/systemd/nextpresskit-backend@.service`](../../deploy/systemd/nextpresskit-backend@.service) expects:

   - Clone at `/var/www/nextpresskit-backend-<instance>` (for example `<instance>` = `local` → `/var/www/nextpresskit-backend-local`).
   - `ExecStart` = `/var/www/nextpresskit-backend-<instance>/bin/server`
   - `Environment=APP_ENV=%i` — use instance name **`local`** so the process gets `APP_ENV=local` (this key overrides `APP_ENV` from `.env` if both are set).

   **Layout matching the template:** clone directly into `/var/www/nextpresskit-backend-local` (a real directory, not a symlink into `$HOME`, so a recursive `chown` for `www-data` does not touch your home tree). Then:

   ```bash
   cd /var/www/nextpresskit-backend-local
   make build
   sudo cp deploy/systemd/nextpresskit-backend@.service /etc/systemd/system/
   sudo chown -R www-data:www-data /var/www/nextpresskit-backend-local
   ```

   If the project must stay under your home directory, **do not** point the stock unit at a symlink and then `chown -R` (that can follow into `$HOME`). Instead copy the unit to a dedicated file (for example `/etc/systemd/system/nextpresskit-backend-local.service`), set `User=` / `Group=` to your login user, and set `WorkingDirectory`, `EnvironmentFile`, and `ExecStart` to that path. Use the stock template and [DEPLOYMENT.md](../DEPLOYMENT.md) as references.

3. **Enable and start** (instance name `local`):

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable --now nextpresskit-backend@local
   ```

4. **Logs and status**

   ```bash
   sudo systemctl status nextpresskit-backend@local
   journalctl -u nextpresskit-backend@local -f
   ```

After code changes: `make build && sudo systemctl restart nextpresskit-backend@local`. When models change, run `make migrate-up`. For a local data reset: `make db-fresh` then `make seed` ([COMMANDS.md](../COMMANDS.md#database-and-seed-data)).

## Optional: Nginx in front (local)

Use this to terminate HTTP or HTTPS on a chosen port, serve `/uploads/` from disk, and reverse-proxy to the Go process (`APP_PORT`, default **9090**). Pair with [HTTPS locally (recommended)](#https-locally-recommended) for TLS.

1. **Backend must be listening** (foreground `./scripts/run_local.sh`, `make run`, or systemd above).

2. **Start from the dev site config** [`deploy/nginx/dev.conf`](../../deploy/nginx/dev.conf). Copy it and edit paths and ports, for example:

   ```bash
   sudo cp deploy/nginx/dev.conf /etc/nginx/sites-available/nextpresskit-backend-local.conf
   ```

   In that file:

   - Set `server_name` to `localhost`, a `*.local` hostname, or your machine name.
   - Set `proxy_pass` to `http://127.0.0.1:9090` (or whatever `APP_PORT` is in `.env`).
   - Set the `location /uploads/` `alias` to the **absolute** path of `storage/uploads/` in your clone (same idea as [DEPLOYMENT.md](../DEPLOYMENT.md) uploads note).

   To avoid binding port 80 as root for quick tests, use another `listen` port (for example `8080`) in the `server { ... }` block.

3. **Enable and reload**

   ```bash
   sudo ln -sf /etc/nginx/sites-available/nextpresskit-backend-local.conf /etc/nginx/sites-enabled/
   sudo nginx -t && sudo systemctl reload nginx
   ```

4. **Call the API** through Nginx (example): `http://localhost:8080/...` if you used `listen 8080`.

For **HTTPS**, add a `listen 443 ssl` server (or let Certbot modify the site on a **public** hostname); on a laptop, prefer **mkcert** as in [HTTPS locally (recommended)](#https-locally-recommended). Server-grade TLS with Let’s Encrypt is documented in [DEPLOYMENT.md § TLS](../DEPLOYMENT.md#4-tls-https).

## Optional: Elasticsearch

`ELASTICSEARCH_ENABLED=true`, `ELASTICSEARCH_URLS`. With `APP_ENV` `local` or `dev`, index auto-create defaults on. `GET /posts/search` → 501 if disabled. Reindex: `POST /admin/posts/search/reindex` (`posts:write`).

## Optional: GraphQL

`GRAPHQL_ENABLED=true`; playground only for `APP_ENV` `local` or `dev`. After editing `internal/graphql/schema.graphqls`: `make graphql`.

## Tests

```bash
go test ./...
go vet ./...
```

---

**See also:** [Documentation index](../README.md) · [macOS](macos.md) · [Server deployment](../DEPLOYMENT.md) · [Seeding](../SEEDING.md) · [Roadmap](../ROADMAP.md) · [TODO](../TODO.md) (optional ES / GraphQL / tests **`[x]`** / **`[ ]`**)
