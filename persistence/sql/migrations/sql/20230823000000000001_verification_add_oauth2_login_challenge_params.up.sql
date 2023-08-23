ALTER TABLE selfservice_verification_flows ADD COLUMN identity_id UUID;
ALTER TABLE selfservice_verification_flows ADD COLUMN authentication_methods JSON;
