ALTER TABLE identity_credential_identifiers ADD COLUMN identity_id UUID NULL;
ALTER TABLE session_devices ADD COLUMN identity_id UUID NULL;
