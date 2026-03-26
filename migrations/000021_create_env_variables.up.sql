CREATE TABLE env_variables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resource_id UUID NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value_encrypted TEXT NOT NULL,
    is_secret BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_env_variables_resource ON env_variables(resource_id, resource_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_env_variables_org_id ON env_variables(organization_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_env_variables_unique_key ON env_variables(resource_id, resource_type, key) WHERE deleted_at IS NULL;
CREATE INDEX idx_env_variables_deleted_at ON env_variables(deleted_at);
