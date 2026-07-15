# 01 — What Grit is

## In one paragraph

**Grit** is a full-stack meta-framework and CLI. It scaffolds a complete, opinionated
application: a **Go backend** (Gin HTTP framework + GORM ORM), one or more **React
front-ends** (Next.js App Router, or a Vite/TanStack SPA), a **shared TypeScript package**
(Zod schemas + types), and a batteries-included platform layer (auth, file storage, email,
background jobs, cron, caching, AI, observability, security). One command — `grit new` —
produces all of it, wired together, in a pnpm + Turborepo monorepo (or a flat single-binary
project). Because every Grit app has the **same known shape**, tooling can reason about it —
which is the entire premise of Grit Cloud.

**Tagline:** *Go + React. Built with Grit.*

## Why this matters for Orbita

Dokploy/Coolify treat every app as an opaque container, so the user hand-configures the
Dockerfile, ports, build method, env vars, domains, and migrations. **A Grit app is not
opaque.** It declares its shape (`grit.json`), ships its own build recipes (Dockerfiles +
compose), has a fixed env contract, and exposes a known health/migration/observability
surface. Orbita's job is to *read* that shape and wire it up — not to guess.

## The tech stack (fixed — do not assume anything else)

| Layer | Technology | Notes |
|-------|-----------|-------|
| Backend | **Go 1.24** (module per app) | Gin router, GORM ORM |
| DB (prod) | **PostgreSQL 16** | SQLite is used only for quick-start/tests |
| Cache / queue | **Redis 7** | `asynq` for background jobs + cron |
| Frontend | **Next.js 16 (App Router)** or **Vite + TanStack Router** | chosen at scaffold time (`grit.json.frontend`) |
| Admin | Next.js admin panel (Filament-style) | present in `triple` / `+admin` |
| Object storage | **S3-compatible**: MinIO (local), Cloudflare R2, AWS S3, Backblaze B2 | driver-selected via `STORAGE_DRIVER` |
| Email | **Resend** | transactional |
| Package mgr | **pnpm 9** (workspace) + **Turborepo** | monorepo modes only |
| DB browser | **GORM Studio** | embedded, mounted on the API |
| Observability | **Pulse** | embedded, mounted on the API |
| Security | **Sentinel** (WAF/rate-limit/anomaly) | embedded, mounted on the API |

## The Grit CLI (the binary Grit Cloud extends)

Grit Cloud's `grit cloud` / `grit deploy` are **new subcommands of this existing CLI**
(Cobra). Commands that already exist and that you will lean on:

- `grit new <name> [--single|--double|--triple|--api|--mobile] [--desktop] [--next|--vite]`
  — scaffold a project. `--full` = `--triple` + docs.
- `grit generate resource <Name> --fields "..."` — generate a full-stack CRUD feature.
- `grit migrate` — apply the database schema (see `06-migrations.md`).
- `grit seed` — seed data.
- `grit start [server|web|admin|expo|desktop]` — run pieces locally.
- `grit backup` / `grit restore` — full-DB snapshot / restore (relevant to disaster recovery).
- `grit deploy` — **already exists** as a VPS deploy helper (SSH + docker compose). Grit
  Cloud replaces/augments its target with Orbita's HTTPS API. Read the existing command
  before you extend it; don't collide with it.

## The Grit ecosystem (how the pieces relate at deploy time)

| Piece | What it is | Role at deploy |
|-------|-----------|----------------|
| **Grit framework** | the app being deployed | its known shape makes zero-config deploys possible |
| **Orbita** | the PaaS control plane on the VPS | builds, routes, migrates, runs the app |
| **`grit migrate`** | schema migrator (GORM AutoMigrate + hooks) | runs on deploy, under an advisory lock |
| **Pulse** | request/latency/SQL observability, mounted on the API | reachable at `/pulse/ui` |
| **Sentinel** | WAF + rate limiting + anomaly detection, mounted on the API | reachable at `/sentinel/ui` |
| **GORM Studio** | visual DB browser, mounted on the API | reachable at `/studio` |
| **DGateway** | billing gateway | future paid/managed tier only |

**Next:** [`02-architecture-modes.md`](./02-architecture-modes.md) — the deployable shapes.
