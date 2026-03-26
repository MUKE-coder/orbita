<p align="center">
  <h1 align="center">Orbita</h1>
  <p align="center">
    <strong>Self-Hosted PaaS with True Multi-Tenancy</strong>
  </p>
  <p align="center">
    One VPS. Many Clients. Full Isolation.
  </p>
  <p align="center">
    <a href="#features">Features</a> &middot;
    <a href="#quick-start">Quick Start</a> &middot;
    <a href="#deployment-guide">Deployment Guide</a> &middot;
    <a href="#api-reference">API Reference</a> &middot;
    <a href="#architecture">Architecture</a>
  </p>
</p>

---

## What is Orbita?

Orbita is an open-source, self-hosted **Platform-as-a-Service (PaaS)** built in Go, designed for developers, freelancers, and agencies who manage infrastructure for multiple clients on a single server.

It lets you take a single VPS and turn it into a fully isolated hosting environment for multiple client organizations — each with their own dashboard login, projects, environment variables, secrets, logs, domains, and resource budget. Clients never see each other. You, as the super-admin, see everything.

**Ships as a single ~30MB binary** with an embedded React SPA. The control plane uses **under 50MB of RAM** at idle, leaving nearly all server resources for workloads.

### Why Orbita?

| Problem | Orbita's Solution |
|---------|-------------------|
| Dokploy/Coolify lack multi-tenancy | True tenant isolation: separate networks, secrets, volumes, quotas |
| No working invite/RBAC system | 4-role RBAC (Owner/Admin/Developer/Viewer) with email invites |
| High memory overhead (PHP/Node.js) | Go binary — under 50MB idle RAM |
| No resource quotas per client | cgroup v2 slices enforce CPU/memory limits per org |
| No cron job management | Built-in cron scheduler with run history and concurrency policies |
| Vendor lock-in with managed PaaS | 100% self-hosted, no per-seat fees, you own everything |

---

## Features

### Core Platform

- **Multi-Tenancy & Organizations** — Fully isolated tenants with separate Docker networks, encryption keys, volume namespaces, and database-level scoping
- **RBAC (4 Roles)** — Owner, Admin, Developer, Viewer with API-level enforcement
- **Team Invites** — Cryptographically secure invite tokens via email (72h expiry)
- **Projects & Environments** — Logical grouping with auto-created Production + Staging environments
- **Resource Plans** — Super admin defines plans (Free/Starter/Pro/Enterprise) with CPU, RAM, disk, app limits

### Application Deployment

- **Docker Image Deploy** — Deploy from Docker Hub, GHCR, or any registry
- **Git Auto-Deploy** — Connect GitHub, GitLab, or Gitea. Auto-deploy on push via webhooks
- **Build Engine** — Build from Dockerfile or Nixpacks (stubs ready for production)
- **Zero-Downtime Deploys** — Rolling updates via Docker Swarm with rollback to any previous version
- **Deploy History** — Versioned deployment records with status, trigger type, timestamps
- **App Lifecycle** — Start, stop, restart, scale replicas

### Database Management

- **One-Click Provisioning** — PostgreSQL 15/16, MySQL 8, MariaDB 10/11, MongoDB 6/7, Redis 7
- **Auto-Generated Credentials** — Strong passwords, encrypted connection strings
- **Scheduled Backups** — Hourly/daily/weekly with configurable retention
- **Backup & Restore** — Create manual backups, restore from any backup point

### Cron Jobs

- **Scheduled Containers** — Docker containers that run on a cron schedule and exit
- **Concurrency Policies** — Allow, Forbid (skip if running), Replace (kill previous)
- **Run History** — Last 50 executions with status, exit code, duration, logs
- **Manual Trigger** — "Run Now" button for immediate execution
- **Timeout Enforcement** — Configurable max runtime per job

### Domains & SSL

- **Custom Domains** — Assign domains to apps, databases, and services
- **Automatic TLS** — Let's Encrypt certificates via Traefik ACME
- **HTTP → HTTPS Redirect** — Automatic redirect with security headers
- **DNS Verification** — Built-in DNS lookup to verify domain configuration

### Service Marketplace

- **10 Pre-Built Templates** — WordPress, Plausible Analytics, Uptime Kuma, n8n, Metabase, Grafana, MinIO, Gitea, Ghost CMS, Vaultwarden
- **Parameterized Deploys** — Each template has configurable parameters (port, admin password, etc.)
- **One-Click Deploy** — Select template, fill params, deploy

### Monitoring & Observability

- **Real-Time Metrics** — CPU%, memory, network I/O per container
- **Dashboard Overview** — Running apps, databases, cron jobs, recent deploys, resource usage bars
- **Log Streaming** — Real-time log viewer with polling (WebSocket infrastructure ready)
- **In-Browser Terminal** — xterm.js SSH into running containers (WebSocket)
- **Shell Exec** — Run one-off commands in containers via API

### Security

- **JWT Authentication** — 15-minute access tokens + 30-day refresh tokens (httpOnly cookies)
- **bcrypt Password Hashing** — Cost 12
- **AES-256-GCM Encryption** — All secrets encrypted at rest with per-org derived keys (HKDF-SHA256)
- **Rate Limiting** — Redis sliding window (5 req/15min on auth endpoints)
- **CORS** — Restricted to configured origin
- **Org-Scoped Queries** — Every database query includes `organization_id` in WHERE clause
- **Webhook Signatures** — HMAC-SHA256 verification for GitHub webhooks
- **API Keys** — Per-user keys with `orb_` prefix for CI/CD integration

### Notifications & Audit

- **In-App Notifications** — Bell icon with unread count, auto-refresh
- **Audit Logs** — Every action logged with user, action, resource, IP, timestamp
- **Paginated History** — Searchable audit trail per organization

### Infrastructure

- **Multi-Node Support** — Add worker nodes via SSH, Docker Swarm orchestration
- **cgroup Resource Slicing** — Per-org CPU/memory limits via cgroup v2
- **Node Management** — Add, drain, remove nodes from the super admin panel

---

## Tech Stack

### Backend
| Component | Technology |
|-----------|------------|
| Language | Go 1.22+ |
| HTTP Framework | Gin |
| ORM | GORM + PostgreSQL |
| Cache/Queue | Redis 7 |
| Auth | JWT (golang-jwt/jwt v5) |
| Email | Resend API |
| Containers | Docker SDK |
| Reverse Proxy | Traefik v3 |
| Cron | robfig/cron v3 |
| WebSocket | gorilla/websocket |
| Migrations | golang-migrate |
| Logging | zerolog |

### Frontend
| Component | Technology |
|-----------|------------|
| Framework | React 18 + TypeScript |
| Build Tool | Vite |
| Styling | Tailwind CSS v4 |
| Components | shadcn/ui |
| State | Zustand |
| Data Fetching | TanStack Query |
| Routing | React Router v6 |
| Forms | React Hook Form + Zod |
| Terminal | xterm.js |
| Icons | Lucide React |

### Build Output
Single Go binary with embedded React SPA via `//go:embed all:web/dist`

---

## Quick Start

### Prerequisites
- Docker 24+ with Docker Compose v2
- A Linux VPS (Ubuntu 22.04+ recommended)

### 1-Minute Install

```bash
# Clone the repository
git clone https://github.com/MUKE-coder/orbita.git
cd orbita

# Start services
docker compose -f docker/docker-compose.dev.yml up -d

# Copy and configure environment
cp .env.example .env
# Edit .env — set JWT_SECRET, JWT_REFRESH_SECRET, ENCRYPTION_MASTER_KEY

# Run the server
go run ./cmd/server/
```

Open `http://localhost:8080` — the first user to register becomes the super admin.

---

## Deployment Guide

### Step-by-Step VPS/Dedicated Server Deployment

This guide walks you through deploying Orbita on a fresh Linux VPS from scratch. Estimated time: 10-15 minutes.

#### Step 1: Server Preparation

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Install essential tools
sudo apt install -y curl git ufw

# Configure firewall
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw allow 8080/tcp  # Orbita (remove after setting up Traefik)
sudo ufw enable
```

#### Step 2: Install Docker

```bash
# Install Docker using the official script
curl -fsSL https://get.docker.com | sh

# Add your user to the docker group
sudo usermod -aG docker $USER

# Apply group changes (log out and back in, or run:)
newgrp docker

# Verify Docker is working
docker --version
docker compose version

# Initialize Docker Swarm (required for service orchestration)
docker swarm init
```

#### Step 3: Create Project Directory

```bash
sudo mkdir -p /opt/orbita
sudo chown $USER:$USER /opt/orbita
cd /opt/orbita
```

#### Step 4: Create Docker Compose File

```bash
cat > docker-compose.yml << 'COMPOSE_EOF'
version: '3.8'

services:
  orbita:
    image: ghcr.io/muke-coder/orbita:latest
    container_name: orbita
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://orbita:${DB_PASSWORD}@postgres:5432/orbita?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
      - JWT_REFRESH_SECRET=${JWT_REFRESH_SECRET}
      - ENCRYPTION_MASTER_KEY=${ENCRYPTION_MASTER_KEY}
      - APP_BASE_URL=${APP_BASE_URL}
      - RESEND_API_KEY=${RESEND_API_KEY}
      - RESEND_FROM_EMAIL=${RESEND_FROM_EMAIL}
      - TRAEFIK_CONFIG_DIR=/etc/orbita/traefik
      - DOCKER_SOCKET=/var/run/docker.sock
      - IS_PRODUCTION=true
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - traefik_dynamic:/etc/orbita/traefik
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    container_name: orbita-postgres
    environment:
      - POSTGRES_DB=orbita
      - POSTGRES_USER=orbita
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U orbita"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    container_name: orbita-redis
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  traefik:
    image: traefik:v3.0
    container_name: orbita-traefik
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - traefik_dynamic:/etc/orbita/traefik
      - traefik_certs:/etc/traefik/acme
    command:
      - "--api.dashboard=false"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--providers.file.directory=/etc/orbita/traefik/dynamic"
      - "--providers.file.watch=true"
      - "--certificatesresolvers.letsencrypt.acme.email=${ACME_EMAIL}"
      - "--certificatesresolvers.letsencrypt.acme.storage=/etc/traefik/acme/acme.json"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
      - "--log.level=INFO"
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  traefik_dynamic:
  traefik_certs:
COMPOSE_EOF
```

#### Step 5: Generate Secrets and Create .env File

```bash
cat > .env << ENV_EOF
# Database
DB_PASSWORD=$(openssl rand -hex 16)

# JWT Secrets (MUST be unique and random)
JWT_SECRET=$(openssl rand -hex 32)
JWT_REFRESH_SECRET=$(openssl rand -hex 32)

# Encryption key for secrets (32 hex chars = 16 bytes)
ENCRYPTION_MASTER_KEY=$(openssl rand -hex 16)

# Your domain (change this!)
APP_BASE_URL=https://orbita.yourdomain.com

# Email via Resend (get a key at https://resend.com)
RESEND_API_KEY=re_your_api_key_here
RESEND_FROM_EMAIL=noreply@yourdomain.com

# Let's Encrypt email for SSL certificates
ACME_EMAIL=admin@yourdomain.com
ENV_EOF

# Verify the generated secrets
echo "Generated .env file:"
cat .env
```

> **IMPORTANT:** Back up this `.env` file securely. The `ENCRYPTION_MASTER_KEY` is required to decrypt all stored secrets. Losing it means losing access to all encrypted environment variables, database passwords, and API tokens.

#### Step 6: Set Up DNS

Before starting Orbita, configure your DNS:

| Record Type | Name | Value |
|-------------|------|-------|
| A | `orbita.yourdomain.com` | `YOUR_SERVER_IP` |
| A | `*.orbita.yourdomain.com` | `YOUR_SERVER_IP` |

The wildcard record allows automatic subdomain assignment for deployed apps.

Wait for DNS propagation (usually 1-5 minutes, up to 48 hours).

```bash
# Verify DNS is pointing to your server
dig orbita.yourdomain.com +short
# Should return your server IP
```

#### Step 7: Start Orbita

```bash
cd /opt/orbita

# Pull images and start all services
docker compose up -d

# Watch the logs to ensure everything starts correctly
docker compose logs -f orbita
```

You should see:
```
Connected to PostgreSQL
Migrations applied
Connected to Redis
Cron scheduler started  jobs_loaded=0
Starting Orbita server  port=8080
```

#### Step 8: Register Super Admin

1. Open `https://orbita.yourdomain.com` (or `http://YOUR_SERVER_IP:8080` if DNS isn't ready yet)
2. Click **"Get Started"** or navigate to `/register`
3. Create the first account — this user automatically becomes the **super admin**
4. Log in and create your first organization

#### Step 9: Verify Installation

```bash
# Health check
curl -s https://orbita.yourdomain.com/health
# Expected: {"status":"ok","version":"0.1.0"}

# Check all containers are running
docker compose ps
# All should show "Up" status

# Check resource usage
docker stats --no-stream
```

#### Step 10: Post-Deployment Security Hardening

```bash
# 1. Remove direct port 8080 access (traffic should go through Traefik)
sudo ufw delete allow 8080/tcp

# 2. Set up automatic security updates
sudo apt install -y unattended-upgrades
sudo dpkg-reconfigure -plow unattended-upgrades

# 3. Set up PostgreSQL backup cron job
cat > /opt/orbita/backup.sh << 'BACKUP_EOF'
#!/bin/bash
BACKUP_DIR="/opt/orbita/backups"
mkdir -p $BACKUP_DIR
docker exec orbita-postgres pg_dump -U orbita orbita | gzip > $BACKUP_DIR/orbita-$(date +%Y%m%d-%H%M%S).sql.gz
# Keep last 30 days of backups
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete
BACKUP_EOF
chmod +x /opt/orbita/backup.sh

# Add to crontab (daily at 2 AM)
echo "0 2 * * * /opt/orbita/backup.sh" | sudo tee -a /etc/crontab

# 4. Monitor disk space
df -h
```

---

### Updating Orbita

```bash
cd /opt/orbita

# Pull the latest image
docker compose pull orbita

# Restart with the new version (zero downtime)
docker compose up -d orbita

# Check logs
docker compose logs -f orbita
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/MUKE-coder/orbita.git
cd orbita

# Build frontend
cd web && npm install && npm run build && cd ..

# Build Go binary
go build -ldflags="-s -w" -o orbita ./cmd/server/

# Or build Docker image
docker build -t orbita:local .
```

---

## API Reference

All routes are prefixed with `/api/v1`. Authentication via `Authorization: Bearer <token>` header.

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Register new user |
| POST | `/auth/login` | Login |
| POST | `/auth/logout` | Logout |
| POST | `/auth/refresh` | Refresh tokens |
| POST | `/auth/forgot-password` | Request password reset OTP |
| POST | `/auth/reset-password` | Reset password with OTP |
| POST | `/auth/verify-email` | Verify email address |

### User Profile
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/me` | Get profile |
| PUT | `/me` | Update profile |
| POST | `/me/change-password` | Change password |
| GET | `/me/sessions` | List active sessions |
| DELETE | `/me/sessions/:id` | Revoke session |
| GET/POST/DELETE | `/me/api-keys` | Manage API keys |

### Organizations
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/POST | `/orgs` | List/create organizations |
| GET/PUT/DELETE | `/orgs/:slug` | Get/update/delete org |
| GET | `/orgs/:slug/members` | List members |
| POST | `/orgs/:slug/invites` | Send invite |
| GET | `/orgs/:slug/invites` | List pending invites |
| PUT | `/orgs/:slug/members/:userId/role` | Change member role |
| POST | `/orgs/:slug/leave` | Leave organization |

### Projects & Environments
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/POST | `/orgs/:slug/projects` | List/create projects |
| GET/PUT/DELETE | `/orgs/:slug/projects/:id` | Project CRUD |
| GET/POST | `/orgs/:slug/projects/:id/environments` | Environment CRUD |

### Applications
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/POST | `/orgs/:slug/apps` | List/create apps |
| GET/PUT/DELETE | `/orgs/:slug/apps/:id` | App CRUD |
| POST | `/orgs/:slug/apps/:id/deploy` | Trigger deploy |
| POST | `/orgs/:slug/apps/:id/rollback/:deployId` | Rollback |
| POST | `/orgs/:slug/apps/:id/stop\|start\|restart` | Lifecycle |
| GET | `/orgs/:slug/apps/:id/deployments` | Deploy history |
| GET | `/orgs/:slug/apps/:id/logs` | Get logs |
| GET | `/orgs/:slug/apps/:id/metrics` | Get metrics |
| GET | `/orgs/:slug/apps/:id/status` | Get status |
| POST | `/orgs/:slug/apps/:id/exec` | Run command |
| GET | `/orgs/:slug/apps/:id/terminal` | WebSocket terminal |
| GET/POST/DELETE | `/orgs/:slug/apps/:id/env` | Environment variables |
| GET/POST/DELETE | `/orgs/:slug/apps/:id/domains` | Domain management |

### Databases
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/POST | `/orgs/:slug/databases` | List/create |
| GET/DELETE | `/orgs/:slug/databases/:id` | Get/delete |
| POST | `/orgs/:slug/databases/:id/restart\|stop\|start` | Lifecycle |
| GET/POST | `/orgs/:slug/databases/:id/backups` | Backups |
| POST | `/orgs/:slug/databases/:id/backups/:id/restore` | Restore |

### Cron Jobs
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/POST | `/orgs/:slug/cron-jobs` | List/create |
| GET/PUT/DELETE | `/orgs/:slug/cron-jobs/:id` | CRUD |
| POST | `/orgs/:slug/cron-jobs/:id/toggle` | Enable/disable |
| POST | `/orgs/:slug/cron-jobs/:id/run` | Manual trigger |
| GET | `/orgs/:slug/cron-jobs/:id/runs` | Run history |

### Services & Templates
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/templates` | List available templates |
| GET/POST | `/orgs/:slug/services` | List/deploy services |
| GET/DELETE | `/orgs/:slug/services/:id` | Get/delete service |

### Admin (Super Admin Only)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/POST/PUT/DELETE | `/admin/plans` | Resource plan CRUD |
| GET | `/admin/orgs` | List all organizations |
| PUT | `/admin/orgs/:slug/plan` | Assign plan to org |
| GET/POST/DELETE | `/admin/nodes` | Node management |
| GET | `/admin/platform/metrics` | Platform overview |

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Orbita Binary                        │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │  Gin Router  │  │  React SPA   │  │ Cron Scheduler │  │
│  │  (REST API)  │  │  (Embedded)  │  │ (robfig/cron) │  │
│  └──────┬───────┘  └──────────────┘  └───────┬───────┘  │
│         │                                     │          │
│  ┌──────┴───────────────────────────────────┴───────┐   │
│  │              Service Layer                        │   │
│  │  Auth │ Org │ Project │ App │ DB │ Cron │ Domain  │   │
│  └──────┬────────────────────────────────────────────┘   │
│         │                                                │
│  ┌──────┴───────────────────────────────────────────┐   │
│  │            Repository Layer (GORM + OrgScope)     │   │
│  └──────┬────────────────────────────────────────────┘   │
└─────────┼────────────────────────────────────────────────┘
          │
    ┌─────┴─────┐     ┌──────────┐     ┌──────────────┐
    │ PostgreSQL │     │  Redis   │     │ Docker Engine│
    │   (Data)   │     │ (Cache)  │     │ (Containers) │
    └───────────┘     └──────────┘     └──────┬───────┘
                                              │
                                       ┌──────┴───────┐
                                       │   Traefik    │
                                       │ (Reverse     │
                                       │  Proxy + TLS)│
                                       └──────────────┘
```

### Directory Structure

```
orbita/
├── cmd/
│   ├── server/main.go          # Entry point with graceful shutdown
│   └── migrate/main.go         # Migration CLI
├── internal/
│   ├── api/                    # Gin router + handlers
│   │   └── handlers/           # One file per resource (auth, org, app, db...)
│   ├── auth/                   # JWT, bcrypt, AES-256-GCM, HKDF
│   ├── config/                 # Environment config loader
│   ├── cron/                   # Scheduler + executor
│   ├── database/               # GORM connection + migrations
│   ├── docker/                 # Docker SDK wrapper
│   ├── mailer/                 # Resend email client
│   ├── middleware/             # Auth, RBAC, rate limiting, org scoping
│   ├── models/                 # GORM models (15 models)
│   ├── orchestrator/           # Deploy, provision, build, cgroup, node
│   ├── queue/                  # Redis deploy queue
│   ├── redis/                  # Redis client
│   ├── repository/             # Data access layer (10 repos, all org-scoped)
│   ├── response/               # Standardized JSON responses
│   ├── service/                # Business logic (10 services)
│   ├── traefik/                # Dynamic config file writer
│   └── websocket/              # WS hub + terminal handler
├── migrations/                 # 22 SQL migration files
├── web/                        # React SPA (Vite + TypeScript)
│   └── src/
│       ├── api/                # Typed API clients
│       ├── components/         # Reusable components
│       ├── pages/              # 25+ pages
│       └── stores/             # Zustand stores
├── docker/                     # Docker Compose + Traefik config
├── static.go                   # go:embed directive
├── Dockerfile                  # Multi-stage production build
├── Makefile                    # Dev/build/test/lint targets
└── .env.example                # Environment template
```

---

## Database Schema

22 tables across 22 migrations:

| Table | Purpose |
|-------|---------|
| `users` | User accounts with bcrypt passwords |
| `sessions` | Refresh token sessions |
| `email_verifications` | Email verification tokens |
| `password_resets` | OTP-based password reset |
| `api_keys` | Personal API keys (`orb_` prefix) |
| `resource_plans` | CPU/RAM/disk/app limits (Free/Starter/Pro/Enterprise) |
| `organizations` | Tenant organizations with slugs |
| `org_members` | User-org membership with roles |
| `org_invites` | Pending invitations (72h expiry) |
| `projects` | Logical resource grouping |
| `environments` | Production/Staging/Custom per project |
| `applications` | Deployed apps with JSONB configs |
| `deployments` | Versioned deployment history |
| `managed_databases` | Provisioned database instances |
| `backups` | Database backup records |
| `backup_schedules` | Automated backup configuration |
| `cron_jobs` | Scheduled container execution |
| `cron_runs` | Cron execution history |
| `domains` | Custom domain assignments |
| `git_connections` | Git provider OAuth tokens (encrypted) |
| `env_variables` | Encrypted environment variables |
| `notifications` | In-app notification records |
| `audit_logs` | Action audit trail |
| `templates` | Service marketplace templates |
| `services` | Deployed template instances |
| `nodes` | Swarm worker nodes |
| `registry_credentials` | Docker registry auth (encrypted) |
| `notification_settings` | Per-user notification preferences |

---

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run tests: `make test`
5. Commit: `git commit -m "Add my feature"`
6. Push: `git push origin feature/my-feature`
7. Open a Pull Request

---

## License

MIT License. See [LICENSE](LICENSE) for details.

---

<p align="center">
  Built with Go, React, and a lot of coffee.
  <br/>
  <a href="https://github.com/MUKE-coder/orbita">github.com/MUKE-coder/orbita</a>
</p>
