# Nginx

One config per environment. Edit `server_name` and `proxy_pass` to match your domain and app port.
Uploads are served via Nginx at `/uploads/` by default (must match `MEDIA_PUBLIC_BASE_URL`).

| File              | Domain (example)            | Port (example) |
|-------------------|----------------------------|----------------|
| `production.conf` | cms.yourdomain.com         | 9090           |
| `staging.conf`    | cms-staging.yourdomain.com | 9091           |
| `dev.conf`        | cms-dev.yourdomain.com     | 9092           |

## Enable

```bash
sudo cp deploy/nginx/production.conf /etc/nginx/sites-available/nextpress-backend-production.conf
sudo ln -sf /etc/nginx/sites-available/nextpress-backend-production.conf /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

Repeat for staging/dev if needed, adjusting filenames.

## Uploads (static files)

The configs include:

- `location /uploads/ { alias .../storage/uploads/; }`

Make sure the `alias` path matches your server folder layout (e.g. `/var/www/nextpress-backend-production/storage/uploads/`)
and that the app uses the same `MEDIA_STORAGE_DIR` (default: `storage/uploads`).

## TLS (Let's Encrypt)

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d <your-domain>
```

Replace `<your-domain>` with the domain in your config (e.g. `cms.yourdomain.com`). certbot will update the Nginx config and reload it.

