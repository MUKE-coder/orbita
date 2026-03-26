# Orbita — Project Build Phases

## Overview

This document defines the exact sequence of phases and tasks for building Orbita. Each phase builds on the previous one. Phases are ordered so that a working, testable system exists at the end of every phase. **Do not skip phases or tasks within a phase.**

---

## Phase 0: Project Scaffold & Tooling

**Goal**: Repository structure, toolchain, Docker environment, and CI-ready project skeleton are in place. Nothing runs yet, but everything compiles.

### Tasks

- [ ] **0.1 — Directory structure**
  Create the following layout:
  ```
  orbita/
    cmd/server/main.go
    internal/
      api/          (Gin router, handlers)
      auth/         (JWT, bcrypt, sessions)
      config/       (env config loader)
      database/     (GORM setup, migrations)
      docker/       (Docker SDK wrapper)
      mailer/       (Resend client)
      middleware/   (Gin middlewares)
      models/       (GORM models)
      orchestrator/ (deploy engine)
      queue/        (Redis worker jobs)
      redis/        (Redis client)
      repository/   (data access layer)
      service/      (business logic layer)
      traefik/      (Traefik config manager)
      cron/         (cron job scheduler)
      websocket/    (WS hub)
    web/            (React SPA — Vite project)
    migrations/     (SQL migration files)
    docker/         (Dockerfile, docker-compose.dev.yml)
    .env.example
    Makefile
    go.mod
    go.sum
  ```

- [ ] **0.2 — Go modules**
  Initialize `go.mod` with module name `github.com/orbita-sh/orbita`. Add all dependencies:
  - `github.com/gin-gonic/gin`
  - `gorm.io/gorm`
  - `gorm.io/driver/postgres`
  - `github.com/go-redis/redis/v9`
  - `github.com/golang-jwt/jwt/v5`
  - `github.com/google/uuid`
  - `github.com/docker/docker` (docker/client)
  - `github.com/gorilla/websocket`
  - `github.com/robfig/cron/v3`
  - `github.com/resend/resend-go/v2`
  - `github.com/golang-migrate/migrate/v4`
  - `github.com/rs/zerolog`
  - `github.com/joho/godotenv`
  - `golang.org/x/crypto`
  - `github.com/gin-contrib/cors`
  - `github.com/gin-contrib/requestid`

- [ ] **0.3 — Config loader**
  `internal/config/config.go`: Load all config from env vars. Struct fields:
  - ServerPort, DatabaseURL, RedisURL
  - JWTSecret, JWTRefreshSecret
  - ResendAPIKey, ResendFromEmail
  - DockerSocket, TraefikConfigDir
  - EncryptionMasterKey (32-byte hex)
  - CORSOrigins, AppBaseURL
  - IsProduction bool
  - SuperAdminEmail (first run bootstrap)

- [ ] **0.4 — Database setup**
  `internal/database/db.go`: GORM connection with pgx driver, connection pool (max 25 open, 5 idle, 5min lifetime). Auto-ping on startup. Fail fast if DB unreachable.

- [ ] **0.5 — Redis setup**
  `internal/redis/client.go`: Redis client with ping on startup.

- [ ] **0.6 — Migration system**
  `migrations/` directory with SQL migration files. First file: `000001_create_extensions.sql` (enable uuid-ossp extension). Create `internal/database/migrate.go` that runs migrations on startup using golang-migrate.

- [ ] **0.7 — Gin router skeleton**
  `internal/api/router.go`: Create Gin engine with:
  - zerolog request logger middleware
  - CORS middleware (gin-contrib/cors)
  - Request ID middleware
  - Recovery middleware
  - Health endpoint: `GET /health` → `{"status":"ok","version":"0.1.0"}`
  - Route groups: `/api/v1/auth`, `/api/v1/me`, `/api/v1/orgs`, `/api/v1/admin`, `/api/v1/webhooks`
  - Static file serving: embed `web/dist` and serve React SPA for all non-API routes

- [ ] **0.8 — React frontend scaffold**
  Inside `web/`:
  - `npm create vite@latest . -- --template react-ts`
  - Install: `tailwindcss`, `@shadcn/ui`, `zustand`, `@tanstack/react-query`, `react-router-dom`, `lucide-react`, `react-hook-form`, `zod`, `@hookform/resolvers`, `axios`, `recharts`
  - Setup Tailwind config
  - Initialize shadcn/ui
  - Create basic `App.tsx` with React Router
  - Create placeholder pages: Login, Dashboard
  - `vite.config.ts`: proxy `/api` to `http://localhost:8080` for dev

- [ ] **0.9 — Makefile**
  Targets: `make dev` (air hot-reload), `make build` (go build + vite build + embed), `make migrate`, `make migrate-down`, `make test`, `make lint`, `make docker-up`, `make docker-down`

- [ ] **0.10 — Docker development environment**
  `docker/docker-compose.dev.yml`: PostgreSQL 15 + Redis 7 + (optional) Traefik dev container. Mount volumes for data persistence.

- [ ] **0.11 — Embed directive**
  `cmd/server/main.go`: `//go:embed all:web/dist` static embed. Wire up config, DB, Redis, router, start HTTP server with graceful shutdown (SIGINT/SIGTERM → 30s drain).

---

## Phase 1: Authentication System

**Goal**: Users can register, login, logout, reset passwords. JWT is issued and validated. Sessions tracked.

### Tasks

- [ ] **1.1 — User model & migration**
  GORM model + SQL migration for `users` table:
  columns: id (uuid), email (unique), password_hash, name, avatar_url, is_super_admin, is_email_verified, totp_secret (nullable), totp_enabled, created_at, updated_at, deleted_at (soft delete)

- [ ] **1.2 — Sessions model & migration**
  `sessions` table: id, user_id, refresh_token_hash, device_info, ip_address, expires_at, created_at

- [ ] **1.3 — Password utilities**
  `internal/auth/password.go`: `HashPassword(plain string) (string, error)` using bcrypt cost 12. `CheckPassword(plain, hash string) bool`.

- [ ] **1.4 — JWT utilities**
  `internal/auth/jwt.go`:
  - `GenerateAccessToken(userID, email string, isSuperAdmin bool) (string, error)` — 15 min TTL, signed with JWTSecret
  - `GenerateRefreshToken(userID string, sessionID uuid.UUID) (string, error)` — 30 day TTL
  - `ValidateAccessToken(tokenString string) (*Claims, error)`
  - `ValidateRefreshToken(tokenString string) (*RefreshClaims, error)`
  - Claims struct includes: UserID, Email, IsSuperAdmin, OrgID (optional, set after org selection)

- [ ] **1.5 — Encryption utilities**
  `internal/auth/crypto.go`:
  - `Encrypt(plaintext string, key []byte) (string, error)` — AES-256-GCM, returns base64 ciphertext
  - `Decrypt(ciphertext string, key []byte) (string, error)`
  - `DeriveOrgKey(masterKey []byte, orgID uuid.UUID) []byte` — HKDF-SHA256

- [ ] **1.6 — Mailer service**
  `internal/mailer/mailer.go`: Resend client wrapper. Methods:
  - `SendEmailVerification(to, name, verifyURL string) error`
  - `SendPasswordReset(to, name, otp string) error`
  - `SendInvite(to, orgName, inviterName, acceptURL string) error`
  - `SendDeployNotification(to, appName, status, orgName string) error`
  - HTML email templates (inline CSS, readable without images)

- [ ] **1.7 — Auth repository**
  `internal/repository/user_repo.go`:
  - `CreateUser(user *models.User) error`
  - `FindUserByEmail(email string) (*models.User, error)`
  - `FindUserByID(id uuid.UUID) (*models.User, error)`
  - `UpdateUser(user *models.User) error`
  - `CreateSession(session *models.Session) error`
  - `FindSessionByTokenHash(hash string) (*models.Session, error)`
  - `DeleteSession(id uuid.UUID) error`
  - `DeleteUserSessions(userID uuid.UUID) error`

- [ ] **1.8 — Email verification**
  `email_verifications` table: id, user_id, token_hash, expires_at, used_at. Methods in repo and service.

- [ ] **1.9 — Password reset**
  `password_resets` table: id, user_id, otp_hash, expires_at (10 min), used_at. 6-digit OTP.

- [ ] **1.10 — Auth service**
  `internal/service/auth_service.go`:
  - `Register(email, password, name string) (*models.User, string, string, error)` — create user, send verify email, return access+refresh tokens
  - `Login(email, password, deviceInfo, ip string) (*models.User, string, string, error)`
  - `Logout(sessionID uuid.UUID) error`
  - `RefreshTokens(refreshToken string) (string, string, error)`
  - `ForgotPassword(email string) error` — generate OTP, send email
  - `ResetPassword(email, otp, newPassword string) error`
  - `VerifyEmail(token string) error`

- [ ] **1.11 — Auth handlers**
  `internal/api/handlers/auth_handler.go`: POST /auth/login, /register, /logout, /refresh, /forgot-password, /reset-password, /verify-email
  - Input validation with Gin binding + custom validators
  - Rate limit: 5 requests per 15 min per IP on login and forgot-password endpoints
  - Return access token in JSON body, refresh token in httpOnly cookie

- [ ] **1.12 — Auth middleware**
  `internal/middleware/auth.go`:
  - `RequireAuth()` — validates Bearer token from Authorization header, sets user in Gin context
  - `RequireSuperAdmin()` — additionally checks IsSuperAdmin
  - `OptionalAuth()` — parses token if present but doesn't block

- [ ] **1.13 — Me endpoints**
  GET /me (get profile), PUT /me (update name/avatar), GET /me/sessions, DELETE /me/sessions/:id, POST /me/change-password

- [ ] **1.14 — Frontend: Auth pages**
  React pages with React Hook Form + Zod validation:
  - `/login` — email/password form, "Forgot password?" link, submit → store access token in memory (Zustand), refresh token is in httpOnly cookie
  - `/register` — name/email/password, terms checkbox
  - `/forgot-password` — email input → OTP sent
  - `/reset-password` — OTP + new password
  - `/verify-email` — auto-verify from URL token
  - Protected route wrapper: redirect to /login if no valid token
  - Axios interceptor: auto-refresh access token on 401 using refresh cookie

- [ ] **1.15 — API key model & endpoints**
  `api_keys` table: id, user_id, org_id (nullable), name, key_hash, key_prefix (first 8 chars shown), scopes, last_used_at, expires_at.
  Endpoints: GET /me/api-keys, POST /me/api-keys (returns full key once), DELETE /me/api-keys/:id
  `ApiKeyAuth()` middleware for API key authentication (header: `Authorization: Bearer orb_...`)

---

## Phase 2: Organizations & Multi-Tenancy

**Goal**: Organizations exist, members can be invited, RBAC is enforced on all routes.

### Tasks

- [ ] **2.1 — Organization models & migrations**
  Tables: `organizations`, `org_members`, `org_invites`, `resource_plans`
  All columns as defined in project-description.md data models section.

- [ ] **2.2 — Org repository**
  Full CRUD for orgs, member management, invite management.

- [ ] **2.3 — Org service**
  `internal/service/org_service.go`:
  - `CreateOrganization(ownerID uuid.UUID, name, slug string) (*models.Organization, error)` — also creates owner membership, creates Docker network `orbita-org-{slug}`
  - `GetOrganization(orgSlug string) (*models.Organization, error)`
  - `UpdateOrganization(orgSlug string, updates map[string]interface{}) error`
  - `DeleteOrganization(orgSlug string) error` — teardown all resources, remove network
  - `ListUserOrgs(userID uuid.UUID) ([]models.Organization, error)`
  - `InviteMember(orgID uuid.UUID, email string, role string, inviterID uuid.UUID) error` — generate token, send email
  - `AcceptInvite(token string, userID uuid.UUID) error` — validate token, add member, mark used
  - `RevokeInvite(inviteID uuid.UUID, revokedBy uuid.UUID) error`
  - `UpdateMemberRole(orgID, targetUserID, requesterID uuid.UUID, newRole string) error`
  - `RemoveMember(orgID, targetUserID, requesterID uuid.UUID) error`
  - `GetMemberRole(orgID, userID uuid.UUID) (string, error)`

- [ ] **2.4 — Org middleware**
  `internal/middleware/org.go`:
  - `RequireOrgMember(minRole string)` — parses `:orgSlug` from URL, fetches org, checks user membership and role, sets org+role in context
  - `RequireOrgRole(role string)` — wrapper for specific role (e.g., "admin")
  - Helper `GetOrgFromContext(c *gin.Context) *models.Organization`
  - Helper `GetMemberRoleFromContext(c *gin.Context) string`

- [ ] **2.5 — Org handlers**
  All routes under `/api/v1/orgs`:
  - GET /orgs — list user's orgs
  - POST /orgs — create org (any authenticated user can create one org; super admin unlimited)
  - GET /orgs/:orgSlug — get org details + member count + resource usage summary
  - PUT /orgs/:orgSlug — update name/description (admin+)
  - DELETE /orgs/:orgSlug — delete org (owner only)
  - GET /orgs/:orgSlug/members — list members with roles
  - POST /orgs/:orgSlug/invites — send invite (admin+)
  - GET /orgs/:orgSlug/invites — list pending invites (admin+)
  - DELETE /orgs/:orgSlug/invites/:id — revoke invite (admin+)
  - PUT /orgs/:orgSlug/members/:userId/role — change role (owner only)
  - DELETE /orgs/:orgSlug/members/:userId — remove member (owner/admin)
  - POST /orgs/:orgSlug/leave — leave org (any member, owner cannot leave)
  - GET /api/v1/join — get invite info by token (public)
  - POST /api/v1/join — accept invite by token (authenticated)

- [ ] **2.6 — Resource plans (super admin)**
  `resource_plans` table. Admin endpoints to create/update/delete plans. `GET /admin/plans`, `POST /admin/plans`, etc.
  Endpoint to assign plan to org: `PUT /admin/orgs/:orgSlug/plan`

- [ ] **2.7 — Docker network management**
  `internal/docker/network.go`:
  - `CreateOrgNetwork(orgSlug string) error` — creates overlay network if Swarm mode, bridge otherwise
  - `DeleteOrgNetwork(orgSlug string) error`
  - `GetOrgNetworkName(orgSlug string) string` — returns `orbita-org-{slug}`

- [ ] **2.8 — Org tenant scoping utility**
  `internal/repository/scopes.go`: GORM scope function `OrgScope(orgID uuid.UUID)` that adds `WHERE organization_id = ?` to every query. Applied at repository layer — never optional.

- [ ] **2.9 — Frontend: Org management**
  - Organization switcher in sidebar (dropdown showing all user's orgs, "Create Organization" option)
  - `/orgs/new` — create org form (name auto-generates slug, editable)
  - `/orgs/:orgSlug/settings/members` — member list, invite form, pending invites, role dropdown, remove button
  - `/join` page — accept invite (shows org name, inviter, role; sign in or sign up form)
  - Org settings page: name, slug (read-only after creation), danger zone (delete org)
  - Role-based UI: hide action buttons based on current user's role

---

## Phase 3: Projects & Environments

**Goal**: Projects and environments exist and can be CRUD'd. Navigation is in place.

### Tasks

- [ ] **3.1 — Project & environment models + migrations**
  Tables: `projects`, `environments`

- [ ] **3.2 — Project/env repository & service**
  Standard CRUD. Service enforces org ownership. Env must belong to project which must belong to org.

- [ ] **3.3 — Project/env handlers**
  Full CRUD endpoints under `/api/v1/orgs/:orgSlug/projects` and `/api/v1/orgs/:orgSlug/projects/:projectId/environments`

- [ ] **3.4 — Frontend: Projects**
  - `/orgs/:orgSlug/projects` — project list (cards with emoji, name, resource count)
  - `/orgs/:orgSlug/projects/new` — create project form
  - `/orgs/:orgSlug/projects/:projectId` — project overview with environment tabs
  - Environment tab: resource counts per environment (apps, databases, services, cron jobs)
  - "Add environment" modal, rename/delete environment
  - Sidebar navigation: org → projects → project → environments

---

## Phase 4: Application Deployment (Core)

**Goal**: Apps can be created from Docker images, deployed, restarted, stopped, and deleted. Logs work.

### Tasks

- [ ] **4.1 — Application & deployment models + migrations**
  Tables: `applications`, `deployments`

- [ ] **4.2 — Docker service wrapper**
  `internal/docker/client.go`: Initialize Docker client from socket. Methods:
  - `PullImage(ctx, imageRef, registryAuth string) (io.ReadCloser, error)`
  - `CreateService(ctx context.Context, spec types.ServiceSpec) (string, error)` — Swarm service
  - `UpdateService(ctx, serviceID string, spec types.ServiceSpec) error`
  - `RemoveService(ctx, serviceID string) error`
  - `GetServiceLogs(ctx, serviceID string, tail int) (io.ReadCloser, error)`
  - `InspectService(ctx, serviceID string) (swarm.Service, error)`
  - `ListServiceTasks(ctx, serviceID string) ([]swarm.Task, error)`
  - `GetContainerStats(ctx, containerID string) (*types.StatsJSON, error)`
  - `ExecInContainer(ctx, containerID, cmd string) (io.ReadCloser, error)`

- [ ] **4.3 — Orchestrator**
  `internal/orchestrator/orchestrator.go`:
  - `DeployApplication(ctx context.Context, app *models.Application, deployment *models.Deployment) error`
    - Build or pull image
    - Create Docker Swarm service spec with resource limits from org plan
    - Apply cgroup parent `orbita-org-{slug}`
    - Attach to org network
    - Update Traefik config (call traefik package)
    - Update deployment status in DB
  - `StopApplication(ctx, app *models.Application) error`
  - `StartApplication(ctx, app *models.Application) error`
  - `RestartApplication(ctx, app *models.Application) error`
  - `RemoveApplication(ctx, app *models.Application) error`
  - `RollbackApplication(ctx, app *models.Application, deploymentID uuid.UUID) error`
  - `GetApplicationStatus(ctx, app *models.Application) (string, error)`

- [ ] **4.4 — App repository & service**
  Repository: CRUD with org scope.
  Service: business logic wrapping orchestrator calls, deployment record creation, status updates.

- [ ] **4.5 — App handlers**
  Routes under `/api/v1/orgs/:orgSlug/apps`:
  - GET /apps — list all apps in org (optionally filtered by env)
  - POST /apps — create app (source type: docker-image; git support in Phase 5)
  - GET /apps/:appId — get app detail with current status
  - PUT /apps/:appId — update app config
  - DELETE /apps/:appId — remove app and its containers
  - POST /apps/:appId/deploy — trigger deploy
  - POST /apps/:appId/restart — restart service
  - POST /apps/:appId/stop — scale to 0 replicas
  - POST /apps/:appId/start — scale back to configured replicas
  - GET /apps/:appId/deployments — deployment history
  - POST /apps/:appId/rollback/:deploymentId — rollback
  - GET /apps/:appId/status — live status check

- [ ] **4.6 — Real-time log streaming**
  `internal/websocket/hub.go`: WebSocket hub managing per-resource log rooms. When client connects to `/api/v1/orgs/:orgSlug/apps/:appId/logs`, authenticate WS handshake (token in query param), subscribe to log stream.
  
  `internal/websocket/log_streamer.go`: goroutine reading Docker service logs and fanning out to all subscribed clients for that app.

- [ ] **4.7 — Metrics collection**
  `internal/service/metrics_service.go`: Background goroutine collecting Docker stats per service every 5s, storing in Redis as time-series (`orbita:metrics:{serviceID}:{timestamp}`). Evict entries older than 24h.
  
  GET /apps/:appId/metrics endpoint returns last N data points.

- [ ] **4.8 — Frontend: Apps**
  - App list page: table/card view, status badge (running/stopped/deploying/error), quick actions
  - "New App" modal: select environment, enter name, select source type (Docker image for now), enter image, port, env vars
  - App detail page with tabs:
    - **Overview**: status, replicas, last deploy info, quick action buttons
    - **Deployments**: list of past deploys with status, commit hash, duration, rollback button
    - **Logs**: log viewer (real-time WebSocket), search input, auto-scroll toggle
    - **Metrics**: CPU/memory/network charts (Recharts LineChart), 1h/6h/24h selector
    - **Environment**: env var key-value editor, secret toggle, bulk import button
    - **Domains**: domain list, add domain form (Phase 8)
    - **Settings**: resource limits, restart policy, replicas, delete app

- [ ] **4.9 — Status badges and real-time updates**
  Polling every 10s for app status on list page. WebSocket notification when deploy status changes (push via notification hub).

---

## Phase 5: Git Integration & Auto-Deploy

**Goal**: Apps can deploy from GitHub/GitLab repos with auto-deploy on push.

### Tasks

- [ ] **5.1 — Git connection models + migrations**
  Tables: `git_connections`, add `source_config` JSON to applications (repo_url, branch, root_dir, build_method: dockerfile|nixpacks|buildpack, dockerfile_path, build_args, build_command, install_command, start_command)

- [ ] **5.2 — GitHub OAuth & App integration**
  - GitHub OAuth2 flow: GET /auth/github, GET /auth/github/callback
  - GitHub App installation flow (optional, for better rate limits)
  - Store encrypted access token in `git_connections`
  - GET /orgs/:orgSlug/git-connections — list connected providers
  - GET /orgs/:orgSlug/git-connections/:id/repos — list accessible repos
  - GET /orgs/:orgSlug/git-connections/:id/repos/:owner/:repo/branches — list branches

- [ ] **5.3 — GitLab OAuth integration**
  Same pattern as GitHub.

- [ ] **5.4 — Gitea integration**
  Personal access token (not OAuth). Store base URL + token.

- [ ] **5.5 — Build engine**
  `internal/orchestrator/builder.go`:
  - `BuildFromDockerfile(ctx, buildContext BuildContext) (imageRef string, logsCh chan string, err error)`
    - Clone repo to temp dir (`git clone --depth 1 --branch {branch}`)
    - Build with Docker BuildKit API
    - Stream build logs to channel
    - Push to local registry: `localhost:5000/orbita/{orgSlug}/{appName}:{commit-sha}`
    - Clean up temp dir
  - `BuildWithNixpacks(ctx, buildContext BuildContext) (imageRef string, logsCh chan string, err error)`
    - Run nixpacks CLI in Docker container
  - `BuildFromCompose(ctx, composeYAML string) error`

- [ ] **5.6 — Build log streaming**
  Build logs streamed real-time via WebSocket (same hub as runtime logs, different channel prefix `build:{deploymentID}`)

- [ ] **5.7 — Webhook handler**
  `internal/api/handlers/webhook_handler.go`:
  - `POST /webhooks/github` — verify HMAC-SHA256 signature, parse push event, find matching app by repo+branch, trigger deploy via queue
  - `POST /webhooks/gitlab` — verify X-Gitlab-Token, same flow
  - `POST /webhooks/gitea` — same pattern
  
  On app creation with git source: auto-register webhook on the repo via provider API.

- [ ] **5.8 — Internal Docker registry**
  Deploy Docker Registry v2 as a Swarm service on the primary node. Register it as a service in DB. Expose internally only (not to internet). All built images pushed here.

- [ ] **5.9 — Frontend: Git source type**
  - "New App" wizard: add "Git Repository" source type option
  - Step 1: choose provider (GitHub/GitLab/Gitea), connect if not already
  - Step 2: select repository, branch, root directory
  - Step 3: build method selection (auto-detect, Dockerfile, Nixpacks)
  - Build args input (key-value)
  - Auto-deploy toggle with webhook URL display

- [ ] **5.10 — Deploy queue**
  `internal/queue/deploy_worker.go`: Redis-backed job queue (simple list-based FIFO per org). Deploy jobs processed sequentially per org (no two deploys for same app simultaneously). Job payload: app_id, deployment_id, build_context.

- [ ] **5.11 — PR Preview environments (GitHub)**
  On PR open webhook: create new ephemeral environment named `pr-{number}`, deploy app there with unique subdomain `{appName}-pr-{number}.{baseDomain}`. On PR close/merge: destroy environment and all resources.

---

## Phase 6: Database Management

**Goal**: Managed databases can be provisioned, connected, backed up, and restored.

### Tasks

- [ ] **6.1 — Database models + migrations**
  Tables: `managed_databases`, `backup_schedules`, `backups`

- [ ] **6.2 — Database provisioner**
  `internal/orchestrator/db_provisioner.go`:
  - `ProvisionDatabase(ctx, db *models.ManagedDatabase) error`
    - Select Docker image for engine+version (postgres:15, mysql:8, mongo:7, redis:7, mariadb:11)
    - Generate strong random password (crypto/rand, 32 chars)
    - Create named volume `{orgSlug}_{dbName}_data`
    - Create Swarm service with resource limits, cgroup parent, org network
    - Store encrypted connection config in DB
  - `RemoveDatabase(ctx, db *models.ManagedDatabase) error`
  - `RestartDatabase(ctx, db *models.ManagedDatabase) error`
  - `GetDatabaseLogs(ctx, db *models.ManagedDatabase) (io.ReadCloser, error)`

- [ ] **6.3 — Backup engine**
  `internal/service/backup_service.go`:
  - `BackupPostgres(ctx, db *models.ManagedDatabase, dest BackupDestination) (*models.Backup, error)` — run `pg_dump` in sidecar container, pipe to gzip, optionally encrypt, upload to destination
  - `BackupMySQL(ctx, db *models.ManagedDatabase, dest BackupDestination) (*models.Backup, error)` — mysqldump
  - `BackupMongo(ctx, db *models.ManagedDatabase, dest BackupDestination) (*models.Backup, error)` — mongodump
  - `RestorePostgres(ctx, db *models.ManagedDatabase, backup *models.Backup) error`
  - `RestoreMySQL`, `RestoreMongo` similarly
  - Backup destinations: local filesystem, S3-compatible (AWS S3, Cloudflare R2, MinIO, Backblaze B2) using `aws-sdk-go-v2`, SFTP using `golang.org/x/crypto/ssh`

- [ ] **6.4 — Backup scheduler**
  On backup schedule create/update: register with robfig/cron. On schedule fire: enqueue backup job to queue worker.

- [ ] **6.5 — Database handlers**
  Routes under `/api/v1/orgs/:orgSlug/databases`:
  - GET /databases — list
  - POST /databases — provision (engine, version, name, resource limits)
  - GET /databases/:dbId — detail with connection info (masked password)
  - DELETE /databases/:dbId
  - POST /databases/:dbId/restart
  - POST /databases/:dbId/stop / /start
  - GET /databases/:dbId/logs (WebSocket)
  - GET /databases/:dbId/metrics
  - POST /databases/:dbId/backups — manual backup
  - GET /databases/:dbId/backups — list backups
  - POST /databases/:dbId/backups/:backupId/restore — restore
  - GET /databases/:dbId/backup-schedule
  - PUT /databases/:dbId/backup-schedule

- [ ] **6.6 — Frontend: Databases**
  - Database list: engine icon, name, status, storage used
  - "New Database" modal: engine selector (pg icon, mysql icon, etc.), version dropdown, name, resource limits
  - Database detail tabs:
    - **Connection**: masked DSN, copy button, reveal button, internal hostname
    - **Backups**: backup list with size/date, restore button, schedule settings, manual backup button
    - **Logs**: same log viewer component as apps
    - **Metrics**: connections, memory, ops/sec charts
    - **Settings**: resource limits, delete

---

## Phase 7: Cron Jobs

**Goal**: Cron jobs can be created, scheduled, triggered manually, and their run history viewed.

### Tasks

- [ ] **7.1 — Cron job models + migrations**
  Tables: `cron_jobs`, `cron_runs`

- [ ] **7.2 — Cron scheduler**
  `internal/cron/scheduler.go`:
  - On startup: load all enabled cron jobs from DB, register each with robfig/cron
  - `AddJob(cronJob *models.CronJob) error` — adds to live scheduler
  - `RemoveJob(cronJobID uuid.UUID)` — removes from live scheduler
  - `UpdateJob(cronJob *models.CronJob) error` — remove and re-add
  - `TriggerJob(cronJob *models.CronJob) error` — immediate execution (bypass schedule)
  - Each scheduled execution creates a `cron_run` record and calls `executeRun`

- [ ] **7.3 — Cron executor**
  `internal/cron/executor.go`:
  - `executeRun(ctx, cronJob, cronRun) error`
  - Pull image if needed
  - Check concurrency policy: if Forbid and previous run still active, mark as skipped; if Replace, stop previous container
  - Create Docker container (not service — run-to-completion): `docker run --rm --network orbita-org-{slug} --cgroup-parent orbita-org-{slug} ...`
  - Apply timeout context
  - Stream and capture logs (store first 100KB in `cron_runs.logs`)
  - Record exit code, duration, status
  - On failure: retry N times with 5s backoff
  - Send failure notification if configured

- [ ] **7.4 — Cron handlers**
  Routes under `/api/v1/orgs/:orgSlug/cron-jobs`:
  - GET /cron-jobs
  - POST /cron-jobs
  - GET /cron-jobs/:cronId
  - PUT /cron-jobs/:cronId
  - DELETE /cron-jobs/:cronId
  - POST /cron-jobs/:cronId/toggle (enable/disable)
  - POST /cron-jobs/:cronId/run (manual trigger)
  - GET /cron-jobs/:cronId/runs
  - GET /cron-jobs/:cronId/runs/:runId/logs

- [ ] **7.5 — Frontend: Cron Jobs**
  - Cron job list: name, schedule (human-readable), last run status badge, next run time, enable/disable toggle
  - "New Cron Job" form:
    - Name
    - Schedule input with visual helper (preset buttons: @daily, @hourly, @weekly + custom cron expression input + human-readable preview using `cronstrue` npm lib)
    - Source: Docker image or select existing app image
    - Command override
    - Environment variables (reuse EnvVarEditor component)
    - Timeout, retries, concurrency policy
    - Resource limits
  - Cron job detail page:
    - Run history table: started, duration, status, exit code, "View Logs" button
    - Logs modal: show captured log output for a specific run
    - "Run Now" button with confirmation
    - Edit settings
    - Next scheduled runs preview (show next 5 runs)

---

## Phase 8: Domains & Traefik Integration

**Goal**: Custom domains with automatic TLS work for all resource types.

### Tasks

- [ ] **8.1 — Domain models + migrations**
  Table: `domains`

- [ ] **8.2 — Traefik config manager**
  `internal/traefik/manager.go`:
  - Traefik watches a directory for dynamic config files (`TraefikConfigDir/dynamic/`)
  - `UpsertRoute(resource TraefikResource) error` — writes/updates JSON config file for a resource
  - `RemoveRoute(resourceID uuid.UUID) error` — deletes config file
  - Config file per resource: `{resourceID}.json` containing router, service, TLS config
  - Router name: `orbita-{resourceID}`
  - Service name: `orbita-svc-{resourceID}`
  - TLS: Let's Encrypt ACME resolver configured in Traefik static config
  - Middleware: HTTPS redirect, security headers per org

- [ ] **8.3 — Domain service**
  `internal/service/domain_service.go`:
  - `AddDomain(resourceID uuid.UUID, resourceType, domain string, orgID uuid.UUID, sslEnabled bool) error` — validate domain not used elsewhere, add to DB, update Traefik config
  - `RemoveDomain(domainID uuid.UUID) error` — remove from DB, remove Traefik config
  - `VerifyDomain(domain string) (bool, error)` — DNS lookup to check CNAME/A record points to server IP

- [ ] **8.4 — Domain handlers**
  Routes: GET, POST, DELETE under `/api/v1/orgs/:orgSlug/apps/:appId/domains` and equivalent for databases (TCP routing) and services.
  Also: GET `/api/v1/orgs/:orgSlug/domains` — list all domains in org.

- [ ] **8.5 — Traefik setup script**
  `docker/traefik.yml` static config: ACME Let's Encrypt, file provider watching `dynamic/` dir, Docker provider for automatic container discovery, dashboard (password protected by super admin credentials).

- [ ] **8.6 — Frontend: Domains**
  - Domain list per resource (in app/service settings)
  - "Add Domain" form: domain input, HTTPS toggle, DNS verification helper (show required DNS record)
  - Domain status: pending DNS check, active, error
  - Certificate status: issued, renewing, error
  - Remove domain with confirmation

---

## Phase 9: Services (One-Click Templates)

**Goal**: Users can deploy pre-configured services from a template marketplace.

### Tasks

- [ ] **9.1 — Template model + migration**
  Table: `templates`, `services` (deployed instances)

- [ ] **9.2 — Template seed data**
  Seed 20 templates as defined in project-description.md. Each template is a JSON document containing:
  - `name`, `description`, `category`, `icon_url`
  - `compose_template` — Docker Compose YAML with `{{.Param}}` placeholders
  - `params_schema` — JSON Schema describing configurable params (domain, admin email, storage size, etc.)

- [ ] **9.3 — Service deployer**
  `internal/orchestrator/service_deployer.go`:
  - `DeployService(ctx, service *models.Service, params map[string]string) error`
    - Render compose template with params (Go `text/template`)
    - Parse rendered compose file
    - Create all required volumes, networks (scoped to org)
    - Deploy all services as Swarm services with org resource limits
    - Register domains in Traefik

- [ ] **9.4 — Service handlers**
  Full CRUD under `/api/v1/orgs/:orgSlug/services`
  GET `/api/v1/templates` — list templates with categories

- [ ] **9.5 — Frontend: Services / Marketplace**
  - Template gallery: grid of cards with icon, name, description, category filter tabs
  - Template detail: description, parameters form (rendered from params_schema), one-click deploy
  - Deployed service detail: same tabs as app (logs, metrics, settings, domains)

---

## Phase 10: Resource Slicing & Node Management

**Goal**: Super admin can manage Swarm nodes and resource plans, with per-tenant isolation enforced.

### Tasks

- [ ] **10.1 — Node model + migration**
  Table: `nodes`

- [ ] **10.2 — Node manager**
  `internal/orchestrator/node_manager.go`:
  - `AddNode(ctx, name, ip string, sshPort int, sshPrivateKey string) (*models.Node, error)`
    - SSH into node, verify Docker is installed
    - Run `docker swarm join --token {workerToken} {primaryIP}:2377`
    - Label node in Swarm: `orbita.node.id={nodeID}`
    - Start Orbita agent process on node (or use existing Docker remote API)
  - `DrainNode(ctx, nodeID uuid.UUID) error` — Docker Swarm drain, reschedule services
  - `RemoveNode(ctx, nodeID uuid.UUID) error` — leave Swarm, delete from DB
  - `GetNodeMetrics(ctx, nodeID uuid.UUID) (*NodeMetrics, error)` — SSH exec for system metrics OR docker stats API

- [ ] **10.3 — Cgroup slice manager**
  `internal/orchestrator/cgroup_manager.go`:
  - `EnsureOrgSlice(orgSlug string, plan *models.ResourcePlan) error`
    - Create `/sys/fs/cgroup/orbita.slice/orbita-org-{slug}.slice/`
    - Write `memory.max` = plan.MaxRAMMB * 1024 * 1024
    - Write `cpu.weight` = plan.CPUWeight (proportional shares)
    - Write `cgroup.subtree_control` = "+cpu +memory"
  - `UpdateOrgSlice(orgSlug string, plan *models.ResourcePlan) error` — update limits in place
  - `RemoveOrgSlice(orgSlug string) error`
  - All Docker services for an org use `--cgroup-parent=orbita.slice/orbita-org-{slug}.slice`

- [ ] **10.4 — Node & plan admin handlers**
  Routes under `/api/v1/admin/`:
  - GET/POST/PUT/DELETE `/admin/nodes`
  - GET /admin/nodes/:nodeId/metrics
  - GET/POST/PUT/DELETE `/admin/plans`
  - PUT `/admin/orgs/:orgSlug/plan`
  - GET `/admin/orgs` — list all orgs with usage stats
  - GET `/admin/platform/metrics` — platform-wide resource usage

- [ ] **10.5 — Frontend: Nodes (super admin)**
  - Node list: hostname, IP, role (primary/worker), status (online/offline/draining), CPU/memory usage bars
  - "Add Node" form: name, IP, SSH port, SSH private key (paste)
  - Node detail: metrics dashboard, labels, drain/remove buttons
  - Plans management: create/edit resource plans with sliders for CPU cores, RAM GB, max apps, etc.

---

## Phase 11: Environment Variables & Secrets

**Goal**: Full env var and secret management with encryption and bulk operations.

### Tasks

- [ ] **11.1 — Env var model + migration**
  Table: `env_variables`

- [ ] **11.2 — Env var service**
  `internal/service/env_service.go`:
  - `SetEnvVar(resourceID uuid.UUID, resourceType, key, value string, isSecret bool, orgID uuid.UUID) error` — encrypt if secret using org-derived AES key
  - `GetEnvVars(resourceID uuid.UUID, resourceType string, orgID uuid.UUID) ([]models.EnvVariable, error)` — return with values (decrypted for secrets)
  - `GetEnvVarMap(resourceID uuid.UUID, resourceType string, orgID uuid.UUID) (map[string]string, error)` — for passing to Docker
  - `DeleteEnvVar(envVarID uuid.UUID, orgID uuid.UUID) error`
  - `BulkSetEnvVars(resourceID uuid.UUID, resourceType string, envVars map[string]string, orgID uuid.UUID) error` — upsert all
  - `ImportFromDotenv(content string) (map[string]string, error)` — parse .env format

- [ ] **11.3 — Secret redaction in logs**
  Log stream middleware that replaces secret values with `[REDACTED]` before sending to WebSocket clients.

- [ ] **11.4 — Frontend: Env var editor**
  Reusable `EnvVarEditor` component:
  - Table of key-value rows
  - Add row button
  - Secret toggle per row (eye icon → masked input)
  - Inline edit (click to edit)
  - Delete row
  - "Bulk Import" button → textarea for pasting .env content
  - "Export as .env" button (excludes secret values)
  - "Save Changes" button (batch update)
  - Diff indicator: show which vars are new/modified/deleted before saving

---

## Phase 12: Notifications & Audit Logs

**Goal**: In-app and email notifications work. Audit log records all actions.

### Tasks

- [ ] **12.1 — Notification & audit models + migrations**
  Tables: `notifications`, `notification_settings`, `audit_logs`

- [ ] **12.2 — Notification service**
  `internal/service/notification_service.go`:
  - `CreateNotification(userID, orgID uuid.UUID, notifType, title, body string) error` — insert + push via WebSocket
  - `SendDeployNotification(orgID uuid.UUID, app *models.Application, deployment *models.Deployment) error` — check settings, send in-app + email
  - `SendCronFailureNotification`, `SendBackupFailureNotification`, `SendQuotaAlertNotification`
  - `SendWebhookNotification(webhookURL string, payload interface{}) error` — POST JSON

- [ ] **12.3 — WebSocket notification push**
  Extend WS hub to support per-user notification channel. When notification created, push to all WS connections for that user.

- [ ] **12.4 — Audit logger**
  `internal/service/audit_service.go`:
  - `Log(ctx context.Context, action, resourceType string, resourceID uuid.UUID, metadata interface{}) error`
  - Called after every significant handler action (create, update, delete, deploy, rollback, member changes)
  - IP and user pulled from Gin context

- [ ] **12.5 — Notification & audit handlers**
  - GET /orgs/:orgSlug/notifications — list unread + recent
  - PUT /orgs/:orgSlug/notifications/:id/read
  - PUT /orgs/:orgSlug/notifications/read-all
  - GET /orgs/:orgSlug/notification-settings
  - PUT /orgs/:orgSlug/notification-settings
  - GET /orgs/:orgSlug/audit-logs — paginated, filterable

- [ ] **12.6 — Frontend: Notifications**
  - Bell icon in topbar with unread count badge
  - Notification dropdown: list of recent notifications, mark read, "View all" link
  - Notification settings page: toggle per event type, webhook URL input
  - Audit log page: table with timestamp, user, action, resource, IP; search + date filter

---

## Phase 13: SSH Terminal & Advanced App Features

**Goal**: In-browser terminal, Docker Compose apps, and advanced deployment features.

### Tasks

- [ ] **13.1 — SSH terminal WebSocket**
  `internal/websocket/terminal_handler.go`:
  - On WS connect: authenticate user + check Developer+ role for org
  - Find running container ID for app (via Docker API task lookup)
  - `docker exec -it {containerID} /bin/sh` (or /bin/bash if available)
  - Pipe stdin/stdout/stderr between WS and exec stream
  - Log in audit trail: "User X opened terminal in app Y"
  - Force close if user loses membership

- [ ] **13.2 — Frontend: Terminal**
  `xterm.js` + `@xterm/addon-fit` in a modal or full-screen page:
  - On open: WebSocket connection to `/api/v1/orgs/:orgSlug/apps/:appId/terminal?token={accessToken}`
  - Resize messages (`{type:"resize",cols:N,rows:M}`) sent on window resize
  - Terminal renders ANSI escape codes

- [ ] **13.3 — Docker Compose source type**
  Extend application creation: "Docker Compose" source type.
  - Input: paste or upload `docker-compose.yml`
  - Parse and show services list with deploy confirmation
  - Deploy all services in the compose file as individual Swarm services scoped to org
  - Link them into a shared internal stack network (still within org network)

- [ ] **13.4 — Blue-Green deploy**
  `internal/orchestrator/bluegreen.go`:
  - Maintain two Swarm services per app: `-blue` and `-green`
  - Deploy to inactive slot, run health checks, then switch Traefik to point to new slot
  - Old slot kept running for 10 min (configurable) then removed
  - Toggle in app settings to enable Blue-Green mode

- [ ] **13.5 — Shell exec (run-once commands)**
  POST /apps/:appId/exec — `{command: "python manage.py migrate"}` — creates a temporary container with same image/env, runs command, streams output, exits. Returns in response body (not WebSocket for simple commands).

- [ ] **13.6 — Health check configuration**
  In app settings: HTTP health check (path, port, expected status, interval, timeout, retries) or TCP check. Orbita polls and updates app status. Restarts container if health check fails N consecutive times.

---

## Phase 14: Dashboard & Monitoring Overview

**Goal**: The main dashboard gives a useful overview. Org-level metrics work.

### Tasks

- [ ] **14.1 — Dashboard data endpoints**
  - GET /orgs/:orgSlug/dashboard — summary: total apps (by status), databases, cron jobs, recent deploys, recent notifications, quota usage
  - GET /orgs/:orgSlug/metrics/overview — aggregated resource usage across all org containers

- [ ] **14.2 — Frontend: Dashboard**
  - Overview cards: # Running apps, # Databases, # Active cron jobs, # Failed deploys (last 24h)
  - Recent activity feed (last 10 deploys/events with status)
  - Resource usage bars: CPU% of quota, Memory% of quota, Disk% of quota
  - Recent notifications list (top 5 unread)
  - Quick action buttons: "Deploy New App", "New Database", "Add Service"

- [ ] **14.3 — Super admin dashboard**
  - Platform totals: # orgs, # nodes, # total apps/databases, platform CPU/memory aggregate
  - Node health grid: each node as a card with utilization bars
  - Org resource usage table: sortable by CPU/memory, with "over quota" highlight

- [ ] **14.4 — Frontend: Settings pages**
  - Org settings: name, description, danger zone
  - Profile settings: name, email, avatar, password change, 2FA setup, sessions list
  - Billing/Plan: current plan, usage vs limits, upgrade request button (link to super admin)
  - API Keys: list, create, revoke

---

## Phase 15: Polish, Security & Production Hardening

**Goal**: The system is production-ready with proper error handling, security hardening, rate limiting, and API docs.

### Tasks

- [ ] **15.1 — Rate limiting**
  Gin middleware using Redis sliding window: configurable per route group. Auth endpoints: 5/15min. API endpoints: 1000/min per API key. WebSocket: max 10 concurrent connections per user.

- [ ] **15.2 — Input validation**
  Audit all handlers: ensure every input is validated (Gin binding tags + custom validators). Sanitize strings for HTML (even though React handles output escaping, backend should not trust input).

- [ ] **15.3 — Error handling standardization**
  `internal/api/errors.go`: standard error response format `{"error": {"code": "RESOURCE_NOT_FOUND", "message": "..."}}`. All handlers use helper `RespondError(c, err)` which maps domain errors to HTTP status codes.

- [ ] **15.4 — TOTP / 2FA**
  `internal/auth/totp.go` using `github.com/pquerna/otp/totp`. Endpoints: POST /me/totp/enable (returns QR code), POST /me/totp/verify (confirms setup), DELETE /me/totp (disable). Require TOTP code on login if enabled.

- [ ] **15.5 — API documentation**
  Swagger/OpenAPI using `github.com/swaggo/gin-swagger` with annotations on all handlers. Served at `/api/docs`.

- [ ] **15.6 — Graceful shutdown**
  On SIGTERM: stop accepting new connections, wait for in-flight requests (30s max), stop cron scheduler, flush Redis queue, close Docker connections, close DB.

- [ ] **15.7 — Structured logging**
  All services log with zerolog: request ID, org ID, user ID, resource ID in every log entry. Log levels configurable (DEBUG/INFO/WARN/ERROR). Production mode: JSON logs. Dev mode: pretty logs.

- [ ] **15.8 — Frontend: 404 and error pages**
  Custom 404 (resource not found), 403 (no permission), 500 (server error) pages with helpful messages and navigation back to dashboard.

- [ ] **15.9 — First-run setup**
  If no users exist in DB, serve a `/setup` page (override routing):
  - Create super admin account (email + password)
  - Enter server IP / base domain
  - Enter Resend API key (optional, skip for local)
  - System creates Docker networks, ensures Traefik is running
  - Redirects to login

- [ ] **15.10 — Single binary production build**
  `make build`:
  1. `cd web && npm run build` → `web/dist/`
  2. `go build -ldflags="-s -w" -o orbita ./cmd/server/`
  Result: single `orbita` binary (~30MB) containing the entire application.
  
  Install script: `curl -fsSL https://get.orbita.sh | sh` — downloads binary, creates systemd service, runs setup.

- [ ] **15.11 — Docker image**
  Multi-stage Dockerfile: Node 20 for frontend build, Go 1.22 for backend build, distroless or alpine for runtime. Published to GHCR.

- [ ] **15.12 — E2E test coverage**
  Key happy paths tested:
  - User registers, creates org, invites member, member accepts
  - User creates app from Docker image, deploys, views logs, rolls back
  - User creates PostgreSQL database, takes backup, restores
  - User creates cron job, triggers manually, views run log
  - Super admin creates plan, assigns to org, verifies cgroup slice created

---

## Phase Order Summary

| Phase | Focus | Deliverable |
|---|---|---|
| 0 | Scaffold | Compiles, health endpoint works |
| 1 | Auth | Login/register/JWT/email verify |
| 2 | Multi-tenancy | Orgs, RBAC, invites fully working |
| 3 | Projects | Project/environment CRUD |
| 4 | Apps (Docker image) | Deploy, logs, metrics, rollback |
| 5 | Git + Auto-deploy | Build from git, webhooks, PR previews |
| 6 | Databases | Provision, backup, restore |
| 7 | Cron Jobs | Schedule, run history, manual trigger |
| 8 | Domains & SSL | Traefik integration, Let's Encrypt |
| 9 | Services | Template marketplace, one-click deploy |
| 10 | Nodes & Slicing | Swarm nodes, cgroup resource quotas |
| 11 | Env Vars & Secrets | Encryption, bulk import, redaction |
| 12 | Notifications & Audit | Real-time push, email alerts, audit log |
| 13 | Terminal & Compose | xterm.js SSH, Docker Compose source |
| 14 | Dashboard | Overview metrics, admin panel |
| 15 | Hardening | Rate limiting, 2FA, docs, single binary |
