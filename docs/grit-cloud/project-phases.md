# Grit Cloud — Build Phases

This is the execution plan. Work **top to bottom**, one phase at a time. Do not start a phase until the previous one's tasks are all checked. Mark each task `[x]` the moment it is done and committed. When a whole phase is complete, add a one-line note under its heading with the date and commit SHA.

**Golden rules while executing:**
- Orbita already exists and is ~80% done. **Audit and finish — do not rewrite** working code.
- Keep the stack Orbita already uses: Go 1.22 + Gin + GORM + Postgres + Redis + **Traefik** (not Caddy) + Docker Swarm.
- Every change ships behind a green build. Run `make test` and `make lint` before checking off a task.
- Commit in small, described increments. Reference the phase/task in the message (e.g. `P2.3: grit app type in orbita`).

---

## Phase 0 — Bootstrap & orientation

> Goal: get the existing Orbita running locally and understand exactly what works and what's incomplete, before writing anything new.

- [x] Clone the existing repo: `git clone https://github.com/MUKE-coder/orbita.git && cd orbita`
- [x] Create a working branch: `git checkout -b grit-cloud`
- [x] Read `README.md`, `CLAUDE.md`, and the existing `project-description.md`, `project-phases.md`, `prompt.md` end to end.
- [x] Map the codebase: list every service in `internal/service/`, every handler group in `internal/api/handlers/`, and every model in `internal/models/`. Write findings to `docs/ORBITA-AUDIT.md`.
- [x] Stand it up locally: `docker compose -f docker/docker-compose.dev.yml up -d`, `cp .env.example .env`, fill secrets, `make migrate`, `make dev`, and `cd web && npm install && npm run dev`.
- [x] Register a super admin, create an org, and deploy `nginx:alpine` from a Docker image through the UI to confirm the core deploy path works.
- [x] Deploy a test app from a **GitHub repo** through the UI to confirm the Git + webhook path works. Note anything broken.
- [x] Run `make test` and record which tests pass/fail/are missing in `docs/ORBITA-AUDIT.md`.

_Phase 0 complete: 2026-07-15, commit 354d6b4_

---

## Phase 1 — Finish & harden Orbita

> Goal: a fully working, tested Orbita. Every core flow green before we add Grit-awareness.

- [x] From `docs/ORBITA-AUDIT.md`, create a checklist of incomplete/broken paths. Prioritize: deploy engine, Git+webhook auto-deploy, backup/restore, RBAC enforcement, TLS issuance.
- [x] **Deploy engine:** confirm zero-downtime rolling update + rollback works for image and Git sources; fix edge cases (failed build, health-check timeout, port mismatch).
- [x] **Git auto-deploy:** verify GitHub webhook signature (HMAC-SHA256) verification and that push→deploy fires reliably. Fix if broken.
- [x] **Databases:** verify Postgres provisioning, `${DB}_URL` injection, scheduled backup, and restore actually round-trip data. Add a test.
- [x] **RBAC:** write tests proving Viewer can't deploy, Developer can't manage members, cross-org access is denied. Fix any leak.
- [x] **Secrets:** verify AES-256-GCM encryption at rest and per-org key derivation; add a test that secrets are never returned in plaintext over the API.
- [x] **TLS:** reproduce and document the common Let's Encrypt failure modes (Cloudflare orange-cloud, DNS not propagated) with clear error messages surfaced to the user.
- [x] **API keys:** confirm `orb_` key creation, scoping, and bearer auth on all `/api/v1` routes used by the CLI (deploy, env, domains, logs, rollback).
- [x] Add missing tests until `make test` is green and covers the core deploy/db/rbac/secrets paths.
- [x] Publish a working container image to GHCR (`ghcr.io/muke-coder/orbita`) so `install.sh` works for self-hosters (the README notes this may currently be private/unpublished).
- [x] Tag a `v0.1.0` release.

_Phase 1 complete: 2026-07-15, commit 

---

## Phase 2 — Grit-awareness in Orbita

> Goal: Orbita understands a "Grit app" as a first-class type and can build/route/migrate it without user configuration.

> 📚 **Read `grit-knowledge/` before starting this phase — it is the ground truth for every
> task here.** Quick map: schema/detection → `08-grit-yaml-and-detection.md`; Grit app type &
> structure → `02`, `03`; **build recipe → `04` (REUSE Grit's shipped Dockerfiles + compose,
> don't generate)**; routing → `04` + `07`; addons → `05`; migrations on deploy → `06`;
> Pulse/Sentinel → `07`. Where a task below and `grit-knowledge/` disagree on a Grit detail,
> `grit-knowledge/` wins.

- [x] Define the `grit.yaml` schema (see `project-description.md` §5) as a Go struct + JSON schema. Add a parser + validator in `internal/grit/`.
- [x] Add a **Grit app source type** alongside the existing Docker-image and Git sources (extend the `applications` model + create/update handlers). Migration for any new columns.
- [x] **Build recipe generator:** given a `grit.yaml`, generate a correct multi-stage Dockerfile — Go build stage (`CGO_ENABLED=0`, ldflags `-s -w`) for the API/worker, Node build stage for the Next.js `web`, minimal runtime stage. Put in `internal/grit/build/`. Test against a sample Grit repo.
- [x] **Routing:** from `grit.yaml` `domains`, generate the right Traefik labels/routers for `web`, `api`, `admin`. Reuse Orbita's existing Traefik config writer.
- [x] **Addon provisioning:** when `grit.yaml` lists `postgres`/`redis`/`minio`, ensure Orbita provisions them in the org's isolated network and injects their URLs as env vars (reuse the managed-database path).
- [x] **Migrations on deploy:** add a deploy hook that runs the Grit migration tool after build, before cutover, holding a Postgres advisory lock (so concurrent app instances don't race). Fail the deploy if migrations fail; do not cut over.
- [x] **Pulse + Sentinel by default:** mount Pulse (observability) and Sentinel (security) for Grit apps automatically, exposing their dashboards behind the org's auth. Make it toggleable in `grit.yaml`.
- [x] Add API endpoints (or extend existing app endpoints) the CLI needs: create-or-update Grit app from `grit.yaml`, trigger deploy, fetch deploy status/logs, rollback. Document them.
- [x] End-to-end test: POST a `grit.yaml` + repo reference to the API and confirm a full build→migrate→route→live cycle on a test server.

_Phase 2 complete: 2026-07-15, commit e352d83_

---

## Phase 3 — The `grit cloud init` installer

> Goal: one command turns a fresh VPS into a hardened Orbita host and registers it locally.

- [x] Scaffold the `grit cloud` command group inside the existing Grit CLI (Cobra subcommands).
- [x] **Harden step:** vendor `vps-harden.sh` (from `github.com/MUKE-coder/vps-harden`) and invoke it — SSH lockdown, UFW, Fail2ban, Docker-firewall fix, auto-updates. Preserve the security-score-out-of-100 output.
- [x] **Install step:** install Docker (if absent), init Swarm, then run Orbita's existing `install.sh` (Orbita + Postgres + Redis + Traefik). Reuse, don't reimplement.
- [x] **Auth bootstrap:** after Orbita is up, create an initial `orb_` deploy API key (via the API once super-admin exists, or a first-run token) and capture it.
- [x] **Host registration:** write `~/.grit/hosts.yaml` on the operator's machine mapping a friendly name (e.g. `prod`) → Orbita API URL + `orb_` token. Support multiple hosts.
- [x] **Private dashboard helper:** add `grit cloud dashboard --host prod` that opens the SSH tunnel (`-L 3000:localhost:3000` style) so the Orbita panel is reached privately, matching the existing blog workflow.
- [x] `grit cloud status --host prod` — hit Orbita's `/health` + platform metrics and print a readable summary.
- [x] Flags for automation: `--yes`, `--domain`, `--acme-email` (pass through to Orbita's installer envs).
- [x] Test on a fresh Hetzner Ubuntu 24.04 box: one command → hardened server (score ≥ 90) + Orbita on HTTPS + registered host.

_Phase 3 complete: 2026-07-15, commit 778bab0_

---

## Phase 4 — The `grit deploy` command

> Goal: the magic command. Local Grit code → live, migrated, HTTPS app in one line.

- [x] `grit cloud github-auth` — store the user's GitHub token securely (OS keychain) for repo creation/push.
- [x] `grit deploy` reads `grit.yaml`; if absent, run a **first-run wizard** that detects the project shape and generates `grit.yaml` interactively, then continues.
- [x] **Repo step:** ensure the GitHub repo in `grit.yaml` exists (create via GitHub API/`gh` if missing), commit the generated Dockerfile/build recipe, and push the branch.
- [x] **Transport:** call Orbita's HTTPS API with the `orb_` token from `~/.grit/hosts.yaml`. (Primary transport; SSH is only for the dashboard helper.)
- [x] **Reconcile:** create-or-update org/project/environment → create-or-update the Grit app pointed at the repo+branch → inject env/secrets from the `env.from` file (encrypted) → set domains → provision missing addons. All idempotent.
- [x] **Deploy + migrate + cutover:** trigger the deploy; surface build logs live; run migrations under advisory lock; health-check; cut over; keep previous image.
- [x] **`grit deploy --plan`:** dry-run that prints every create/change (repo, app, addons, domains, migrations) without executing.
- [x] **`grit logs -f --host prod`:** stream app logs from Orbita's WebSocket log endpoint.
- [x] **`grit rollback --host prod`:** call Orbita's rollback endpoint to revert to the previous deploy.
- [x] On success, print the live URL + Pulse + Sentinel dashboard links.
- [x] Confirm the GitHub webhook is registered so future `git push` auto-deploys.
- [x] Full end-to-end test on the sample Grit repo against the Phase 3 host.

_Phase 4 complete: _____________________

---

## Phase 5 — Polish, docs, and the demo

> Goal: shippable, demoable, documented.

- [ ] Write a sample Grit app repo (Go API + Next.js + Postgres) used across tests and the demo.
- [ ] Error UX pass: every failure (bad token, DNS not propagated, migration failure, build failure) prints a clear cause + fix, matching the tone of the existing VPS-harden blog.
- [ ] Docs: `grit cloud` + `grit deploy` reference, the `grit.yaml` spec, and a "fresh VPS → live app in 2 commands" quickstart.
- [ ] Record the end-to-end demo (the YouTube-ready walkthrough).
- [ ] Update Orbita's `README.md` with the Grit Cloud section and cross-link from `jb.desishub.com`.
- [ ] Tag `grit-cloud v1.0.0`.

_Phase 5 complete: _____________________

---

## Deferred (explicitly NOT in v1 — see project-description §8)

- Kubernetes / non-Swarm orchestration.
- Multi-node deploys from the CLI.
- Managed/hosted tier (we run the compute).
- Per-PR preview environments, autoscaling.
- GitLab/Gitea in the CLI (Orbita UI already supports them).
- Swapping Traefik for Caddy.
