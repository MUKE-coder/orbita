CREATE TABLE domains (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resource_id UUID NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    domain VARCHAR(512) NOT NULL UNIQUE,
    ssl_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ssl_config JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_domains_resource ON domains(resource_id, resource_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_domains_org_id ON domains(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_domains_domain ON domains(domain) WHERE deleted_at IS NULL;
CREATE INDEX idx_domains_deleted_at ON domains(deleted_at);
