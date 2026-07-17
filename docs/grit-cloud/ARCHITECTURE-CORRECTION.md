# Architecture correction — Orbita is an independent PaaS, not a Grit subcommand

> **From the maintainer. Read this before your next task.** It supersedes the CLI/naming
> parts of `project-description.md`, `project-phases.md` (Phases 3–5) and every doc that says
> `grit cloud …`. Where this file and those disagree, **this file wins**.
>
> The work you shipped is good — Phases 0–2 (audit, hardening, Grit-awareness) are keepers.
> The problem is **where the CLI lives and what it's called**, and that flows into the docs.

---

## 1. The bug that triggered this

A user followed the Quickstart on a real machine with the real Grit CLI installed:

```
$ grit cloud init
Error: unknown command "cloud" for "grit"
```

This is not a typo or a PATH issue. **It cannot ever work as built.** Here's why:

- The real Grit CLI is a different repo and module: `github.com/MUKE-coder/grit/v3`,
  a single `cmd/grit/main.go` (~1,950 lines) registering ~30 commands via
  `rootCmd.AddCommand(...)`. It ships as prebuilt cross-platform binaries from that repo's
  `release.yml`. **It has no `cloud` command and never imported anything from Orbita.**
- What you built is a **second, unrelated binary that is also called `grit`**, living at
  `orbita/cmd/grit/` in module `github.com/orbita-sh/orbita`. It defines `cloud`, `deploy`,
  `logs`, `rollback`.
- Your own `cmd/grit/main.go` says the quiet part out loud:

  > *"In production these `cloud` and `deploy` subcommands are part of the existing Grit
  > framework CLI; here they are built as a standalone binary against the Orbita control plane."*

  That "in production" merge **never happened and must not happen** (see §2).
- So `grit cloud init` only works if the user builds `./grit` from the Orbita repo
  (`make build-cli`) and puts it on PATH **shadowing their real Grit CLI**. Two different
  binaries named `grit` fighting over one name.
- `docs/QUICKSTART.md` then papers over it:
  > *"The `grit` CLI on your laptop (`make build-cli` → `./grit`, **or install the Grit CLI**)"*

  The "or install the Grit CLI" branch is **false**. That sentence is what sent the user
  down a dead end.

### Two things you could not have known

1. **Go's `internal/` rule makes the "merge later" plan impossible as written.**
   `github.com/orbita-sh/orbita/internal/grit` and `orbita/cmd/grit/internal/*` are
   importable **only** from within the `orbita` module. The Grit CLI could never import them.
   The plan required a refactor nobody had specified.
2. **The name `deploy` is already taken.** The real CLI has shipped `grit deploy`
   ("Deploy application to a remote server" — build, upload via SSH, systemd unit,
   optional Caddy + auto-TLS, driven by `--host`/`DEPLOY_HOST`) for many versions. It is a
   completely different mechanism from Orbita's control-plane deploy. Merging would have
   silently broken existing users.

**Good news: both problems disappear entirely under the correct architecture below.**

---

## 2. The correct positioning (this is the product decision)

**Orbita is an independent, general-purpose PaaS — a better Dokploy.** It deploys *any*
containerisable app: Laravel, Django, Rails, Next.js, a static HTML site, a raw Docker image,
a Compose file, or a repo built with Nixpacks. A Laravel developer must be able to adopt
Orbita **without ever hearing the word "Grit"**.

**Grit is a first-class citizen, not a prerequisite.** Because we also own Grit, Orbita
detects Grit apps and gives them the optimised path — reuse the shipped Dockerfiles, correct
build contexts, migrations under an advisory lock, Pulse/Sentinel wired automatically. That's
a *fast path for a recognised app type*, exactly like Nixpacks is a fast path for a Rails app.

**Dependency direction — memorise this:**

```
  Orbita  ──depends on──>  NOTHING from Grit.       (it must stand alone)
  Grit    ──optionally──>  Orbita.                  (grit new can offer Orbita deploys)
```

Orbita never imports the Grit module. Grit may, later and optionally, talk to Orbita — and
**that work happens in the Grit repo, by the Grit maintainer, not by you.**

### What this settles

| Question | Answer |
|---|---|
| Does `grit cloud` go into the Grit CLI? | **No. Never.** Delete the idea. |
| What is the CLI called? | **`orbita`** — its own binary, its own installer, its own docs. |
| The `grit deploy` clash? | **Gone.** Orbita has `orbita deploy`. Nothing collides. |
| The Go `internal/` blocker? | **Gone.** Nothing outside Orbita imports Orbita. `internal/grit` stays internal. |
| Who owns grit.json/grit.yaml parsing? | **Orbita**, in `internal/grit`. It must detect Grit apps standalone. |

---

## 3. Concrete changes

### 3.1 Rename the CLI: `grit cloud …` → `orbita …`

Move `cmd/grit/` → `cmd/orbita/`. The binary is `orbita`. Flatten the `cloud` group — with a
dedicated binary, `orbita cloud init` is redundant noise.

| Now (broken) | Correct |
|---|---|
| `grit cloud init` | `orbita init` |
| `grit cloud status --host prod` | `orbita status --host prod` |
| `grit cloud dashboard --host prod` | `orbita dashboard --host prod` |
| `grit cloud hosts` | `orbita hosts` |
| `grit cloud github-auth` | `orbita github-auth` |
| `grit deploy --host prod` | `orbita deploy --host prod` |
| `grit logs -f --host prod` | `orbita logs -f --host prod` |
| `grit rollback --host prod` | `orbita rollback --host prod` |

Keep every flag, behaviour and test. This is a rename + reparent, not a rewrite.

### 3.2 Decouple the config directory

`~/.grit/hosts.yaml` → **`~/.orbita/hosts.yaml`** (5 references). A Laravel user has no
`~/.grit` and never will. Orbita owns its own host registry.

### 3.3 Rename the deploy manifest: `grit.yaml` → `orbita.yaml`

~30 references. **A Laravel developer cannot be asked to write a file called `grit.yaml`.**
The manifest is Orbita's contract for *any* app:

```yaml
app: my-shop
repo: acme/my-shop
branch: main
# build: omit and Orbita detects (grit.json -> Grit fast path; Dockerfile; else Nixpacks)
addons: [postgres, redis]
domains:
  web: shop.example.com
env:
  from: .env.production
```

**`grit.json` detection stays exactly as built** — that file is Grit's own marker, emitted by
`grit new`, and reading it is precisely the Grit-awareness we want. Keep `internal/grit`,
keep the detection algorithm, keep the shipped-Dockerfile reuse. Only the *user-authored
manifest* is renamed. (v1.0.0 has no real users yet, so do a clean rename — no alias needed.)

### 3.4 Make the non-Grit paths real, not theoretical

You already have the source types (`docker-image`, `docker-compose`, `git`, `grit`) and a
builder. The positioning now demands each is genuinely tested:

- **Grit** → detect `grit.json`, reuse shipped Dockerfiles + contexts (`grit-knowledge/04`).
- **Dockerfile** → repo has a `Dockerfile`: build it, no guessing.
- **Nixpacks** → no Dockerfile: Laravel/Django/Rails/Node get built anyway. **This is the
  Dokploy-parity feature and it must actually work**, not just appear in a switch statement.
- **Docker image / Compose** → as today.

Detection order: `grit.json` → `Dockerfile` → `orbita.yaml` explicit build → Nixpacks.

### 3.5 Installer

Orbita needs its own one-liner, independent of Go, Grit, or a clone:

```bash
curl -sSL https://get.orbita.sh | sh          # or the GitHub raw URL you already use
```

plus `go install github.com/orbita-sh/orbita/cmd/orbita@latest` for Go users, and the
prebuilt cross-platform binaries from a release workflow. Whatever you choose, **the docs
must state exactly one supported way and it must be tested on a clean machine.**

---

## 4. Documentation rules (this is where it actually broke)

The code was defensible; the docs made a promise the code couldn't keep. Non-negotiable now:

1. **Every command must be copy-pasteable and real.** If a reader pastes it in order from a
   fresh box, it works. No `grit cloud init` when no such command exists.
2. **No "or …" escape hatches.** The line *"`make build-cli` → `./grit`, or install the Grit
   CLI"* is exactly the bug. One install path. If there are genuinely two, both get tested
   and are shown as explicit labelled tabs.
3. **Label every block with where it runs** — `Local terminal (your PC)` vs
   `Server terminal (SSH)`. The install page already does this well; do it everywhere.
4. **Test on the real server, top to bottom, in order.** The maintainer has a real box now.
   A step you have not executed is not documented — it's a guess. Paste real output.
5. **State the prerequisite explicitly**: Orbita requires **no Grit installation**. Say so on
   the Quickstart, because the whole point is that a Laravel dev can use this.
6. **Lead with the general case.** The Quickstart should deploy *any* app. Grit gets its own
   page ("Deploying Grit apps — zero config") as the showcase, not the entry requirement.
7. **`grit-knowledge/` stays authoritative** for what a Grit app is and how to build it.
   Nothing in it changes — it describes the framework, not the CLI.

---

## 5. Migration checklist

Do these in order, green build at each step:

- [ ] `git mv cmd/grit cmd/orbita`; module path `github.com/orbita-sh/orbita/cmd/orbita`.
- [ ] Flatten `cloud`: `cmd/cloud_init.go` → `init`, `cloud_status.go` → `status`, etc.
      Delete `cloud.go` (the group). Keep `deploy`/`logs`/`rollback` top-level.
- [ ] Delete the "in production these are part of the Grit CLI" comment in `main.go` — it is
      now wrong and will mislead the next reader.
- [ ] `~/.grit/hosts.yaml` → `~/.orbita/hosts.yaml` (`internal/hosts`).
- [ ] `grit.yaml` → `orbita.yaml` across schema, validator, CLI, testdata, docs (~30 refs).
      Keep `grit.json` detection untouched.
- [ ] `make build-cli` → builds `./orbita`. Update the Makefile target name.
- [ ] Prove Nixpacks: deploy a real Laravel **and** a static HTML repo end-to-end on the box.
- [ ] Rewrite Quickstart around `orbita init` + `orbita deploy` for a **non-Grit** app;
      add a separate Grit page. Execute every step on the real server and paste real output.
- [ ] Grep the whole repo + website for `grit cloud`, `grit deploy`, `grit.yaml`, `.grit/`.
      Zero hits outside the Grit-specific page and `grit-knowledge/`.
- [ ] Update `project-phases.md` §3–5 to match this file, and tick the boxes honestly.

---

## 6. Out of scope for you — do not do this

- **Do not add commands to the Grit CLI.** Do not open a PR against
  `github.com/MUKE-coder/grit`. Do not import the Grit module.
- **Do not build a `grit` binary.** The name belongs to the framework CLI. Shipping a second
  one shadows a tool people already have installed.
- The Grit-side integration ("when you create a new app it asks whether you'll deploy with
  Orbita") is **the Grit maintainer's task, in the Grit repo.** It will call Orbita's
  documented HTTP API or shell out to the `orbita` binary. Your job is to make that API and
  binary stable and documented — nothing more.

---

## 7. Why this is better (so it sticks)

- **A bigger market.** "Deploys any Docker app, and it's *magic* for Grit" beats "a deploy
  tool for one small framework."
- **No name collisions, no shadowed binaries, no Go `internal/` gymnastics, no breaking
  `grit deploy`.** Every hard problem in §1 evaporates.
- **Independent release cadence.** Orbita ships without waiting on a Grit release.
- **Grit still wins.** `grit new` can offer "deploy with Orbita?" and a Grit app still goes
  from code to live with zero hand-written Docker — the demo is unchanged. It's now a
  *showcase* of Orbita rather than its only reason to exist.
