-- Grit-awareness: a Grit app fans out to one application row per deployable
-- service (api/web/admin/docs, or a single "app"). These columns group those
-- rows and identify their role so the deploy engine can build each with the
-- right recipe and run migrations against the API service.
ALTER TABLE applications ADD COLUMN IF NOT EXISTS grit_app VARCHAR(255);
ALTER TABLE applications ADD COLUMN IF NOT EXISTS grit_role VARCHAR(20);

CREATE INDEX IF NOT EXISTS idx_applications_grit_app
    ON applications(environment_id, grit_app) WHERE deleted_at IS NULL AND grit_app IS NOT NULL;
