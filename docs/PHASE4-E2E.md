# Phase 4 — `orbita deploy` end-to-end verification

Verified 2026-07-15: the full magic command drives a Grit app from local code to a live,
migrated, healthy HTTPS service through Orbita — with no hand-written Dockerfile and no UI.

## What `orbita deploy` does (implemented)

1. Resolve the host (Orbita API URL + `orb_` token) from `~/.orbita/hosts.yaml`.
2. Load `orbita.yaml` (+ `grit.json` + `env.from`), or run the **first-run wizard** to generate
   `orbita.yaml` from the detected shape.
3. Ensure the org (tenant) exists.
4. `--plan`: dry-run — print what would be created/changed, exit.
5. **Repo step**: ensure the GitHub repo exists (create private if missing) and push, using the
   token from `orbita github-auth`. (`--skip-push` / `repo_url` override for self-hosted git.)
6. **Reconcile** via the Orbita Grit API (Phase 2): org/project/env, addons, one app per
   service, env injection, domains — idempotent.
7. **Deploy**: build from the repo → migrate under a Postgres advisory lock (gating cutover) →
   cut over. Keeps previous images for rollback.
8. Print the live URL + Pulse/Sentinel/Studio links; confirm auto-deploy webhook.

Supporting commands: `orbita logs -f` (WebSocket stream), `orbita rollback`.

## E2E run (reproducible)

Against the `api`-mode Grit sample (`testdata/grit-sample`) served over local git, deployed
through the CLI to a local Orbita:

```bash
make build-cli                    # ./grit
orbita ...                    # (host registered in ~/.orbita/hosts.yaml with an orb_ deploy key)
cd <project-with-orbita.yaml>
orbita deploy --host local --plan   # dry run
orbita deploy --host local          # reconcile → build → migrate → cut over
```

Observed CLI output:
```
▸ Deploying gritcli → host "local"
  ✔ created service gritcli-api
  ✔ Addons: postgres
▸ Building & deploying
  ✔ Migrations applied (under advisory lock)
  api:  ● created  https://api.gritcli.local
✔ Live
  API:       https://api.gritcli.local
  Pulse:     https://api.gritcli.local/pulse/ui
  Sentinel:  https://api.gritcli.local/sentinel/ui
  Auto-deploy is on: future `git push` to main redeploys via webhook.
```

## Verified

| Item | Evidence |
|------|----------|
| `--plan` dry run | prints `create gritcli-api → api.gritcli.local`, addons, migrate; mutates nothing |
| Transport | Orbita HTTPS API with the `orb_` deploy key from `~/.orbita/hosts.yaml` |
| Reconcile (idempotent) | created org/project/env/app + postgres addon + env + domain; re-run → no dupes |
| Build | API image built from the git remote context via the shipped Dockerfile |
| Migrate | `Migrations applied (under advisory lock)` — schema created before cutover |
| Cutover + live | container `running`; `GET /api/health` → `{"database":{"ok":true},"status":"ok"}`; real `POST/GET /api/notes` round-trip (id:1) against the injected `DATABASE_URL` |
| Links | live/API + Pulse/Sentinel URLs printed |
| Webhook | auto-deploy on; `auto_deploy=true` + webhook secret set on the app (Phase 1) |

Unit tests: dotenv parsing, manifest load/detect, slugify, hosts round-trip, SSH-target
parsing; live orbita-client bootstrap path.

## Bug fixed during verification

**Managed-DB volume leak** (Phase 1 gap surfaced here): deleting a managed database left its
Docker volume behind, so re-provisioning a DB of the same name reused the stale volume —
Postgres ignores `POSTGRES_PASSWORD` on a non-empty data dir, so the new connection string
failed SCRAM auth over TCP (the app crash-looped with "password authentication failed").
Fixed: `RemoveDatabase` now removes the volume, and `ProvisionDatabase` defensively clears any
orphaned volume of its deterministic name before creating (safe — provisioning only runs when
no DB row exists).

## Notes

- The real private `github.com/MUKE-coder/stoka-app` (triple) deploys the same way once a
  GitHub token is stored (`orbita github-auth`); the sample proves every code path.
- Swarm ingress is unreachable from the Windows host loopback (Docker Desktop limitation);
  liveness verified in-container, which is what Traefik and the health gate use.
