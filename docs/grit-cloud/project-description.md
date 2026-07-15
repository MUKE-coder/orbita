# Grit Cloud — Project Description

> **One line:** Turn a single VPS into a secure, Grit-aware deployment platform, and let a developer ship a full Grit app (Go API + Next.js + Postgres/Redis/MinIO + migrations) with **one command** from the terminal.

This document is the north star. It explains *what we are building, why, how it should feel, and what "done" looks like.* Read this before `project-phases.md` and `prompt.md`.

---

## 1. The dream

A developer buys a fresh VPS (Hetzner, Contabo, DigitalOcean). Two commands later they have a hardened, production-grade PaaS running, and their Grit app is live on HTTPS with a database, a domain, and migrations applied:

```bash
# On the fresh server (or from laptop, targeting the server)
grit cloud init                 # harden + install Orbita + register the host

# From the project directory on the laptop
grit deploy --host prod         # build recipe, push to GitHub, deploy via Orbita, migrate, live
```

No Dockerfile hand-writing. No clicking through a dashboard. No wrestling with Traefik labels or Nixpacks guessing. Because a Grit app has a **known shape**, the tooling already understands it and does the wiring for the developer.

After the first deploy, every `git push` auto-deploys via webhook — the CLI becomes optional.

---

## 2. Why this exists

The self-hosting PaaS space (Dokploy, Coolify) has two problems for our users:

1. **No multi-tenancy.** Freelancers and agencies (our core market — East Africa and beyond) run many clients on one cheap VPS. Dokploy/Coolify treat everything as one flat space. **Orbita already solves this** with true tenant isolation (separate Docker networks, per-org encryption keys, cgroup quotas, 4-role RBAC).

2. **Generic, opaque deploys.** They treat every app as a mystery container, so the user still hand-configures Dockerfiles, ports, build methods, env vars, domains, and migrations through a UI. **Grit Cloud closes this gap:** because the app is Grit, the CLI generates the build, wires the services, provisions the addons, and runs migrations — automatically.

**Grit Cloud is the delivery vehicle that makes the whole Grit ecosystem cohere:**

| Ecosystem piece | Role in Grit Cloud |
| --- | --- |
| Grit Framework | The app being deployed; its known shape is what makes zero-config deploys possible |
| Orbita | The control plane / PaaS that runs on the VPS and executes deploys |
| `vps-harden.sh` | The hardening step inside `grit cloud init` |
| Migration tool | Runs on deploy, under advisory lock |
| Pulse | Built-in observability panel, mounted by default |
| Sentinel | Built-in security layer, mounted by default |
| DGateway | Powers billing for a future paid/managed tier |

We are **not building a PaaS from scratch.** Orbita already exists and is ~80% of a working PaaS. We are (a) finishing and hardening Orbita, (b) teaching Orbita to understand Grit apps natively, and (c) building the `grit cloud` / `grit deploy` CLI that sits on top.

---

## 3. What already exists (Orbita today)

Orbita is a real, running self-hosted PaaS. Do **not** rebuild these — audit, finish, and test them:

- **Multi-tenancy:** organizations with isolated Docker networks, per-org HKDF-derived AES-256 keys, cgroup v2 CPU/RAM quotas.
- **RBAC:** Owner / Admin / Developer / Viewer, enforced at the API layer, with signed email invites (72h expiry, via Resend).
- **Deploys:** from Docker image *or* Git repo (GitHub/GitLab/Gitea) with webhook auto-deploy on push; build via Dockerfile or Nixpacks; zero-downtime rolling updates on Docker Swarm; versioned deploy history with one-click rollback.
- **Databases:** one-click Postgres/MySQL/MariaDB/Mongo/Redis with encrypted credentials, scheduled backups, restore.
- **Cron jobs:** scheduled containers with concurrency policies and run history.
- **Domains & TLS:** custom domains, automatic Let's Encrypt via Traefik ACME, HTTP→HTTPS redirect.
- **Marketplace:** 10 one-click templates (WordPress, Plausible, n8n, Grafana, MinIO, etc.).
- **Monitoring:** live logs (WebSocket), CPU/mem/network metrics, in-browser xterm.js terminal, exec API.
- **Security:** JWT (15m access / 30d refresh, httpOnly), bcrypt cost 12, AES-256-GCM secrets at rest, Redis rate limiting, org-scoped queries, HMAC webhook signatures, `orb_` API keys for CI/CD.
- **Shipping:** single ~30MB Go binary with embedded React SPA, <50MB idle RAM, `install.sh` one-liner.
- **Stack:** Go 1.22 + Gin + GORM + Postgres 15 + Redis 7 + Traefik v3; React 18 + Vite + Tailwind v4 + shadcn/ui + Zustand + TanStack Query + xterm.js.

**Existing API surface we build on:** `/api/v1` with endpoints for orgs, projects, environments, apps (`/deploy`, `/rollback`, `/env`, `/domains`, `/logs`, `/metrics`), databases, cron, git-connections, and `orb_` API keys.

---

## 4. What we are adding (the Grit Cloud layer)

Three things, in priority order:

### 4a. Finish & harden Orbita
Audit the existing repo for incomplete or untested paths (deploy engine edge cases, backup/restore correctness, webhook signature verification, RBAC enforcement gaps, TLS issuance failures) and get a **fully tested green build**. See phases for the checklist.

### 4b. Teach Orbita to understand Grit ("Grit-awareness")
A Grit app is not an opaque container. Orbita gains a **Grit app type** that knows:
- The shape: a Go API service (Gin+GORM), an optional Next.js `web`, optional workers, and addons (postgres/redis/minio).
- How to build it: generate a correct multi-stage Dockerfile (Go build + Next.js build) instead of guessing with Nixpacks.
- How to route it: map `api.*`, `admin.*`, root domain per the `grit.yaml` contract.
- How to migrate it: run the Grit migration tool on deploy, under a Postgres advisory lock, before cutover.
- How to observe/secure it: mount Pulse and Sentinel dashboards by default.

### 4c. The `grit cloud` / `grit deploy` CLI (subcommands of the existing Grit CLI)
- `grit cloud init` — harden the server (vendor `vps-harden.sh`), install Docker + Traefik + Orbita, bootstrap an `orb_` deploy token, register the host in `~/.grit/hosts.yaml`.
- `grit deploy --host <name>` — the magic command (flow in §6).
- `grit logs -f`, `grit cloud status`, `grit rollback`, `grit secrets` — supporting commands.

---

## 5. The `grit.yaml` contract

The single source of truth in the user's project root. The CLI and Orbita both read it. Example:

```yaml
app: rental-manager
repo: MUKE-coder/rental-manager     # GitHub owner/repo; created if missing
branch: main
services:
  api:    { type: go,     path: ./cmd/api, port: 8080 }
  web:    { type: nextjs, path: ./web }
  worker: { type: go,     path: ./cmd/worker }
addons:
  - postgres
  - redis
  - minio
domains:
  web:   hmkestates.com
  api:   api.hmkestates.com
  admin: admin.hmkestates.com
migrate: true          # run the migration tool on deploy
env:
  from: .env.production # local file, values encrypted into Orbita, never committed
```

Design principle: **the user writes this once; everything else is inferred.** If a field is omitted, sane Grit defaults apply (api on :8080, web on :3000, production environment, migrate: true).

---

## 6. The deploy flow (the experience that beats Dokploy)

Decisions locked for v1: **transport = Orbita HTTPS API + `orb_` token** (SSH tunnel kept only for private dashboard viewing); **build source = GitHub**, deployed from `grit.yaml` like Dokploy but fully driven from the terminal.

`grit deploy --host prod` does, in order:

1. **Read `grit.yaml`** and resolve `prod` → Orbita API URL + `orb_` token from `~/.grit/hosts.yaml`.
2. **Ensure the GitHub repo exists.** Using the user's stored GitHub token (added once via `grit cloud github-auth`), if the repo in `grit.yaml` doesn't exist, create it with `gh`/the GitHub API and push the current project. If it exists, push the current branch.
3. **Generate the build recipe.** Because the app is Grit, write a correct multi-stage Dockerfile (or Orbita build manifest) for the Go API + Next.js web, and commit it. No Nixpacks guessing, no user-authored Dockerfile.
4. **Reconcile with Orbita via API:** ensure org/project/environment exist → create-or-update the Grit app (source = the GitHub repo + branch) → inject env/secrets (encrypted) → set domains → provision missing addons (postgres/redis/minio containers in the org's isolated network).
5. **Deploy + migrate.** Trigger the deploy; Orbita builds from the repo, then runs the migration tool **under a Postgres advisory lock** before health-check and Traefik cutover. Keep the previous image for rollback.
6. **Stream logs** back to the terminal (`grit logs -f --host prod`), with Pulse + Sentinel now live on their dashboards.
7. **Hand off to webhooks.** Orbita registers a GitHub webhook, so every future `git push` auto-deploys — the CLI is now optional.

**Better-than-Dokploy touches to build in:**
- **First-run wizard:** `grit deploy` with no `grit.yaml` present generates one interactively from the detected project shape, then proceeds.
- **Dry run:** `grit deploy --plan` prints exactly what will be created/changed in Orbita (repo, app, addons, domains, migrations) without executing.
- **Preview URL immediately:** print the live URL and the Pulse/Sentinel dashboard links the moment cutover succeeds.
- **One `env` source of truth:** values come from a local `.env.production` referenced in `grit.yaml`, encrypted into Orbita — the user never pastes secrets into a UI.

---

## 7. Value proposition (per audience)

- **Freelancers / agencies (core):** run all client apps on one €4.5/mo Hetzner box with true isolation; deploy any client's Grit app in one command; bill clients later via DGateway.
- **Solo Grit developers:** the fastest path from `git init` to a live, HTTPS, migrated, observable app — faster than Vercel+Railway and self-hosted so it's nearly free.
- **The Grit ecosystem:** Grit Cloud is the reason to adopt Grit — framework, migrations, Sentinel, Pulse and DGateway all pay off at deploy time.
- **Teaching / YouTube funnel:** "secure a VPS and deploy a full-stack app in 2 commands" is a killer demo and drives Grit adoption.

---

## 8. Non-goals for v1 (scope discipline — the PaaS graveyard is real)

Explicitly **out** of the first shippable version:
- Kubernetes or any orchestrator beyond Docker Swarm (Orbita already uses Swarm — keep it).
- Multi-server / multi-node deploys from the CLI (Orbita supports nodes; the CLI targets one host in v1).
- A managed/hosted tier where we run the compute (self-hosted only first).
- Autoscaling, preview environments per-PR, blue/green beyond Orbita's existing rolling+rollback.
- Non-GitHub git providers in the CLI (Orbita supports GitLab/Gitea in its UI; the CLI is GitHub-only in v1).
- Rewriting anything Orbita already does well (Traefik stays; do **not** swap to Caddy).

---

## 9. Definition of done (v1)

On a fresh Ubuntu 24.04 Hetzner VPS:

1. `grit cloud init` produces a hardened server (security score ≥ 90) with Orbita running on HTTPS and a registered host in `~/.grit/hosts.yaml`.
2. In a sample Grit repo (Go API + Next.js + Postgres), `grit deploy --host prod` takes it from local code to a live HTTPS URL with the database provisioned and migrations applied — **without the user writing a Dockerfile or touching the Orbita UI.**
3. `git push` on that repo triggers an auto-deploy.
4. Pulse and Sentinel dashboards are reachable for the deployed app.
5. `grit rollback --host prod` reverts to the previous deploy.
6. The whole path is documented and reproducible from the two commands.
