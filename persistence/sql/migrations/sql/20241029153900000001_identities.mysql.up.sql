CREATE INDEX identities_nid_idx ON identities (nid ASC);
DROP INDEX identities_id_nid_idx ON identities;
DROP INDEX identities_nid_id_idx ON identities;

CREATE INDEX identity_recovery_addresses_identity_id_idx ON identity_recovery_addresses (identity_id ASC);
DROP INDEX identity_recovery_addresses_status_via_idx ON identity_recovery_addresses;
DROP INDEX identity_recovery_addresses_nid_identity_id_idx ON identity_recovery_addresses;
DROP INDEX identity_recovery_addresses_nid_id_idx ON identity_recovery_addresses;
DROP INDEX identity_recovery_addresses_id_nid_idx ON identity_recovery_addresses;

CREATE INDEX identity_verifiable_addresses_identity_id_idx ON identity_verifiable_addresses (identity_id ASC);
DROP INDEX identity_verifiable_addresses_status_via_idx ON identity_verifiable_addresses;
DROP INDEX identity_verifiable_addresses_nid_identity_id_idx ON identity_verifiable_addresses;
DROP INDEX identity_verifiable_addresses_nid_id_idx ON identity_verifiable_addresses;
DROP INDEX identity_verifiable_addresses_id_nid_idx ON identity_verifiable_addresses;
