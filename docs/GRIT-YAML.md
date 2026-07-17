# The `orbita.yaml` spec

`orbita.yaml` is the single deploy manifest at your Grit project root. You write it once;
everything about *what services exist and how to build them* is derived from `grit.json` + the
repository. Parsed and validated by `internal/grit` (shared by Orbita and the CLI).

```yaml
app: rental-manager              # required — app name (also the default org/project name)
repo: MUKE-coder/rental-manager  # required — GitHub owner/name (created + pushed if missing)
branch: main                     # default: main

# Optional — override the clone URL Orbita builds from. Normally derived from `repo` as
# https://github.com/<owner>/<name>.git. Set for a self-hosted / non-GitHub git host.
repo_url: ""

addons:                          # subset of {postgres, redis, minio}; provisioned in the
  - postgres                     #   org's isolated network, URLs injected into the API env
  - redis

domains:                         # bare hostnames (no scheme, port, or path)
  web:   rental.example.com      # root site (Next.js web, or the single-mode container)
  admin: admin.rental.example.com
  api:   api.rental.example.com  # Go API — also serves /pulse/ui, /sentinel/ui, /studio, /docs
  docs:  docs.rental.example.com # only if the app has a docs site

migrate: true                    # default true — run cmd/migrate under a Postgres advisory
                                 #   lock before cutover; false = you manage schema yourself

# Embedded dashboards (grit-knowledge/07). Defaults: observability & security ON, studio OFF.
observability: true              # Pulse   (PULSE_ENABLED)
security: true                   # Sentinel (SENTINEL_ENABLED)
studio: false                    # GORM Studio — off in prod is safer (it edits live data)

env:
  from: .env.production          # local file; values encrypted into Orbita, never committed
```

## What's derived (you don't write these)

- **Services**: from `grit.json.architecture` — `single` (1 container), `api` (Go API),
  `double` (api+web), `triple` (api+web+admin, +docs if present). `mobile` is rejected;
  `expo`/`desktop` are never deployed.
- **Build recipe**: the Dockerfiles Grit already ships, with the correct contexts (api builds
  from `apps/api`; Next.js apps from the repo root with `NEXT_PUBLIC_API_URL` baked from
  `domains.api`).
- **Addon env**: `DATABASE_URL`, `REDIS_URL`, `STORAGE_DRIVER`/`MINIO_*` — set from the
  provisioned addons.
- **Platform env**: `APP_ENV=production`, `API_URL`, `CORS_ORIGINS`, `OAUTH_FRONTEND_URL`,
  dashboard `*_ENABLED` + generated `*_PASSWORD`.

## Defaults when a field is omitted

| Field | Default |
|-------|---------|
| `branch` | `main` |
| `migrate` | `true` |
| `observability` / `security` | `true` |
| `studio` | `false` |
| `repo_url` | derived from `repo` |
| service ports | api `8080`, web/admin `3000`, docs `3002` |
| environment | `production` |

## Validation rules

- `app` and `repo` are required; `repo` must be `owner/name`.
- `domains.*` must be bare hostnames (no `https://`, no `:port`, no `/path`, and FQDN).
- `addons` ⊆ `{postgres, redis, minio}`.
- The app's architecture must be VPS-deployable (`single`/`double`/`triple`/`api`).
