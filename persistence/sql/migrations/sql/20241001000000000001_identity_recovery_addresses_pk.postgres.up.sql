CREATE UNIQUE INDEX identity_recovery_addresses_id_uq_idx ON identity_recovery_addresses (id);

ALTER TABLE identity_recovery_codes
    DROP CONSTRAINT identity_recovery_codes_identity_recovery_addresses_id_fk;

ALTER TABLE identity_recovery_tokens
    DROP CONSTRAINT identity_recovery_tokens_identity_recovery_address_id_fkey;

ALTER TABLE identity_recovery_addresses
    DROP CONSTRAINT identity_recovery_addresses_pkey,
    ADD PRIMARY KEY (identity_id, id);

ALTER TABLE identity_recovery_codes
    ADD CONSTRAINT identity_recovery_codes_identity_recovery_addresses_id_fk FOREIGN KEY (identity_recovery_address_id) REFERENCES identity_recovery_addresses(id) ON DELETE CASCADE;

ALTER TABLE identity_recovery_tokens
    ADD CONSTRAINT identity_recovery_tokens_identity_recovery_address_id_fkey FOREIGN KEY (identity_recovery_address_id) REFERENCES identity_recovery_addresses(id) ON DELETE CASCADE;
