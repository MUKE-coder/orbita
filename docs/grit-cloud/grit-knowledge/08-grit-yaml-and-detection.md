# 08 — `grit.yaml` contract, detection & deriving everything

This file reconciles the **`grit.yaml` deploy manifest** (the contract in
`project-description.md` §5) with the **real Grit app** on disk, and tells you how to derive
what `grit.yaml` doesn't say. This feeds Phase 2.1 (schema + validator) and the whole deploy
reconcile.

## The manifest is small; the app is self-describing — prefer detection

Design principle from the vision: *the user writes `grit.yaml` once; everything else is
inferred.* The best source of truth for "what services exist and how to build them" is **not**
the hand-written `grit.yaml` `services` block — it's **`grit.json` + the filesystem**, which
Grit itself generates and keeps correct. So:

> **Derive the service map, paths, Dockerfiles, and ports from `grit.json` + the repo. Use
> `grit.yaml` for the things only the user knows: repo, branch, domains, addons, env source,
> migrate toggle.**

### ⚠️ Path correction (important)

The example in `project-description.md` §5 writes:

```yaml
services:
  api:    { type: go,     path: ./cmd/api, port: 8080 }   # ← illustrative, NOT real Grit paths
  web:    { type: nextjs, path: ./web }
```

Real Grit paths are **`./apps/api`** (entry `apps/api/cmd/server`) and **`./apps/web`** /
**`./apps/admin`** (see `03-project-structure.md`). Either (a) **ignore user-written
`services.*.path` and derive from the mode**, or (b) if you keep an explicit `services` block,
**default its paths to the real Grit locations** and validate them against the filesystem.
Don't hard-code `./cmd/api` / `./web`.

## Detection algorithm (implement this in `internal/grit/`)

```
1. Is this a Grit app?           → grit.json exists at repo root. Else: not Grit.
2. mode      = grit.json.architecture      # single | double | triple | api | mobile
3. frontend  = grit.json.frontend          # next | vite
4. hasDocs   = grit.json.apps.docs
   (ignore grit.json.apps.expo / .desktop  → never VPS-deployed)
5. Build the service list from the mode:
     single → [ app(root Dockerfile, :8080, serves SPA+API) ]
     api    → [ api(apps/api, :8080) ]
     double → [ api(apps/api, :8080), web(apps/web, :3000) ]
     triple → [ api(apps/api, :8080), web(apps/web, :3000), admin(apps/admin, :3000) ]
              (+ docs(apps/docs, :3002) if hasDocs)
6. For each service, resolve the Dockerfile + build context (see 04):
     single  → ./Dockerfile,               context = .
     api     → apps/api/Dockerfile,         context = ./apps/api
     web/admin/docs → apps/<svc>/Dockerfile, context = .   (repo root!)
   If a Dockerfile is missing (rare / hand-written app), generate one matching 04's shape.
7. mode == mobile → refuse VPS deploy with a clear message ("mobile apps ship to stores").
```

## Recommended `grit.yaml` (matches reality)

```yaml
app: rental-manager
repo: MUKE-coder/rental-manager     # GitHub owner/repo; created + pushed if missing
branch: main

# Optional — usually OMITTED. If present, paths are the REAL Grit paths and are
# validated against the filesystem. When omitted, derive from grit.json (preferred).
# services:
#   api:   { path: ./apps/api, port: 8080 }
#   web:   { path: ./apps/web, port: 3000 }
#   admin: { path: ./apps/admin, port: 3000 }

addons:
  - postgres
  - redis
  - minio            # omit if using S3/R2 (set STORAGE_DRIVER + keys in env instead)

domains:
  web:   hmkestates.com
  admin: admin.hmkestates.com
  api:   api.hmkestates.com

migrate: true        # default true; runs cmd/migrate under advisory lock before cutover

# Optional dashboards (default on):
# observability: true   # Pulse   (PULSE_ENABLED)
# security: true        # Sentinel (SENTINEL_ENABLED)
# studio: false         # GORM Studio — off by default in prod is safer

env:
  from: .env.production  # local file; values encrypted into Orbita, never committed
```

### Defaults when a field is omitted
- `branch` → `main`. `migrate` → `true`. `observability`/`security` → `true`. `studio` → `false`.
- `services` → derived from `grit.json` (preferred). `port` → api `8080`, web/admin `3000`, docs `3002`.
- environment → `production`. Addon hosts → the in-network service names (see 05).

## `grit.yaml` field → Orbita action

| Field | Orbita does |
|-------|-------------|
| `app` | app name within the org/project |
| `repo` / `branch` | ensure GitHub repo exists (create+push if missing); source = repo+branch |
| `services` (or derived) | one container per service; build with the shipped Dockerfile + correct context (04) |
| `addons` | provision postgres/redis/minio in the org's isolated network; inject URLs (05) |
| `domains.{web,admin,api}` | Traefik routers → `web:3000` / `admin:3000` / `api:8080` (04); ACME TLS |
| `migrate` | run `cmd/migrate` under advisory lock, gate cutover on success (06) |
| `observability`/`security`/`studio` | set `*_ENABLED` + generate strong passwords; expose sub-paths (07) |
| `env.from` | read the local file, encrypt values into Orbita, inject at runtime |

## Validation (Phase 2.1)

Validate a submitted `grit.yaml`:
- `app` and `repo` present; `repo` is `owner/name`.
- `domains` values are bare hostnames (no scheme/path).
- `addons` ⊆ `{postgres, redis, minio}`.
- If `services` present, each `path` exists in the repo and matches the mode.
- `grit.json` exists and `architecture` ∈ `{single, double, triple, api}` for a VPS deploy
  (reject `mobile` with a helpful message).

---

That's the full picture. With `grit.json` + these eight files, Orbita can classify any Grit
repo, build it with the files Grit already ships, wire its addons + domains + dashboards, run
its migrations safely, and cut over — no Nixpacks guessing and no user-authored Dockerfile.
