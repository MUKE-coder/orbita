CREATE TABLE managed_databases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    engine VARCHAR(50) NOT NULL,
    version VARCHAR(50) NOT NULL,
    connection_config TEXT,
    volume_name VARCHAR(255),
    docker_service_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'creating',
    port INTEGER,
    cpu_limit INTEGER NOT NULL DEFAULT 0,
    memory_limit INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_managed_databases_org_id ON managed_databases(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_managed_databases_env_id ON managed_databases(environment_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_managed_databases_deleted_at ON managed_databases(deleted_at);

CREATE TABLE backups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_id UUID NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    size_bytes BIGINT NOT NULL DEFAULT 0,
    storage_path TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_backups_source ON backups(source_id, source_type);
CREATE INDEX idx_backups_org_id ON backups(organization_id);

CREATE TABLE backup_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_id UUID NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    frequency VARCHAR(50) NOT NULL DEFAULT 'daily',
    retention_count INTEGER NOT NULL DEFAULT 7,
    destination_config TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_backup_schedules_source ON backup_schedules(source_id, source_type);
