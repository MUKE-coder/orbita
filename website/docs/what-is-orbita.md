# What is Orbita?

**Orbita is an open-source, self-hosted Platform-as-a-Service (PaaS)** written in Go. It turns a
single VPS into a fully isolated hosting environment for many client organizations — each with
their own dashboard, projects, environment variables, secrets, logs, domains, databases, and
resource quotas. Clients never see each other; you, the super-admin, see everything.

It ships as a **single ~30 MB binary** with the React dashboard embedded, and idles under
**50 MB of RAM** — leaving nearly all of your server for the apps you actually run.

::: tip The one-liner
One VPS. Many clients. Full isolation. Deploy a full Grit app in two commands.
:::

## Two things in one

**1. A multi-tenant PaaS.** Deploy from a Docker image or a Git repo, with managed databases,
cron jobs, custom domains + automatic TLS, live logs, metrics, an in-browser terminal, and
per-organization resource quotas — all from one dashboard or the API.

**2. The control plane for Grit Cloud.** Because a [Grit](https://github.com/MUKE-coder) app has
a *known shape* (declared in `grit.json`), Orbita can build, route, migrate, and observe it with
**zero hand-configuration**. The `grit cloud` / `grit deploy` CLI sits on top and drives it.

## What makes it different

- **True multi-tenancy** — isolated Docker networks, per-org encryption keys, cgroup v2 quotas,
  and 4-role RBAC. This is the thing Dokploy and Coolify don't have.
- **Grit-awareness** — Orbita doesn't treat a Grit app as an opaque container. It reads the
  shape and wires the build, addons, domains, and migrations itself.
- **Tiny footprint** — a single Go binary, not a PHP/Node control plane.

## Who it's for

- **Freelancers & agencies** running many client apps on one €4.5/mo box with true isolation.
- **Solo Grit developers** who want the fastest path from `git init` to a live, migrated,
  observable app — self-hosted, so it's nearly free.

## Next

- [Why Orbita exists](./why) — the problem it solves.
- [Architecture](./architecture) — how it's built.
- [Quickstart](./quickstart) — a live app in two commands.
