---
layout: home

hero:
  name: Orbita
  text: One control plane for your whole stack.
  tagline: A self-hosted, multi-tenant PaaS in a single Go binary — and the control plane for Grit Cloud. Go from a bare VPS to a live, migrated, HTTPS app in two commands.
  image:
    src: /logo.svg
    alt: Orbita
  actions:
    - theme: brand
      text: Quickstart →
      link: /docs/quickstart
    - theme: alt
      text: What is Orbita?
      link: /docs/what-is-orbita
    - theme: alt
      text: GitHub
      link: https://github.com/MUKE-coder/orbita

features:
  - icon: 🏢
    title: True multi-tenancy
    details: Isolated Docker networks, per-org AES-256 keys, cgroup v2 CPU/RAM quotas, and 4-role RBAC. Run every client on one cheap box — they never see each other.
  - icon: ⚡
    title: Grit-aware, zero-config deploys
    details: A Grit app has a known shape, so Orbita reads it — builds each service from the Dockerfiles Grit ships, provisions Postgres/Redis/MinIO, wires domains, and migrates. No hand-written Dockerfile.
  - icon: 🔒
    title: Migrations under an advisory lock
    details: Deploys build, then run cmd/migrate under a Postgres advisory lock before cutover. A failed migration aborts the deploy — never a schema-mismatched cutover.
  - icon: 🌐
    title: HTTPS by default
    details: Traefik v3 with automatic Let's Encrypt, HTTP→HTTPS redirect, and per-app routing generated from grit.yaml. Nothing binds the public host but the proxy.
  - icon: 📈
    title: Observable from day one
    details: Pulse (latency/SQL/errors) and Sentinel (WAF/rate-limit/anomaly) mount on every Grit app by default. Live logs, metrics, and an in-browser terminal in the dashboard.
  - icon: 🪶
    title: One ~30 MB binary
    details: The Go control plane embeds the React dashboard and idles under 50 MB of RAM — leaving nearly all of your server for the apps you actually run.
---

<div style="max-width: 960px; margin: 4rem auto 0; padding: 0 24px;">

## From bare server to live app — two commands

```bash
# 1. Harden the VPS, install Orbita + Traefik, register the host
grit cloud init --server root@1.2.3.4 --domain orbita.example.com \
  --acme-email you@example.com --admin-email admin@example.com

# 2. From your Grit project: build → migrate → route → live
grit deploy --host prod
```

```ansi
✔ Live
  App:       https://rental.example.com
  API:       https://api.rental.example.com
  Pulse:     https://api.rental.example.com/pulse/ui
  Sentinel:  https://api.rental.example.com/sentinel/ui
  Auto-deploy is on: future `git push` to main redeploys via webhook.
```

After the first deploy, every `git push` auto-deploys via webhook — the CLI becomes optional.

</div>
