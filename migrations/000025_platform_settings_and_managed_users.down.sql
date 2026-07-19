ALTER TABLE users DROP COLUMN IF EXISTS created_by;
ALTER TABLE users DROP COLUMN IF EXISTS must_change_password;
DROP TABLE IF EXISTS platform_settings;
