# Deploy a Grit app

`grit deploy` takes local Grit code to a live, migrated, HTTPS app through Orbita — no
hand-written Dockerfile, no touching the dashboard.

## What happens, in order

1. **Resolve the host** — Orbita API URL + `orb_` token from `~/.grit/hosts.yaml`.
2. **Load `grit.yaml`** (+ `grit.json` + the `env.from` file), or run the **first-run wizard** to
   generate `grit.yaml` from the detected app shape.
3. **Ensure the GitHub repo** exists (create private if missing) and push your code.
4. **Reconcile** — org/project/environment, addons (Postgres/Redis/MinIO), env + secrets, domains
   — idempotent.
5. **Build → migrate → cut over** — build each service from the shipped Dockerfiles, run
   migrations under a Postgres advisory lock, then cut over. Previous images are kept for rollback.

## One-time setup

```bash
grit cloud github-auth      # store a GitHub token (repo + admin:repo_hook scope)
```

## Deploy

```bash
cd my-grit-app
grit deploy --host prod
```

Preview first with a dry run:

```bash
grit deploy --plan --host prod
```

```ansi
▸ Plan (dry run — nothing will be changed)
  App:       rental
  Mode:      triple
  Migrate:   true
  Addons:    postgres, redis

  create  rental-api    → api.rental.example.com
  create  rental-web    → rental.example.com
  create  rental-admin  → admin.rental.example.com
```

## The `grit.yaml` manifest

You write this once; everything else is inferred from `grit.json` + the repo. See the full
[grit.yaml spec](./grit-yaml).

```yaml
app: rental
repo: MUKE-coder/rental
branch: main
addons: [postgres, redis]
domains:
  web: rental.example.com
  admin: admin.rental.example.com
  api: api.rental.example.com
migrate: true
env:
  from: .env.production   # values encrypted into Orbita, never committed
```

::: tip What's derived (you don't write it)
Services, build contexts, ports, and the `NEXT_PUBLIC_API_URL` build arg all come from
`grit.json` + the mode. Orbita reuses the Dockerfiles Grit ships — it never generates one or
guesses with Nixpacks.
:::

## Migrations gate cutover

::: warning A failed migration aborts the deploy
Orbita runs `cmd/migrate` in a one-off container **under a Postgres advisory lock** before
cutover. If it exits non-zero, the deploy stops — you never cut over to a schema-mismatched
image. The previous version keeps serving.
:::

Common migration failure: the app's `go.sum` isn't committed, so `go run ./cmd/migrate` can't
resolve modules. Commit `go.sum` (real Grit apps ship it).

## After deploy

```bash
grit logs -f --host prod          # stream logs (WebSocket)
grit rollback --host prod         # revert to the previous deploy
grit cloud status --host prod     # health + platform metrics
```

Every future `git push` to your branch auto-deploys via webhook.

## Deploy a non-Grit app

Orbita also deploys any Docker image or Git repo through the dashboard:

::: code-group

```text [Docker image]
Apps → Create App → Source: Docker Image
  Image: nginx:alpine   Port: 80   → Deploy
```

```text [Git repository]
Settings → Git Connections → add a token (repo + admin:repo_hook)
Apps → Create App → Source: Git Repository
  Pick repo + branch → build via Dockerfile or Nixpacks → Deploy
  (every push auto-deploys via webhook)
```

:::

## Next

- [grit.yaml spec](./grit-yaml)
- [CLI reference](./cli)
- [Troubleshooting](./troubleshooting)
