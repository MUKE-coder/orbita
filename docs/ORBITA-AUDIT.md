# Orbita Audit — Phase 0 (Grit Cloud)

Date: 2026-07-15 · Branch: `grit-cloud` · Base commit: `6db197e`

Purpose: map the existing codebase, verify what actually works vs. what is stubbed, and
produce the prioritized finish-list for Phase 1. Method: full read of README/CLAUDE/planning
docs, file-by-file inspection of `internal/`, grep for stubs/TODOs, plus local stand-up.

---

## 1. Codebase map

### Services (`internal/service/`) — 12 files
| Service | File | State |
|---|---|---|
| App (deploy/rollback/lifecycle) | `app_service.go` | REAL (metrics endpoint fake) |
| Auth (register/login/JWT/reset) | `auth_service.go` | REAL |
| Cron | `cron_service.go` | REAL CRUD; execution STUB |
| Managed databases | `db_service.go` | PARTIAL (backup/restore STUB) |
| Domains | `domain_service.go` | REAL |
| Env vars/secrets | `env_service.go` | REAL (injection into deploys MISSING) |
| Git adapter + service | `git_adapter.go`, `git_service.go` | REAL (GitHub); GitLab/Gitea partial |
| Notifications | `notification_service.go` | PARTIAL (webhook POST stub) |
| Orgs/members/invites | `org_service.go` | REAL |
| Projects/environments | `project_service.go` | REAL |
| Marketplace templates | `template_service.go` | PARTIAL (deploys first image only) |

### Handler groups (`internal/api/handlers/`) — 19 files
admin (+linux/other variants), app, auth, cron, dashboard, db, domain, env, exec (STUB),
git, me, node, notification, org, project, service, webhook.

### Models (`internal/models/`) — 17 files
api_key, application, cron_job, domain, email_verification, env_variable, git_connection,
health_check (**unused**), managed_database, node, notification, organization,
password_reset, project, service_template, session, user.

### Other layers
- `internal/repository/` — 13 repos + `scopes.go` (org scoping) — REAL.
- `internal/orchestrator/` — orchestrator.go REAL; bluegreen.go STUB; builder.go STUB
  (dead — real builds use Docker remote git context in orchestrator.go); cgroup_manager.go
  REAL (no-ops off Linux); db_provisioner.go PARTIAL; node_manager.go STUB.
- `internal/docker/` — real Docker SDK wrapper (client.go, network.go).
- `internal/traefik/manager.go` — real file-provider config writer.
- `internal/websocket/` — hub.go real but **unused**; terminal + log stream handlers STUB.
- `internal/queue/deploy_worker.go` — Redis queue implemented but **never started/used**.
- `internal/cron/` — scheduler REAL, executor STUB.
- Migrations: 23 numbered pairs (`000001`–`000023`), run at startup via golang-migrate.

### Toolchain notes (differ from README)
- go.mod: **Go 1.25** (README says 1.22+); local dev toolchain 1.26 works.
- Frontend: **React 19** + Vite 8 + Tailwind v4 (README says React 18; CLAUDE.md is correct).
- `web/package.json` name is `web-temp` — intentional, do not "fix".

---

## 2. What actually works (verified by code inspection)

- **Auth stack**: JWT 15m/30d, bcrypt cost 12, sessions, email verification, password
  reset OTP, rate limiting (Redis sliding window) — real.
- **Multi-tenancy**: org CRUD, member roles, signed invites (72h), per-org Docker network,
  HKDF-SHA256 per-org keys + AES-256-GCM env encryption at rest — real.
- **RBAC**: enforced via route groups in `router.go` (viewer < developer < admin < owner
  + super-admin axis). Grouping is sane (deploy = developer+, members = admin/owner).
- **Deploys**: docker-image source pulls + creates/updates Swarm service; git source builds
  via Docker remote-git build context (PAT embedded for private repos), tags
  `orbita-<org>-<app>:v<N>`; deployment history rows written; **rollback re-deploys the
  previous image ref** — real.
- **GitHub webhook**: HMAC-SHA256 verified (constant-time) when a secret exists; push →
  synchronous deploy — real.
- **Traefik**: per-resource JSON to file-provider dir, certResolver letsencrypt,
  HTTP→HTTPS redirect — real (but only invoked when a domain is added).
- **DB provisioning**: image pull + Swarm service + strong password + encrypted connection
  string, engines pg/mysql/mariadb/mongo/redis — real (volume caveat below).
- **install.sh / Dockerfile / docker-compose.prod.yml**: consistent
  (ghcr.io/muke-coder/orbita, Traefik v3.6.14, ACME HTTP-01).

## 3. Gaps found (prioritized for Phase 1)

### P0 — correctness/security gaps in "working" paths
1. **Env vars are never injected into deployed containers.** `DeployApplication` never
   sets `spec.EnvVars`; `EnvService.GetEnvVarMap` has no callers. The whole env/secrets
   UI is write-only theatre. (orchestrator.go / app_service.go)
2. **`orb_` API-key auth is dead code.** `ApiKeyAuth` middleware exists but is applied to
   zero routes; the CLI cannot authenticate. Keys can be created but never used.
   (middleware/auth.go:96, router.go)
3. **Webhook unsigned bypass**: signature only verified *if* `app.WebhookSecret` is set;
   otherwise pushes deploy unauthenticated. (webhook_handler.go:83)
4. **Managed-DB data is not persisted**: `VolumeName` recorded but no volume/mount is
   attached to the service spec — data lost on container restart. (db_provisioner.go)
5. **No `${NAME}_URL` injection** into apps (README claims it) — nothing wires managed-DB
   URLs into app env.
6. **No zero-downtime update config**: Swarm `UpdateConfig` (start-first) and container
   healthchecks are absent; updates are stop-first with Swarm defaults; `health_check.go`
   model unused. (docker/client.go buildSwarmServiceSpec)
7. **Traefik route not written on deploy** — only when a domain is added; apps are
   published via Swarm ingress mode instead.

### P1 — stubbed features presented as working
8. **WebSocket log streaming** emits fake "Application running normally" lines; terminal
   is an echo loop; hub.go never started. (websocket/)
9. **Cron executor** never runs a container; every run reports success. (cron/executor.go:90)
10. **Backups/restore**: backup marks itself completed (fake 1MB); restore is a no-op;
    schedules are stored but nothing fires them. (db_service.go:189–222)
11. **App metrics endpoint** returns hardcoded numbers; dashboard overview passes a
    *service* ID to `ContainerStatsOneShot` so real stats silently read 0.
12. **Exec endpoint** returns stub output. (exec_handler.go)
13. **GitLab/Gitea webhooks**: TODO stubs (out of CLI scope for v1, UI claims support).
14. **Node manager (multi-node)**: DB-only stubs + fake metrics (deferred in v1 anyway).
15. **Deploy queue**: implemented, never started; deploys are synchronous in-request.

### P2 — hardening/misc
16. `ENCRYPTION_MASTER_KEY` not required at startup (empty default silently weakens all
    org keys); install.sh seeds only 16 bytes hex.
17. Docker daemon reachability never pinged at startup — client silently degrades to a
    stub that errors on first use.
18. Blue-green deploy is a no-op falling back to rolling.
19. WS origin check `return true` (terminal_handler.go:17).
20. Rollback doesn't set `TriggeredBy`/timestamps as completely as Deploy.

## 4. Test status

**There are zero Go test files in the repository** (`find . -name '*_test.go'` → none).
`make test` (`go test ./... -v -race`) passes vacuously — every package reports "no test
files". `make lint` requires golangci-lint. Phase 1 must add tests for: deploy engine,
webhook signature, RBAC matrix, secrets round-trip, backup/restore, API-key auth.

`make test` run 2026-07-15: green (vacuous — no tests exist). See §6 for local stand-up.

## 5. Planning-doc deltas worth knowing

- Repo's own `project-description.md`/`project-phases.md`/`prompt.md` are the **original
  Orbita build plan** (checkboxes all unticked despite work done). The Grit Cloud docs at
  the workspace root supersede them for this effort.
- README overstates: `${DB}_URL` injection, real-time log streaming, in-browser terminal,
  backup/restore, metrics — all stubbed or missing (see §3).
- CLAUDE.md is accurate about layering, org-scoping, RBAC grouping, migrations discipline.

## 6. Local stand-up log (Phase 0) — verified 2026-07-15

Environment: Windows 11 + Docker Desktop 29.5.3 (Swarm initialized), Go 1.26.4, Node 26.
`make` and `golangci-lint` are not installed on this dev box — the Makefile targets were run
via their underlying commands (`go test ./... -race`, `go vet`, `go run ./cmd/migrate`).

What was done (API-driven; same handlers/services the UI calls — headless environment):
1. `docker compose -f docker/docker-compose.dev.yml up -d` → Postgres 15 + Redis 7 up.
2. `.env` from `.env.example` with generated secrets; `DOCKER_SOCKET=npipe:////./pipe/docker_engine`
   (client accepts any `scheme://` host — works on Windows), local `TRAEFIK_CONFIG_DIR`.
3. `go run ./cmd/migrate up` → all 23 migrations applied.
4. `web`: `npm install && npm run build` (embed requires `web/dist`), then server built+run →
   `/health` OK, SPA served.
5. **Register** first user → `is_super_admin=true` in JWT. **Org** `desishub` created (network,
   default plan assigned). **Project** auto-creates Production+Staging environments.
6. **Docker-image deploy** (`nginx:alpine`, port 80) → Swarm service `orbita-<id>` 1/1 running,
   content verified in-container. NOTE: published ingress port unreachable from host loopback —
   Docker Desktop/Windows Swarm limitation, not an Orbita bug.
7. **Git deploy** from public GitHub repo (`traefik/whoami`) → image built via Docker remote git
   context (`orbita-desishub-<id>:v1`), deployment `success`, service running.
8. **Webhook**: simulated GitHub push → deploy v2 triggered and succeeded — but only after
   manual DB fixes (see new findings below).
9. Auth rate limiting verified live (6th login → `RATE_LIMIT_EXCEEDED`).
10. `go test ./... -race`: **green but vacuous — every package "[no test files]"**. `go vet` clean.

### Additional gaps found during live verification (append to §3 P0 list)
21. **Git-source app creation requires `git_connection_id` even for public repos** — the
    orchestrator supports tokenless public clones but the handler rejects them. (app_service)
22. **Webhook matching can never fire for apps created via the API**: matcher is
    `source_config->>'repo_url' = <github clone_url>` (git_repo.go:54) but app creation stores
    `repo_url: ""` when only `repo_full_name` is given. Must derive/normalize repo_url.
23. **`auto_deploy` defaults to `false` and no API/UI field can enable it** — `UpdateAppRequest`
    only has name/port/replicas. Webhook auto-deploy is unreachable end-to-end without SQL.
24. **Webhook deploys pass empty org slug** — `Deploy(ctx, app.ID, app.OrganizationID, "", nil)`
    → image tagged `orbita--<id>:v2`, and network/cgroup names would derive from the empty slug
    on a first-time create. Also recorded as `trigger_type=manual`, not `webhook`.
25. **No webhook auto-registration on GitHub** at app creation (docs claim it; nothing calls the
    GitHub hooks API), and `webhook_secret` is never generated — so gap #3 (unsigned bypass) is
    the *default* state.
