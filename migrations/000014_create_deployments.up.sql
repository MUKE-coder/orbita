CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1,
    image_ref VARCHAR(512) NOT NULL,
    deploy_config JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    triggered_by UUID REFERENCES users(id),
    trigger_type VARCHAR(50) NOT NULL DEFAULT 'manual',
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deployments_app_id ON deployments(app_id);
CREATE INDEX idx_deployments_status ON deployments(status);
