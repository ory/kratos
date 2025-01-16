CREATE INDEX identities_id_nid_idx ON identities (id ASC, nid ASC);

CREATE INDEX identity_recovery_addresses_status_via_idx ON identity_recovery_addresses (nid ASC, via ASC, value ASC);
-- While this index did not exist in the past, it is needed in MySQL for foreign key relations. We accept
-- that this index is "unaccounted" for if we execute down and then up migrations on MySQL.
CREATE INDEX identity_recovery_addresses_identity_id_fk_idx ON identity_recovery_addresses (identity_id ASC);
CREATE INDEX identity_recovery_addresses_nid_id_idx ON identity_recovery_addresses (nid ASC, id ASC);
CREATE INDEX identity_recovery_addresses_id_nid_idx ON identity_recovery_addresses (id ASC, nid ASC);
DROP INDEX identity_recovery_addresses_identity_id_idx ON identity_recovery_addresses;

CREATE INDEX identity_verifiable_addresses_status_via_idx ON identity_verifiable_addresses (nid ASC, via ASC, value ASC);
-- While this index did not exist in the past, it is needed in MySQL for foreign key relations. We accept
-- that this index is "unaccounted" for if we execute down and then up migrations on MySQL.
CREATE INDEX identity_verifiable_addresses_identity_id_fk_idx ON identity_verifiable_addresses (identity_id ASC);
CREATE INDEX identity_verifiable_addresses_nid_id_idx ON identity_verifiable_addresses (nid ASC, id ASC);
CREATE INDEX identity_verifiable_addresses_id_nid_idx ON identity_verifiable_addresses (id ASC, nid ASC);
DROP INDEX identity_verifiable_addresses_identity_id_idx ON identity_verifiable_addresses;
