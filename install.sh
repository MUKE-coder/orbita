#!/bin/bash
set -e

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
mkdir -p /opt/orbita
cd /opt/orbita

echo -e "${GREEN}[3/6]${NC} Generating secrets..."
DB_PASSWORD=$(openssl rand -hex 16)
JWT_SECRET=$(openssl rand -hex 32)
JWT_REFRESH_SECRET=$(openssl rand -hex 32)
ENCRYPTION_MASTER_KEY=$(openssl rand -hex 16)

echo -e "${GREEN}[4/6]${NC} Creating configuration..."

# Prompt for domain
read -p "Enter your domain (e.g., orbita.example.com): " DOMAIN
DOMAIN=${DOMAIN:-localhost}

read -p "Enter email for Let's Encrypt SSL (e.g., admin@example.com): " ACME_EMAIL
ACME_EMAIL=${ACME_EMAIL:-admin@localhost}

# Create .env
cat > .env << EOF
DB_PASSWORD=$DB_PASSWORD
JWT_SECRET=$JWT_SECRET
JWT_REFRESH_SECRET=$JWT_REFRESH_SECRET
ENCRYPTION_MASTER_KEY=$ENCRYPTION_MASTER_KEY
APP_BASE_URL=https://$DOMAIN
RESEND_API_KEY=re_placeholder
RESEND_FROM_EMAIL=noreply@$DOMAIN
ACME_EMAIL=$ACME_EMAIL
EOF

# Create docker-compose.yml
cat > docker-compose.yml << 'DCEOF'
version: '3.8'
services:
  orbita:
    image: ghcr.io/muke-coder/orbita:latest
    container_name: orbita
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://orbita:${DB_PASSWORD}@postgres:5432/orbita?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
      - JWT_REFRESH_SECRET=${JWT_REFRESH_SECRET}
      - ENCRYPTION_MASTER_KEY=${ENCRYPTION_MASTER_KEY}
      - APP_BASE_URL=${APP_BASE_URL}
      - RESEND_API_KEY=${RESEND_API_KEY}
      - RESEND_FROM_EMAIL=${RESEND_FROM_EMAIL}
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
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    container_name: orbita-postgres
    environment:
      - POSTGRES_DB=orbita
      - POSTGRES_USER=orbita
      - POSTGRES_PASSWORD=${DB_PASSWORD}
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
      - "--providers.file.directory=/etc/orbita/traefik/dynamic"
      - "--providers.file.watch=true"
      - "--certificatesresolvers.letsencrypt.acme.email=${ACME_EMAIL}"
      - "--certificatesresolvers.letsencrypt.acme.storage=/etc/traefik/acme/acme.json"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  traefik_dynamic:
  traefik_certs:
DCEOF

echo -e "${GREEN}[5/6]${NC} Starting Orbita..."
docker compose up -d

echo -e "${GREEN}[6/6]${NC} Waiting for services to start..."
sleep 10

# Health check
if curl -s http://localhost:8080/health | grep -q '"ok"'; then
  echo ""
  echo -e "${GREEN}════════════════════════════════════════════${NC}"
  echo -e "${GREEN}  Orbita is running!${NC}"
  echo -e "${GREEN}════════════════════════════════════════════${NC}"
  echo ""
  echo "  Dashboard:  http://$DOMAIN:8080"
  echo "              (or https://$DOMAIN once DNS is configured)"
  echo ""
  echo "  Next steps:"
  echo "  1. Point your domain's A record to this server's IP"
  echo "  2. Open the dashboard and register the first user (super admin)"
  echo "  3. Create your first organization"
  echo ""
  echo "  Config:     /opt/orbita/.env"
  echo "  Logs:       docker compose logs -f orbita"
  echo "  Update:     docker compose pull && docker compose up -d"
  echo ""
else
  echo -e "${RED}Health check failed. Check logs:${NC}"
  echo "  docker compose logs orbita"
fi
