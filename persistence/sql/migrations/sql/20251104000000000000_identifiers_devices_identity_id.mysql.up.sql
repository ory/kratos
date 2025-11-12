ALTER TABLE identity_credential_identifiers ADD COLUMN identity_id char(36) NULL;
ALTER TABLE session_devices ADD COLUMN identity_id char(36) NULL;
