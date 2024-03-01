-- Remove unused index
DROP INDEX IF EXISTS identity_credential_identifiers_nid_id_idx;
DROP INDEX IF EXISTS identity_credential_identifiers_id_nid_idx;
DROP INDEX IF EXISTS session_devices_id_nid_idx;

CREATE INDEX IF NOT EXISTS identity_login_codes_identity_id_idx ON identity_login_codes (identity_id ASC);
CREATE INDEX IF NOT EXISTS identity_login_codes_flow_id_idx ON identity_login_codes (selfservice_login_flow_id ASC);
CREATE INDEX IF NOT EXISTS identity_registration_codes_flow_id_idx ON identity_registration_codes (selfservice_registration_flow_id ASC);
CREATE INDEX IF NOT EXISTS identity_recovery_codes_flow_id_idx ON identity_recovery_codes (selfservice_recovery_flow_id ASC);
CREATE INDEX IF NOT EXISTS identity_verification_codes_flow_id_idx ON identity_verification_codes (selfservice_verification_flow_id ASC);
