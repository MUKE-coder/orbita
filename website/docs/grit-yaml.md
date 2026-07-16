# `grit.yaml` spec

`grit.yaml` is the single deploy manifest at your Grit project root. You write it once;
everything about *what services exist and how to build them* is derived from `grit.json` + the
repository.

## Full example

```yaml
app: rental-manager              # required — app name (also the default org/project name)
repo: MUKE-coder/rental-manager  # required — GitHub owner/name (created + pushed if missing)
branch: main                     # default: main

# Optional — override the clone URL Orbita builds from. Normally derived from `repo`.
# Set for a self-hosted / non-GitHub git host.
repo_url: ""

addons:                          # subset of {postgres, redis, minio}; provisioned in the
  - postgres                     #   org's isolated network, URLs injected into the API env
  - redis

domains:                         # bare hostnames (no scheme, port, or path)
  web:   rental.example.com      # root site (Next.js web, or the single-mode container)
  admin: admin.rental.example.com
  api:   api.rental.example.com  # Go API — also serves /pulse/ui, /sentinel/ui, /studio, /docs
  docs:  docs.rental.example.com # only if the app has a docs site

migrate: true                    # default true — run cmd/migrate under an advisory lock

# Embedded dashboards. Defaults: observability & security ON, studio OFF.
observability: true              # Pulse
security: true                   # Sentinel
studio: false                    # GORM Studio — off in prod is safer (it edits live data)

env:
  from: .env.production          # local file; values encrypted into Orbita, never committed
```

## What's derived (you don't write these)

- **Services** — from `grit.json.architecture`:

  | Mode | Containers |
  |---|---|
  | `single` | 1 — one Go binary with the SPA embedded, on `:8080` |
  | `api` | 1 — the Go API on `:8080` |
  | `double` | 2 — `api` + `web` |
  | `triple` | 3 — `api` + `web` + `admin` (+ `docs` if present) |

  `mobile` is rejected; `expo`/`desktop` are never deployed.

- **Build recipe** — the Dockerfiles Grit ships, with the correct contexts (api builds from
  `apps/api`; Next.js apps from the repo root with `NEXT_PUBLIC_API_URL` baked from `domains.api`).
- **Addon env** — `DATABASE_URL`, `REDIS_URL`, `STORAGE_DRIVER` / `MINIO_*`.
- **Platform env** — `APP_ENV=production`, `API_URL`, `CORS_ORIGINS`, dashboard `*_ENABLED` +
  generated `*_PASSWORD`.

## Defaults

| Field | Default |
|---|---|
| `branch` | `main` |
| `migrate` | `true` |
| `observability` / `security` | `true` |
| `studio` | `false` |
| `repo_url` | derived from `repo` |
| ports | api `8080`, web/admin `3000`, docs `3002` |

## Validation rules

::: warning
- `app` and `repo` are required; `repo` must be `owner/name`.
- `domains.*` must be **bare hostnames** — no `https://`, no `:port`, no `/path`, and FQDN.
- `addons` ⊆ `{postgres, redis, minio}`.
- The architecture must be VPS-deployable (`single`/`double`/`triple`/`api`).
:::
