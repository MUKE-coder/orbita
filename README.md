<p align="center">
  <img src="images/orbita_banner.png" alt="Orbita" width="100%" />
</p>

<p align="center">
  <img src="images/orbita_logo.png" alt="Orbita" width="110" height="110" />
</p>

<h1 align="center">Orbita</h1>

<p align="center"><strong>Self-hosted, multi-tenant PaaS — and the control plane for Grit Cloud.</strong></p>
<p align="center">One VPS. Many clients. Full isolation. Deploy a full Grit app in two commands.</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go" alt="Go" />
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react" alt="React" />
  <img src="https://img.shields.io/badge/Docker-Swarm-2496ED?style=flat-square&logo=docker" alt="Docker" />
  <img src="https://img.shields.io/badge/Traefik-v3-24A1C1?style=flat-square&logo=traefikproxy" alt="Traefik" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="MIT" />
</p>

<p align="center">
  <a href="#what-is-orbita">What</a> ·
  <a href="#why-orbita-exists">Why</a> ·
  <a href="#architecture">Architecture</a> ·
  <a href="#install-on-a-fresh-server">Install</a> ·
  <a href="#deploy-a-grit-app">Deploy a Grit app</a> ·
  <a href="#troubleshooting--common-errors">Troubleshooting</a>
</p>

---

## What is Orbita?

**Orbita is an open-source, self-hosted Platform-as-a-Service (PaaS)** written in Go. It turns a
single VPS into a fully isolated hosting environment for many client organizations — each with
their own dashboard, projects, environment variables, secrets, logs, domains, databases, and
resource quotas. Clients never see each other; you, the super-admin, see everything.

It ships as a **single ~30 MB binary** with the React dashboard embedded, and idles under
**50 MB of RAM** — leaving nearly all of your server for the apps you actually run.

Orbita is also the **control plane for Grit Cloud**: because a [Grit](https://github.com/MUKE-coder)
app has a *known shape*, Orbita can build, route, migrate, and observe it with zero
hand-configuration. Two commands take you from a bare server to a live, migrated, HTTPS app:

```bash
grit cloud init  --server root@<ip> --domain orbita.example.com --acme-email you@example.com --admin-email admin@example.com
grit deploy      --host prod        # build → migrate (advisory lock) → route → live
```

After the first deploy, every `git push` auto-deploys via webhook.

---

## Why Orbita exists

Freelancers and agencies — Orbita's core audience — run many clients on one cheap box. The
existing self-hosted PaaS tools don't fit that, and managed platforms are expensive and opaque:

| Problem | How Orbita solves it |
|---|---|
| **Dokploy / Coolify have no multi-tenancy** | True tenant isolation: per-org Docker networks, per-org encryption keys, cgroup v2 CPU/RAM quotas, 4-role RBAC |
| **No resource caps per client** | cgroup v2 slices enforce CPU/memory limits per organization |
| **Heavy runtimes** (PHP/Node control planes) | A single Go binary, <50 MB idle |
| **Generic, opaque deploys** — you still hand-write Dockerfiles, ports, env, migrations | **Grit-awareness**: Orbita reads the app's shape and wires the build, addons, domains, and migrations itself |
| **Vendor lock-in / per-seat fees** | 100% self-hosted, you own the data, no per-seat pricing |

### How it compares

| | Orbita | Dokploy | Coolify |
|---|:---:|:---:|:---:|
| Multi-tenancy | ✅ | ❌ | ❌ |
| Resource quotas (cgroups) | ✅ | ❌ | ❌ |
| RBAC (4 roles) + invites | ✅ | ❌ | Partial |
| Cron manager | ✅ | ❌ | ❌ |
| Grit-aware zero-config deploys | ✅ | ❌ | ❌ |
| Single-binary control plane | ✅ | ❌ | ❌ |
| Idle memory | **<50 MB** | ~200 MB | ~500 MB |
| Written in | Go | Node.js | PHP/Laravel |

---

## Architecture

Orbita is a strictly layered Go service with an embedded SPA, orchestrating Docker Swarm and
Traefik on the host:

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

**Layers (never skipped):** `Gin router → handlers → services → repositories → PostgreSQL`.
Handlers only validate and call services; business logic lives in services; repositories run
org-scoped GORM queries. The **orchestrator** wraps the Docker SDK (build, deploy, provision,
cgroup slicing); the **Traefik writer** emits dynamic routing config.

**Key components**

- **Control plane** — the Go binary: REST API (`/api/v1`), embedded React 19 dashboard, cron
  scheduler, WebSocket hub (logs + terminal).
- **PostgreSQL** — all metadata: orgs, apps, deployments, encrypted secrets, audit logs.
- **Redis** — cache, rate limiting, deploy queue.
- **Docker Swarm** — runs every workload as a service; rolling zero-downtime updates + rollback.
- **Traefik v3** — reverse proxy + automatic Let's Encrypt TLS, driven by Orbita's dynamic config.

**Multi-tenancy invariants:** every tenant query is scoped by `organization_id`; each org gets
its own Docker network, cgroup slice, and AES-256 key (HKDF-derived from a master key + org ID);
volumes, container names, and Traefik routers are prefixed with the org slug.

**Grit-awareness layer:** `internal/grit` detects a Grit app from `grit.json`, derives its
services and build recipe (reusing the Dockerfiles Grit ships), and the `GritService` reconciles
addons + env + domains and runs migrations under a Postgres advisory lock before cutover.

**Tech stack:** Go 1.25 · Gin · GORM · PostgreSQL 16 · Redis 7 · Docker SDK + Swarm · Traefik v3 ·
JWT · Resend · React 19 + Vite + Tailwind v4 + shadcn/ui · xterm.js.

---

## Install on a fresh server

The fast path is `grit cloud init` (below). This section is the **manual, step-by-step install**
for a fresh **Ubuntu 24.04** VPS, so you understand exactly what happens.

### 0. Prerequisites

- A fresh VPS (Hetzner CX22 — 2 vCPU / 4 GB / €4.5/mo — is plenty). Ports **80, 443, 8080** open.
- A domain. Point an `A` record and a wildcard at the server IP **and wait for DNS to resolve**:

  | Type | Name | Value |
  |---|---|---|
  | A | `orbita.example.com` | `YOUR_SERVER_IP` |
  | A | `*.example.com` (or per-app records) | `YOUR_SERVER_IP` |

> **Verify DNS before installing.** `dig orbita.example.com +short` must return your server IP.
> Let's Encrypt cannot issue a certificate until it does.

### 1. Connect and update

```bash
ssh root@YOUR_SERVER_IP
apt update && apt upgrade -y
apt install -y curl git ufw
```

### 2. Firewall

```bash
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
ufw status
```

### 3. Install Docker + Swarm

```bash
curl -fsSL https://get.docker.com | sh
docker --version          # 24+ expected
docker swarm init         # required — Orbita runs workloads as Swarm services
```

### 4. Install Orbita (one line)

```bash
curl -sSL https://raw.githubusercontent.com/MUKE-coder/orbita/main/install.sh \
  | sudo ORBITA_DOMAIN=orbita.example.com ORBITA_ACME_EMAIL=you@example.com bash -s -- --yes
```

The installer: installs Docker if missing, generates secrets, writes `docker-compose.yml` + `.env`,
starts **Orbita + PostgreSQL + Redis + Traefik**, and health-checks. Traefik requests a Let's
Encrypt certificate on first request.

### 5. Verify

```bash
curl -s http://localhost:8080/health          # {"status":"ok","version":"0.1.0"}
docker compose -f /opt/orbita/docker-compose.yml ps   # all 4 services Up
curl -I https://orbita.example.com            # 200 from the outside world
```

Open `https://orbita.example.com`, register the first user (**they become super-admin**), and
create your first organization.

> **Do it all at once with the CLI.** `grit cloud init` runs the hardening, this install, the
> super-admin + token bootstrap, and host registration in a single command — see
> [Deploy a Grit app](#deploy-a-grit-app).

---

## Deploy a Grit app

From bare server to a live, migrated, observable Grit app in **two commands**.

### Command 1 — provision the host

On your **laptop** (with the `grit` CLI and SSH access to the server):

```bash
grit cloud init \
  --server root@YOUR_SERVER_IP \
  --name prod \
  --domain orbita.example.com \
  --acme-email you@example.com \
  --admin-email admin@example.com
```

This hardens the server (security score printed), installs Docker + Orbita + Traefik, creates your
super-admin and an `orb_` deploy token, and registers the host in `~/.grit/hosts.yaml`. Verify:

```bash
grit cloud status --host prod          # ● ok
```

### Command 2 — deploy

From your Grit project directory (it contains a `grit.json`):

```bash
grit cloud github-auth                 # once: store a GitHub token (repo + admin:repo_hook)
grit deploy --host prod
```

If there's no `grit.yaml`, a first-run wizard creates one from your app's detected shape. Then
`grit deploy`:

1. ensures the GitHub repo exists and pushes your code,
2. reconciles org/project/environment, provisions **Postgres/Redis/MinIO** addons, injects env +
   secrets, sets domains,
3. builds each service from the Dockerfiles Grit already ships,
4. runs migrations **under a Postgres advisory lock** (aborting the deploy if they fail),
5. cuts over and prints your links:

```
✔ Live
  App:       https://rental.example.com
  API:       https://api.rental.example.com
  Pulse:     https://api.rental.example.com/pulse/ui
  Sentinel:  https://api.rental.example.com/sentinel/ui
  Auto-deploy is on: future `git push` to main redeploys via webhook.
```

### `grit.yaml` — the deploy manifest

You write this once; everything else is inferred from `grit.json` + the repo.

```yaml
app: rental                       # app name
repo: MUKE-coder/rental           # GitHub owner/name (created + pushed if missing)
branch: main
addons: [postgres, redis]         # provisioned in the org's isolated network
domains:
  web: rental.example.com
  api: api.rental.example.com
migrate: true                     # run cmd/migrate under an advisory lock before cutover
env:
  from: .env.production           # values encrypted into Orbita, never committed
```

Supporting commands:

```bash
grit deploy --plan --host prod    # dry run: show what would change, don't apply
grit logs -f --host prod          # stream logs
grit rollback --host prod         # revert to the previous deploy
grit cloud dashboard --host prod  # private SSH tunnel to the Orbita panel
```

### Deploy a non-Grit app

Orbita also deploys any Docker image or Git repo through the dashboard: **Apps → Create App →**
choose **Docker Image** (e.g. `nginx:alpine`) or **Git Repository** (build via Dockerfile or
Nixpacks, auto-deploy on push).

---

## Troubleshooting / common errors

Ordered most-common first. Most installer issues are handled automatically; this is for the rest.

### Install fails: `denied` pulling `ghcr.io/muke-coder/orbita`

The image is private or unpublished for your fork. Build locally and point the installer at it:

```bash
cd /opt && git clone https://github.com/MUKE-coder/orbita.git orbita-src
cd orbita-src && docker build -t orbita:local .
ORBITA_IMAGE=orbita:local bash <(curl -sSL https://raw.githubusercontent.com/MUKE-coder/orbita/main/install.sh)
```

### `address already in use` on port 80 / 443

Something else (usually Apache2 on Contabo/DO images, sometimes nginx) holds the port:

```bash
ss -ltnp 'sport = :80'
systemctl stop apache2 && systemctl disable apache2
apt purge -y apache2 apache2-utils apache2-bin apache2-data && apt autoremove -y
cd /opt/orbita && docker compose up -d
```

### SSL certificate never issues (padlock stays broken)

```bash
cd /opt/orbita && docker compose logs traefik --tail 50 | grep -iE "acme|certificate|error"
```

| Log symptom | Cause → Fix |
|---|---|
| `DNS problem` / `NXDOMAIN` | DNS not propagated. `dig orbita.example.com +short` must return the server IP. Wait, retry. |
| `HTTP 404` from challenge | **Cloudflare orange cloud** proxying the record. Set it to **DNS only** (gray) until the cert issues. |
| `connection refused` on :80 | Firewall. `ufw allow 80/tcp && ufw allow 443/tcp && ufw reload` |
| `rate limit exceeded` | Let's Encrypt limit (5 certs/week/domain). Wait an hour; avoid delete/re-add loops. |
| nothing ACME-related | Traefik isn't reading labels. Confirm `--providers.docker=true` in the compose `traefik` command. |

Orbita's own check surfaces the exact cause: `GET /api/v1/orgs/:org/domains/verify?domain=...`
returns resolved IPs, Cloudflare-proxy detection, and copy-ready guidance.

### `ERR_TOO_MANY_REDIRECTS` on the dashboard

Browser cache (try Incognito), or an old image before the SPA-serving fix:
`cd /opt/orbita && docker compose pull orbita && docker compose up -d orbita`.

### Traefik: `client version 1.24 is too old`

Docker Engine 28+ with an old Traefik. Fixed in the current template (`traefik:v3.6.14`):
`sed -i 's|image: traefik:v3.0|image: traefik:v3.6.14|' docker-compose.yml && docker compose up -d --force-recreate traefik`.

### Containers restart in a loop

```bash
docker compose ps
docker compose logs <service> --tail 100
```

- **orbita** → DB connection failed (check `DATABASE_URL`) or a migration error.
- **postgres** → old volume with mismatched credentials. `docker compose down -v` wipes data, then reinstall.
- **traefik** → config error; re-generate the compose via the installer.

### Grit deploy: app crash-loops with `password authentication failed`

A managed database was deleted and re-created with the same name, reusing a stale volume (Postgres
ignores `POSTGRES_PASSWORD` on a non-empty data dir). Fixed in current Orbita — `RemoveDatabase`
now removes the volume and provisioning clears orphaned volumes. If you hit an old orphan:
`docker volume rm <orgslug>_<dbname>_data` after removing the DB service.

### Grit deploy: migrations failed, deploy aborted

By design — Orbita will not cut over to a schema-mismatched image. Run `grit logs -f --host prod`
to see the migrator output, fix the migration, and redeploy. Common cause: the app's `go.sum`
isn't committed, so `go run ./cmd/migrate` can't resolve modules.

### General diagnostics

```bash
cd /opt/orbita
docker compose ps                       # all 4 healthy?
docker compose logs orbita --tail 50    # backend
docker compose logs traefik --tail 50   # routing + TLS
ss -ltnp 'sport = :443'                 # Traefik bound?
curl -I http://localhost:8080/health    # backend up?
curl -I https://orbita.example.com      # reachable from outside?
```

---

## Documentation

- **[Quickstart](docs/QUICKSTART.md)** — fresh VPS → live app in 2 commands
- **[grit.yaml spec](docs/GRIT-YAML.md)** · **[grit cloud CLI](docs/GRIT-CLOUD-CLI.md)** · **[Grit Cloud API](docs/GRIT-API.md)**
- **[TLS troubleshooting](docs/TLS-TROUBLESHOOTING.md)** · **[Demo walkthrough](docs/DEMO.md)**
- **Website + full docs:** the [`website/`](website/) directory (VitePress — landing page + docs).

---

## Development

```bash
git clone https://github.com/MUKE-coder/orbita.git && cd orbita
make docker-up                 # Postgres + Redis for local dev
cp .env.example .env           # set JWT_SECRET, JWT_REFRESH_SECRET, ENCRYPTION_MASTER_KEY
make dev                       # backend with Air hot-reload
cd web && npm install && npm run dev   # dashboard on :5173

make build       # build web + embed + Go binary (~30 MB)
make build-cli   # build the grit CLI
make test        # go test ./... -race
```

| Path | What |
|---|---|
| `cmd/server` | control-plane entry point |
| `cmd/grit` | the `grit` CLI (cloud + deploy) |
| `internal/{api,service,repository,orchestrator}` | layered core |
| `internal/grit` | Grit-awareness (detect / build recipe / plan) |
| `internal/docker`, `internal/traefik` | Docker SDK + dynamic routing |
| `migrations/` | numbered SQL migrations (source of truth for schema) |
| `web/` | React 19 dashboard (embedded at build time) |
| `website/` | landing page + docs (VitePress) |

---

## License

MIT — see [LICENSE](LICENSE).

<p align="center"><sub>Orbita — self-hosted PaaS with true multi-tenancy, and the control plane for Grit Cloud.</sub></p>
