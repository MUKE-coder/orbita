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

echo -e "${GREEN}[1/6]${NC} Checking prerequisites..."

# Install Docker if not present
if ! command -v docker &> /dev/null; then
  echo -e "${YELLOW}Docker not found. Installing...${NC}"
  curl -fsSL https://get.docker.com | sh
  systemctl enable docker
  systemctl start docker
  echo -e "${GREEN}Docker installed successfully${NC}"
else
  echo -e "${GREEN}Docker is already installed${NC}"
fi

# Initialize Swarm if not already
if ! docker info 2>/dev/null | grep -q "Swarm: active"; then
  echo -e "${YELLOW}Initializing Docker Swarm...${NC}"
  docker swarm init 2>/dev/null || true
fi

echo -e "${GREEN}[2/6]${NC} Creating Orbita directory..."
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

echo -e "${GREEN}[3/6]${NC} Generating secrets..."

# Preserve existing secrets on re-run
if [ -f .env ]; then
  echo -e "${YELLOW}Existing .env found — keeping existing secrets${NC}"
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

echo -e "${GREEN}[4/6]${NC} Creating configuration..."

# Prompt for domain (allow non-interactive via env)
if [ -n "${ORBITA_DOMAIN:-}" ]; then
  DOMAIN="$ORBITA_DOMAIN"
elif [ -n "$DOMAIN_EXISTING" ]; then
  DOMAIN="$DOMAIN_EXISTING"
else
  read -rp "Enter your domain (e.g., orbita.example.com) [localhost]: " DOMAIN
  DOMAIN=${DOMAIN:-localhost}
fi

if [ -n "${ORBITA_ACME_EMAIL:-}" ]; then
  ACME_EMAIL="$ORBITA_ACME_EMAIL"
elif [ -n "$ACME_EMAIL_EXISTING" ]; then
  ACME_EMAIL="$ACME_EMAIL_EXISTING"
else
  read -rp "Enter email for Let's Encrypt SSL (e.g., admin@example.com) [admin@$DOMAIN]: " ACME_EMAIL
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
    image: traefik:v3.0
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

echo -e "${GREEN}[5/6]${NC} Pulling images..."
if ! docker compose pull; then
  echo -e "${RED}Failed to pull images.${NC}"
  echo "  The Orbita image ($IMAGE) may be private or not yet published."
  echo "  Try: docker login ghcr.io   (if the image is private)"
  echo "  Or:  ORBITA_IMAGE=<your-image> bash install.sh"
  exit 1
fi

echo -e "${GREEN}Starting Orbita...${NC}"
docker compose up -d

echo -e "${GREEN}[6/6]${NC} Waiting for services to become healthy..."
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
    echo "  Also on:    http://$(hostname -I | awk '{print $1}'):8080"
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
