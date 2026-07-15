DROP INDEX IF EXISTS idx_applications_grit_app;
ALTER TABLE applications DROP COLUMN IF EXISTS grit_role;
ALTER TABLE applications DROP COLUMN IF EXISTS grit_app;
