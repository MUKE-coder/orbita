# Phase 2 — Grit-awareness end-to-end verification

Verified 2026-07-15 against a real Grit-shaped app, driving the whole Grit-aware cycle through
Orbita's API: **detect → reconcile → build → migrate → route → live**.

## Test fixture

`testdata/grit-sample/` is a minimal but real `api`-mode Grit app (also the Phase 5.1 sample):
`grit.json` (`architecture: api`), `apps/api/` Go module with `cmd/server` (Gin + GORM, serves
`/api/health` + `/api/notes` CRUD), `cmd/migrate` (GORM `AutoMigrate` over `models.Models()`),
and the shipped-style 2-stage Dockerfile (CGO off, `-ldflags="-s -w"`, non-root, `chown` before
`USER`). It ships `go.sum`, exactly like a real Grit app.

## How the E2E was run (reproducible)

Orbita builds from a git remote context. Instead of a GitHub repo, the fixture was served
locally so the whole cycle runs offline:

```bash
# 1. serve the fixture over git:// (Docker build supports git remote contexts)
cd testdata && git init grit-sample && (cd grit-sample && git add -A && git commit -m init)
git daemon --base-path=$PWD --export-all --listen=0.0.0.0 --port=9418

# 2. reconcile an api-mode grit.yaml (postgres addon, api domain, migrate: true)
POST /api/v1/orgs/:org/grit/reconcile   # creates project/env, provisions postgres, creates the
                                        # grit api app, injects env, attaches the domain

# 3. point the app at the local git daemon (the CLI would push to GitHub; here we serve locally)
UPDATE applications SET source_config = jsonb_set(source_config,'{repo_url}',
  '"git://host.docker.internal:9418/grit-sample"') WHERE grit_app='gritsample';

# 4. deploy: build the API from the git context, migrate under advisory lock, cut over
POST /api/v1/orgs/:org/grit/deploy  {"grit_app":"gritsample"}
```

## What was verified

| Step | Evidence |
|------|----------|
| **Detect** | `grit.json` `architecture: api` → one `api` service (build context `apps/api`, Dockerfile `Dockerfile`, :8080). Verified against real stoka too (triple → api/web/admin/docs). |
| **Reconcile (idempotent)** | Created project + prod env, `gritsample-postgres` addon (running), the grit `api` application, injected env, attached `api.gritsample.local`. Re-reconcile → `created=false`, stable counts. |
| **Env injection** | API env carried `DATABASE_URL` (addon, masked), `APP_ENV=production`, `API_URL`, `CORS_ORIGINS`, `PULSE_ENABLED`/`SENTINEL_ENABLED=true` + generated passwords, `GORM_STUDIO_ENABLED=false`. |
| **Build** | Image `orbita-desishub-<id>:v2` built from the git remote context using the shipped Dockerfile + correct `apps/api` context. |
| **Migrate (advisory lock, gates cutover)** | Migration hook took `pg_advisory_lock`, ran `go run ./cmd/migrate` → `notes` table created in the `gritsample-postgres` addon. An earlier run with a broken repo (missing go.sum) **failed the migrator and aborted the deploy without cutting over** — the gate works in both directions. |
| **Route** | `api.gritsample.local` domain attached; Traefik dynamic route written on deploy. |
| **Live** | Container `1/1 running`; `GET /api/health` → `{"database":{"ok":true},"status":"ok"}`; real CRUD: `POST /api/notes` created id:1, `GET /api/notes` read it back — the app is talking to the provisioned addon via the injected `DATABASE_URL`. |

## Notes

- The real private app `github.com/MUKE-coder/stoka-app` (triple mode) can be deployed once a
  GitHub token is stored (the `grit cloud github-auth` flow in Phase 4) so Orbita can clone it;
  detection/plan were verified against its shape. A full triple build (Go + 3 Next.js
  pnpm-workspace images) is heavier but exercises the same code paths proven here.
- Swarm ingress publishing is unreachable from the Windows host loopback (a Docker
  Desktop/Windows limitation seen since Phase 0); liveness was therefore verified in-container,
  which is what Traefik and the health gate use.
