# systemd unit (template + instances)

This folder contains a single **template unit**: `nextpress-backend@.service`.

- **Template**: `nextpress-backend@.service` (install once)
- **Instances**: `nextpress-backend@production`, `nextpress-backend@staging`, `nextpress-backend@dev`
- **Folder layout**: `/var/www/nextpress-backend-%i` (instance name is substituted for `%i`)
- **APP_ENV**: the unit sets `APP_ENV=%i` (for example `APP_ENV=staging`)

For the full server walkthrough (Nginx, TLS, deploy script) see `docs/DEPLOYMENT.md` and `docs/deployment/*.md`.

## Install (one-time)

```bash
sudo cp deploy/systemd/nextpress-backend@.service /etc/systemd/system/
sudo systemctl daemon-reload
```

## Enable / start an environment

```bash
sudo systemctl enable nextpress-backend@<env>
sudo systemctl start nextpress-backend@<env>
```

Examples:

- `sudo systemctl enable nextpress-backend@production && sudo systemctl start nextpress-backend@production`
- `sudo systemctl enable nextpress-backend@staging && sudo systemctl start nextpress-backend@staging`
- `sudo systemctl enable nextpress-backend@dev && sudo systemctl start nextpress-backend@dev`

## Deploy interaction

`scripts/deploy` (run from the environment folder) will:

- pull the correct branch (`main` / `staging` / `dev`)
- build `bin/server` + `bin/migrate` + `bin/seed`
- run migrations (`migrate up`)
- optionally run seeders if `.env` contains `RUN_SEED_ON_DEPLOY=true`
- restart `nextpress-backend@<env>` if the service exists

