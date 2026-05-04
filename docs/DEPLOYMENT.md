# Deployment

[← Documentation index](README.md) · [Quick start (run locally)](../README.md#quick-start) · [Command reference](COMMANDS.md)

How to run NextPressKit on Ubuntu behind Nginx and systemd, using `scripts/deploy`. For development on your own machine (foreground `make run` or optional local Nginx/systemd), see [Local development](deployment/local.md) and [macOS](deployment/macos.md).

## Quick command meanings

- `./scripts/deploy`: interactive wizard that asks questions and writes ready-to-install nginx/systemd snippets.
- `make deploy`: convenience wrapper around the same deploy wizard.
- `./scripts/nextpresskit deploy`: same deploy flow through the unified command runner.

This deploy flow does not replace local bootstrap. For a fresh local clone, use `./scripts/nextpresskit setup` first.

Database: after deploy, run migrate-up so the live schema matches the code. seed is optional (often used on staging, rarely on production). db-fresh is for local machines only, not production. See [COMMANDS.md](COMMANDS.md#database-and-seed-data) and [SEEDING.md](SEEDING.md).

## Already running the project?

Existing clones keep working: HTTP Nginx configs, `APP_PORT`, and `scripts/deploy` are unchanged. HTTPS is optional (Let’s Encrypt below). If you add TLS, update `CORS_ORIGINS` and client base URLs to `https://` where applicable. Local HTTPS with mkcert is in [deployment/local.md](deployment/local.md).

## Contents

- [Interactive deploy (`scripts/deploy`)](#interactive-deploy-scriptsdeploy)
- [Git branches](#git-branches)
- [Environment matrix](#environment-matrix)
- [Prerequisites](#server-prerequisites)
- [First clone](#first-clones)
- [Per-environment setup](#per-environment-setup-repeat-for-each-tier)
- [Config templates](#config-templates)

---

## Interactive deploy (`scripts/deploy`)

**Always interactive** (no `production`/`staging` CLI arguments). From the **clone root**:

| Platform | Command |
|----------|---------|
| Linux / macOS / Git Bash | `./scripts/deploy` or `make deploy` |
| Windows (PowerShell) | `.\scripts\deploy.ps1` |

The wizard asks for **tier** (production / staging / dev / local), whether to **generate Nginx** (+ optional **systemd** on Linux), **server_name**, **paths**, **APP_PORT**, **Nginx listen port**, and **TLS mode** (HTTP only, Let’s Encrypt with optional `certbot --nginx`, or HTTPS with PEM paths). It writes **`deploy/generated/<slug>/`** (gitignored) with `nginx-nextpresskit-backend-<tier>.conf`, optional `nextpresskit-backend@<tier>.service`, and a **README** with install commands.

On **Linux**, you can opt in to **`sudo`** copy into `/etc/nginx` and reload Nginx. You can optionally run **`certbot --nginx`** when TLS mode is Let’s Encrypt.

The wizard can also run a **release** step on this machine: update git to the right branch, **`go build`**, **`bin/migrate -command=up`**, optionally **`bin/seed`** if you use `RUN_SEED_ON_DEPLOY` ([SEEDING.md](SEEDING.md)), then restart systemd if configured. That flow does **not** include **`db-fresh`** (dev only—[COMMANDS.md](COMMANDS.md#database-and-seed-data)).

You can still use the checked-in files under `deploy/nginx/` and `deploy/systemd/`; the wizard is optional.

---

## Git branches

| Branch | Role | Typical clone path |
|--------|------|--------------------|
| `dev` | Integration | `/var/www/nextpresskit-backend-dev` |
| `staging` | Pre-production | `/var/www/nextpresskit-backend-staging` |
| `main` | Production | `/var/www/nextpresskit-backend-production` |

Promotion: merge and push `dev` → `staging` → `main`.

```bash
git checkout staging && git pull && git merge dev && git push
git checkout main && git pull && git merge staging && git push
```

After releases, back-merge `main` into `staging` and `dev`. Avoid force-push to `main` / `staging`. Exercise migrations on dev/staging before production.

---

## Environment matrix

One row per deployed instance. Ports are examples when several envs share a host.

| Tier | Branch | Directory | `APP_ENV` | systemd | Nginx config | Example `APP_PORT` |
|------|--------|-----------|-----------|---------|--------------|------------------|
| Production | `main` | `/var/www/nextpresskit-backend-production` | `production` | `nextpresskit-backend@production` | `deploy/nginx/production.conf` | 9090 |
| Staging | `staging` | `/var/www/nextpresskit-backend-staging` | `staging` | `nextpresskit-backend@staging` | `deploy/nginx/staging.conf` | 9091 |
| Dev | `dev` | `/var/www/nextpresskit-backend-dev` | `dev` | `nextpresskit-backend@dev` | `deploy/nginx/dev.conf` | 9092 |

---

## Server prerequisites

Ubuntu 22.04+ (or similar), Go (`go.mod`), PostgreSQL, Nginx, Git, systemd, deploy user with repo access.

---

## First clone(s)

```bash
sudo mkdir -p /var/www && sudo chown "$USER" /var/www
git clone <repo-url> /var/www/nextpresskit-backend-production
```

Repeat paths for staging/dev if needed.

---

## Per-environment setup (repeat for each tier)

### 1. Environment file

```bash
cd /var/www/nextpresskit-backend-<tier>   # production | staging | dev
cp .env.example .env
```

Set `APP_PORT`, `DB_*`, `JWT_SECRET`, and **`APP_ENV`** (`production` | `staging` | `dev`).

For browser clients, configure **`JWT_AUTH_SOURCE`** (`cookie` default vs `header`), **`JWT_COOKIE_*`**, and an explicit **`CORS_ORIGINS`** when using cross-site HttpOnly cookies (see [SECURITY.md](SECURITY.md)).

### 2. systemd (install template once per machine)

Repository file: `deploy/systemd/nextpresskit-backend@.service` (`WorkingDirectory=/var/www/nextpresskit-backend-%i`, `APP_ENV=%i`, `EnvironmentFile`, `ExecStart=.../bin/server`).

```bash
sudo cp deploy/systemd/nextpresskit-backend@.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable nextpresskit-backend@<tier>
sudo systemctl start nextpresskit-backend@<tier>
```

`<tier>` ∈ `production`, `staging`, `dev`.

### 3. Nginx

```bash
sudo cp deploy/nginx/production.conf /etc/nginx/sites-available/nextpresskit-backend-production.conf
sudo ln -sf /etc/nginx/sites-available/nextpresskit-backend-production.conf /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

Use `staging.conf` / `dev.conf` for other tiers. Edit **`server_name`** and **`proxy_pass`** → `http://127.0.0.1:<APP_PORT>`.

**Uploads:** align `.env` (`MEDIA_PUBLIC_BASE_URL=/uploads`, `MEDIA_STORAGE_DIR=storage/uploads`) with Nginx `alias` to the absolute `storage/uploads` path.

### 4. TLS (HTTPS)

Terminate TLS at **Nginx** so browsers talk `https://` to your `server_name`. The Go process stays on **`http://127.0.0.1:<APP_PORT>`**; keep **`proxy_set_header X-Forwarded-Proto $scheme`** as in [`deploy/nginx/production.conf`](../deploy/nginx/production.conf) so the scheme reflects the client connection after Certbot adds `listen 443 ssl`.

#### Prerequisites

- **DNS:** `A` (and `AAAA` if you use IPv6) records for the hostname point to this server.
- **Firewall:** allow **80** (ACME HTTP-01 challenge and redirect) and **443**.
- **Nginx:** site enabled with the real `server_name` (matches the certificate you request).

#### Certbot (recommended)

Install the Nginx plugin and obtain a certificate; Certbot will add a `listen 443 ssl` server (or extend your `server` block) and wire `ssl_certificate` paths.

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d api.yourdomain.com
```

Repeat `-d` for each hostname on the same cert, or run Certbot separately per tier (`api-staging.yourdomain.com`, etc.). Use **staging** first if you are validating a new setup.

Certbot usually adds **HTTP → HTTPS** redirect (`return 301` or equivalent). If you manage configs by hand, add a small `listen 80` server that only redirects:

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;
    return 301 https://$host$request_uri;
}
```

#### Renewal

Let’s Encrypt certificates are short-lived. On Ubuntu, `certbot` typically installs a **systemd timer**; check with:

```bash
systemctl list-timers | grep certbot
sudo certbot renew --dry-run
```

If `renew` succeeds, production renewals should run unattended. After manual cert changes, `sudo nginx -t && sudo systemctl reload nginx`.

#### Verification

```bash
curl -sI https://api.yourdomain.com/health
openssl s_client -connect api.yourdomain.com:443 -servername api.yourdomain.com </dev/null 2>/dev/null | openssl x509 -noout -dates
```

#### Manual certificates (no Certbot)

If you use another CA or internal PKI, see the commented template **[`deploy/nginx/production.ssl-snippet.conf.example`](../deploy/nginx/production.ssl-snippet.conf.example)** for `listen 443 ssl`, certificate paths, and the same `proxy_pass` / `X-Forwarded-Proto` pattern.

### 5. Run the interactive deploy wizard

From that environment’s clone (after `.env`, Nginx, and systemd are in place):

```bash
chmod +x scripts/deploy
./scripts/deploy
```

Use the prompts to refresh generated configs, run **Let’s Encrypt** if needed, and optionally **build / migrate / restart** the service.

---

## Config templates

`deploy/nginx/` and `deploy/systemd/` contain only snippets; procedural steps are in this document. Optional manual TLS comments: [`deploy/nginx/production.ssl-snippet.conf.example`](../deploy/nginx/production.ssl-snippet.conf.example).

---

**See also:** [Documentation index](README.md) · [Local development](deployment/local.md) · [macOS](deployment/macos.md) · [TODO](TODO.md) (ops / platform **`[ ]`** items)
