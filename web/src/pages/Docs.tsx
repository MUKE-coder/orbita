import { Link } from "react-router-dom";
import {
  Rocket,
  ArrowLeft,
  Server,
  Download,
  Settings,
  Shield,
  GitFork,
  Terminal,
  CheckCircle,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

function Docs() {
  return (
    <div className="min-h-screen bg-background">
      {/* Nav */}
      <nav className="border-b">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <Link to="/" className="flex items-center gap-2 text-xl font-bold">
            <Rocket className="h-6 w-6" /> Orbita
          </Link>
          <div className="flex items-center gap-4">
            <a href="https://github.com/MUKE-coder/orbita" target="_blank" rel="noopener noreferrer" className="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1">
              <GitFork className="h-4 w-4" /> GitHub
            </a>
            <Link to="/login"><Button size="sm">Sign In</Button></Link>
          </div>
        </div>
      </nav>

      <div className="mx-auto max-w-4xl px-6 py-12 space-y-12">
        <div>
          <Link to="/" className="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1 mb-4">
            <ArrowLeft className="h-4 w-4" /> Back to home
          </Link>
          <h1 className="text-4xl font-bold">Installation Guide</h1>
          <p className="mt-2 text-muted-foreground">
            Deploy Orbita on your VPS or dedicated server in under 10 minutes.
          </p>
        </div>

        {/* Prerequisites */}
        <section>
          <h2 className="text-2xl font-bold flex items-center gap-2 mb-4">
            <Server className="h-6 w-6 text-primary" /> Prerequisites
          </h2>
          <Card>
            <CardContent className="pt-6 space-y-3">
              <div className="grid gap-3 sm:grid-cols-2">
                <div className="flex items-start gap-2">
                  <CheckCircle className="h-5 w-5 text-green-500 shrink-0 mt-0.5" />
                  <div><strong>Linux VPS</strong><br /><span className="text-sm text-muted-foreground">Ubuntu 22.04+, Debian 12+, or any modern Linux with systemd</span></div>
                </div>
                <div className="flex items-start gap-2">
                  <CheckCircle className="h-5 w-5 text-green-500 shrink-0 mt-0.5" />
                  <div><strong>Docker 24+</strong><br /><span className="text-sm text-muted-foreground">With Docker Compose v2. Swarm mode for multi-node.</span></div>
                </div>
                <div className="flex items-start gap-2">
                  <CheckCircle className="h-5 w-5 text-green-500 shrink-0 mt-0.5" />
                  <div><strong>1 vCPU / 1GB RAM</strong><br /><span className="text-sm text-muted-foreground">Minimum. 2 vCPU / 4GB recommended for production.</span></div>
                </div>
                <div className="flex items-start gap-2">
                  <CheckCircle className="h-5 w-5 text-green-500 shrink-0 mt-0.5" />
                  <div><strong>Domain Name</strong><br /><span className="text-sm text-muted-foreground">Point an A record to your server IP for SSL.</span></div>
                </div>
              </div>
            </CardContent>
          </Card>
        </section>

        <Separator />

        {/* Method 1: Docker */}
        <section>
          <h2 className="text-2xl font-bold flex items-center gap-2 mb-4">
            <Download className="h-6 w-6 text-primary" /> Method 1: Docker Compose (Recommended)
          </h2>
          <div className="space-y-4">
            <Step n={1} title="Install Docker">
              <Code>{`curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
# Log out and back in, then:
docker swarm init`}</Code>
            </Step>

            <Step n={2} title="Create the project directory">
              <Code>{`mkdir -p /opt/orbita && cd /opt/orbita`}</Code>
            </Step>

            <Step n={3} title="Create docker-compose.yml">
              <Code>{`cat > docker-compose.yml << 'EOF'
version: '3.8'
services:
  orbita:
    image: ghcr.io/muke-coder/orbita:latest
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://orbita:orbita@postgres:5432/orbita?sslmode=disable
      REDIS_URL: redis://redis:6379
      JWT_SECRET: \${JWT_SECRET:-$(openssl rand -hex 32)}
      JWT_REFRESH_SECRET: \${JWT_REFRESH_SECRET:-$(openssl rand -hex 32)}
      ENCRYPTION_MASTER_KEY: \${ENCRYPTION_MASTER_KEY:-$(openssl rand -hex 16)}
      APP_BASE_URL: https://your-domain.com
      RESEND_API_KEY: \${RESEND_API_KEY:-re_placeholder}
      RESEND_FROM_EMAIL: noreply@your-domain.com
      TRAEFIK_CONFIG_DIR: /etc/orbita/traefik
      IS_PRODUCTION: "true"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - traefik_config:/etc/orbita/traefik
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: orbita
      POSTGRES_USER: orbita
      POSTGRES_PASSWORD: orbita
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    restart: unless-stopped

  traefik:
    image: traefik:v3.0
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - traefik_config:/etc/orbita/traefik
      - traefik_certs:/etc/traefik/acme
    command:
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--providers.file.directory=/etc/orbita/traefik/dynamic"
      - "--providers.file.watch=true"
      - "--certificatesresolvers.letsencrypt.acme.email=admin@your-domain.com"
      - "--certificatesresolvers.letsencrypt.acme.storage=/etc/traefik/acme/acme.json"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  traefik_config:
  traefik_certs:
EOF`}</Code>
            </Step>

            <Step n={4} title="Generate secrets and start">
              <Code>{`# Generate secure secrets
export JWT_SECRET=$(openssl rand -hex 32)
export JWT_REFRESH_SECRET=$(openssl rand -hex 32)
export ENCRYPTION_MASTER_KEY=$(openssl rand -hex 16)

# Start everything
docker compose up -d

# Check logs
docker compose logs -f orbita`}</Code>
            </Step>

            <Step n={5} title="Access Orbita">
              <p className="text-sm text-muted-foreground">
                Open <code className="bg-muted px-1.5 py-0.5 rounded">http://your-server-ip:8080</code> in your browser.
                The first user to register automatically becomes the super admin.
              </p>
            </Step>
          </div>
        </section>

        <Separator />

        {/* Method 2: Binary */}
        <section>
          <h2 className="text-2xl font-bold flex items-center gap-2 mb-4">
            <Terminal className="h-6 w-6 text-primary" /> Method 2: Single Binary
          </h2>
          <div className="space-y-4">
            <Step n={1} title="Download the binary">
              <Code>{`# AMD64
curl -L -o orbita https://github.com/MUKE-coder/orbita/releases/latest/download/orbita-linux-amd64
chmod +x orbita

# ARM64
curl -L -o orbita https://github.com/MUKE-coder/orbita/releases/latest/download/orbita-linux-arm64
chmod +x orbita`}</Code>
            </Step>

            <Step n={2} title="Set up PostgreSQL and Redis">
              <Code>{`# Install PostgreSQL
sudo apt install -y postgresql
sudo -u postgres createuser orbita
sudo -u postgres createdb orbita -O orbita
sudo -u postgres psql -c "ALTER USER orbita PASSWORD 'your-db-password';"

# Install Redis
sudo apt install -y redis-server
sudo systemctl enable redis-server`}</Code>
            </Step>

            <Step n={3} title="Create .env file">
              <Code>{`cat > /opt/orbita/.env << 'EOF'
SERVER_PORT=8080
APP_BASE_URL=https://your-domain.com
IS_PRODUCTION=true
DATABASE_URL=postgres://orbita:your-db-password@localhost:5432/orbita?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=$(openssl rand -hex 32)
JWT_REFRESH_SECRET=$(openssl rand -hex 32)
ENCRYPTION_MASTER_KEY=$(openssl rand -hex 16)
RESEND_API_KEY=re_your_key
RESEND_FROM_EMAIL=noreply@your-domain.com
DOCKER_SOCKET=/var/run/docker.sock
TRAEFIK_CONFIG_DIR=/etc/orbita/traefik
EOF`}</Code>
            </Step>

            <Step n={4} title="Create systemd service">
              <Code>{`sudo cat > /etc/systemd/system/orbita.service << 'EOF'
[Unit]
Description=Orbita PaaS
After=network.target postgresql.service redis.service docker.service
Requires=docker.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/orbita
ExecStart=/opt/orbita/orbita
Restart=always
RestartSec=5
EnvironmentFile=/opt/orbita/.env

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable orbita
sudo systemctl start orbita`}</Code>
            </Step>

            <Step n={5} title="Check status">
              <Code>{`sudo systemctl status orbita
curl http://localhost:8080/health`}</Code>
            </Step>
          </div>
        </section>

        <Separator />

        {/* Post-install */}
        <section>
          <h2 className="text-2xl font-bold flex items-center gap-2 mb-4">
            <Settings className="h-6 w-6 text-primary" /> Post-Installation
          </h2>
          <div className="space-y-4">
            <Card>
              <CardHeader><CardTitle className="text-base">1. Register Super Admin</CardTitle></CardHeader>
              <CardContent className="text-sm text-muted-foreground">
                Navigate to your Orbita URL and register the first account. This user automatically becomes the super admin with full platform access.
              </CardContent>
            </Card>
            <Card>
              <CardHeader><CardTitle className="text-base">2. Create Your First Organization</CardTitle></CardHeader>
              <CardContent className="text-sm text-muted-foreground">
                Click "Create Organization" and give it a name and slug. Each org gets its own isolated Docker network, encryption keys, and resource quotas.
              </CardContent>
            </Card>
            <Card>
              <CardHeader><CardTitle className="text-base">3. Set Up DNS</CardTitle></CardHeader>
              <CardContent className="text-sm text-muted-foreground">
                Point your domain (and wildcard *.domain) to your server IP with an A record. Traefik will handle SSL certificates automatically via Let's Encrypt.
              </CardContent>
            </Card>
            <Card>
              <CardHeader><CardTitle className="text-base">4. Configure Email (Optional)</CardTitle></CardHeader>
              <CardContent className="text-sm text-muted-foreground">
                Sign up for <a href="https://resend.com" className="text-primary hover:underline" target="_blank" rel="noopener noreferrer">Resend</a> and set your <code className="bg-muted px-1 rounded">RESEND_API_KEY</code> for invite emails, password resets, and deploy notifications.
              </CardContent>
            </Card>
          </div>
        </section>

        <Separator />

        {/* Security */}
        <section>
          <h2 className="text-2xl font-bold flex items-center gap-2 mb-4">
            <Shield className="h-6 w-6 text-primary" /> Security Checklist
          </h2>
          <div className="grid gap-2 sm:grid-cols-2">
            {[
              "Generate unique JWT_SECRET and JWT_REFRESH_SECRET (64+ chars)",
              "Generate unique ENCRYPTION_MASTER_KEY (32 hex chars)",
              "Enable firewall: allow only 80, 443, 22",
              "Set IS_PRODUCTION=true in production",
              "Use HTTPS (Traefik + Let's Encrypt)",
              "Set up automated backups for PostgreSQL",
              "Keep Docker and Orbita updated",
              "Enable 2FA for the super admin account",
            ].map((item) => (
              <div key={item} className="flex items-start gap-2 text-sm">
                <CheckCircle className="h-4 w-4 text-green-500 shrink-0 mt-0.5" />
                {item}
              </div>
            ))}
          </div>
        </section>

        <Separator />

        {/* Need help */}
        <section className="text-center py-8">
          <h2 className="text-2xl font-bold mb-2">Need Help?</h2>
          <p className="text-muted-foreground mb-4">
            Check the GitHub repository for issues, discussions, and updates.
          </p>
          <a href="https://github.com/MUKE-coder/orbita" target="_blank" rel="noopener noreferrer">
            <Button variant="outline">
              <GitFork className="mr-2 h-4 w-4" /> View on GitHub
            </Button>
          </a>
        </section>
      </div>
    </div>
  );
}

function Step({ n, title, children }: { n: number; title: string; children: React.ReactNode }) {
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <Badge variant="secondary" className="h-6 w-6 rounded-full p-0 flex items-center justify-center text-xs">
          {n}
        </Badge>
        <h3 className="font-semibold">{title}</h3>
      </div>
      <div className="ml-8">{children}</div>
    </div>
  );
}

function Code({ children }: { children: string }) {
  return (
    <pre className="rounded-lg bg-gray-950 p-4 text-xs text-green-400 font-mono overflow-x-auto">
      {children}
    </pre>
  );
}

export default Docs;
