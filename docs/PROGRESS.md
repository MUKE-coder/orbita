# Grit Cloud — Progress Log

Running log per work session. Newest entry first.

---

## 2026-07-15 — Session 2: Phase 1 (finish & harden Orbita) — COMPLETE

Phase 1 done end-to-end. Every gap from the audit's P0/P1 list is closed, each change
verified live against a running Orbita + Docker Swarm, and the build is green
(`go test ./... -race`). Tag `v0.1.0` created. Commits `085325b`…`541f372`.

**Deploy engine (P1.2):** env vars (secrets decrypted) injected into containers; Swarm
zero-downtime updates (start-first + rollback-on-failure) with convergence gating so a
bad image fails the deploy while the previous version keeps serving; Traefik routes
written/removed across the app lifecycle (per-domain files); Docker daemon pinged at
startup. *Verified: env in container; rolling update converges; bad tag → failed deploy,
old container still serving.*

**Git auto-deploy (P1.3):** repo_url derived so webhook matching works; git apps default
`auto_deploy=true` with an always-generated webhook secret (unsigned deliveries rejected);
webhook deploys carry the real org slug + `trigger_type=webhook`; public repos build with
no connection. *Verified: 401 unsigned / 401 bad sig / 200 signed → webhook deploy.*
Unit tests for HMAC verification.

**Databases (P1.4):** named volumes (data survives restart); DB ports no longer public;
`<NAME>_URL` injected into same-env apps; real backups (engine dump → gzip → BACKUP_DIR)
and restore; schedule runner with retention pruning. *Verified: pg16 provision → SMOKEDB_URL
in app → write → backup(566B) → drop → restore → row back → survives restart.*

**Secrets (P1.5):** crypto test suite (AES-256-GCM round-trip, nonce uniqueness, per-org
key isolation, tamper rejection); `ENCRYPTION_MASTER_KEY` now mandatory (≥32 chars);
install.sh generates 32 bytes.

**RBAC + API keys (P1.6):** `orb_` key auth wired onto org routes (was dead code) with
expiry + scope (read/deploy/admin) + org-binding enforcement; `HasMinRole` unknown-role
bug fixed. *Verified live matrix: read-key 200/403, deploy-key 200, bogus 401, viewer
deploy 403, developer invite 403, cross-org 403.* Unit tests for hierarchy + scopes.

**TLS (P1.7):** `/domains/verify` returns resolved IPs + Cloudflare-proxy detection +
per-failure guidance; `docs/TLS-TROUBLESHOOTING.md`.

**Observability (P1.8):** real WS log streaming (docker service logs), real terminal
(docker exec TTY + resize, self-authed + role-checked), real exec, real metrics (fixed
service-vs-container-ID bug), real cron executor (run-to-completion container, captured
logs/exit/retries). *Verified: streamed nginx logs; terminal echo; exec id→root; metrics
14MB; cron container printed output, exit 0.*

**Release (P1.9):** CI workflow (vet + test -race + lint), `.golangci.yml`, production
image builds (70MB) and boots, `v0.1.0` tag. GHCR push fires via existing `release.yml`
when the tag is pushed to origin (needs maintainer push access).

**Decisions/notes:**
- HTTP write timeout raised 15s→10m (synchronous git-build deploys were truncated).
- GitLab/Gitea webhooks, blue-green, multi-node, deploy-queue activation left deferred
  per project-description §8.
- `golangci-lint` not installed on this dev box; config added and CI runs it.

**Next:** Phase 2 — Grit-awareness. Read `grit-knowledge/` fully first; reuse Grit's
shipped Dockerfiles (don't generate). Reference app: `D:\LEARNING\Grit Framework\stoka`
(or `grit new sample --full`).

---

## 2026-07-15 — Session 1: Phase 0 bootstrap

**Done:**
- Cloned `MUKE-coder/orbita`, created `grit-cloud` branch.
- Read repo README.md, CLAUDE.md, and the repo's original planning docs end to end.
- Full codebase audit → `docs/ORBITA-AUDIT.md` (committed `7f71ea1`). Headline: core
  deploy/auth/multi-tenancy/RBAC/Traefik are real; env-var injection into containers,
  `orb_` API-key auth wiring, log streaming, terminal, cron execution, backups/restore,
  and metrics are stubbed or missing. Zero Go tests exist.
- Copied the Grit Cloud planning docs into `docs/grit-cloud/` (canonical copy for
  checkbox tracking — the repo root's same-named files are the original Orbita plan and
  were left untouched).

**Decisions:**
- Checkboxes are maintained in `docs/grit-cloud/project-phases.md` (committed with each
  task) since the workspace-root copies live outside the git repo.
- Local verification of UI deploy paths is API-driven (headless environment); flows are
  exercised through the same handlers/services the UI calls.

**Next:** stand up Orbita locally (P0.5), verify docker-image and git deploy paths
(P0.6–P0.7), run `make test` (P0.8), close Phase 0.

**Blockers:** none yet.
