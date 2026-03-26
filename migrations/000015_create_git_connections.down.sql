ALTER TABLE applications DROP COLUMN IF EXISTS webhook_secret;
ALTER TABLE applications DROP COLUMN IF EXISTS auto_deploy;
DROP TABLE IF EXISTS registry_credentials;
DROP TABLE IF EXISTS git_connections;
