# Grit Cloud — demo walkthrough (YouTube-ready)

The "secure a VPS and deploy a full-stack app in 2 commands" demo. This is the script; record
it against a real Hetzner box + the `stoka-app` repo.

## Setup (before recording)

- Fresh Hetzner CX22, Ubuntu 24.04, IP noted.
- DNS A-records pointing at the IP (resolved): `orbita.<domain>`, `stoka.<domain>`,
  `admin.stoka.<domain>`, `api.stoka.<domain>`.
- `grit` CLI built (`make build-cli`).
- A GitHub token ready (repo + admin:repo_hook).

## Beat 1 — "A fresh, empty server" (15s)

```bash
ssh root@<ip> 'uptime && docker --version || echo "no docker"'
```
Show: it's a blank box. Nothing installed.

## Beat 2 — Command 1: harden + install + register (2–3 min, fast-forward the build)

```bash
grit cloud init \
  --server root@<ip> \
  --name prod \
  --domain orbita.example.com \
  --acme-email you@example.com \
  --admin-email admin@example.com
```
Narrate the steps as they stream:
- **Hardening** — SSH lockdown, UFW, Fail2ban, auto-updates, **security score (aim ≥ 90)**.
- **Install** — Docker + Swarm + Orbita + Traefik, Let's Encrypt cert issued.
- **Bootstrap** — super-admin + `orb_` deploy token, host written to `~/.grit/hosts.yaml`.

Then:
```bash
grit cloud status --host prod          # ● ok
open https://orbita.example.com        # the dashboard is live with a valid cert
```

## Beat 3 — Command 2: deploy the Grit app (2–4 min, fast-forward the build)

```bash
cd stoka-app
grit cloud github-auth                 # paste token (once)
grit deploy --host prod --plan         # show the dry run: services, addons, domains, migrate
grit deploy --host prod
```
Narrate:
- ensure repo + push,
- reconcile (org/project/env, **Postgres/Redis provisioned**, env injected, domains set),
- **build** each service from Grit's shipped Dockerfiles,
- **migrate under an advisory lock** (gates cutover),
- cut over → **live links printed** (App, API, Pulse, Sentinel).

## Beat 4 — "It's real" (30s)

```bash
open https://stoka.example.com                 # the app
open https://api.stoka.example.com/api/health  # {"status":"ok","database":{"ok":true}}
open https://api.stoka.example.com/pulse/ui    # observability
open https://api.stoka.example.com/sentinel/ui # security
```

## Beat 5 — "And it keeps deploying" (30s)

```bash
# make a trivial change, commit, push
git commit -am "tweak" && git push
```
Show the webhook auto-deploy fire in `grit logs -f --host prod`. Then:
```bash
grit rollback --host prod              # instant revert to the previous deploy
```

## Close

Two commands: a hardened server and a live, migrated, observable, secured full-stack app —
self-hosted on a €4.5/mo box. That's Grit Cloud.

---

A fully local, no-VPS version of this cycle (for CI / a quick smoke test) is in
`docs/PHASE4-E2E.md`.
