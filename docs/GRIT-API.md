# Grit Cloud API (Orbita Grit-awareness endpoints)

These are the endpoints the `grit deploy` / `grit cloud` CLI calls. All are under
`/api/v1/orgs/:orgSlug/grit/...`, authenticate with a JWT **or** an `orb_` API key
(`Authorization: Bearer <token>`), and are org-scoped by RBAC.

The CLI reads `grit.yaml` (the deploy manifest) and the repo's `grit.json` (the architecture
marker), plus the local `.env.production` referenced by `grit.yaml`'s `env.from`, and submits
them. Orbita derives the service map, build recipe, addons, routing, and migrations from those
— the user never hand-configures a Dockerfile or the UI.

## Detection & derivation (server-side)

- A repo is a Grit app iff `grit.json` exists. `architecture` ∈ {single, double, triple, api}
  is deployable; `mobile` is rejected.
- Services are derived from the mode (grit-knowledge/02–04): `api` builds from `./apps/api`
  with the root Dockerfile; `web`/`admin`/`docs` build from the **repo root** with their own
  Dockerfile and get `NEXT_PUBLIC_API_URL` (from `domains.api`) as a build arg. `single` is
  one container serving SPA+API on :8080. `expo`/`desktop` are never deployed.

---

## POST `/grit/plan` — dry run (viewer+)

Validates the manifest + grit.json and returns the full plan **without mutating anything**.
Backs `grit deploy --plan`.

Request:
```json
{
  "grit_yaml": "<contents of grit.yaml>",
  "grit_json": "<contents of grit.json>",
  "env_values": { "JWT_SECRET": "...", "RESEND_API_KEY": "..." },
  "git_connection_id": "<uuid, optional>"
}
```

Response (`data`):
```json
{
  "grit_app": "stoka",
  "mode": "triple",
  "project_id": "<uuid or empty if new>",
  "environment_id": "<uuid or empty>",
  "services": [
    {"role":"api","app_name":"stoka-api","app_id":"","domain":"api.stoka.com","created":true},
    {"role":"web","app_name":"stoka-web","domain":"stoka.com","created":true},
    {"role":"admin","app_name":"stoka-admin","domain":"admin.stoka.com","created":true},
    {"role":"docs","app_name":"stoka-docs","created":true}
  ],
  "addons": ["postgres","redis"],
  "migrate": true,
  "dashboard_urls": {"pulse":"https://api.stoka.com/pulse/ui", "...": "..."}
}
```

## POST `/grit/reconcile` — create-or-update (developer+)

Idempotently creates-or-updates every Orbita resource for the Grit app: project + production
environment, addons (postgres/redis/minio) in the org's isolated network, one application per
service (grit source type), env injection (user env + addon URLs + generated dashboard creds +
platform overrides), and domains. Re-running reconciles to the same state — no duplicates.

Same request shape as `/grit/plan`. Response: same shape as plan, with real `app_id`s and
`created` reflecting what happened.

## POST `/grit/deploy` — build → migrate → cut over (developer+)

Deploys an already-reconciled Grit app: builds+deploys the API first, runs migrations under a
Postgres advisory lock (gating cutover — a migration failure aborts before any front-end cuts
over), then deploys the front-ends. Keeps previous images for rollback.

Request:
```json
{ "grit_app": "stoka", "environment_id": "<uuid, optional>" }
```

Response (`data`):
```json
{
  "grit_app": "stoka",
  "migrated": true,
  "services": [
    {"role":"api","app_id":"...","status":"running","url":"https://api.stoka.com"},
    {"role":"web","app_id":"...","status":"running","url":"https://stoka.com"}
  ],
  "live_url": "https://stoka.com",
  "api_url": "https://api.stoka.com",
  "dashboard_urls": {"pulse":"https://api.stoka.com/pulse/ui","sentinel":".../sentinel/ui","health":".../api/health"}
}
```

## GET `/grit/:gritApp/status` — status + links (viewer+)

Returns each service's live status and its URL, plus the app/api/dashboard links. Optional
`?environment_id=<uuid>`.

## POST `/grit/:gritApp/rollback` — revert (developer+)

Reverts every service to its previous successful deploy. No "migrate down" — Grit migrations
are additive, so older code tolerates the newer schema (grit-knowledge/06). Optional
`?environment_id=<uuid>`.

---

## Auth for CI/CD

Create an `orb_` key (`POST /me/api-keys`, scope `deploy`) and pass it as the bearer token.
`deploy` scope permits reconcile/deploy/rollback; `read` permits plan/status only.

## grit.yaml contract

See `docs/grit-cloud/grit-knowledge/08-grit-yaml-and-detection.md`. Minimal example:
```yaml
app: stoka
repo: MUKE-coder/stoka
branch: main
addons: [postgres, redis]
domains:
  web: stoka.com
  admin: admin.stoka.com
  api: api.stoka.com
migrate: true
env:
  from: .env.production
```
