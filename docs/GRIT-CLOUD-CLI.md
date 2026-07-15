# grit cloud — CLI reference (Phase 3)

`grit cloud` turns a fresh VPS into a hardened, Grit-aware Orbita host and registers it in
`~/.grit/hosts.yaml` so `grit deploy --host <name>` (Phase 4) can ship to it.

Build: `make build-cli` → `./grit`. In production these are subcommands of the Grit framework
CLI; here they build as a standalone binary against the Orbita control plane, reusing Orbita's
`internal/grit` package for grit.yaml/grit.json logic.

## `grit cloud init` — provision a fresh VPS

```
grit cloud init \
  --server root@203.0.113.10 \
  --name prod \
  --domain orbita.example.com \
  --acme-email you@example.com \
  --admin-email admin@example.com \
  [--ssh-key ~/.ssh/id_ed25519] [--yes] [--skip-harden]
```

Over SSH, in order:
1. **Harden** — uploads the vendored `vps-harden.sh` and runs it `--no-dokploy --yes`
   (Orbita replaces Dokploy): SSH lockdown, UFW, Fail2ban, Docker-firewall fix, auto-updates,
   and prints a **security score out of 100**. Skip with `--skip-harden` if already hardened.
2. **Install** — installs Docker (if absent), inits Swarm, runs Orbita's `install.sh`
   non-interactively with `ORBITA_DOMAIN` + `ORBITA_ACME_EMAIL` (Orbita + Postgres + Redis +
   Traefik with Let's Encrypt).
3. **Bootstrap token** — waits for Orbita to answer over HTTPS, registers the first user
   (super admin; password from `GRIT_ADMIN_PASSWORD` or generated + printed once), and creates
   an `orb_` deploy key.
4. **Register** — writes the host to `~/.grit/hosts.yaml` (`api_url` + `orb_` token + `ssh`
   target), 0600 perms.

Result: `grit deploy --host prod` is ready.

### Required flags
`--server`, `--domain`, `--acme-email`, `--admin-email`. Missing ones are reported together.

### Env vars
- `GRIT_ADMIN_PASSWORD` — super-admin password (else a strong one is generated and printed once).
- `GRIT_HOSTS_FILE` — override the hosts file path (tests).

## `grit cloud status --host prod`

Hits Orbita's `/health` and, if the key is super-admin-scoped, platform metrics; prints a
readable summary (health, version, org/app/db/node counts).

## `grit cloud dashboard --host prod`

Opens an SSH tunnel to the Orbita dashboard for private access (`-L 8080:localhost:8080`),
using the `ssh` target stored at init. Flags: `--local-port`, `--remote-port`.

## `grit cloud hosts` / `grit cloud hosts remove <name>`

List or remove registered hosts. Multiple hosts are supported (prod/staging/…).

## `grit cloud github-auth`

Stores a GitHub token (`repo` + `admin:repo_hook`) in `~/.grit/github` (0600) for the Phase 4
repo-create/push flow.

---

## Manual verification on a fresh Ubuntu 24.04 VPS (P3.9)

Not runnable in CI (needs a real box + DNS). Steps to verify the definition of done:

```bash
# 0. Point orbita.example.com (A record) at the VPS IP; wait for DNS.
# 1. From your laptop:
grit cloud init --server root@<vps-ip> --name prod \
  --domain orbita.example.com --acme-email you@example.com --admin-email admin@example.com
# Expect: security score ≥ 90 printed by the harden step; Orbita reachable at
#         https://orbita.example.com; "Registered host prod" written to ~/.grit/hosts.yaml.
# 2. Verify:
grit cloud status --host prod          # health ● ok, version
grit cloud dashboard --host prod       # tunnel opens; visit http://localhost:8080
cat ~/.grit/hosts.yaml                 # prod → api_url + orb_ token (0600)
```

## What was verified automatically (this environment)

- CLI builds (`make build-cli`) and all subcommands wire up; help + flag validation.
- `~/.grit/hosts.yaml` round-trip (Set/Load/Resolve, multi-host, 0600 on Unix) — unit tested.
- SSH target parsing (`user@host:port` forms) — unit tested.
- **The exact bootstrap path init uses** (health → login → create `orb_` deploy key → the key
  authenticates) — verified live against a running Orbita.
- `grit cloud status --host local` against a live Orbita using an `orb_` key — health + version.

The SSH-driven harden/install steps run real remote commands (vendored `vps-harden.sh` +
Orbita `install.sh`); they require a target VPS and are covered by the manual steps above.
