ALTER TABLE selfservice_verification_flows ADD COLUMN identity_id VARCHAR(36);
ALTER TABLE selfservice_verification_flows ADD COLUMN authentication_methods TEXT;
