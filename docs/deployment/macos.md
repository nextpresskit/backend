# Local development on macOS

**How-to** for running **NextPressKit** on a Mac: toolchain install, typical paths, and pointers to **HTTPS** with mkcert. For stack behavior, migrations, Nginx layout, and cookie auth notes, see [Local development](local.md). For Linux servers, see [DEPLOYMENT.md](../DEPLOYMENT.md).

## Prerequisites

- [Homebrew](https://brew.sh/)
- Go, PostgreSQL, Git (below)

## Install toolchain (Homebrew)

```bash
brew install go postgresql@16 git nginx mkcert nss
```

PostgreSQL version may vary; follow `brew info postgresql` for the current formula name. **Start PostgreSQL** when you need it:

```bash
brew services start postgresql@16
```

Create a database user and database matching `DB_*` in `.env` (for example with `createuser` / `createdb` or a GUI).

## Project setup

Same as the shared tutorial in [local.md § Setup](local.md#setup): clone, `go mod download`, `cp .env.example .env`, set `APP_ENV=local` and credentials.

Run the API:

```bash
make migrate-up
make seed
make run
```

Default URL: `http://localhost:9090` (or your `APP_PORT`). Background mode: `make start` / `make stop` from the root [README](../../README.md).

## Interactive Nginx config

From the repo root, `make deploy` (or `./scripts/deploy`) asks for your hostname, TLS mode, and ports, then writes snippets under `deploy/generated/` with Homebrew-oriented **README** steps. systemd is not used on macOS.

## Nginx paths on Apple Silicon vs Intel

Homebrew’s prefix differs by CPU:

| | Apple Silicon | Intel |
|--|---------------|--------|
| Prefix | `/opt/homebrew` | `/usr/local` |
| Nginx config root (typical) | `/opt/homebrew/etc/nginx` | `/usr/local/etc/nginx` |

Site snippets often go in `servers/` or `sites-enabled/` depending on your `nginx.conf` `include`. Test and reload (Homebrew’s `nginx` binary; use `sudo` for `-t` if your install binds privileged ports):

```bash
"$(brew --prefix nginx)/bin/nginx" -t
brew services restart nginx
```

If `nginx` is on your `PATH`, `nginx -t` is enough. Use the prefix that matches your machine (`brew --prefix nginx`).

## mkcert and HTTPS

1. Install the local CA once:

   ```bash
   mkcert -install
   ```

   This adds trust in the **login** keychain; restart the browser if it still warns on first use.

2. Generate certs and configure Nginx as in [local.md § HTTPS locally (recommended)](local.md#https-locally-recommended). Point `ssl_certificate` and `ssl_certificate_key` at the `.pem` files mkcert created.

3. Optional **custom hostname** (for example `api.nextpress.local`):

   ```bash
   sudo sh -c 'echo "127.0.0.1 api.nextpress.local" >> /etc/private/hosts'
   ```

   Use that name in `server_name` and in `mkcert api.nextpress.local`.

## systemd

macOS does not use systemd. Use foreground **`make run`**, **`make start`** for background, or define a **LaunchAgent** if you need login-level autostart (out of scope here; mirror the working directory and `ExecStart` from [`deploy/systemd/nextpress-backend@.service`](../../deploy/systemd/nextpress-backend@.service) conceptually).

## Server deployment

Production/staging on Ubuntu with Nginx, systemd, and Let’s Encrypt is documented in [DEPLOYMENT.md](../DEPLOYMENT.md); this page is for **local** Mac development only.

---

**See also:** [Local development](local.md) · [Documentation index](../README.md) · [SECURITY.md](../SECURITY.md) (CORS, cookies)
