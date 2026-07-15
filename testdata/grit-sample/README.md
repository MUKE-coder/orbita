# grit-sample — Grit Cloud test/demo app

A minimal but real **`api`-mode Grit app** used by Grit Cloud's automated tests and the local
end-to-end demo. It mirrors the shape of a real Grit app so Orbita's detection, build, addon,
migration, and routing paths run against something authentic.

## Shape

```
grit.json                       # {"architecture":"api"} — the detection marker
apps/api/
  go.mod  go.sum                # Go module (at apps/api, like real Grit)
  Dockerfile                    # shipped-style 2-stage build (CGO off, non-root, chown before USER)
  cmd/server/main.go            # Gin API: GET /api/health, GET/POST /api/notes
  cmd/migrate/main.go           # GORM AutoMigrate over models.Models()
  internal/{config,database,models}/
```

## What it exercises

- **Detection**: `grit.json` → `api` mode → one service (build context `apps/api`, :8080).
- **Build**: reuses the shipped Dockerfile via Orbita's git remote-context build.
- **Addon + env**: a `postgres` addon → `DATABASE_URL` injected; the API connects at startup.
- **Migration hook**: `cmd/migrate` runs under a Postgres advisory lock, creating the `notes`
  table before cutover.
- **Health/route/live**: `/api/health` is the readiness gate; `/api/notes` proves a real CRUD
  round-trip against the provisioned database.

## Run it (local, no VPS)

See `docs/PHASE4-E2E.md` — serve this over local git and `grit deploy` it through a local
Orbita. A `grit.yaml` for it looks like:

```yaml
app: gritsample
repo: MUKE-coder/grit-sample
addons: [postgres]
domains:
  api: api.gritsample.local
migrate: true
env:
  from: .env.production
```

For the full **triple** shape (Go API + Next.js web + admin), use a `grit new sample --full`
app or the real `github.com/MUKE-coder/stoka-app`.
