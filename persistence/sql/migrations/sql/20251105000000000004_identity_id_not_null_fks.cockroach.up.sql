CREATE INDEX IF NOT EXISTS identity_credential_identifiers_identities_id_fk_idx ON identity_credential_identifiers (identity_id);
CREATE INDEX IF NOT EXISTS session_devices_identities_id_fk_idx ON session_devices (identity_id);
