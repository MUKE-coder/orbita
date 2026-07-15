# Grit Cloud — Progress Log

Running log per work session. Newest entry first.

---

## 2026-07-15 — Session 4: Phase 3 (grit cloud init installer) — COMPLETE

The `grit cloud` CLI now provisions a fresh VPS into a hardened Orbita host and registers it
locally. All 9 tasks done. Commit `778bab0`; Phase 3 complete: `778bab0`.

Built as a standalone Cobra binary at `cmd/grit/` inside the Orbita module so it reuses
`internal/grit` (Phase 2's schema/detection). `make build-cli` → `./grit`.

- **`grit cloud init`** (SSH-driven): harden (vendored `vps-harden.sh --no-dokploy`, preserves
  the 0–100 score) → install Docker+Swarm+Orbita (reuses `install.sh`) → wait healthy →
  register the first user (super admin) + create an `orb_` deploy key → write
  `~/.grit/hosts.yaml`. Required flags reported together; `--yes`/`--skip-harden`;
  `GRIT_ADMIN_PASSWORD`.
- **`grit cloud status/dashboard/hosts/github-auth`**: health+metrics summary; SSH `-L`
  dashboard tunnel; multi-host registry (0600); GitHub token store for Phase 4.
- Packages: `internal/orbita` (API client), `internal/sshx` (x/crypto/ssh exec+upload),
  `internal/hosts`, `internal/ui` (design-guide §9 colors/step timeline), `internal/assets`
  (embedded harden script).

**Verified:** CLI builds + help + flag validation; hosts round-trip + SSH-target parsing (unit);
`grit cloud status` against live Orbita with an `orb_` key; and **the exact init bootstrap path**
(health → login → create `orb_` key → the key authenticates) live against Orbita. The
SSH-driven harden/install run real remote commands — covered by the fresh-VPS manual steps in
`docs/GRIT-CLOUD-CLI.md` (P3.9; needs a real box + DNS).

**Next:** Phase 4 — `grit deploy`: read grit.yaml, ensure the GitHub repo, commit the build
recipe, push, then reconcile+deploy via the Orbita Grit API (Phase 2 endpoints), stream logs,
`--plan`, rollback. The Phase-2 reconcile/deploy engine + Phase-3 hosts/client/github-auth are
the foundation.

---

## 2026-07-15 — Session 3: Phase 2 (Grit-awareness) — COMPLETE

Orbita now understands a Grit app as a first-class type and can build/route/migrate it with
zero user configuration. All 9 tasks done, verified live end-to-end. Commits `00c683b`…`e352d83`.
Phase 2 complete: `e352d83`.

**Ground truth first:** read all of `grit-knowledge/` and cross-checked against the real
`stoka` app on disk — detection/build recipe match its `docker-compose.prod.yml` exactly.

- **P2.1 detection/schema** (`internal/grit`): parse grit.yaml + grit.json; derive the service
  map + build recipe from the architecture mode (single/api/double/triple, +docs); reject
  mobile; validate. `grit.json` + filesystem is the source of truth, not hand-written paths.
- **P2.2 Grit source type**: `source_type=grit`; each deployable service is one application row
  grouped by `grit_app`/`grit_role` (migration 000024); build args threaded into the git build.
- **P2.3 build recipe**: reuse the shipped Dockerfiles with the correct contexts (api →
  `apps/api`; web/admin/docs → repo root + own Dockerfile + `NEXT_PUBLIC_API_URL`); fallback
  generators match the shipped shapes.
- **P2.4–2.7 reconcile+deploy** (`GritService`): idempotent create-or-update of
  project/env/addons/apps/env/domains; postgres/redis via the managed-DB path + minio
  provisioner (`DATABASE_URL`/`REDIS_URL`/`MINIO_*` injected); migration hook runs
  `cmd/migrate` in a one-off container under a `pg_advisory_lock`, gating cutover; Pulse/Sentinel
  on by default with generated passwords (studio off).
- **P2.8 CLI API**: `/grit/plan|reconcile|deploy|:app/status|:app/rollback` (`orb_` key works);
  `docs/GRIT-API.md`. Fixed a pre-existing `EnvRepo.Upsert` duplicate-key bug exposed by
  re-reconcile.
- **P2.9 end-to-end**: built a real `api`-mode Grit sample (`testdata/grit-sample`), served it
  over local git, and drove the whole cycle through Orbita's API — reconcile → build from git
  context → migrate under advisory lock (created the `notes` table; a broken repo aborted
  cutover) → route → live: `/api/health` db-ok and a real `POST/GET /api/notes` round-trip
  against the injected `DATABASE_URL`. `docs/PHASE2-E2E.md`.

**Decisions/notes:**
- A Grit app fans out to one Orbita application per service — the Phase-1 deploy engine is
  reused unchanged (no rewrite).
- Migrator image is `golang:1.25-alpine` (covers Grit deps needing a newer toolchain than the
  1.24 the Dockerfiles pin).
- Real private `stoka-app` deploy is token-gated (needs `grit cloud github-auth`, Phase 4);
  detection/plan verified against its shape. Swarm ingress unreachable from Windows host
  loopback → liveness verified in-container (what Traefik/health use).

**Next:** Phase 3 — `grit cloud init` installer (Cobra subcommands, vps-harden, install Orbita,
`orb_` token bootstrap, `~/.grit/hosts.yaml`).

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
