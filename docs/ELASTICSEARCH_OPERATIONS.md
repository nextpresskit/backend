# Elasticsearch operations runbook

[← Documentation index](README.md) · [Command reference](COMMANDS.md)

Operational guidance for running NextPressKit search in production-like environments.

## Quick start (local/dev)

1. Set `ELASTICSEARCH_ENABLED=true` and `ELASTICSEARCH_URLS=...` in `.env`.
2. Run the API (`make run` or `./scripts/nextpresskit run`).
3. Use `GET /posts/search` and admin reindex `POST /admin/posts/search/reindex` when needed.

Scope:
- Index templates and mappings strategy
- Upgrade and reindex workflow
- Multi-cluster patterns

Related docs: [Local development](deployment/local.md) · [Deployment](DEPLOYMENT.md) · [TODO](TODO.md)

---

## Current behavior (as implemented)

- Elasticsearch is optional; PostgreSQL remains source of truth.
- The posts index name is `<ELASTICSEARCH_INDEX_PREFIX>_posts` (default prefix `nextpresskit`; see `.env.example`).
- When enabled, indexing happens on post save and on scheduled publish promotion.
- Public search route: `GET /posts/search`
- Admin reindex route: `POST /admin/posts/search/reindex` (`posts:write`)
- In `local`/`dev`, auto-create index defaults to on unless explicitly overridden:
  - `ELASTICSEARCH_AUTO_CREATE_INDEX=true|false`

---

## Config checklist

Minimum:
- `ELASTICSEARCH_ENABLED=true`
- `ELASTICSEARCH_URLS=https://es-node-1:9200,https://es-node-2:9200`
- `ELASTICSEARCH_INDEX_PREFIX=nextpresskit`

Auth (choose one):
- API key: `ELASTICSEARCH_API_KEY=...`
- Basic auth: `ELASTICSEARCH_USERNAME=...`, `ELASTICSEARCH_PASSWORD=...`

TLS:
- Keep `ELASTICSEARCH_INSECURE_SKIP_VERIFY=false` in production.
- Use `true` only for local/dev or temporary debugging.

---

## Index template and mapping policy

The app currently creates index mappings directly when auto-create is enabled. For production, prefer explicit templates:

1. Create a composable template for `nextpresskit*_posts*` (or your prefix pattern).
2. Keep fields used by app code compatible:
   - `id`: `keyword`
   - `title`, `excerpt`, `content`: `text`
   - `slug`, `status`: `keyword`
   - `published_at`: `date`
3. Version template changes with semantic names (example: `nextpresskit-posts-v1`, `v2`).
4. Roll template updates in staging first, then production.

Rule: do not deploy mapping changes without a reindex plan.

---

## Upgrade and reindex workflow

Use this when upgrading Elasticsearch major versions, changing analyzers, or changing mapping semantics.

1. Prepare target index
   - New index name with version suffix (example: `nextpresskit_posts_v2`)
   - Apply desired settings/mappings/template
2. Reindex data
   - Trigger app-level repopulation via `POST /admin/posts/search/reindex`
   - Or use Elasticsearch `_reindex` if doing index-to-index migration
3. Validate
   - Check document counts vs published posts in PostgreSQL
   - Spot-check search relevance and sorting
4. Cut over
   - If using aliases, move read alias to new index
   - If not using aliases yet, update `ELASTICSEARCH_INDEX_PREFIX` and restart app
5. Observe and rollback window
   - Keep old index read-only for a rollback period
   - Remove old index only after validation completes

Rollback:
- Restore previous alias/prefix and restart API if needed.
- Re-run admin reindex endpoint if recovery is partial.

---

## Multi-cluster strategy

Recommended baseline:
- One write/read cluster per environment (`dev`, `staging`, `prod`)
- Keep `ELASTICSEARCH_INDEX_PREFIX` environment-specific to avoid collisions

Cross-region or split-read patterns:
- Option A (simplest): one primary cluster per environment, app points only there.
- Option B: replicate indices externally (CCR/snapshots), app still writes to one primary cluster.
- Option C: per-region clusters with region-local app instances and independent indexing.

Application note:
- `ELASTICSEARCH_URLS` can include multiple nodes for one cluster.
- Current app design expects a single logical cluster endpoint set per deployment.

---

## Observability and operations checks

At startup:
- Confirm app logs include Elasticsearch integration active and index name.
- Alert on repeated index/search request failures in logs.

Routine checks:
- Cluster health (green/yellow policy based on SLOs)
- Search latency (`/posts/search`)
- Admin reindex duration and failure rate
- Index size and shard growth trends

Incident checklist:
1. Verify cluster reachability and auth.
2. Verify mapping compatibility for indexed fields.
3. Run admin reindex endpoint for repair.
4. If still failing, cut over to known-good index/alias and investigate offline.

