# CLI reference

The `grit cloud` and `grit deploy` commands. Build with `make build-cli` → `./grit`, or install
the Grit CLI.

## `grit cloud init`

Harden a fresh VPS, install Orbita, and register it locally.

```bash
grit cloud init \
  --server root@203.0.113.10 \
  --name prod \
  --domain orbita.example.com \
  --acme-email you@example.com \
  --admin-email admin@example.com \
  [--ssh-key ~/.ssh/id_ed25519] [--yes] [--skip-harden]
```

Over SSH: hardens (vendored `vps-harden.sh`, prints a 0–100 score) → installs Docker + Swarm +
Orbita → waits healthy → creates the super-admin + an `orb_` deploy key → writes
`~/.grit/hosts.yaml`.

| Flag | Meaning |
|---|---|
| `--server` | target VPS as `user@ip[:port]` (required) |
| `--domain` | Orbita dashboard domain (required) |
| `--acme-email` | Let's Encrypt email (required) |
| `--admin-email` | super-admin email for the first-run account (required) |
| `--ssh-key` | SSH private key (defaults to agent / `~/.ssh`) |
| `--yes` | non-interactive |
| `--skip-harden` | skip `vps-harden` (already hardened) |

::: tip Env vars
`GRIT_ADMIN_PASSWORD` sets the super-admin password (else a strong one is generated + printed
once). `GRIT_HOSTS_FILE` overrides the hosts file path.
:::

## `grit deploy`

Deploy a Grit app (build → migrate → route → live).

```bash
grit deploy --host prod
grit deploy --plan --host prod    # dry run — show changes, don't apply
grit deploy --skip-push --host prod   # skip the GitHub ensure/push step
```

| Flag | Meaning |
|---|---|
| `--host` | registered host name (default `prod`) |
| `--org` | org slug (defaults to the app name) |
| `--dir` | project directory (defaults to cwd) |
| `--plan` | dry run |
| `--skip-push` | skip GitHub ensure/push |

## `grit logs`

```bash
grit logs -f --host prod [--app rental] [--role api]
```

Streams a service's logs over WebSocket. `--role` picks which service (`api`/`web`/`admin`/`docs`).

## `grit rollback`

```bash
grit rollback --host prod [--app rental]
```

Reverts every service to its previous deploy. No "migrate down" — Grit migrations are additive,
so older code tolerates a newer schema.

## `grit cloud` helpers

::: code-group

```bash [status]
grit cloud status --host prod       # health + platform metrics
```

```bash [dashboard]
grit cloud dashboard --host prod    # private SSH tunnel to the Orbita panel
```

```bash [hosts]
grit cloud hosts                    # list registered hosts
grit cloud hosts remove staging     # remove one
```

```bash [github-auth]
grit cloud github-auth              # store a GitHub token for repo create/push
```

:::

## `~/.grit/hosts.yaml`

Written by `grit cloud init`; read by every other command. Holds tokens — `0600` perms.

```yaml
hosts:
  prod:
    api_url: https://orbita.example.com
    token: orb_...
    ssh: root@203.0.113.10
  staging:
    api_url: https://staging.example.com
    token: orb_...
```

## Auth for CI/CD

Create an `orb_` key (dashboard → API Keys, or the init flow) with scope `deploy` and use it as
the bearer token. `deploy` permits reconcile/deploy/rollback; `read` permits plan/status only.
