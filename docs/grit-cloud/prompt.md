# Prompt — Building Orbita (for Claude Code)

You are the engineer building **Orbita**: an independent, general-purpose PaaS (a better
Dokploy) that deploys **any** containerised app — Laravel, Django, Next.js, static HTML, raw
Docker images, Compose, or anything Nixpacks can build — with **Grit apps as a first-class,
zero-config fast path**. Orbita ships **its own `orbita` CLI**. Read this file first, then
follow it exactly.

---

## Step 1 — Read these files in order, fully, before writing any code

0. **`ARCHITECTURE-CORRECTION.md`** — ⚠️ **READ FIRST. It outranks every other file here.**
   Orbita is independent and depends on **nothing** from Grit; Grit is a first-class citizen,
   never a prerequisite. There is **no `grit cloud` command** and there must never be one — it
   does not exist in the real Grit CLI (`grit cloud init` → *unknown command*), and shipping a
   second binary named `grit` shadows the framework CLI users already have installed. Wherever
   any file below says `grit cloud …`, `grit deploy`, `grit.yaml` or `~/.grit/`, the
   correction's `orbita …`, `orbita.yaml` and `~/.orbita/` win.
1. **`project-description.md`** — the vision, the deploy flow, the manifest contract, what already exists in Orbita, what we're adding, and the v1 definition of done. The north star **as corrected by `ARCHITECTURE-CORRECTION.md`** (the manifest is `orbita.yaml`; Orbita never depends on Grit).
2. **`project-phases.md`** — the ordered execution plan with checkboxes. This is your task list.
3. **`design-style-guide.md`** — the UI/UX rules for the Orbita dashboard and CLI output. Consult it for anything visual.
4. **`grit-knowledge/`** — the authoritative description of **what a Grit app is and how it is
   built, configured, migrated, and deployed**, written by the Grit framework maintainer. It
   is the ground truth for the whole "Grit-awareness" layer. **Read `grit-knowledge/README.md`
   now for orientation, and read the full folder before you start Phase 2** — you cannot build
   the Grit app type, build recipe, addon wiring, or migration hook correctly without it.

Do not skim. These files are the contract. If anything you're about to do contradicts them,
stop and re-read.

**Precedence when they disagree:** `ARCHITECTURE-CORRECTION.md` > `grit-knowledge/` >
everything else.
- On **CLI naming, positioning, the manifest, and Orbita's independence** →
  `ARCHITECTURE-CORRECTION.md` wins (the other files predate the decision and still say
  `grit cloud` / `grit.yaml`).
- On any **Grit-specific detail** (e.g. Phase 2.3 says "generate a Dockerfile" but Grit already
  ships one, or example paths differ from real Grit paths) → `grit-knowledge/` wins; it
  reflects the actual framework.

---

## Step 2 — Understand the ground truth

- **Orbita already exists and is ~80% complete.** It lives at `https://github.com/MUKE-coder/orbita`. Your **first task (Phase 0) is to clone it** and get it running locally. You are auditing and finishing a real codebase, **not building a PaaS from scratch.**
- **Never rewrite working code.** Audit → identify gaps → finish → test. Preserve Orbita's stack: Go 1.22 + Gin + GORM + Postgres + Redis + **Traefik** (not Caddy) + Docker Swarm; React 18 + Vite + Tailwind v4 + shadcn/ui.
- The four documents (`project-description.md`, `project-phases.md`, `design-style-guide.md`, and this `prompt.md`) live in the Orbita repo alongside its existing docs.

---

## Step 3 — How to work through the phases

- Work **strictly top to bottom** in `project-phases.md`. Do not start Phase N+1 until every task in Phase N is checked and committed.
- Complete **one task at a time.** For each task:
  1. State which task you're doing (e.g. "P2.3 — Grit build recipe generator").
  2. Make the change in small, focused edits.
  3. Run `make test` and `make lint`. Do not proceed on a red build.
  4. Commit with a message referencing the task: `P2.3: generate multi-stage Dockerfile for Grit apps`.
  5. **Edit `project-phases.md` and change that task's `[ ]` to `[x]`.** Commit that too.
- When a full phase is done, fill in the `_Phase N complete: _` line with the date and the final commit SHA.

---

## Step 4 — Rules of engagement

- **Test as you go.** Every finished flow must have a test or a documented manual verification. Phase 1 is explicitly about getting Orbita green before adding Grit features.
- **Idempotency.** Everything the CLI does against Orbita (create org/project/app/addons/domains) must be safe to run repeatedly. Re-running `grit deploy` should reconcile, not duplicate.
- **Ask before assuming on irreversible or ambiguous choices.** If a task is under-specified (a schema decision, an API contract, a destructive migration), pause and ask rather than guessing. Small, reversible choices: proceed and note them.
- **Respect the scope guardrails** in `project-description.md` §8 and the Deferred section of `project-phases.md`. If you find yourself reaching for Kubernetes, multi-node CLI deploys, a managed tier, or swapping Traefik for Caddy — stop. That's out of v1.
- **Follow the design guide** for any UI or CLI output. Dark-first, hairline borders, monospace for real values, status by color+label, alive deploy screen, teaching empty states.
- **Keep a running log.** Maintain `docs/PROGRESS.md` with a short dated entry per work session: what you finished, what's next, any decisions or blockers.

---

## Step 5 — Definition of done (what you're driving toward)

You are done with v1 when, on a fresh Ubuntu 24.04 Hetzner VPS:

1. `grit cloud init` → hardened server (security score ≥ 90) + Orbita on HTTPS + host registered in `~/.grit/hosts.yaml`.
2. `grit deploy --host prod` in a sample Grit repo (Go API + Next.js + Postgres) → live HTTPS URL, database provisioned, migrations applied — **no hand-written Dockerfile, no touching the Orbita UI.**
3. `git push` auto-deploys.
4. Pulse + Sentinel dashboards reachable for the app.
5. `grit rollback --host prod` reverts to the previous deploy.
6. The whole path is documented and reproducible.

---

## Start now

Begin with **Phase 0, task 1**: clone `https://github.com/MUKE-coder/orbita`, create the `grit-cloud` branch, and read the existing repo. Report what you find in `docs/ORBITA-AUDIT.md` before moving on. Work one task at a time, check the boxes as you go, and keep the build green.
