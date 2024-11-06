CREATE INDEX IF NOT EXISTS selfservice_login_flows_nid_id_idx ON selfservice_login_flows (nid ASC, id ASC);
CREATE INDEX IF NOT EXISTS selfservice_login_flows_id_nid_idx ON selfservice_login_flows (id ASC, nid ASC);
DROP INDEX IF EXISTS selfservice_login_flows_nid_idx;

CREATE INDEX IF NOT EXISTS selfservice_errors_errors_nid_id_idx ON selfservice_errors (nid ASC, id ASC);
DROP INDEX IF EXISTS selfservice_errors_nid_idx;

CREATE INDEX IF NOT EXISTS selfservice_recovery_flows_nid_id_idx ON selfservice_recovery_flows (nid ASC, id ASC);
CREATE INDEX IF NOT EXISTS selfservice_recovery_flows_id_nid_idx ON selfservice_recovery_flows (id ASC, nid ASC);
CREATE INDEX IF NOT EXISTS selfservice_recovery_flows_recovered_identity_id_nid_idx ON selfservice_recovery_flows (recovered_identity_id ASC, nid ASC);
DROP INDEX IF EXISTS selfservice_recovery_flows_nid_idx;
DROP INDEX IF EXISTS selfservice_recovery_flows_recovered_identity_id_idx;

CREATE INDEX IF NOT EXISTS selfservice_registration_flows_nid_id_idx ON selfservice_registration_flows (nid ASC, id ASC);
CREATE INDEX IF NOT EXISTS selfservice_registration_flows_id_nid_idx ON selfservice_registration_flows (id ASC, nid ASC);
DROP INDEX IF EXISTS selfservice_registration_flows_nid_idx;

CREATE INDEX IF NOT EXISTS selfservice_settings_flows_nid_id_idx ON selfservice_settings_flows (nid ASC, id ASC);
CREATE INDEX IF NOT EXISTS selfservice_settings_flows_id_nid_idx ON selfservice_settings_flows (id ASC, nid ASC);
CREATE INDEX IF NOT EXISTS selfservice_settings_flows_identity_id_nid_idx ON selfservice_settings_flows (identity_id ASC, nid ASC);
DROP INDEX IF EXISTS selfservice_settings_flows_nid_idx;
DROP INDEX IF EXISTS selfservice_settings_flows_identity_id_idx;

CREATE INDEX IF NOT EXISTS selfservice_verification_flows_nid_id_idx ON selfservice_verification_flows (nid ASC, id ASC);
CREATE INDEX IF NOT EXISTS selfservice_verification_flows_id_nid_idx ON selfservice_verification_flows (id ASC, nid ASC);
DROP INDEX IF EXISTS selfservice_verification_flows_nid_idx;
