# 03 — Real project structure (actual on-disk paths)

> ⚠️ **Correction to watch for.** The example `grit.yaml` in `project-description.md` §5 uses
> illustrative paths like `path: ./cmd/api` and `path: ./web`. **Real Grit apps do not use
> those paths.** In the monorepo modes the API is at **`apps/api`** (entry
> `apps/api/cmd/server/main.go`) and the front-ends are at **`apps/web`** / **`apps/admin`**.
> Derive paths from the mode + `grit.json` (see `08-grit-yaml-and-detection.md`); don't trust
> hand-written ones blindly, and default them correctly.

## `triple` (and `double`, `--full`) — the pnpm monorepo

```
myapp/
├── grit.json                     # architecture marker (read this first)
├── grit.config.ts                # framework config (TS)
├── package.json                  # pnpm workspace root
├── pnpm-workspace.yaml           # workspaces: apps/*, packages/*
├── turbo.json
├── docker-compose.yml            # DEV infra (postgres/redis/minio/mailhog, host ports)
├── docker-compose.prod.yml       # PROD stack (api/web/admin/postgres/redis/minio, expose only)
├── .env  /  .env.example         # the env contract (see 05)
├── .dockerignore
├── apps/
│   ├── api/                      # ← the Go backend (its own Go module)
│   │   ├── go.mod  go.sum
│   │   ├── cmd/server/main.go    # ← entry point (built by the Dockerfile as ./cmd/server)
│   │   ├── internal/             # config, database, models, handlers, services,
│   │   │                         #   middleware, routes, backup, jobs, cron, sync, ...
│   │   ├── Dockerfile            # ← shipped, production-ready (2-stage Go build)
│   │   └── tmp/                  # air hot-reload output (dev only; ignored)
│   ├── web/                      # ← Next.js App Router (public site)
│   │   ├── app/  components/  hooks/  lib/
│   │   ├── next.config.ts        # output: 'standalone'  ← required for the Docker runtime
│   │   ├── package.json
│   │   └── Dockerfile            # ← shipped (pnpm-workspace multi-stage, standalone)
│   ├── admin/                    # ← Next.js admin panel (present in triple / +admin)
│   │   └── ... + Dockerfile
│   ├── docs/                     # ← Next.js docs (only when apps.docs = true, e.g. --full)
│   │   └── ... + Dockerfile
│   ├── expo/                     # ← React Native (only if apps.expo). NOT VPS-DEPLOYED.
│   └── desktop/                  # ← Wails desktop (only if apps.desktop). NOT VPS-DEPLOYED.
└── packages/
    └── shared/                   # Zod schemas + TS types + constants, imported as @repo/shared
```

Key facts:
- **The Go module lives under `apps/api/`** (not the repo root). Its entry is
  `apps/api/cmd/server/main.go`; the Dockerfile builds it as `go build ./cmd/server`.
- **`apps/web` and `apps/admin` are Next.js apps with `output: 'standalone'`.** Their
  Dockerfiles build from the **monorepo root context** (they need the root `package.json`,
  `pnpm-workspace.yaml`, and `packages/shared`). See `04-build-and-deploy.md` — this is the
  single most important build detail.
- `packages/shared` is a workspace package (`@repo/shared`) consumed by every front-end.

## `single` — flat, one binary, embedded SPA

```
myapp/
├── grit.json                     # "architecture": "single"
├── go.mod  go.sum                # ← Go module at the REPO ROOT
├── main.go                       # ← entry at root; contains //go:embed all:frontend/dist
├── internal/                     # Go code (config, database, models, handlers, ...)
├── frontend/                     # ← Vite SPA
│   ├── src/
│   ├── package.json  pnpm-lock.yaml
│   └── dist/                     # built SPA (created during Docker build; embedded)
├── Dockerfile                    # ← at the ROOT (3-stage: build SPA → build Go → runtime)
├── .env  /  .env.example
└── Makefile                      # make dev / make build
```

- **No `apps/` folder, no `docker-compose.prod.yml`.** One container serves everything on
  `:8080`.

## `api` — Go API only

```
myapp/
├── grit.json                     # "architecture": "api"
├── package.json  pnpm-workspace.yaml   # minimal workspace (no web/admin/shared)
├── docker-compose.prod.yml       # api + postgres + redis + minio
├── .env / .env.example
└── apps/
    └── api/                      # the Go backend (same layout as triple's apps/api)
        ├── go.mod  cmd/server/main.go  internal/  Dockerfile
```

- No `web` / `admin` / `packages/shared`. Serves JSON + Scalar API docs at `/docs`.

## Two config files you can read

- **`grit.json`** — the machine marker (mode, frontend, which apps). *Always present. Parse
  this.*
- **`grit.config.ts`** — a TypeScript config (project name, feature toggles). Human-facing;
  you generally don't need to parse it for deploy — `grit.json` + the filesystem is enough.

**Next:** [`04-build-and-deploy.md`](./04-build-and-deploy.md) — the shipped Dockerfiles,
build contexts, ports, health, and routing.
