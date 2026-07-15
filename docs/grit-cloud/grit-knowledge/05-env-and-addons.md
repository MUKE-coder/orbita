# 05 — Environment contract & addons

A Grit app reads a **fixed, known set of environment variables** (`.env` at the repo root;
`.env.example` documents them). Orbita injects these as encrypted secrets and — for the
addons it provisions — sets the connection URLs itself. **The user never pastes secrets into
a UI**; values come from their local `.env.production` referenced by `grit.yaml`
(`env.from`), plus the addon URLs Orbita fills in.

## The env contract (what the Go API reads)

### Database (Postgres) — required
```
POSTGRES_USER=grit
POSTGRES_PASSWORD=<generated>
POSTGRES_DB=<appname>
POSTGRES_HOST=localhost      # ← Orbita/compose overrides to the postgres SERVICE NAME
POSTGRES_PORT=5434           # ← dev host port; Orbita/compose overrides to 5432 (in-network)
# OR a single URL, which WINS over the parts above:
# DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require
# DATABASE_URL=sqlite:./app.db     # pure-Go SQLite (no CGO) — used for quick-start/tests
```
The Go config builds `DATABASE_URL` from the `POSTGRES_*` parts when `DATABASE_URL` is empty.
**When Orbita provisions the postgres addon, set `POSTGRES_HOST=<db-service>`,
`POSTGRES_PORT=5432`, and the user/password/db** (or set `DATABASE_URL` directly).

### Auth / JWT — required
```
JWT_SECRET=<random>          # generate a strong one per environment
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h
# Optional OAuth:
GOOGLE_CLIENT_ID=   GOOGLE_CLIENT_SECRET=
GITHUB_CLIENT_ID=   GITHUB_CLIENT_SECRET=
OAUTH_FRONTEND_URL=https://app.example.com    # where the OAuth callback redirects
```

### Redis — required (jobs, cache, cron)
```
REDIS_URL=redis://localhost:6380      # ← Orbita/compose overrides to redis://<redis-service>:6379
```

### URLs / CORS
```
API_URL=https://api.example.com               # the API's public URL
NEXT_PUBLIC_ADMIN_URL=https://admin.example.com
CORS_ORIGINS=https://example.com,https://admin.example.com   # comma-separated allowed origins
# NEXT_PUBLIC_API_URL is a BUILD ARG for the Next.js apps (see 04), not a runtime env.
```

### Object storage — driver-selected
```
STORAGE_DRIVER=minio          # minio | s3 | r2 | b2
# MinIO (local/self-hosted):
MINIO_ENDPOINT=http://localhost:9002   # ← compose overrides to http://<minio-service>:9000
MINIO_ACCESS_KEY=  MINIO_SECRET_KEY=  MINIO_BUCKET=<appname>-uploads  MINIO_REGION=us-east-1  MINIO_USE_SSL=false
# AWS S3:            S3_ENDPOINT= S3_ACCESS_KEY= S3_SECRET_KEY= S3_BUCKET= S3_REGION=
# Cloudflare R2:     R2_ENDPOINT= R2_ACCESS_KEY= R2_SECRET_KEY= R2_BUCKET= R2_REGION=auto
# Backblaze B2:      B2_ENDPOINT= B2_ACCESS_KEY= B2_SECRET_KEY= B2_BUCKET= B2_REGION=
```
Most production deploys use **R2 or S3** (`STORAGE_DRIVER=s3|r2`) and drop the minio addon.
When Orbita provisions the **minio** addon, set `STORAGE_DRIVER=minio` +
`MINIO_ENDPOINT=http://<minio-service>:9000` + keys + bucket.

### Email (Resend)
```
RESEND_API_KEY=re_...        MAIL_FROM=noreply@example.com     SUPPORT_EMAIL=
```

### Embedded dashboards (all served BY the Go API — see 07)
```
GORM_STUDIO_ENABLED=true  GORM_STUDIO_USERNAME=admin  GORM_STUDIO_PASSWORD=<set>
PULSE_ENABLED=true        PULSE_USERNAME=admin        PULSE_PASSWORD=<set>
SENTINEL_ENABLED=true     SENTINEL_USERNAME=admin     SENTINEL_PASSWORD=<set>
```

### AI (optional) + TOTP
```
AI_GATEWAY_API_KEY=          AI_GATEWAY_MODEL=anthropic/claude-sonnet-4-6   AI_GATEWAY_URL=https://ai-gateway.vercel.sh/v1
TOTP_ISSUER=<appname>
APP_ENV=production           # set by Orbita/compose in prod
PORT=8080                    # the API port (compose/Dockerfile default 8080)
```

## Addons → env mapping (what Orbita must set when it provisions each)

| `grit.yaml` addon | Orbita provisions | Env it must inject (in-network) |
|-------------------|-------------------|---------------------------------|
| `postgres` | postgres:16-alpine, encrypted creds, backups | `POSTGRES_HOST=<svc>`, `POSTGRES_PORT=5432`, `POSTGRES_USER/PASSWORD/DB` (or `DATABASE_URL`) |
| `redis` | redis:7-alpine | `REDIS_URL=redis://<svc>:6379` |
| `minio` | minio, bucket `<app>-uploads` | `STORAGE_DRIVER=minio`, `MINIO_ENDPOINT=http://<svc>:9000`, `MINIO_ACCESS_KEY/SECRET_KEY/BUCKET` |

This mirrors Orbita's existing managed-database URL-injection path — **reuse it**, don't
build a parallel one. If the addon list omits `minio` but the app uses storage, the user is
expected to supply `STORAGE_DRIVER=s3|r2` + keys via `env.from`.

## Precedence (important)

1. `docker-compose.prod.yml` `environment:` overrides (APP_ENV, POSTGRES_HOST/PORT, REDIS_URL,
   MINIO_ENDPOINT) — **highest**.
2. `.env` / injected secrets.
3. Built-in defaults in the Go config.

So Orbita's in-network overrides for the addon hosts always win over whatever the user left in
`.env` — which is exactly what you want.

**Next:** [`06-migrations.md`](./06-migrations.md) — running migrations on deploy.
