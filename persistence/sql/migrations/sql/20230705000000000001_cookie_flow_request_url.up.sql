ALTER TABLE selfservice_login_flows ALTER COLUMN request_url TYPE TEXT;
ALTER TABLE selfservice_recovery_flows ALTER COLUMN request_url TYPE TEXT;
ALTER TABLE selfservice_registration_flows ALTER COLUMN request_url TYPE TEXT;
ALTER TABLE selfservice_settings_flows ALTER COLUMN request_url TYPE TEXT;
ALTER TABLE selfservice_verification_flows ALTER COLUMN request_url TYPE TEXT;
