# 06 — Migrations (running them on deploy)

## How Grit migrates

Grit uses **GORM `AutoMigrate`** over a registered model list — not versioned SQL files.

- The registry: `models.Models()` returns every model struct; `models.Migrate(db)` runs
  `AutoMigrate` for each (creating tables, adding columns/indexes). Each `grit generate
  resource` adds its model here automatically.
- **It is idempotent and additive.** Re-running it only creates missing tables and adds
  missing columns — it never drops data. This is exactly what you want for a re-runnable
  deploy: `grit deploy` twice = migrate twice = same result.
- **The API server does NOT migrate on boot.** Migration is an **explicit, separate step**.
  (Confirmed: a freshly started server serves requests but will error on a not-yet-created
  table until migrate runs.)

## The entrypoint: `cmd/migrate`

Every Grit API ships a standalone program at **`apps/api/cmd/migrate/main.go`** (or
`cmd/migrate/main.go` in `single` mode). It:

1. Loads config from the **same environment the server uses** (`DATABASE_URL`, or the
   `POSTGRES_*` parts).
2. Connects to Postgres.
3. Runs `models.Migrate(db)`.
4. Exits 0 on success, non-zero on failure.

Flags: `--fresh` **drops all tables first** — ⚠️ **destructive, never use in a deploy.**

`grit migrate` (the CLI) simply runs `go run cmd/migrate/main.go` locally. In a container
deploy you don't want `go run` at runtime — instead **build a `migrate` binary** (below).

## Recommended deploy hook (Phase 2.6)

Run migrations **after build, before cutover**, under a **Postgres advisory lock**, and
**fail the deploy if they fail** (do not cut over to a schema-mismatched image):

1. **Produce a `migrate` binary.** The shipped API Dockerfile builds only `./cmd/server`. For
   Grit apps, extend the build so the same image also contains the migrator — add to the Go
   build stage:
   ```dockerfile
   RUN CGO_ENABLED=0 GOOS=linux go build -o /out/migrate ./cmd/migrate
   ```
   and copy `/out/migrate` into the runtime image next to `./server`. (This is the one small,
   safe augmentation to the shipped Dockerfile that Orbita's Grit build recipe should make.
   Alternatively run a one-off `golang:1.24-alpine` container over the repo that does
   `go run ./cmd/migrate` — heavier, but no Dockerfile change.)
2. **Take a Postgres advisory lock** so concurrent app instances / overlapping deploys don't
   race (`AutoMigrate` itself does not lock):
   ```sql
   SELECT pg_advisory_lock(hashtext('grit_migrate'));   -- or a fixed bigint key
   -- ... run the migrator ...
   SELECT pg_advisory_unlock(hashtext('grit_migrate'));
   ```
   You can wrap the migrate run in this from Orbita's deploy engine, or have the migrator
   itself acquire it — either is fine as long as exactly one migrate runs at a time per DB.
3. **Run the migrator once**, as a one-off container on the org's network with the production
   env (same `DATABASE_URL`/`POSTGRES_*` the API will use):
   ```
   docker run --rm --network <org-net> --env-file <app.env> <app-image> ./migrate
   ```
4. **Gate cutover on success.** Only start/route the new app version if migrate exited 0.
   Keep the previous image for rollback.

## When `grit.yaml` says `migrate: true` (the default)

Run the hook above on every deploy. When `migrate: false`, skip it (the user is managing
schema themselves). Default to **true** if the field is omitted.

## Rollback note

Because migrations are additive (`AutoMigrate` doesn't drop), a **code rollback** to the
previous image is safe against a newer schema (extra columns are ignored by older code). You
do **not** need to "migrate down" — Grit has no down-migrations. `grit rollback` just
re-points to the previous image; leave the schema as-is.

**Next:** [`07-pulse-sentinel-studio.md`](./07-pulse-sentinel-studio.md).
