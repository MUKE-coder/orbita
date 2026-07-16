# Why Orbita exists

Freelancers and agencies — Orbita's core audience — run many clients on one cheap box. The
existing self-hosted PaaS tools don't fit that, and managed platforms are expensive and opaque.

## The problems

| Problem | How Orbita solves it |
|---|---|
| **Dokploy / Coolify have no multi-tenancy** | True tenant isolation: per-org Docker networks, per-org encryption keys, cgroup v2 quotas, 4-role RBAC |
| **No resource caps per client** | cgroup v2 slices enforce CPU/memory limits per organization |
| **Heavy control-plane runtimes** | A single Go binary, <50 MB idle |
| **Generic, opaque deploys** — you still hand-write Dockerfiles, ports, env, migrations | **Grit-awareness**: Orbita reads the app's shape and wires build, addons, domains, and migrations itself |
| **Vendor lock-in / per-seat fees** | 100% self-hosted, you own the data, no per-seat pricing |

## How it compares

| | Orbita | Dokploy | Coolify |
|---|:---:|:---:|:---:|
| Multi-tenancy | ✅ | ❌ | ❌ |
| Resource quotas (cgroups) | ✅ | ❌ | ❌ |
| RBAC (4 roles) + invites | ✅ | ❌ | Partial |
| Cron manager | ✅ | ❌ | ❌ |
| Grit-aware zero-config deploys | ✅ | ❌ | ❌ |
| Single-binary control plane | ✅ | ❌ | ❌ |
| Idle memory | **<50 MB** | ~200 MB | ~500 MB |
| Written in | Go | Node.js | PHP/Laravel |

## The bigger picture: Grit Cloud

Orbita is the delivery vehicle that makes the whole Grit ecosystem cohere:

- **Grit framework** — the app being deployed; its known shape makes zero-config deploys possible.
- **Orbita** — the control plane that runs on the VPS and executes deploys.
- **Pulse** — built-in observability, mounted on every Grit app by default.
- **Sentinel** — built-in security (WAF/rate-limit/anomaly), mounted by default.
- **`grit cloud` / `grit deploy`** — the CLI that ties it together from the terminal.

::: tip The pitch
"Secure a VPS and deploy a full-stack app in two commands." That's the demo — and it's real.
:::

## Next

- [Architecture](./architecture) — how the pieces fit.
- [Quickstart](./quickstart) — try it.
