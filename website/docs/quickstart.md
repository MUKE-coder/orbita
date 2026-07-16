# Quickstart

From a fresh Ubuntu 24.04 VPS to a live, migrated, HTTPS Grit app — **two commands**.

## Prerequisites

- A fresh **Ubuntu 24.04** VPS with a public IP (Hetzner CX22 is plenty).
- A domain you control. Point A-records at the VPS IP and wait for DNS to resolve:
  `orbita.example.com`, plus your app's domains (e.g. `api.rental.example.com`, `rental.example.com`).
- The `grit` CLI on your laptop, and SSH access (`ssh root@<ip>` works with your key).

::: warning DNS must resolve first
`dig orbita.example.com +short` must return your server IP before you start — Let's Encrypt
can't issue a certificate until it does.
:::

## Command 1 — provision the host

```bash
grit cloud init \
  --server root@YOUR_SERVER_IP \
  --name prod \
  --domain orbita.example.com \
  --acme-email you@example.com \
  --admin-email admin@example.com
```

This hardens the server (security score printed), installs Docker + Orbita + Traefik with
automatic Let's Encrypt, creates your super-admin + an `orb_` deploy token, and registers the
host in `~/.grit/hosts.yaml`.

```bash
grit cloud status --host prod        # ● ok, version
```

## Command 2 — deploy your Grit app

From your Grit project directory (it has a `grit.json`):

```bash
grit cloud github-auth               # once: store a GitHub token (repo + admin:repo_hook)
grit deploy --host prod
```

If there's no `grit.yaml`, a first-run wizard creates one. Then `grit deploy` ensures the repo,
reconciles infrastructure, builds, migrates under an advisory lock, and cuts over:

```ansi
✔ Live
  App:       https://rental.example.com
  API:       https://api.rental.example.com
  Pulse:     https://api.rental.example.com/pulse/ui
  Sentinel:  https://api.rental.example.com/sentinel/ui
  Auto-deploy is on: future `git push` to main redeploys via webhook.
```

## After that

::: tip You're done — it keeps deploying
Every `git push` to your branch auto-deploys via webhook. The CLI is now optional.
:::

```bash
grit deploy --plan --host prod    # preview changes without applying
grit logs -f --host prod          # stream logs
grit rollback --host prod         # revert to the previous deploy
grit cloud dashboard --host prod  # private SSH tunnel to the Orbita panel
```

## Next

- [Install on a fresh server](./install) — the manual path, step by step.
- [Deploy a Grit app](./deploy) — the full deploy flow explained.
- [grit.yaml spec](./grit-yaml) — the deploy manifest.
