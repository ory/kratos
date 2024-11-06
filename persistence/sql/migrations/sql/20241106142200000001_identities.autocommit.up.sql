CREATE INDEX IF NOT EXISTS identity_recovery_addresses_identity_id_id_idx ON identity_recovery_addresses(identity_id ASC, id ASC);
CREATE INDEX IF NOT EXISTS identity_verifiable_addresses_identity_id_id_idx ON identity_verifiable_addresses (identity_id ASC, id ASC);

DROP INDEX IF EXISTS identity_recovery_addresses_identity_id_idx;
DROP INDEX IF EXISTS identity_verifiable_addresses_identity_id_idx;
