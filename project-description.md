# Orbita — Self-Hosted PaaS with True Multi-Tenancy

## What is Orbita?

Orbita is an open-source, self-hosted Platform-as-a-Service (PaaS) built in Go, designed for developers and freelancers who manage infrastructure for multiple clients on a single server. It is a spiritual successor to tools like Dokploy and Coolify, but built from the ground up to solve the three problems they have not: **true multi-tenancy with client isolation**, **resource-capped node slicing on a single VPS**, and **a reliable team invite and RBAC system**.

Orbita lets you take a single 48GB VPS and turn it into a fully isolated hosting environment for up to N client organizations — each with their own dashboard login, their own projects, their own environment variables, their own logs, their own domains, and their own resource budget. Clients never see each other. You, as the super-admin, see everything.

Built on Go (Gin, GORM), PostgreSQL, Redis, Docker SDK, and Traefik, Orbita ships as a single binary with an embedded React SPA frontend. The entire control plane uses under 50MB of RAM at idle, leaving nearly all server memory for workloads.

---

## Positioning

- **Target user**: Freelancers, agencies, small hosting providers, indie hackers who self-host for multiple clients
- **Core value prop**: One VPS, many clients, full isolation — at a fraction of the complexity of Kubernetes
- **Against Dokploy**: Adds multi-tenancy, resource quotas, working invite system, cron job manager
- **Against Coolify**: Docker-native (no PHP/Laravel overhead), lower memory footprint, proper tenant network isolation
- **Against managed PaaS (Heroku, Render)**: 100% self-hosted, no per-seat or per-service fees, you own the data

---

## Tech Stack

### Backend
- **Language**: Go 1.22+
- **HTTP Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL 15+
- **Cache / Queue**: Redis 7+
- **Authentication**: JWT (access + refresh tokens)
- **Email**: Resend API
- **Container Runtime**: Docker SDK (docker/client)
- **Orchestration**: Docker Swarm (multi-node)
- **Reverse Proxy**: Traefik v3 (via dynamic HTTP config API)
- **SSH**: golang.org/x/crypto/ssh (for remote node execution)
- **Cron**: robfig/cron v3
- **WebSockets**: gorilla/websocket (real-time logs and events)
- **Migrations**: golang-migrate/migrate

### Frontend
- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS v3
- **State**: Zustand
- **Data Fetching**: TanStack Query (React Query)
- **UI Components**: shadcn/ui (Radix primitives)
- **Icons**: Lucide React
- **WebSocket**: native browser WebSocket API
- **Terminal**: xterm.js (in-browser SSH terminal)
- **Charts**: Recharts (metrics dashboards)
- **Forms**: React Hook Form + Zod

### DevOps
- **Embed**: `//go:embed all:web/dist` (React SPA served from Go binary)
- **Config**: environment variables + .env file (godotenv)
- **Logging**: zerolog (structured JSON logs)
- **Secrets**: encrypted at rest in PostgreSQL (AES-256)

---

## Core Concepts

### Hierarchy
```
Super Admin (platform owner)
  └── Organizations (tenants / clients)
        └── Teams (groups within an org)
              └── Members (users with roles)
                    └── Projects
                          └── Environments (production, staging, preview)
                                └── Resources (apps, databases, services, cron jobs)
```

### Resource Types
1. **Application** — deployed from Git repo, Docker image, or Docker Compose file
2. **Database** — managed PostgreSQL, MySQL, MongoDB, Redis, MariaDB instances
3. **Service** — pre-configured one-click templates (WordPress, Plausible, n8n, etc.)
4. **Cron Job** — scheduled tasks running as Docker containers on a cron schedule
5. **Volume** — persistent storage attached to resources

### Node Types
1. **Primary Node** — the VPS where Orbita itself runs (always exists)
2. **Worker Node** — additional VPS/dedicated servers added to the Swarm cluster
3. **Virtual Slice** — a resource-capped partition of a node assigned to a tenant (cgroup-based)

---

## Feature Specification

### 1. Authentication & User Management

#### Login / Registration
- Email + password login with bcrypt hashing
- JWT access token (15 min TTL) + refresh token (30 day TTL, stored in httpOnly cookie)
- "Remember me" option extending refresh token to 90 days
- Password reset via Resend email (6-digit OTP, 10 min TTL)
- Email verification on signup (Resend)
- Two-factor authentication (TOTP via authenticator app) — optional per user

#### User Profile
- Change name, email, avatar (uploaded or gravatar)
- Change password (requires current password)
- View active sessions, revoke individual sessions
- Generate personal API keys (for CLI or CI/CD integration)

#### Super Admin Capabilities
- See all organizations and their resource usage
- Impersonate any organization (for support)
- Set platform-wide resource limits
- Manage global Docker registry credentials
- View platform-wide audit logs
- Manage Traefik configuration
- Manage Swarm nodes
- Configure SMTP/Resend settings
- Configure OAuth providers (GitHub, GitLab)
- Platform health dashboard

---

### 2. Multi-Tenancy & Organizations

#### Organization (Tenant)
- Each organization is a fully isolated tenant
- Has its own: projects, environments, resources, domains, env vars, secrets, logs, members, billing data
- Super admin creates organizations and assigns an owner
- Organization has a unique slug (used in URLs and Docker network names)
- Organization-level resource quota (max CPU cores, max RAM, max disk, max apps, max databases)
- Organization has a resource plan (Free, Starter, Pro, Enterprise — super admin defines plans)

#### Isolation Mechanisms
- **Database isolation**: Every DB query includes `organization_id` scoping. GORM scopes enforce this automatically.
- **Network isolation**: Each organization gets a dedicated Docker overlay network (`orbita-org-{slug}`). Apps within an org can talk to each other by service name. Apps in different orgs cannot.
- **Secret isolation**: Environment variables and secrets are encrypted per-organization using a per-org AES-256 key derived from a master key.
- **Volume isolation**: Named volumes are prefixed with org slug (`{slug}_{volume_name}`).
- **Log isolation**: Deployment and runtime logs are partitioned by org in the database.
- **Domain isolation**: A domain can only be assigned to resources within one org.

#### Members & Roles (RBAC)
Roles within an organization:
- **Owner**: Full control of the org, can delete the org, manage billing, transfer ownership
- **Admin**: Manage all projects, members, deployments, domains, secrets
- **Developer**: Create and deploy apps, view logs, manage own app secrets, cannot manage members or billing
- **Viewer**: Read-only access to projects, logs, metrics (cannot deploy or change config)

Roles are enforced at the API level via Gin middleware that checks `organization_id` + `role` from the JWT claims.

#### Team Invite System
- Owner/Admin clicks "Invite Member" → enters email + selects role
- System generates a cryptographically secure token (32 bytes, base64url)
- Token stored in `organization_invites` table with: token hash, email, org_id, role, expires_at (72h), used_at
- Resend sends branded invite email with `Accept Invitation` button linking to `/join?token=<token>`
- Recipient lands on accept page: if no account, prompted to create one; if account exists, immediately joins org
- Token is invalidated after use or expiry
- Pending invites visible in Members settings, can be revoked
- Re-invite if expired

---

### 3. Projects & Environments

#### Projects
- A project is a logical grouping of resources under an organization
- Has a name, description, and icon/emoji
- Projects have one or more environments

#### Environments
- Default environments: **Production**, **Staging**
- Custom environments can be added (e.g., Preview, QA, Dev)
- Each environment is isolated: separate containers, separate env vars, separate domains
- Environment-level env vars that override project-level vars
- One-click clone environment (copies all resource configs, not data)

---

### 4. Application Deployment

#### Supported Source Types
1. **Git Repository** (GitHub, GitLab, Bitbucket, Gitea, or any public git URL)
   - Connect via OAuth (GitHub App, GitLab OAuth)
   - Select repo + branch
   - Auto-deploy on push (webhook)
   - Manual deploy button
   - Build detection: Nixpacks, Dockerfile, Buildpack
   - Build arguments
   - Custom build command
   - Root directory override (monorepo support)

2. **Docker Image**
   - Public or private registry (Docker Hub, GHCR, custom)
   - Registry credentials stored per org (encrypted)
   - Image tag pinning or `latest` with auto-pull
   - Digest-based deploys for immutability

3. **Docker Compose**
   - Paste or upload `docker-compose.yml`
   - Multi-service stacks
   - Named volumes, custom networks (scoped to org)
   - Per-service resource limits

#### Build Process
- Build runs in isolated Docker container (buildkit)
- Build logs streamed in real time via WebSocket to frontend
- Build cache per repo/branch
- Build timeout configurable (default 30 min)
- On success: image pushed to local registry or specified registry
- On failure: previous deployment remains active (zero downtime)

#### Deploy Process
- Zero-downtime rolling deploy using Docker Swarm update policy
- Health check configuration (HTTP path, TCP, command)
- Readiness probe before traffic cutover
- Replica count (1–N)
- Deploy strategy: Rolling, Blue-Green (toggle), Recreate
- Rollback to any previous deployment (list of last 20 deploys)
- Deploy hooks: pre-deploy command, post-deploy command (runs in container)

#### Application Configuration
- **Environment Variables**: Key-value pairs, supports secret references, bulk import from `.env` file
- **Secrets**: Sensitive values, masked in UI, encrypted at rest
- **Ports**: Internal port → external domain mapping
- **Volumes**: Named volumes or bind mounts (within allowed paths)
- **Domains**: Custom domains with auto-TLS via Let's Encrypt
- **Resource Limits**: CPU shares, memory limit, memory reservation
- **Restart Policy**: Always, on-failure (with max retries), unless-stopped
- **Network**: Which org network to attach to (default: org's network)
- **Labels**: Custom Docker labels

#### Runtime Features
- **Logs**: Real-time log streaming (tail, search, filter by level)
- **Terminal**: In-browser SSH terminal into running container (xterm.js)
- **Metrics**: CPU%, memory%, network I/O, disk I/O per container
- **Shell exec**: Run one-off commands in container
- **Process list**: View running processes inside container

---

### 5. Database Management

#### Supported Engines
- PostgreSQL 15, 16
- MySQL 8
- MariaDB 10, 11
- MongoDB 6, 7
- Redis 7

#### Provisioning
- One-click spin up in Docker container
- Auto-generated strong password (stored as secret)
- Configurable version tag
- Custom init SQL/script on first start
- Persistent volume auto-created and named

#### Connection Details
- Connection string displayed in UI (masked password, toggle to reveal)
- Internal hostname (accessible from apps in same org network by service name)
- Optional external access via TCP port on Traefik (with IP allowlist)

#### Database Operations
- View logs
- Restart
- Resource limits (CPU, memory)
- Pause/resume
- Backup (manual + scheduled — see Backups)
- Restore from backup
- Database metrics (connections, query stats via pg_stat for PostgreSQL)

---

### 6. Services / One-Click Templates

Pre-built templates for common self-hosted apps. Each template is a Docker Compose config with parameterized fields.

**Available templates (expandable):**
- WordPress (+ MySQL)
- Ghost CMS
- Plausible Analytics
- n8n (workflow automation)
- Metabase
- Grafana
- Uptime Kuma
- Minio (S3-compatible object storage)
- Gitea
- Outline (wiki)
- Nocodb (Airtable alternative)
- Pocketbase
- Appsmith
- Directus
- Umami Analytics
- Cal.com
- Listmonk (email marketing)
- Infisical (secrets manager)
- Vaultwarden (Bitwarden)
- SearXNG

Templates have configurable parameters (domain, admin email, storage size, etc.) rendered as a form in the UI.

Custom templates can be added by super admin (JSON/YAML format stored in DB).

---

### 7. Cron Jobs

#### What It Is
A cron job in Orbita is a Docker container that runs on a schedule and exits. Unlike long-running apps, cron containers spin up, execute, and terminate.

#### Configuration
- **Name**: Human-readable label
- **Schedule**: Standard cron expression (with helper UI: `@daily`, `@hourly`, `@weekly`, `@monthly`, or custom `* * * * *`)
- **Source**: Docker image OR Git repo (build + run)
- **Command**: Override command/entrypoint
- **Environment Variables**: Same as apps (can reference org secrets)
- **Timeout**: Max runtime before container is killed (default: 1h)
- **Concurrency Policy**: 
  - `Allow`: Multiple instances can run simultaneously
  - `Forbid`: Skip if previous run still running
  - `Replace`: Kill previous and start new
- **Resource Limits**: CPU and memory
- **Retry on failure**: Number of retries (0–5)

#### Cron Dashboard
- List of all cron jobs with: name, schedule, last run time, last run status, next run time
- Run history: last 50 executions with status (success/failed/timeout/skipped), duration, exit code
- Real-time logs per execution
- Manual trigger button ("Run Now")
- Enable/disable toggle (pauses scheduling without deleting)

#### How It Works Internally
- Go's `robfig/cron` v3 library manages scheduling in-process
- On trigger: create Docker container from image, set resource limits, attach to org network for secrets resolution
- Container logs captured and stored in DB (capped at 100KB per run)
- Exit code determines success/failure
- Notifications on failure (email via Resend, webhook)

---

### 8. Domain & SSL Management

#### Domain Assignment
- Each resource (app, service, database with TCP) can have one or more domains
- Domains are validated to belong to the org (no cross-org domain stealing)
- Wildcard subdomains supported
- Internal service-to-service communication uses Docker DNS (no external domain needed)

#### SSL / TLS
- Automatic Let's Encrypt certificate via Traefik ACME resolver
- HTTP challenge (default) or DNS challenge (for wildcard certs)
- Certificate auto-renewal (Traefik handles)
- Custom certificate upload (paste PEM)
- Force HTTPS redirect toggle
- Certificate status visible in UI

#### Traefik Integration
- Orbita manages Traefik's dynamic configuration via its File Provider (writes JSON to a watched directory) or HTTP Provider
- Each deploy creates/updates Traefik router + service + middleware entries
- Middleware per org: basic auth, IP allowlist, rate limiting, custom headers
- Path-based routing (e.g., route `/api` to one service, `/` to another)

---

### 9. Resource Slicing & Node Management

#### Virtual Slices (Single-Node)
- Super admin can define "Resource Plans" (e.g., "Starter: 1 CPU / 2GB RAM", "Pro: 4 CPU / 8GB RAM")
- When an org is assigned a plan, Orbita writes a cgroup v2 slice file:  
  `/sys/fs/cgroup/orbita-org-{slug}/memory.max` and `cpu.weight`
- All Docker containers for that org run under `--cgroup-parent=orbita-org-{slug}`
- If an org exceeds its memory quota, Linux OOM killer acts within that cgroup
- CPU weight controls relative CPU share (not hard limit, but fair share)
- Per-resource limits are additive within the org's total budget

#### Multi-Node (Docker Swarm)
- Add worker nodes via the Nodes UI
- Enter IP, SSH port, SSH key → Orbita connects and initializes Docker Swarm worker
- Nodes are labeled (e.g., `region=us`, `tier=pro`)
- Resources can be pinned to a node or node label via placement constraints
- Swarm services used for all long-running apps (not bare Docker containers)
- Node health visible in dashboard (CPU, memory, disk, Docker state)
- Drain/remove node gracefully (reschedules workloads)

#### Resource Dashboard
- Per-org: real-time CPU%, memory%, disk usage, network I/O vs quota
- Per-node: full utilization metrics
- Alerts when org exceeds 80% of quota (in-app notification + optional email)

---

### 10. Backups

#### Database Backups
- Schedule: manual, hourly, daily, weekly
- Backup targets: local disk, AWS S3, Cloudflare R2, Backblaze B2, SFTP
- Retention policy: keep last N backups
- Backup format: native dump (pg_dump, mysqldump, mongodump)
- Compressed with gzip
- Backup encryption (AES-256) before upload to remote
- Restore: pick a backup → restores to running database instance
- Backup status and size visible per database

#### Volume Backups
- Tar + gzip Docker named volumes
- Same destination targets as database backups
- Can schedule alongside database backups

---

### 11. Monitoring & Observability

#### Metrics (per resource)
- CPU usage %
- Memory usage (used / limit)
- Network: bytes in / bytes out
- Disk I/O: read / write bytes per second
- Container restarts counter
- Uptime since last deploy
- Request count / response time (if HTTP healthcheck configured)

Collected via Docker Stats API (streaming), stored in Redis time-series (recent 24h), PostgreSQL for historical (7 days free, configurable).

#### Logs
- Real-time log streaming via WebSocket (gorilla/websocket on server, native WS on client)
- Historical logs stored in PostgreSQL (last 10,000 lines per resource, configurable)
- Search/filter: keyword, log level (auto-detected), timestamp range
- Log download as .txt or .json
- Multi-source log merge (e.g., view all containers in a Compose stack together)

#### Notifications
- In-app notification bell (WebSocket push)
- Email notifications (Resend):
  - Deploy success / failure
  - Cron job failure
  - Resource down (health check failing)
  - Backup failure
  - Org approaching quota limit (80%, 95%, 100%)
- Webhook notifications: POST JSON payload to user-configured URL on events
- Notification settings per org, per user

#### Audit Log
- Every action tracked: who did what, to which resource, when, from which IP
- Viewable per org (admins) and globally (super admin)
- Exported as CSV

---

### 12. Git Provider Integration

#### GitHub
- GitHub App installation (preferred) or OAuth App
- Auto-detect repos the user has access to
- Branch list auto-populated from GitHub API
- Webhook auto-configured on repo for push events
- Pull Request preview environments (create ephemeral environment per PR)
- Commit status checks (post deploy status back to GitHub)

#### GitLab
- GitLab OAuth2 integration
- Group/project listing
- Merge Request preview environments
- Pipeline trigger integration

#### Gitea / Forgejo
- Self-hosted Gitea support
- Enter base URL + personal access token
- Webhook auto-registration

#### Generic Git
- Any public git URL
- Polling-based auto-deploy (check for new commits every N minutes)

---

### 13. Private Docker Registry

- Built-in Docker registry (runs as a service on primary node)
- All built images pushed to this registry
- Registry authentication (username/password, per org)
- Image list and tag management per resource
- Garbage collection (prune old/unused images)
- External registry support (Docker Hub, GHCR, ECR, GCR, custom)
- Registry credentials stored encrypted per org

---

### 14. API & CLI

#### REST API
- Full REST API (same endpoints the frontend uses)
- API key authentication (per user, per org, scoped to permissions)
- Documented with OpenAPI/Swagger (served at `/api/docs`)
- Rate limiting: 1000 req/min per API key (configurable by super admin)
- Webhook signatures for incoming events (HMAC-SHA256)

#### CLI (Future Phase)
- `orbita deploy` — trigger deploy
- `orbita logs <app>` — stream logs
- `orbita exec <app> <cmd>` — run command in container
- `orbita env set <key> <value>` — set environment variable
- `orbita db backup` — trigger backup
- Written in Go, distributed as single binary

---

### 15. Security

- All passwords hashed with bcrypt (cost 12)
- JWT secrets rotated on startup if not set
- Secrets (env vars) encrypted with AES-256-GCM using per-org derived key
- Traefik enforces HTTPS everywhere
- Rate limiting on auth endpoints (5 attempts per 15 min per IP)
- CORS policy (only allow frontend origin)
- SQL injection: GORM parameterized queries
- XSS: React handles escaping; CSP headers via Traefik middleware
- Container isolation: no privileged containers by default (configurable per resource for super admin only)
- SSH terminal: limited to org members with Developer+ role, logged in audit trail
- API key scopes: read, deploy, admin
- IP allowlist per org (restrict access to dashboard from specific IPs)

---

## Data Models (Key Tables)

```
users                   — id, email, password_hash, name, avatar_url, is_super_admin, totp_secret, created_at
organizations           — id, name, slug, plan_id, created_by, created_at
org_members             — org_id, user_id, role, joined_at
org_invites             — id, org_id, email, role, token_hash, expires_at, used_at
resource_plans          — id, name, max_cpu_cores, max_ram_mb, max_disk_gb, max_apps, max_databases
projects                — id, org_id, name, description, emoji, created_at
environments            — id, project_id, name, type (production|staging|custom)
applications            — id, env_id, org_id, name, source_type, source_config (JSON), build_config (JSON), deploy_config (JSON), status, created_at
databases               — id, env_id, org_id, name, engine, version, connection_config (JSON encrypted), volume_name, status
services                — id, env_id, org_id, template_id, name, config (JSON), status
cron_jobs               — id, env_id, org_id, name, schedule, image, command, env_config (JSON encrypted), timeout, concurrency_policy, enabled, last_run_at, next_run_at
cron_runs               — id, cron_job_id, started_at, finished_at, status, exit_code, log_snippet
deployments             — id, app_id, version, image_ref, deploy_config (JSON), status, started_at, finished_at, triggered_by, trigger_type (push|manual|webhook)
domains                 — id, resource_id, resource_type, org_id, domain, ssl_enabled, ssl_config (JSON)
env_variables           — id, resource_id, resource_type, org_id, key, value_encrypted, is_secret
volumes                 — id, org_id, name, driver, size_gb
nodes                   — id, name, ip, ssh_port, ssh_key_id, role (primary|worker), status, labels (JSON), created_at
backups                 — id, source_id, source_type, org_id, status, size_bytes, storage_path, created_at, expires_at
backup_schedules        — id, source_id, source_type, org_id, frequency, retention_count, destination_config (JSON encrypted)
audit_logs              — id, org_id, user_id, action, resource_type, resource_id, metadata (JSON), ip, created_at
notifications           — id, user_id, org_id, type, title, body, read, created_at
notification_settings   — org_id, user_id, event_type, email_enabled, webhook_url, webhook_enabled
sessions                — id, user_id, refresh_token_hash, device_info, ip, expires_at, created_at
api_keys                — id, user_id, org_id, name, key_hash, scopes, last_used_at, expires_at
registry_credentials    — id, org_id, registry_url, username, password_encrypted
git_connections         — id, org_id, provider (github|gitlab|gitea), access_token_encrypted, refresh_token_encrypted, metadata (JSON)
templates               — id, name, description, category, compose_template, params_schema (JSON), icon_url, is_active
```

---

## Frontend Structure

```
web/
  src/
    pages/
      auth/          — Login, Register, ForgotPassword, AcceptInvite
      dashboard/     — Overview, activity feed
      projects/      — Project list, project detail
      apps/          — App list, app detail (deploy, logs, terminal, metrics, settings)
      databases/     — Database list, database detail
      services/      — Service list, one-click deploy template gallery
      cron/          — Cron job list, cron job detail, run history
      domains/       — Domain management
      members/       — Team members, invite, roles
      settings/      — Org settings, billing/plan, secrets, env vars, backups
      nodes/         — Node management (super admin)
      admin/         — Super admin: orgs, plans, platform settings, audit log
    components/
      layout/        — Sidebar, Topbar, NotificationBell
      deploy/        — DeployButton, DeployStatus, BuildLogs, RollbackModal
      logs/          — LogViewer (real-time streaming)
      terminal/      — XtermTerminal (WebSocket SSH)
      metrics/       — ResourceChart, NodeMetrics
      forms/         — EnvVarEditor, DomainForm, CronScheduleEditor
      ui/            — shadcn/ui components
    stores/          — Zustand stores (auth, org, notifications)
    hooks/           — useWebSocket, useDeployLogs, useMetrics
    api/             — Typed API client (fetch + React Query)
    utils/           — formatBytes, cronToHuman, relativeTime
```

---

## API Structure

All routes prefixed with `/api/v1`

```
POST   /auth/login
POST   /auth/register
POST   /auth/logout
POST   /auth/refresh
POST   /auth/forgot-password
POST   /auth/reset-password
POST   /auth/verify-email
POST   /auth/totp/enable
POST   /auth/totp/verify

GET    /me
PUT    /me
GET    /me/sessions
DELETE /me/sessions/:id
GET    /me/api-keys
POST   /me/api-keys
DELETE /me/api-keys/:id

GET    /orgs
POST   /orgs
GET    /orgs/:orgSlug
PUT    /orgs/:orgSlug
DELETE /orgs/:orgSlug
GET    /orgs/:orgSlug/members
POST   /orgs/:orgSlug/invites
GET    /orgs/:orgSlug/invites
DELETE /orgs/:orgSlug/invites/:id
PUT    /orgs/:orgSlug/members/:userId/role
DELETE /orgs/:orgSlug/members/:userId

GET    /orgs/:orgSlug/projects
POST   /orgs/:orgSlug/projects
GET    /orgs/:orgSlug/projects/:projectId
PUT    /orgs/:orgSlug/projects/:projectId
DELETE /orgs/:orgSlug/projects/:projectId

GET    /orgs/:orgSlug/projects/:projectId/environments
POST   /orgs/:orgSlug/projects/:projectId/environments
...

GET    /orgs/:orgSlug/apps
POST   /orgs/:orgSlug/apps
GET    /orgs/:orgSlug/apps/:appId
PUT    /orgs/:orgSlug/apps/:appId
DELETE /orgs/:orgSlug/apps/:appId
POST   /orgs/:orgSlug/apps/:appId/deploy
POST   /orgs/:orgSlug/apps/:appId/rollback/:deploymentId
GET    /orgs/:orgSlug/apps/:appId/deployments
GET    /orgs/:orgSlug/apps/:appId/logs            (WebSocket)
GET    /orgs/:orgSlug/apps/:appId/metrics
POST   /orgs/:orgSlug/apps/:appId/restart
POST   /orgs/:orgSlug/apps/:appId/stop
POST   /orgs/:orgSlug/apps/:appId/start
GET    /orgs/:orgSlug/apps/:appId/terminal        (WebSocket)
GET    /orgs/:orgSlug/apps/:appId/env
POST   /orgs/:orgSlug/apps/:appId/env
PUT    /orgs/:orgSlug/apps/:appId/env/:key
DELETE /orgs/:orgSlug/apps/:appId/env/:key
GET    /orgs/:orgSlug/apps/:appId/domains
POST   /orgs/:orgSlug/apps/:appId/domains
DELETE /orgs/:orgSlug/apps/:appId/domains/:id

GET    /orgs/:orgSlug/databases
POST   /orgs/:orgSlug/databases
GET    /orgs/:orgSlug/databases/:dbId
...similar CRUD + /backups, /restore, /logs, /metrics

GET    /orgs/:orgSlug/cron-jobs
POST   /orgs/:orgSlug/cron-jobs
GET    /orgs/:orgSlug/cron-jobs/:cronId
PUT    /orgs/:orgSlug/cron-jobs/:cronId
DELETE /orgs/:orgSlug/cron-jobs/:cronId
POST   /orgs/:orgSlug/cron-jobs/:cronId/run
GET    /orgs/:orgSlug/cron-jobs/:cronId/runs
GET    /orgs/:orgSlug/cron-jobs/:cronId/runs/:runId/logs

GET    /orgs/:orgSlug/services
POST   /orgs/:orgSlug/services
...

GET    /orgs/:orgSlug/domains
POST   /orgs/:orgSlug/domains
...

GET    /orgs/:orgSlug/notifications
PUT    /orgs/:orgSlug/notifications/:id/read
GET    /orgs/:orgSlug/notification-settings
PUT    /orgs/:orgSlug/notification-settings
GET    /orgs/:orgSlug/audit-logs
GET    /orgs/:orgSlug/metrics/overview

GET    /templates

POST   /webhooks/github
POST   /webhooks/gitlab
POST   /webhooks/gitea

/admin/...   (super admin only — orgs, plans, nodes, platform settings)

GET    /health
GET    /api/docs
```

---

## Non-Functional Requirements

- **Performance**: API response time < 200ms p95 for non-deploy operations
- **Reliability**: Deploy failures must not affect currently running services
- **Security**: Zero cross-tenant data leakage by design
- **Scalability**: Support 50+ orgs on a single 16GB node; scale out via Swarm for more
- **Observability**: Every error logged with zerolog, correlation IDs on requests
- **Upgradability**: DB migrations versioned with golang-migrate; backward-compatible API
- **Portability**: Runs on any Linux VPS with Docker installed; ARM64 + AMD64 supported
