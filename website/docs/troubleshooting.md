# Troubleshooting

Ordered most-common first. Most installer issues are handled automatically; this is for the rest.

## Install fails: `denied` pulling the image

`error from registry: denied` pulling `ghcr.io/muke-coder/orbita`. The image is private or
unpublished for your fork. Build locally and point the installer at it:

```bash
cd /opt && git clone https://github.com/MUKE-coder/orbita.git orbita-src
cd orbita-src && docker build -t orbita:local .
ORBITA_IMAGE=orbita:local bash <(curl -sSL https://raw.githubusercontent.com/MUKE-coder/orbita/main/install.sh)
```

## `address already in use` on port 80 / 443

Something else holds the port — usually Apache2 (Contabo/DigitalOcean images) or nginx:

```bash
ss -ltnp 'sport = :80'
systemctl stop apache2 && systemctl disable apache2
apt purge -y apache2 apache2-utils apache2-bin apache2-data && apt autoremove -y
cd /opt/orbita && docker compose up -d
```

## SSL certificate never issues

The padlock stays broken / "Not secure". Check Traefik's ACME logs:

```bash
cd /opt/orbita && docker compose logs traefik --tail 50 | grep -iE "acme|certificate|error"
```

| Log symptom | Cause → Fix |
|---|---|
| `DNS problem` / `NXDOMAIN` | DNS not propagated. `dig orbita.example.com +short` must return the server IP. Wait, retry. |
| `HTTP 404` from challenge | **Cloudflare orange cloud** proxying the record. Set it to **DNS only** (gray) until the cert issues. |
| `connection refused` on :80 | Firewall. `ufw allow 80/tcp && ufw allow 443/tcp && ufw reload` |
| `rate limit exceeded` | Let's Encrypt limit (5 certs/week/domain). Wait an hour; avoid delete/re-add loops. |
| nothing ACME-related | Traefik isn't reading labels. Confirm `--providers.docker=true` in the compose `traefik` command. |

::: tip Built-in diagnosis
`GET /api/v1/orgs/:org/domains/verify?domain=app.example.com` returns resolved IPs,
Cloudflare-proxy detection, and copy-ready guidance for the exact failure mode.
:::

## `ERR_TOO_MANY_REDIRECTS` on the dashboard

Browser cache (try Incognito), or an old image before the SPA-serving fix:

```bash
cd /opt/orbita && docker compose pull orbita && docker compose up -d orbita
```

## Traefik: `client version 1.24 is too old`

Docker Engine 28+ with an old Traefik SDK. Fixed in the current template (`traefik:v3.6.14`):

```bash
sed -i 's|image: traefik:v3.0|image: traefik:v3.6.14|' docker-compose.yml
docker compose up -d --force-recreate traefik
```

## Containers restart in a loop

```bash
docker compose ps
docker compose logs <service> --tail 100
```

- **orbita** → DB connection failed (check `DATABASE_URL`) or a migration error.
- **postgres** → old volume with mismatched credentials. `docker compose down -v` wipes data, then reinstall.
- **traefik** → config error; re-generate the compose via the installer.

## Grit deploy: app crash-loops with `password authentication failed`

::: warning Stale managed-DB volume
A managed database was deleted and re-created with the same name, reusing a stale volume —
Postgres ignores `POSTGRES_PASSWORD` on a non-empty data dir, so the new connection string fails
SCRAM auth over TCP. Current Orbita removes the volume on delete and clears orphans on provision.
For an old orphan: `docker volume rm <orgslug>_<dbname>_data` after removing the DB service.
:::

## Grit deploy: migrations failed, deploy aborted

**By design** — Orbita won't cut over to a schema-mismatched image. Run `grit logs -f --host prod`
to see the migrator output, fix the migration, and redeploy. Common cause: the app's `go.sum`
isn't committed, so `go run ./cmd/migrate` can't resolve modules.

## Grit deploy: `WaitForServiceConverged: timed out`

The image built but the container won't stay up. Check its logs:

```bash
CID=$(docker ps -a --filter "label=orbita.org=<org>" --format '{{.Names}}' | grep -v db | head -1)
docker logs "$CID" --tail 20
```

Usually the app can't reach an addon (wrong `DATABASE_URL`/`REDIS_URL`) or crashes on a missing
required env var. Add it to the `env.from` file and redeploy.

## General diagnostics

```bash
cd /opt/orbita
docker compose ps                       # all 4 healthy?
docker compose logs orbita --tail 50    # backend
docker compose logs traefik --tail 50   # routing + TLS
ss -ltnp 'sport = :443'                 # Traefik bound?
curl -I http://localhost:8080/health    # backend up?
curl -I https://orbita.example.com      # reachable from outside?
```
