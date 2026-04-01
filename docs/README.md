# Documentation

This folder contains the longer-form docs. For quick start, configuration, and common commands see the repository root `README.md`.

## Start here

| Document | Use it when you need… |
|----------|------------------------|
| [PHASES.md](PHASES.md) | The roadmap: what’s implemented and what to build next |
| [openapi.yaml](openapi.yaml) | The API contract (routes + schemas) |
| [SEEDING.md](SEEDING.md) | RBAC defaults, seeded role/permission codes, and seeding commands |

## Working with branches / environments

| Document | Use it when you need… |
|----------|------------------------|
| [GIT_FLOW.md](GIT_FLOW.md) | How `dev` / `staging` / `main` map to deployments; promotion and hotfix flows |

## Deployment

| Document | Use it when you need… |
|----------|------------------------|
| [DEPLOYMENT.md](DEPLOYMENT.md) | Deployment hub (Ubuntu, systemd, Nginx, TLS) |
| [deployment/local.md](deployment/local.md) | Run locally (no systemd/Nginx) |
| [deployment/dev.md](deployment/dev.md) | Deploy the `dev` environment |
| [deployment/staging.md](deployment/staging.md) | Deploy the `staging` environment |
| [deployment/production.md](deployment/production.md) | Deploy the `production` environment (`main`) |

## Notes

- Deployment templates live in `deploy/` (with their own READMEs).
