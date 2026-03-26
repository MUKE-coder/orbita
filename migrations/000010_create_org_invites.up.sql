CREATE TABLE org_invites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'developer',
    token_hash VARCHAR(255) NOT NULL,
    invited_by UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_org_invites_org_id ON org_invites(org_id);
CREATE INDEX idx_org_invites_token_hash ON org_invites(token_hash);
CREATE INDEX idx_org_invites_email ON org_invites(email);
