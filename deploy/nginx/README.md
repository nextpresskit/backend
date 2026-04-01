# Nginx configs (reverse proxy + uploads)

This folder contains one Nginx site config per environment. They are designed to:

- terminate HTTP(S) and reverse-proxy to the Go API (`proxy_pass` → `APP_PORT`)
- serve uploaded media files directly from disk under `/uploads/`

For the full server walkthrough (systemd, TLS, folder layout) see `docs/DEPLOYMENT.md` and `docs/deployment/*.md`.

## Files

| File | Environment | Example domain | Example app port |
|------|-------------|----------------|------------------|
| `production.conf` | production (`main`) | `cms.yourdomain.com` | 9090 |
| `staging.conf` | staging (`staging`) | `cms-staging.yourdomain.com` | 9091 |
| `dev.conf` | dev (`dev`) | `cms-dev.yourdomain.com` | 9092 |

## Install / enable (example: production)

```bash
sudo cp deploy/nginx/production.conf /etc/nginx/sites-available/nextpress-backend-production.conf
sudo ln -sf /etc/nginx/sites-available/nextpress-backend-production.conf /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

Repeat for staging/dev if you deploy multiple environments.

## Required edits

In the copied config file:

- set `server_name` to your domain
- set `proxy_pass` to the correct upstream port (must match `APP_PORT` in that environment’s `.env`)

## Uploads (static files)

The configs expose uploads at `/uploads/`. Keep these values consistent:

- `.env`: `MEDIA_PUBLIC_BASE_URL=/uploads`
- `.env`: `MEDIA_STORAGE_DIR=storage/uploads` (relative to the app working directory)
- Nginx `alias`: must point at the absolute path of that folder on disk, e.g.:
  - `/var/www/nextpress-backend-production/storage/uploads/`

If you change the public URL (for example to `/media/`), update both `MEDIA_PUBLIC_BASE_URL` and the Nginx `location` accordingly.

## TLS (Let’s Encrypt)

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d <your-domain>
```

certbot will update the Nginx site config and reload it.

