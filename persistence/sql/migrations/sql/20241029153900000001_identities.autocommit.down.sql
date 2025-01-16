CREATE INDEX IF NOT EXISTS identities_id_nid_idx ON identities (id ASC, nid ASC);

CREATE INDEX IF NOT EXISTS identity_recovery_addresses_status_via_idx ON identity_recovery_addresses (nid ASC, via ASC, value ASC);
CREATE INDEX IF NOT EXISTS identity_recovery_addresses_nid_identity_id_idx ON identity_recovery_addresses (identity_id ASC, nid ASC);
CREATE INDEX IF NOT EXISTS identity_recovery_addresses_nid_id_idx ON identity_recovery_addresses (nid ASC, id ASC);
CREATE INDEX IF NOT EXISTS identity_recovery_addresses_id_nid_idx ON identity_recovery_addresses (id ASC, nid ASC);
DROP INDEX IF EXISTS identity_recovery_addresses_identity_id_idx;

CREATE INDEX IF NOT EXISTS identity_verifiable_addresses_status_via_idx ON identity_verifiable_addresses (nid ASC, via ASC, value ASC);
CREATE INDEX IF NOT EXISTS identity_verifiable_addresses_nid_identity_id_idx ON identity_verifiable_addresses (identity_id ASC, nid ASC);
CREATE INDEX IF NOT EXISTS identity_verifiable_addresses_nid_id_idx ON identity_verifiable_addresses (nid ASC, id ASC);
CREATE INDEX IF NOT EXISTS identity_verifiable_addresses_id_nid_idx ON identity_verifiable_addresses (id ASC, nid ASC);
DROP INDEX IF EXISTS identity_verifiable_addresses_identity_id_idx;
