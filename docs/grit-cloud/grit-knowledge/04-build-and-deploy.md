# 04 — Build & deploy (the shipped Docker files, ports, health, routing)

> **Confirmed direction:** Orbita **reuses the Dockerfiles and `docker-compose.prod.yml` that
> Grit already committed to the repo.** They are production-grade, mode-aware, and
> battle-tested (they carry fixes for real deploy failures — pnpm pinning, `chown` before
> `USER` for embedded SQLite, Next.js standalone output, non-root users). Do **not** replace
> them with Nixpacks or a hand-rolled generator. Only generate a Dockerfile if one is
> genuinely absent (a hand-written or older Grit app) — and when you do, match the shapes
> below exactly.

## The rule

1. Read `grit.json` → get `architecture`.
2. Use the files already in the repo:
   - `single` → `./Dockerfile` (one image).
   - `double`/`triple`/`api` → `./docker-compose.prod.yml` + `apps/*/Dockerfile`.
3. Build each service with the **exact build context** below (this is the part naive tooling
   gets wrong).

## Build contexts and ports — get these exactly right

| Service | Dockerfile | **Build context** | Build args | Listens on |
|---------|-----------|-------------------|-----------|-----------|
| `single` app | `./Dockerfile` | **repo root** (`.`) | — | **8080** |
| `api` | `apps/api/Dockerfile` | **`./apps/api`** | — | **8080** |
| `web` (Next.js) | `apps/web/Dockerfile` | **repo root** (`.`) | `NEXT_PUBLIC_API_URL` | **3000** |
| `admin` (Next.js) | `apps/admin/Dockerfile` | **repo root** (`.`) | `NEXT_PUBLIC_API_URL` | **3000** |
| `docs` (Next.js) | `apps/docs/Dockerfile` | **repo root** (`.`) | — | **3002** |

⚠️ **The API builds from `./apps/api`, but the Next.js apps build from the REPO ROOT** with
`-f apps/web/Dockerfile`. The Next.js Dockerfiles copy the root `package.json`,
`pnpm-workspace.yaml`, and `packages/shared` to install the workspace — they cannot build
from inside `apps/web`. In compose terms:

```yaml
api:
  build: { context: ./apps/api, dockerfile: Dockerfile }
  expose: ["8080"]
web:
  build:
    context: .                       # ← repo root, not ./apps/web
    dockerfile: apps/web/Dockerfile
    args: { NEXT_PUBLIC_API_URL: "https://api.example.com" }
  expose: ["3000"]
```

`NEXT_PUBLIC_API_URL` is a **build-time arg** (baked into the Next.js bundle) — it must be the
app's public API URL (e.g. `https://api.example.com`), set from the `grit.yaml` `domains.api`.
Rebuild the front-ends if the API domain changes.

## What each shipped Dockerfile does (so a generator can match it)

- **`single`** — 3 stages: `node:22-alpine` builds `frontend/dist` (pnpm pinned to `9.15.0`) →
  `golang:1.24-alpine` builds the Go binary with `CGO_ENABLED=0 -ldflags="-s -w"`, dropping
  `frontend/dist` where `//go:embed all:frontend/dist` expects it → `alpine:3.19` runtime,
  non-root `app` user, `chown -R app:app /app` **before** `USER` (Pulse/Sentinel embedded
  SQLite needs write access), `EXPOSE 8080`, `CMD ["./server"]`.
- **`api`** — 2 stages: `golang:1.24-alpine` builds `./cmd/server` (`CGO_ENABLED=0`) →
  `alpine:3.19` runtime, non-root, `chown` before `USER`, `EXPOSE 8080`, `CMD ["./server"]`.
- **`web`/`admin`/`docs` (Next.js)** — multi-stage pnpm workspace build: `deps` installs from
  the root lockfile + that app's `package.json` + `packages/shared`; `builder` runs
  `pnpm --filter <app> build` with `NEXT_PUBLIC_API_URL` as an arg; `runner` copies the
  Next.js **standalone** output (`.next/standalone`, `.next/static`, `public`), non-root user,
  `EXPOSE 3000`, `CMD ["node", "apps/<app>/server.js"]`.

## The production compose topology (`docker-compose.prod.yml`)

Present for `double`/`triple`/`api` (never for `single`). Its defining property:

> **Nothing uses `ports:` — only `expose:`.** No service binds the public host interface.
> Traffic reaches the app **only** through a reverse proxy on the same Docker network.
> Postgres/Redis have no host binding at all.

Services (subset depends on mode): `api`, `web`, `admin`, `docs`, `postgres` (postgres:16-alpine,
healthchecked), `redis` (redis:7-alpine, healthchecked), `minio` (optional — many prod
deploys use R2/S3 instead). The `api` service sets these **production overrides** in its
`environment:` (they win over `.env`):

```yaml
environment:
  APP_ENV: production
  POSTGRES_HOST: postgres          # hit the DB container by service name
  POSTGRES_PORT: "5432"            # container listens on 5432 internally
  REDIS_URL: redis://redis:6379
  MINIO_ENDPOINT: http://minio:9000
```

## Health check

The Go API exposes **`GET /api/health`** — use it as the container/Traefik health check and
as the migrate→cutover gate. It returns JSON like:

```json
{"status":"ok","database":{"ok":true},"redis":{"ok":true},"jobs":{"ok":true}}
```

(HTTP 200 when healthy.) For Next.js services, a `GET /` 200 is a sufficient liveness check.

## Routing (map `grit.yaml domains` → services)

Orbita already owns Traefik; generate routers from the app's domains:

| `grit.yaml` domain | → service:port | notes |
|--------------------|----------------|-------|
| `domains.web` (root / `www`) | `web:3000` | `single`: the one container `:8080` instead |
| `domains.admin` (`admin.`) | `admin:3000` | triple / +admin only |
| `domains.api` (`api.`) | `api:8080` | JSON API; also serves `/pulse/ui`, `/sentinel/ui`, `/studio`, `/docs` |
| `docs.` (optional) | `docs:3002` | when `apps.docs` |

Use Let's Encrypt via Traefik ACME (already in Orbita). HTTP→HTTPS redirect on. Never serve
the API over plain HTTP.

> **`single` mode routing:** there's exactly one container on `:8080`. Point the app's domain
> at it. The SPA is served at `/` and the API at `/api/*` by the same binary — no separate
> `web` router.

**Next:** [`05-env-and-addons.md`](./05-env-and-addons.md) — the env contract + addon wiring.
