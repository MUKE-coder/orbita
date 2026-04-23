#!/bin/bash
set -euo pipefail

echo ""
echo "  ╔═══════════════════════════════════════╗"
echo "  ║          Orbita Installer              ║"
echo "  ║    Self-Hosted PaaS with Multi-Tenancy ║"
echo "  ╚═══════════════════════════════════════╝"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

INSTALL_DIR="/opt/orbita"
IMAGE="${ORBITA_IMAGE:-ghcr.io/muke-coder/orbita:latest}"
# ORBITA_AUTO_CLEAN=yes to skip confirmation prompts for destructive cleanup

# ---------- Parse arguments ----------

FORCE_RESET=false

usage() {
  cat <<HELP
Usage: install.sh [OPTIONS]

Options:
  --force-reset        Wipe all Orbita data (containers + volumes + .env) before
                       installing. DESTRUCTIVE — deletes your Postgres metadata,
                       encryption secrets, and TLS certs.
  --yes, -y            Skip all confirmation prompts (same as ORBITA_AUTO_CLEAN=yes).
  --help, -h           Show this help and exit.

Environment variables:
  ORBITA_DOMAIN        Non-interactive domain (e.g., orbita.example.com).
  ORBITA_ACME_EMAIL    Email for Let's Encrypt SSL.
  ORBITA_IMAGE         Override the Docker image
                       (default: ghcr.io/muke-coder/orbita:latest).
  ORBITA_AUTO_CLEAN    Set to "yes" to auto-approve cleanup of port conflicts
                       and stale containers.

Examples:
  sudo bash install.sh
  sudo bash install.sh --force-reset --yes
  ORBITA_DOMAIN=orbita.example.com ORBITA_ACME_EMAIL=me@example.com \\
    sudo -E bash install.sh --yes
HELP
}

while [ $# -gt 0 ]; do
  case "$1" in
    --force-reset) FORCE_RESET=true ;;
    --yes|-y)      export ORBITA_AUTO_CLEAN=yes ;;
    --help|-h)     usage; exit 0 ;;
    *) echo -e "${RED}Unknown option: $1${NC}" >&2; usage; exit 1 ;;
  esac
  shift
done

# ---------- Helpers ----------

confirm() {
  if [ "${ORBITA_AUTO_CLEAN:-}" = "yes" ]; then
    echo "    (auto-clean: proceeding)"
    return 0
  fi
  local prompt="$1"
  local reply
  read -rp "  $prompt [y/N]: " reply
  [[ "$reply" =~ ^[Yy]$ ]]
}

# Returns "<pid>|<binary>" for the process holding the given TCP port, or empty.
get_port_holder() {
  local port="$1"
  ss -ltnpH "sport = :$port" 2>/dev/null | awk '
    {
      for (i=1; i<=NF; i++) {
        if ($i ~ /^users:/) {
          match($i, /\(\("([^"]+)",pid=([0-9]+)/, a)
          if (a[1] != "") { print a[2]"|"a[1]; exit }
        }
      }
    }'
}

# If port is held by docker-proxy, print the container name that published it.
docker_container_on_port() {
  local port="$1"
  docker ps --filter "publish=$port" --format '{{.Names}}' 2>/dev/null | head -1
}

pkg_purge() {
  # $1 = service name as it appears to ss (apache2, nginx, ...)
  local svc="$1"
  if command -v apt-get >/dev/null 2>&1; then
    case "$svc" in
      apache2)  apt-get purge -y apache2 apache2-utils apache2-bin apache2-data 2>/dev/null || true ;;
      nginx)    apt-get purge -y nginx nginx-common nginx-core 2>/dev/null || true ;;
      caddy)    apt-get purge -y caddy 2>/dev/null || true ;;
      lighttpd) apt-get purge -y lighttpd 2>/dev/null || true ;;
      httpd)    apt-get purge -y apache2 2>/dev/null || true ;;
    esac
    apt-get autoremove -y 2>/dev/null || true
  elif command -v dnf >/dev/null 2>&1; then
    dnf remove -y "$svc" 2>/dev/null || true
  elif command -v yum >/dev/null 2>&1; then
    yum remove -y "$svc" 2>/dev/null || true
  else
    echo -e "${YELLOW}    Couldn't find apt/dnf/yum — please remove $svc manually.${NC}"
  fi
}

# Wipes ALL Orbita data: containers, volumes, .env, compose file. Destructive.
force_reset() {
  echo ""
  echo -e "${RED}══════════════════════════════════════════════════════════${NC}"
  echo -e "${RED}  DESTRUCTIVE: --force-reset will delete all Orbita data${NC}"
  echo -e "${RED}══════════════════════════════════════════════════════════${NC}"
  echo "  This will remove:"
  echo "    • All Orbita containers (orbita, postgres, redis, traefik)"
  echo "    • Named volumes: postgres_data, redis_data, traefik_*"
  echo "    • $INSTALL_DIR/.env (JWT + encryption secrets)"
  echo "    • $INSTALL_DIR/docker-compose.yml"
  echo ""
  echo -e "${YELLOW}  Apps you deployed through Orbita will keep running but become${NC}"
  echo -e "${YELLOW}  orphaned — the new Orbita install won't know about them.${NC}"
  echo ""
  if ! confirm "Proceed with full reset?"; then
    echo -e "${YELLOW}Reset cancelled.${NC}"
    exit 0
  fi

  echo -e "${YELLOW}  Stopping and removing containers...${NC}"
  if [ -f "$INSTALL_DIR/docker-compose.yml" ]; then
    (cd "$INSTALL_DIR" && docker compose down -v --remove-orphans 2>/dev/null) || true
  fi
  for n in orbita orbita-postgres orbita-redis orbita-traefik; do
    docker rm -f "$n" >/dev/null 2>&1 || true
  done

  echo -e "${YELLOW}  Removing named volumes...${NC}"
  for v in orbita_postgres_data orbita_redis_data orbita_traefik_dynamic orbita_traefik_certs; do
    docker volume rm "$v" >/dev/null 2>&1 || true
  done

  echo -e "${YELLOW}  Wiping config...${NC}"
  rm -f "$INSTALL_DIR/.env" "$INSTALL_DIR/docker-compose.yml"

  echo -e "${GREEN}  Reset complete. Continuing with fresh install.${NC}"
  echo ""
}

# Detects and offers to remove leftover Orbita containers from a previous/failed install.
cleanup_stale_orbita() {
  local names=("orbita" "orbita-postgres" "orbita-redis" "orbita-traefik")
  local existing=()
  for n in "${names[@]}"; do
    if docker ps -a --filter "name=^${n}$" --format '{{.Names}}' | grep -q "^${n}$"; then
      existing+=("$n")
    fi
  done

  if [ ${#existing[@]} -eq 0 ]; then
    return 0
  fi

  echo -e "${YELLOW}  Found existing Orbita containers from a previous run:${NC}"
  printf "    - %s\n" "${existing[@]}"
  echo "    (Named volumes with your data will be preserved.)"
  if confirm "Stop and remove these so the installer can start fresh?"; then
    if [ -f "$INSTALL_DIR/docker-compose.yml" ]; then
      (cd "$INSTALL_DIR" && docker compose down --remove-orphans 2>/dev/null) || true
    fi
    # Belt-and-braces: force-remove any named containers that survived
    for n in "${existing[@]}"; do
      docker rm -f "$n" >/dev/null 2>&1 || true
    done
    echo -e "${GREEN}  Cleaned up previous Orbita containers.${NC}"
  else
    echo -e "${RED}  Cannot continue with stale containers present. Aborting.${NC}"
    exit 1
  fi
}

# Checks a single port for conflicts and offers remediation.
check_port() {
  local port="$1"
  local purpose="$2"
  local info
  info=$(get_port_holder "$port")

  if [ -z "$info" ]; then
    return 0
  fi

  local pid="${info%|*}"
  local proc="${info#*|}"

  echo -e "${YELLOW}  Port $port ($purpose) is in use by '$proc' (pid $pid).${NC}"

  case "$proc" in
    docker-proxy)
      local container
      container=$(docker_container_on_port "$port")
      if [[ "$container" =~ ^(orbita|orbita-traefik|orbita-postgres|orbita-redis)$ ]]; then
        echo -e "${GREEN}    Held by existing Orbita container '$container' — compose will manage it.${NC}"
        return 0
      fi
      if [ -n "$container" ]; then
        echo "    Port is held by Docker container: $container"
        if confirm "  Remove container '$container' so Orbita can bind port $port?"; then
          docker rm -f "$container" >/dev/null
          echo -e "${GREEN}    Removed container $container${NC}"
        else
          echo -e "${RED}    Cannot continue with port $port in use. Aborting.${NC}"
          exit 1
        fi
      else
        echo -e "${RED}    Port held by docker-proxy but no container matched — please investigate. Aborting.${NC}"
        exit 1
      fi
      ;;
    apache2|httpd|nginx|caddy|lighttpd)
      echo "    This is a system-managed web server. It must be stopped and uninstalled"
      echo "    so Orbita's Traefik can bind ports 80 and 443."
      if confirm "Stop and PURGE '$proc' now?"; then
        systemctl stop "$proc" 2>/dev/null || service "$proc" stop 2>/dev/null || true
        systemctl disable "$proc" 2>/dev/null || true
        pkg_purge "$proc"
        echo -e "${GREEN}    Stopped and removed $proc${NC}"
      else
        echo -e "${RED}    Cannot continue with port $port in use. Aborting.${NC}"
        exit 1
      fi
      ;;
    *)
      echo "    Orbita cannot bind port $port while '$proc' (pid $pid) holds it."
      echo "    Please stop it manually and re-run the installer:"
      echo "      systemctl stop $proc   (if it's a systemd service)"
      echo "      kill $pid              (otherwise)"
      exit 1
      ;;
  esac
}

# ---------- Preflight ----------

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}Please run as root (sudo)${NC}"
  exit 1
fi

# Check OS
if [ ! -f /etc/os-release ]; then
  echo -e "${RED}Unsupported operating system${NC}"
  exit 1
fi

# ss is required for port detection
if ! command -v ss >/dev/null 2>&1; then
  if command -v apt-get >/dev/null 2>&1; then
    apt-get update -qq && apt-get install -y -qq iproute2
  fi
fi

echo -e "${GREEN}[1/7]${NC} Checking prerequisites..."

# Install Docker if not present
if ! command -v docker &> /dev/null; then
  echo -e "${YELLOW}  Docker not found. Installing...${NC}"
  curl -fsSL https://get.docker.com | sh
  systemctl enable docker
  systemctl start docker
  # Wait for the daemon to accept commands
  for _ in $(seq 1 20); do
    docker info >/dev/null 2>&1 && break
    sleep 1
  done
  echo -e "${GREEN}  Docker installed successfully${NC}"
else
  echo -e "${GREEN}  Docker is already installed${NC}"
fi

# Initialize Swarm if not already
if ! docker info 2>/dev/null | grep -q "Swarm: active"; then
  echo -e "${YELLOW}  Initializing Docker Swarm...${NC}"
  docker swarm init 2>/dev/null || true
fi

if [ "$FORCE_RESET" = true ]; then
  force_reset
fi

echo -e "${GREEN}[2/7]${NC} Checking for conflicts..."
cleanup_stale_orbita
check_port 80   "HTTP / Traefik"
check_port 443  "HTTPS / Traefik"
check_port 8080 "Orbita dashboard"
echo -e "${GREEN}  No blocking conflicts.${NC}"

echo -e "${GREEN}[3/7]${NC} Creating Orbita directory..."
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

echo -e "${GREEN}[4/7]${NC} Generating secrets..."

# Preserve existing secrets on re-run
if [ -f .env ]; then
  echo -e "${YELLOW}  Existing .env found — keeping existing secrets${NC}"
  # shellcheck disable=SC1091
  set -a; . ./.env; set +a
  DB_PASSWORD="${DB_PASSWORD:-$(openssl rand -hex 16)}"
  JWT_SECRET="${JWT_SECRET:-$(openssl rand -hex 32)}"
  JWT_REFRESH_SECRET="${JWT_REFRESH_SECRET:-$(openssl rand -hex 32)}"
  ENCRYPTION_MASTER_KEY="${ENCRYPTION_MASTER_KEY:-$(openssl rand -hex 16)}"
  DOMAIN_EXISTING="${ORBITA_HOST:-}"
  ACME_EMAIL_EXISTING="${ACME_EMAIL:-}"
else
  DB_PASSWORD=$(openssl rand -hex 16)
  JWT_SECRET=$(openssl rand -hex 32)
  JWT_REFRESH_SECRET=$(openssl rand -hex 32)
  ENCRYPTION_MASTER_KEY=$(openssl rand -hex 16)
  DOMAIN_EXISTING=""
  ACME_EMAIL_EXISTING=""
fi

echo -e "${GREEN}[5/7]${NC} Creating configuration..."

# Prompt for domain (allow non-interactive via env)
if [ -n "${ORBITA_DOMAIN:-}" ]; then
  DOMAIN="$ORBITA_DOMAIN"
elif [ -n "$DOMAIN_EXISTING" ]; then
  DOMAIN="$DOMAIN_EXISTING"
else
  read -rp "  Enter your domain (e.g., orbita.example.com) [localhost]: " DOMAIN
  DOMAIN=${DOMAIN:-localhost}
fi

if [ -n "${ORBITA_ACME_EMAIL:-}" ]; then
  ACME_EMAIL="$ORBITA_ACME_EMAIL"
elif [ -n "$ACME_EMAIL_EXISTING" ]; then
  ACME_EMAIL="$ACME_EMAIL_EXISTING"
else
  read -rp "  Enter email for Let's Encrypt SSL (e.g., admin@example.com) [admin@$DOMAIN]: " ACME_EMAIL
  ACME_EMAIL=${ACME_EMAIL:-admin@$DOMAIN}
fi

# Decide whether to enable TLS via Let's Encrypt
if [ "$DOMAIN" = "localhost" ] || [[ "$DOMAIN" =~ ^[0-9.]+$ ]]; then
  TLS_ENABLED=false
  APP_BASE_URL="http://$DOMAIN:8080"
else
  TLS_ENABLED=true
  APP_BASE_URL="https://$DOMAIN"
fi

# Create .env
cat > .env <<EOF
DB_PASSWORD=$DB_PASSWORD
JWT_SECRET=$JWT_SECRET
JWT_REFRESH_SECRET=$JWT_REFRESH_SECRET
ENCRYPTION_MASTER_KEY=$ENCRYPTION_MASTER_KEY
APP_BASE_URL=$APP_BASE_URL
ORBITA_HOST=$DOMAIN
RESEND_API_KEY=re_placeholder
RESEND_FROM_EMAIL=noreply@$DOMAIN
ACME_EMAIL=$ACME_EMAIL
ORBITA_IMAGE=$IMAGE
EOF
chmod 600 .env

# Build Traefik labels only when TLS is enabled
if [ "$TLS_ENABLED" = true ]; then
  ORBITA_LABELS=$(cat <<'LBL'
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.orbita.rule=Host(`${ORBITA_HOST}`)"
      - "traefik.http.routers.orbita.entrypoints=websecure"
      - "traefik.http.routers.orbita.tls.certresolver=letsencrypt"
      - "traefik.http.services.orbita.loadbalancer.server.port=8080"
      - "traefik.http.routers.orbita-http.rule=Host(`${ORBITA_HOST}`)"
      - "traefik.http.routers.orbita-http.entrypoints=web"
      - "traefik.http.routers.orbita-http.middlewares=orbita-redirect"
      - "traefik.http.middlewares.orbita-redirect.redirectscheme.scheme=https"
      - "traefik.http.middlewares.orbita-redirect.redirectscheme.permanent=true"
LBL
)
else
  ORBITA_LABELS='    labels:
      - "traefik.enable=false"'
fi

# Create docker-compose.yml
cat > docker-compose.yml <<DCEOF
services:
  orbita:
    image: \${ORBITA_IMAGE}
    container_name: orbita
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://orbita:\${DB_PASSWORD}@postgres:5432/orbita?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=\${JWT_SECRET}
      - JWT_REFRESH_SECRET=\${JWT_REFRESH_SECRET}
      - ENCRYPTION_MASTER_KEY=\${ENCRYPTION_MASTER_KEY}
      - APP_BASE_URL=\${APP_BASE_URL}
      - RESEND_API_KEY=\${RESEND_API_KEY}
      - RESEND_FROM_EMAIL=\${RESEND_FROM_EMAIL}
      - TRAEFIK_CONFIG_DIR=/etc/orbita/traefik
      - DOCKER_SOCKET=/var/run/docker.sock
      - IS_PRODUCTION=true
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - traefik_dynamic:/etc/orbita/traefik
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
${ORBITA_LABELS}
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    container_name: orbita-postgres
    environment:
      - POSTGRES_DB=orbita
      - POSTGRES_USER=orbita
      - POSTGRES_PASSWORD=\${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U orbita"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    container_name: orbita-redis
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  traefik:
    image: traefik:v3.6.14
    container_name: orbita-traefik
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - traefik_dynamic:/etc/orbita/traefik
      - traefik_certs:/etc/traefik/acme
    command:
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--providers.file.directory=/etc/orbita/traefik/dynamic"
      - "--providers.file.watch=true"
      - "--certificatesresolvers.letsencrypt.acme.email=\${ACME_EMAIL}"
      - "--certificatesresolvers.letsencrypt.acme.storage=/etc/traefik/acme/acme.json"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge=true"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
      - "--log.level=INFO"
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  traefik_dynamic:
  traefik_certs:
DCEOF

echo -e "${GREEN}[6/7]${NC} Pulling images and starting..."
if ! docker compose pull; then
  echo -e "${RED}Failed to pull images.${NC}"
  echo "  The Orbita image ($IMAGE) may be private or not yet published."
  echo "  Try: docker login ghcr.io   (if the image is private)"
  echo "  Or:  ORBITA_IMAGE=<your-image> bash install.sh"
  exit 1
fi

docker compose up -d

# Pre-create Traefik's dynamic config directory inside the shared volume so
# Traefik's file provider doesn't error on startup. Orbita writes per-app
# routing config here later.
docker exec orbita-traefik mkdir -p /etc/orbita/traefik/dynamic 2>/dev/null || true

echo -e "${GREEN}[7/7]${NC} Waiting for services to become healthy..."
# Retry health check for up to 2 minutes
ATTEMPTS=60
HEALTHY=false
for i in $(seq 1 $ATTEMPTS); do
  if curl -fsS http://localhost:8080/health 2>/dev/null | grep -q '"ok"'; then
    HEALTHY=true
    break
  fi
  sleep 2
  if [ $((i % 10)) -eq 0 ]; then
    echo "  ...still waiting ($((i * 2))s elapsed)"
  fi
done

if [ "$HEALTHY" = true ]; then
  echo ""
  echo -e "${GREEN}════════════════════════════════════════════${NC}"
  echo -e "${GREEN}  Orbita is running!${NC}"
  echo -e "${GREEN}════════════════════════════════════════════${NC}"
  echo ""
  if [ "$TLS_ENABLED" = true ]; then
    echo "  Dashboard:  https://$DOMAIN"
    echo "              (TLS cert will be issued by Let's Encrypt"
    echo "               on first request — DNS must point here)"
    SERVER_IP=$(hostname -I | awk '{print $1}')
    echo "  Also on:    http://$SERVER_IP:8080"
  else
    echo "  Dashboard:  $APP_BASE_URL"
  fi
  echo ""
  echo "  Next steps:"
  echo "  1. Point your domain's A record to this server's IP"
  echo "  2. Open the dashboard and register the first user (super admin)"
  echo "  3. Create your first organization"
  echo ""
  echo "  Config:     $INSTALL_DIR/.env"
  echo "  Logs:       docker compose -f $INSTALL_DIR/docker-compose.yml logs -f orbita"
  echo "  Update:     cd $INSTALL_DIR && docker compose pull && docker compose up -d"
  echo ""
else
  echo -e "${RED}Health check failed after $((ATTEMPTS * 2))s.${NC}"
  echo "  Inspect logs:   docker compose -f $INSTALL_DIR/docker-compose.yml logs orbita"
  echo "  Container state: docker compose -f $INSTALL_DIR/docker-compose.yml ps"
  exit 1
fi
