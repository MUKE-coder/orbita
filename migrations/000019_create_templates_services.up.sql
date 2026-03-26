CREATE TABLE templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL DEFAULT 'other',
    compose_template TEXT NOT NULL,
    params_schema JSONB NOT NULL DEFAULT '[]',
    icon_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_templates_category ON templates(category);

CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    template_id UUID REFERENCES templates(id),
    name VARCHAR(255) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'creating',
    docker_service_ids TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_services_org_id ON services(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_services_deleted_at ON services(deleted_at);

-- Seed templates
INSERT INTO templates (id, name, description, category, compose_template, params_schema, icon_url) VALUES
(uuid_generate_v4(), 'WordPress', 'Popular CMS for blogs and websites', 'cms',
'services:
  wordpress:
    image: wordpress:latest
    ports:
      - "{{.Port}}:80"
    environment:
      WORDPRESS_DB_HOST: wordpress-db
      WORDPRESS_DB_USER: wordpress
      WORDPRESS_DB_PASSWORD: "{{.DBPassword}}"
      WORDPRESS_DB_NAME: wordpress
    volumes:
      - wp_data:/var/www/html
  wordpress-db:
    image: mysql:8
    environment:
      MYSQL_DATABASE: wordpress
      MYSQL_USER: wordpress
      MYSQL_PASSWORD: "{{.DBPassword}}"
      MYSQL_ROOT_PASSWORD: "{{.DBPassword}}"
    volumes:
      - wp_db_data:/var/lib/mysql
volumes:
  wp_data:
  wp_db_data:',
'[{"name":"Port","type":"number","default":"8080","required":true},{"name":"DBPassword","type":"password","default":"","required":true}]',
'https://s.w.org/style/images/about/WordPress-logotype-wmark.png'),

(uuid_generate_v4(), 'Plausible Analytics', 'Privacy-friendly web analytics', 'analytics',
'services:
  plausible:
    image: plausible/analytics:latest
    ports:
      - "{{.Port}}:8000"
    environment:
      BASE_URL: "{{.BaseURL}}"
      SECRET_KEY_BASE: "{{.SecretKey}}"
volumes: {}',
'[{"name":"Port","type":"number","default":"8000","required":true},{"name":"BaseURL","type":"string","default":"http://localhost:8000","required":true},{"name":"SecretKey","type":"password","default":"","required":true}]',
NULL),

(uuid_generate_v4(), 'Uptime Kuma', 'Self-hosted monitoring tool', 'monitoring',
'services:
  uptime-kuma:
    image: louislam/uptime-kuma:latest
    ports:
      - "{{.Port}}:3001"
    volumes:
      - kuma_data:/app/data
volumes:
  kuma_data:',
'[{"name":"Port","type":"number","default":"3001","required":true}]',
NULL),

(uuid_generate_v4(), 'n8n', 'Workflow automation platform', 'automation',
'services:
  n8n:
    image: n8nio/n8n:latest
    ports:
      - "{{.Port}}:5678"
    environment:
      N8N_BASIC_AUTH_ACTIVE: "true"
      N8N_BASIC_AUTH_USER: "{{.AdminUser}}"
      N8N_BASIC_AUTH_PASSWORD: "{{.AdminPassword}}"
    volumes:
      - n8n_data:/home/node/.n8n
volumes:
  n8n_data:',
'[{"name":"Port","type":"number","default":"5678","required":true},{"name":"AdminUser","type":"string","default":"admin","required":true},{"name":"AdminPassword","type":"password","default":"","required":true}]',
NULL),

(uuid_generate_v4(), 'Metabase', 'Business intelligence and analytics', 'analytics',
'services:
  metabase:
    image: metabase/metabase:latest
    ports:
      - "{{.Port}}:3000"
    volumes:
      - metabase_data:/metabase-data
volumes:
  metabase_data:',
'[{"name":"Port","type":"number","default":"3000","required":true}]',
NULL),

(uuid_generate_v4(), 'Grafana', 'Observability and monitoring dashboards', 'monitoring',
'services:
  grafana:
    image: grafana/grafana:latest
    ports:
      - "{{.Port}}:3000"
    environment:
      GF_SECURITY_ADMIN_USER: "{{.AdminUser}}"
      GF_SECURITY_ADMIN_PASSWORD: "{{.AdminPassword}}"
    volumes:
      - grafana_data:/var/lib/grafana
volumes:
  grafana_data:',
'[{"name":"Port","type":"number","default":"3000","required":true},{"name":"AdminUser","type":"string","default":"admin","required":true},{"name":"AdminPassword","type":"password","default":"","required":true}]',
NULL),

(uuid_generate_v4(), 'MinIO', 'S3-compatible object storage', 'storage',
'services:
  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    ports:
      - "{{.Port}}:9000"
      - "{{.ConsolePort}}:9001"
    environment:
      MINIO_ROOT_USER: "{{.AccessKey}}"
      MINIO_ROOT_PASSWORD: "{{.SecretKey}}"
    volumes:
      - minio_data:/data
volumes:
  minio_data:',
'[{"name":"Port","type":"number","default":"9000","required":true},{"name":"ConsolePort","type":"number","default":"9001","required":true},{"name":"AccessKey","type":"string","default":"minioadmin","required":true},{"name":"SecretKey","type":"password","default":"","required":true}]',
NULL),

(uuid_generate_v4(), 'Gitea', 'Lightweight self-hosted Git service', 'devtools',
'services:
  gitea:
    image: gitea/gitea:latest
    ports:
      - "{{.Port}}:3000"
      - "{{.SSHPort}}:22"
    volumes:
      - gitea_data:/data
volumes:
  gitea_data:',
'[{"name":"Port","type":"number","default":"3000","required":true},{"name":"SSHPort","type":"number","default":"2222","required":true}]',
NULL),

(uuid_generate_v4(), 'Ghost CMS', 'Professional publishing platform', 'cms',
'services:
  ghost:
    image: ghost:latest
    ports:
      - "{{.Port}}:2368"
    environment:
      url: "{{.SiteURL}}"
    volumes:
      - ghost_data:/var/lib/ghost/content
volumes:
  ghost_data:',
'[{"name":"Port","type":"number","default":"2368","required":true},{"name":"SiteURL","type":"string","default":"http://localhost:2368","required":true}]',
NULL),

(uuid_generate_v4(), 'Vaultwarden', 'Bitwarden-compatible password manager', 'security',
'services:
  vaultwarden:
    image: vaultwarden/server:latest
    ports:
      - "{{.Port}}:80"
    environment:
      ADMIN_TOKEN: "{{.AdminToken}}"
    volumes:
      - vw_data:/data
volumes:
  vw_data:',
'[{"name":"Port","type":"number","default":"8080","required":true},{"name":"AdminToken","type":"password","default":"","required":true}]',
NULL);
