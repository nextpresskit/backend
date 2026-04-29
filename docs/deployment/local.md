# Local development

**Tutorial / how-to** - run the API on your machine. The default path is `go run` / `make run` with no reverse proxy. Optionally use **systemd** and **Nginx** on the same box to mirror how [server deployment](../DEPLOYMENT.md) runs the binary behind a proxy.

## Prerequisites

- Go (`go.mod`)
- PostgreSQL
- Git

## Setup

```bash
git clone <repo-url> nextpress-backend && cd nextpress-backend
go mod download
cp .env.example .env
```

Configure `DB_*`, `JWT_SECRET`, `APP_ENV=local`. For auth, set `JWT_AUTH_SOURCE` (`cookie` default vs `header`) and `JWT_COOKIE_*` / cookie names as needed; use `CORS_ORIGINS` when testing cookie auth from a separate frontend origin. Optional flags: Elasticsearch, GraphQL - [`.env.example`](../../.env.example).

## Migrate, seed, run

```bash
make migrate-up
make seed
./scripts/run_local.sh   # or: go run ./cmd/api
```

| | |
|--|--|
| Base URL | `http://localhost:<APP_PORT>` (default 9090) |
| Checks | `GET /health`, `GET /ready` |
| REST | [`docs/openapi.yaml`](../openapi.yaml) |

## Optional: systemd (same machine)

Use this when you want the API as a managed service (restart on failure, start on boot) instead of a foreground terminal.

1. **Build a release-style binary** from the repo root (same layout as servers):

   ```bash
   make build
   ```

2. **Choose how paths map to the unit file**

   The checked-in template [`deploy/systemd/nextpress-backend@.service`](../../deploy/systemd/nextpress-backend@.service) expects:

   - Clone at `/var/www/nextpress-backend-<instance>` (for example `<instance>` = `local` → `/var/www/nextpress-backend-local`).
   - `ExecStart` = `/var/www/nextpress-backend-<instance>/bin/server`
   - `Environment=APP_ENV=%i` — use instance name **`local`** so the process gets `APP_ENV=local` (this key overrides `APP_ENV` from `.env` if both are set).

   **Layout matching the template:** clone directly into `/var/www/nextpress-backend-local` (a real directory, not a symlink into `$HOME`, so a recursive `chown` for `www-data` does not touch your home tree). Then:

   ```bash
   cd /var/www/nextpress-backend-local
   make build
   sudo cp deploy/systemd/nextpress-backend@.service /etc/systemd/system/
   sudo chown -R www-data:www-data /var/www/nextpress-backend-local
   ```

   If the project must stay under your home directory, **do not** point the stock unit at a symlink and then `chown -R` (that can follow into `$HOME`). Instead copy the unit to a dedicated file (for example `/etc/systemd/system/nextpress-backend-local.service`), set `User=` / `Group=` to your login user, and set `WorkingDirectory`, `EnvironmentFile`, and `ExecStart` to that path. Use the stock template and [DEPLOYMENT.md](../DEPLOYMENT.md) as references.

3. **Enable and start** (instance name `local`):

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable --now nextpress-backend@local
   ```

4. **Logs and status**

   ```bash
   sudo systemctl status nextpress-backend@local
   journalctl -u nextpress-backend@local -f
   ```

After code changes, rebuild and restart: `make build && sudo systemctl restart nextpress-backend@local`. Run migrations from the repo as usual (`make migrate-up`) before or after restarts as needed.

## Optional: Nginx in front (local)

Use this to terminate HTTP on port 80 (or another port), serve `/uploads/` from disk, and reverse-proxy to the Go process (`APP_PORT`, default **9090**).

1. **Backend must be listening** (foreground `./scripts/run_local.sh`, `make run`, or systemd above).

2. **Start from the dev site config** [`deploy/nginx/dev.conf`](../../deploy/nginx/dev.conf). Copy it and edit paths and ports, for example:

   ```bash
   sudo cp deploy/nginx/dev.conf /etc/nginx/sites-available/nextpress-backend-local.conf
   ```

   In that file:

   - Set `server_name` to `localhost`, a `*.local` hostname, or your machine name.
   - Set `proxy_pass` to `http://127.0.0.1:9090` (or whatever `APP_PORT` is in `.env`).
   - Set the `location /uploads/` `alias` to the **absolute** path of `storage/uploads/` in your clone (same idea as [DEPLOYMENT.md](../DEPLOYMENT.md) uploads note).

   To avoid binding port 80 as root for quick tests, use another `listen` port (for example `8080`) in the `server { ... }` block.

3. **Enable and reload**

   ```bash
   sudo ln -sf /etc/nginx/sites-available/nextpress-backend-local.conf /etc/nginx/sites-enabled/
   sudo nginx -t && sudo systemctl reload nginx
   ```

4. **Call the API** through Nginx (example): `http://localhost:8080/...` if you used `listen 8080`.

TLS on a laptop is optional; on a LAN you can keep HTTP or use `mkcert` / corporate CA. For a public hostname on Ubuntu, [DEPLOYMENT.md](../DEPLOYMENT.md) documents Certbot with Nginx.

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

**See also:** [Documentation index](../README.md) · [Server deployment](../DEPLOYMENT.md) · [Seeding](../SEEDING.md) · [Roadmap](../ROADMAP.md) · [TODO](../TODO.md) (optional ES / GraphQL / tests **`[x]`** / **`[ ]`**)
