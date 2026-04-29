# Deployment

How to run NextPress Backend on **Ubuntu** behind **Nginx** and **systemd**, using **`scripts/deploy`**. For development on your own machine (foreground `make run` or optional local Nginx/systemd), see [Local development](deployment/local.md).

## Contents

- [Git branches](#git-branches)
- [Environment matrix](#environment-matrix)
- [What `scripts/deploy` does](#what-scriptsdeploy-does)
- [Prerequisites](#server-prerequisites)
- [First clone](#first-clones)
- [Per-environment setup](#per-environment-setup-repeat-for-each-tier)
- [Config templates](#config-templates)

---

## Git branches

| Branch | Role | Typical clone path |
|--------|------|--------------------|
| `dev` | Integration | `/var/www/nextpress-backend-dev` |
| `staging` | Pre-production | `/var/www/nextpress-backend-staging` |
| `main` | Production | `/var/www/nextpress-backend-production` |

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
| Production | `main` | `/var/www/nextpress-backend-production` | `production` | `nextpress-backend@production` | `deploy/nginx/production.conf` | 9090 |
| Staging | `staging` | `/var/www/nextpress-backend-staging` | `staging` | `nextpress-backend@staging` | `deploy/nginx/staging.conf` | 9091 |
| Dev | `dev` | `/var/www/nextpress-backend-dev` | `dev` | `nextpress-backend@dev` | `deploy/nginx/dev.conf` | 9092 |

---

## What `scripts/deploy` does

Run from the **root of that environment’s clone** (`.env` required).

| Step | Action |
|------|--------|
| Checkout | `git fetch` + checkout/pull branch for the tier |
| Build | `bin/server`, `bin/migrate`, `bin/seed` |
| Migrate | `./bin/migrate -command=up` |
| Seed | `./bin/seed` only if `RUN_SEED_ON_DEPLOY=true` |
| Restart | `systemctl restart nextpress-backend@<tier>` if unit exists |

```bash
./scripts/deploy                  # production
./scripts/deploy production | staging | dev
```

---

## Server prerequisites

Ubuntu 22.04+ (or similar), Go (`go.mod`), PostgreSQL, Nginx, Git, systemd, deploy user with repo access.

---

## First clone(s)

```bash
sudo mkdir -p /var/www && sudo chown "$USER" /var/www
git clone <repo-url> /var/www/nextpress-backend-production
```

Repeat paths for staging/dev if needed.

---

## Per-environment setup (repeat for each tier)

### 1. Environment file

```bash
cd /var/www/nextpress-backend-<tier>   # production | staging | dev
cp .env.example .env
```

Set `APP_PORT`, `DB_*`, `JWT_SECRET`, and **`APP_ENV`** (`production` | `staging` | `dev`).

For browser clients, configure **`JWT_AUTH_SOURCE`** (`cookie` default vs `header`), **`JWT_COOKIE_*`**, and an explicit **`CORS_ORIGINS`** when using cross-site HttpOnly cookies (see [SECURITY.md](SECURITY.md)).

### 2. systemd (install template once per machine)

Repository file: `deploy/systemd/nextpress-backend@.service` (`WorkingDirectory=/var/www/nextpress-backend-%i`, `APP_ENV=%i`, `EnvironmentFile`, `ExecStart=.../bin/server`).

```bash
sudo cp deploy/systemd/nextpress-backend@.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable nextpress-backend@<tier>
sudo systemctl start nextpress-backend@<tier>
```

`<tier>` ∈ `production`, `staging`, `dev`.

### 3. Nginx

```bash
sudo cp deploy/nginx/production.conf /etc/nginx/sites-available/nextpress-backend-production.conf
sudo ln -sf /etc/nginx/sites-available/nextpress-backend-production.conf /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

Use `staging.conf` / `dev.conf` for other tiers. Edit **`server_name`** and **`proxy_pass`** → `http://127.0.0.1:<APP_PORT>`.

**Uploads:** align `.env` (`MEDIA_PUBLIC_BASE_URL=/uploads`, `MEDIA_STORAGE_DIR=storage/uploads`) with Nginx `alias` to the absolute `storage/uploads` path.

### 4. TLS

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d your.domain.example
```

### 5. Deploy

```bash
chmod +x scripts/deploy
./scripts/deploy <tier>
```

---

## Config templates

`deploy/nginx/` and `deploy/systemd/` contain only snippets; procedural steps are in this document.

---

**See also:** [Documentation index](README.md) · [Local development](deployment/local.md) · [TODO](TODO.md) (ops / platform **`[ ]`** items)
