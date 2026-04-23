# CLAUDE.md

Guidance for Claude Code when working in this repository.

## Project

**Orbita** — self-hosted multi-tenant PaaS. Go backend + embedded React SPA shipped as a single binary (~30MB, <50MB idle RAM). Competes with Dokploy/Coolify on true tenant isolation (per-org Docker networks, encryption keys, cgroup v2 quotas, 4-role RBAC).

- Module path: `github.com/orbita-sh/orbita`
- Go: **1.25** (go.mod says 1.25.0 — README's "1.22+" is a minimum claim, not the toolchain)
- Frontend: React **19** + Vite 8 + Tailwind v4 + shadcn/ui (package name in `web/package.json` is `web-temp` — don't "fix" it)
- Single binary via `//go:embed all:web/dist` in [static.go](static.go)

## Commands

Run from repo root. `make` targets are in [Makefile](Makefile).

| Task | Command |
|------|---------|
| Dev backend (hot reload via Air) | `make dev` |
| Dev frontend | `cd web && npm run dev` (Vite on :5173) |
| Dev DB + Redis | `make docker-up` / `make docker-down` |
| Build production binary | `make build` (builds web first, then Go with `-ldflags="-s -w"`) |
| Run migrations | `make migrate` / `make migrate-down` |
| Tests | `make test` (`-v -race`) |
| Lint | `make lint` (golangci-lint) |

Config loads from `.env` (see [.env.example](.env.example)). Required: `JWT_SECRET`, `JWT_REFRESH_SECRET`. Startup fails hard without them. See [internal/config/config.go](internal/config/config.go).

## Architecture

Strict layering — do not skip layers:

```
Gin Router  →  Handlers  →  Services  →  Repositories (GORM)  →  PostgreSQL
                   ↓            ↓
             Middleware   Orchestrator (Docker SDK)  →  Docker Engine / Swarm
                              ↓
                         Traefik writer  →  /etc/orbita/traefik (dynamic config)
```

- Entry point: [cmd/server/main.go](cmd/server/main.go) — wires all dependencies manually (no DI container).
- Router: [internal/api/router.go](internal/api/router.go) — all routes in one place, grouped by RBAC tier.
- Handlers: [internal/api/handlers/](internal/api/handlers/) — 17 handler files, thin (bind → call service → respond).
- Services: [internal/service/](internal/service/) — business logic. 11 services.
- Repositories: [internal/repository/](internal/repository/) — GORM queries. `scopes.go` holds reusable query scopes.
- Models: [internal/models/](internal/models/) — GORM structs. Migrations in [migrations/](migrations/) are the source of truth for schema, not `AutoMigrate`.
- Orchestrator: [internal/orchestrator/](internal/orchestrator/) — Docker SDK wrapper, build engine (Dockerfile + Nixpacks), blue-green deploys, DB provisioner, cgroup slicing, node manager.
- Auth: [internal/auth/](internal/auth/) — JWT (15min access + 30d refresh), bcrypt cost 12, AES-256-GCM with **per-org HKDF-SHA256 derived keys** (do not encrypt secrets with the master key directly).
- Cron: [internal/cron/](internal/cron/) — `robfig/cron/v3` scheduler + executor. Started from `main.go`, stopped on shutdown.
- WebSocket: [internal/websocket/](internal/websocket/) — terminal + log streaming. JWT comes via query param (browsers can't set headers on WS).

### RBAC

Four roles enforced at the route group level in [router.go](internal/api/router.go): `viewer` < `developer` < `admin` < `owner`. Super admin is a separate axis (`requireSuperAdmin`). When adding an endpoint, put it in the right access group — don't add ad-hoc role checks inside handlers.

### Multi-tenancy invariants

- Every tenant-scoped query **must** be scoped by `organization_id`. Use repository scopes in [internal/repository/scopes.go](internal/repository/scopes.go) — don't write raw `db.Where` that skips org scoping.
- Every org gets its own Docker network, cgroup slice, and encryption key derived via HKDF from the master key + org ID. Provisioning lives in the orchestrator layer.
- Volume namespaces, container names, and Traefik router names are all prefixed with the org slug.

## Frontend

- Source: [web/src/](web/src/) — 25 pages in [web/src/pages/](web/src/pages/), API client in `web/src/api/`, Zustand stores in `web/src/stores/`.
- Build output: `web/dist/` is embedded into the Go binary at compile time. **Always rebuild the frontend before building the Go binary** if you changed web code (`make build` does both).
- React Query for server state, Zustand for client state. React Hook Form + Zod for forms. shadcn/ui components are copied into `web/src/components/ui/`, not imported — edit in place.
- In dev, frontend runs on `:5173` and proxies to the Go backend on `:8080`. CORS is keyed on `CORS_ORIGINS` env.

## Database

- PostgreSQL 15+ with `uuid-ossp` extension.
- 22 migration files in [migrations/](migrations/), numbered `000001_`…`000022_`. Each has `.up.sql` and `.down.sql`.
- Run via `golang-migrate` at startup (see [internal/database/](internal/database/)) — no `AutoMigrate`.
- Adding a schema change: create the next numbered pair, update the matching GORM model, don't edit existing migrations (they've run in prod-like envs).

## Deployment

- [docker/docker-compose.dev.yml](docker/docker-compose.dev.yml) — local Postgres + Redis for `make dev`.
- [docker/docker-compose.prod.yml](docker/docker-compose.prod.yml) — full prod stack (Orbita + Postgres + Redis + Traefik v3). Referenced by [install.sh](install.sh).
- [Dockerfile](Dockerfile) — multi-stage (web build → go build → scratch-ish runtime).
- Docker Swarm mode is required in prod (used for service orchestration and rolling deploys).

## Conventions

- Structured logging: `zerolog` everywhere — `log.Info().Str("key", val).Msg(...)`. Don't use `fmt.Println` or stdlib `log`.
- Errors from handlers go through the shape in [internal/api/errors.go](internal/api/errors.go). Don't hand-roll JSON error responses.
- Request IDs are added by `requestid` middleware; propagate them in logs.
- Rate limiting uses Redis sliding window. Auth endpoints: 5 req / 15 min.
- Webhooks (`/webhooks/github|gitlab|gitea`) verify HMAC-SHA256 signatures — the routes are public but the handler rejects unsigned requests.
- API keys have the `orb_` prefix.

## Things not to do

- Don't bypass the service layer from a handler by calling a repo directly — even for "simple" reads. It breaks audit logging and org scoping.
- Don't add `AutoMigrate` calls. Schema changes go through numbered SQL migrations.
- Don't encrypt secrets with the master key directly — always derive the per-org key via HKDF first.
- Don't add role checks inside handlers — put the route in the right RBAC group in the router.
- Don't commit `.env`, `orbita.exe`, `server.exe`, `migrate.exe` changes as part of feature work (they're build artifacts that live at repo root on Windows dev machines).
- Don't edit files under `web/dist/` or `web/node_modules/` — they're generated.
