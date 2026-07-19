-- Platform-wide settings editable by the super admin from the dashboard, so an
-- operator can turn on email without SSHing in to edit compose and restart.
-- Single-row table: the id CHECK keeps it that way.
CREATE TABLE platform_settings (
    id                     INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    -- Encrypted at rest (AES-256-GCM under a platform-derived key), never
    -- returned to the client — the API reports only whether it is set.
    resend_api_key_enc     TEXT,
    email_from             VARCHAR(255),
    email_from_name        VARCHAR(255),
    updated_by             UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO platform_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

-- When an operator has no email provider configured, the super admin creates
-- accounts directly with a password instead of sending an invite. Those users
-- must choose their own password the first time they sign in, so the
-- handover credential stops being a valid long-term secret.
ALTER TABLE users
    ADD COLUMN must_change_password BOOLEAN NOT NULL DEFAULT FALSE;

-- Distinguishes accounts provisioned by an admin from self-registered ones.
-- Useful for auditing who was handed credentials versus who signed up.
ALTER TABLE users
    ADD COLUMN created_by UUID REFERENCES users(id) ON DELETE SET NULL;
