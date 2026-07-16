# Install on a fresh server

The fastest path is `grit cloud init` (the [Quickstart](./quickstart)). This page is the
**manual, step-by-step install** for a fresh **Ubuntu 24.04** VPS, so you know exactly what
happens.

## 0. Prerequisites

- A fresh VPS with ports **80, 443, 8080** open. 2 vCPU / 4 GB RAM recommended.
- A domain with DNS pointed at the server:

| Type | Name | Value |
|---|---|---|
| A | `orbita.example.com` | `YOUR_SERVER_IP` |
| A | `*.example.com` (or per-app records) | `YOUR_SERVER_IP` |

::: warning Verify DNS first
`dig orbita.example.com +short` must return your server IP. Let's Encrypt cannot issue a
certificate until it does.
:::

## 1. Connect and update

```bash
ssh root@YOUR_SERVER_IP
apt update && apt upgrade -y
apt install -y curl git ufw
```

## 2. Firewall

```bash
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
ufw status
```

## 3. Docker + Swarm

```bash
curl -fsSL https://get.docker.com | sh
docker --version          # 24+ expected
docker swarm init         # required — Orbita runs workloads as Swarm services
```

## 4. Install Orbita

Pick one:

::: code-group

```bash [One-line (recommended)]
curl -sSL https://raw.githubusercontent.com/MUKE-coder/orbita/main/install.sh \
  | sudo ORBITA_DOMAIN=orbita.example.com ORBITA_ACME_EMAIL=you@example.com bash -s -- --yes
```

```bash [Docker Compose (manual)]
mkdir -p /opt/orbita && cd /opt/orbita
curl -sSL https://raw.githubusercontent.com/MUKE-coder/orbita/main/docker/docker-compose.prod.yml -o docker-compose.yml
cat > .env <<EOF
DB_PASSWORD=$(openssl rand -hex 16)
JWT_SECRET=$(openssl rand -hex 32)
JWT_REFRESH_SECRET=$(openssl rand -hex 32)
ENCRYPTION_MASTER_KEY=$(openssl rand -hex 32)
ORBITA_HOST=orbita.example.com
APP_BASE_URL=https://orbita.example.com
ACME_EMAIL=you@example.com
EOF
docker compose up -d
```

```bash [Build from source]
git clone https://github.com/MUKE-coder/orbita.git && cd orbita
docker build -t orbita:local .
ORBITA_IMAGE=orbita:local bash <(curl -sSL https://raw.githubusercontent.com/MUKE-coder/orbita/main/install.sh)
```

:::

The installer generates secrets, writes `docker-compose.yml` + `.env`, and starts **Orbita +
PostgreSQL + Redis + Traefik**. Traefik requests a Let's Encrypt certificate on first request.

::: danger ENCRYPTION_MASTER_KEY
Generate a **32-byte** key (`openssl rand -hex 32`). It derives every org's encryption key — a
weak or missing value silently weakens all tenant secrets. Orbita refuses to start without it.
:::

## 5. Verify

```bash
curl -s http://localhost:8080/health          # {"status":"ok","version":"0.1.0"}
cd /opt/orbita && docker compose ps            # orbita, postgres, redis, traefik → Up
curl -I https://orbita.example.com             # 200 from the outside world
```

Open `https://orbita.example.com`, **register the first user** (they become super-admin), and
create your first organization.

## 6. Configure DNS for apps

Point a wildcard (or per-app A-records) at the server so deployed apps get subdomains:

```bash
dig api.rental.example.com +short   # should return YOUR_SERVER_IP
```

## Managing the install

::: code-group

```bash [Update]
cd /opt/orbita
docker compose pull orbita
docker compose up -d orbita     # migrations run automatically on startup
```

```bash [Back up]
docker exec orbita-postgres pg_dump -U orbita orbita | gzip > orbita-$(date +%F).sql.gz
```

```bash [Stop (keep data)]
cd /opt/orbita && docker compose down
```

```bash [Full reset (destroys data)]
curl -sSL https://raw.githubusercontent.com/MUKE-coder/orbita/main/install.sh | sudo bash -s -- --force-reset
```

:::

## Next

- [Deploy a Grit app](./deploy)
- [Troubleshooting](./troubleshooting) if anything above didn't go clean.
