# Contributing

[Docs index](docs/README.md) · [Commands](docs/COMMANDS.md)

This guide is the shortest path to preparing a clean PR.

## Prerequisites

- Go version in [`go.mod`](go.mod)
- PostgreSQL for integration-style checks (optional for some packages)

## Before you open a PR

```bash
make test
go vet ./...
```

Fix or explain any failures.

## API changes

- Update [`docs/openapi.yaml`](docs/openapi.yaml) for REST.
- GraphQL-only changes: edit `internal/graphql/schema.graphqls`, then `make graphql`.

## Documentation (living docs)

Keep narrative and tasks honest when behavior or priorities change — **prefer the same PR** as the code.

| When you… | Update… |
|-----------|---------|
| Ship or cancel a scoped feature | [`docs/TODO.md`](docs/TODO.md) - set **`[x]`** / **`[ ]`** for the lines you touch (full checklist) |
| Change product direction or major capability | [`docs/ROADMAP.md`](docs/ROADMAP.md) (**Shipped** / **In progress** / **Later**) |
| Add RBAC codes or seed data | [`docs/SEEDING.md`](docs/SEEDING.md) and `pkg/seed` |
| Change deploy steps or branches | [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md) |

See also: [docs/README.md](docs/README.md) — **Pick your path** and the full doc table. Full checklist: [docs/TODO.md](docs/TODO.md).

## Branches and servers

Promotion model and server layout: [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md) (Git branches section).

## Questions

Open a discussion or issue on the repository.
