# Grit Cloud — Quickstart (fresh VPS → live app in 2 commands)

This is the demo path. From a fresh Ubuntu 24.04 VPS to a live, migrated, HTTPS Grit app with
a database and dashboards — two commands, no hand-written Dockerfile, no clicking through a UI.

## Prerequisites

- A fresh **Ubuntu 24.04** VPS (Hetzner CX22 €4.5/mo is plenty) with a public IP.
- A domain you control. Point these A-records at the VPS IP (wait for DNS to resolve):
  - `orbita.example.com` — the Orbita control plane.
  - your app's domains, e.g. `api.rental.example.com`, `rental.example.com`.
- The `grit` CLI on your laptop (`make build-cli` → `./grit`, or install the Grit CLI).
- SSH access to the VPS (`ssh root@<ip>` works with your key).

## Command 1 — provision the host

```bash
grit cloud init \
  --server root@<vps-ip> \
  --name prod \
  --domain orbita.example.com \
  --acme-email you@example.com \
  --admin-email admin@example.com
```

This hardens the server (security score printed), installs Docker + Orbita + Traefik with
automatic Let's Encrypt, creates your super-admin + an `orb_` deploy token, and registers the
host in `~/.grit/hosts.yaml`. When it finishes, `https://orbita.example.com` is live.

Verify:
```bash
grit cloud status --host prod        # ● ok, version
```

## Command 2 — deploy your Grit app

From your Grit project directory (it has a `grit.json`):

```bash
grit cloud github-auth               # once: store a GitHub token (repo + admin:repo_hook)
grit deploy --host prod
```

If there's no `grit.yaml`, the first-run wizard creates one from the detected app shape. Then
`grit deploy`:
1. ensures the GitHub repo exists and pushes your code,
2. reconciles org/project/env, provisions Postgres/Redis/MinIO addons, injects env + secrets,
   sets domains,
3. builds each service from the Dockerfiles Grit already ships,
4. runs migrations under a Postgres advisory lock (aborting if they fail),
5. cuts over and prints your live links:

```
✔ Live
  App:       https://rental.example.com
  API:       https://api.rental.example.com
  Pulse:     https://api.rental.example.com/pulse/ui
  Sentinel:  https://api.rental.example.com/sentinel/ui
  Auto-deploy is on: future `git push` to main redeploys via webhook.
```

## After that

- **`git push`** to your branch auto-deploys (webhook registered on the app).
- `grit deploy --plan` — preview changes without applying.
- `grit logs -f --host prod` — stream logs.
- `grit rollback --host prod` — revert to the previous deploy.
- `grit cloud dashboard --host prod` — private SSH tunnel to the Orbita panel.

## Try it locally first

No VPS handy? `testdata/grit-sample/` is a runnable `api`-mode Grit app. See
`docs/PHASE4-E2E.md` for the fully local run (serve it over git, deploy through a local Orbita)
that exercises the entire build → migrate → route → live cycle.

## The `grit.yaml` contract

See `docs/GRIT-YAML.md` for the full spec, and
`docs/grit-cloud/grit-knowledge/08-grit-yaml-and-detection.md` for the ground truth. Minimal:

```yaml
app: rental
repo: MUKE-coder/rental
branch: main
addons: [postgres, redis]
domains:
  web: rental.example.com
  api: api.rental.example.com
migrate: true
env:
  from: .env.production
```
