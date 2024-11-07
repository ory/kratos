-- DROP INDEXIF EXISTS  identities_id_nid_idx;

CREATE INDEX IF NOT EXISTS identity_recovery_addresses_identity_id_idx ON identity_recovery_addresses(identity_id ASC);
-- DROP INDEXIF EXISTS identity_recovery_addresses_status_via_idx;
-- DROP INDEXIF EXISTS identity_recovery_addresses_nid_identity_id_idx;
-- DROP INDEXIF EXISTS identity_recovery_addresses_nid_id_idx;
-- DROP INDEXIF EXISTS identity_recovery_addresses_id_nid_idx;

CREATE INDEX IF NOT EXISTS identity_verifiable_addresses_identity_id_idx ON identity_verifiable_addresses (identity_id ASC);
-- DROP INDEXIF EXISTS identity_verifiable_addresses_status_via_idx;
-- DROP INDEXIF EXISTS identity_verifiable_addresses_nid_identity_id_idx;
-- DROP INDEXIF EXISTS identity_verifiable_addresses_nid_id_idx;
-- DROP INDEXIF EXISTS identity_verifiable_addresses_id_nid_idx;
