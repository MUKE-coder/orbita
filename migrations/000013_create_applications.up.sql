CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    source_type VARCHAR(50) NOT NULL DEFAULT 'docker-image',
    source_config JSONB NOT NULL DEFAULT '{}',
    build_config JSONB NOT NULL DEFAULT '{}',
    deploy_config JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'created',
    docker_service_id VARCHAR(255),
    replicas INTEGER NOT NULL DEFAULT 1,
    port INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_applications_organization_id ON applications(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_applications_environment_id ON applications(environment_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_applications_status ON applications(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_applications_deleted_at ON applications(deleted_at);
