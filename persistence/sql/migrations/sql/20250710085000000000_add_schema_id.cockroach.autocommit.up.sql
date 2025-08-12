ALTER TABLE selfservice_login_flows ADD COLUMN IF NOT EXISTS identity_schema_id VARCHAR(128) NULL;
ALTER TABLE selfservice_registration_flows ADD COLUMN IF NOT EXISTS identity_schema_id VARCHAR(128) NULL;
