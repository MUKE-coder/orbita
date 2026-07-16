# Architecture

Orbita is a strictly layered Go service with an embedded SPA, orchestrating Docker Swarm and
Traefik on the host.

## The big picture

```
                              ┌──────────────────────────────────────────┐
   grit CLI  ───HTTPS/orb_──▶ │             Orbita binary (~30 MB)         │
   (laptop)                   │                                            │
   browser   ───HTTPS──────▶  │   Gin router → Handlers → Services         │
                              │        │            │                      │
                              │   Middleware   Orchestrator (Docker SDK)    │
                              │   (auth/RBAC)       │        │             │
                              │        │       Traefik writer  cgroup mgr   │
                              │   Repositories (GORM, org-scoped)           │
                              │   Embedded React SPA (//go:embed)           │
                              └──────┬───────────────┬───────────┬─────────┘
                                     │               │           │
                              ┌──────▼─────┐  ┌───────▼────┐  ┌───▼────────────┐
                              │ PostgreSQL │  │   Redis    │  │  Docker Engine  │
                              │ (metadata) │  │ (cache/RL) │  │  + Swarm        │
                              └────────────┘  └────────────┘  └───┬────────────┘
                                                                  │
                                                           ┌──────▼──────┐
                                                           │   Traefik   │
                                                           │  TLS + proxy │
                                                           └─────────────┘
```

## Layered by design

The request path never skips a layer:

```
Gin router → handlers → services → repositories → PostgreSQL
                 │            │
           middleware    orchestrator (Docker SDK) → Docker Engine / Swarm
           (auth/RBAC)        │
                         Traefik writer → dynamic routing config
```

- **Handlers** only validate input and call a service — thin, no business logic.
- **Services** hold the business logic (one per domain: auth, org, app, db, cron, …).
- **Repositories** run GORM queries, always scoped by `organization_id`.
- **Orchestrator** wraps the Docker SDK: build, deploy, provision, cgroup slicing, blue-green.
- **Traefik writer** emits per-resource dynamic config (routers + services + TLS).

## Key components

| Component | Role |
|---|---|
| **Control plane** (Go binary) | REST API `/api/v1`, embedded React 19 dashboard, cron scheduler, WebSocket hub (logs + terminal) |
| **PostgreSQL 16** | All metadata: orgs, apps, deployments, encrypted secrets, audit logs |
| **Redis 7** | Cache, rate limiting, deploy queue |
| **Docker Swarm** | Runs every workload as a service; rolling zero-downtime updates + rollback |
| **Traefik v3** | Reverse proxy + automatic Let's Encrypt TLS, driven by Orbita's dynamic config |

## Multi-tenancy invariants

::: warning These are non-negotiable
- Every tenant query is scoped by `organization_id` (GORM scopes, never optional).
- Each org gets its own Docker network, cgroup slice, and **AES-256 key HKDF-derived** from a
  master key + org ID — secrets are never encrypted with the master key directly.
- Volumes, container names, and Traefik router names are prefixed with the org slug.
:::

## The Grit-awareness layer

This is what makes zero-config deploys possible:

1. **Detect** — a repo is a Grit app iff `grit.json` exists at its root. `architecture` selects
   the deploy strategy (`single` / `double` / `triple` / `api`).
2. **Derive** — from the mode, Orbita builds the service list and the exact build recipe (the
   Dockerfiles Grit already ships, with the correct build contexts). It does **not** generate
   Dockerfiles or guess with Nixpacks.
3. **Reconcile** — provision addons (Postgres/Redis/MinIO), inject env (`DATABASE_URL`,
   `REDIS_URL`, dashboard creds), set domains — all idempotent.
4. **Migrate** — run `cmd/migrate` in a one-off container **under a Postgres advisory lock**,
   gating cutover. A migration failure aborts the deploy.
5. **Observe** — mount Pulse + Sentinel on the API by default.

## Tech stack

**Backend:** Go 1.25 · Gin · GORM · PostgreSQL 16 · Redis 7 · Docker SDK + Swarm · Traefik v3 ·
JWT (golang-jwt v5) · Resend · robfig/cron · gorilla/websocket · golang-migrate · zerolog.

**Frontend:** React 19 + TypeScript · Vite · Tailwind v4 · shadcn/ui · Zustand · TanStack Query ·
React Hook Form + Zod · xterm.js. Built to `web/dist` and embedded via `//go:embed`.

## Next

- [Quickstart](./quickstart) — deploy in two commands.
- [Install on a fresh server](./install) — the manual, step-by-step path.
