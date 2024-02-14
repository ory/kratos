ALTER TABLE identity_recovery_codes
    DROP FOREIGN KEY identity_recovery_codes_identity_recovery_addresses_id_fk,
    DROP FOREIGN KEY identity_recovery_tokens_identity_id_fk; -- this foreign key constraint was previously misnamed

ALTER TABLE identity_recovery_tokens
    DROP FOREIGN KEY identity_recovery_tokens_ibfk_1,
    DROP FOREIGN KEY identity_recovery_tokens_identity_id_fk_idx;

ALTER TABLE identity_recovery_addresses
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (identity_id, id);

ALTER TABLE identity_recovery_codes
    ADD CONSTRAINT identity_recovery_codes_identity_recovery_addresses_id_fk FOREIGN KEY (identity_recovery_address_id) REFERENCES identity_recovery_addresses (id) ON DELETE CASCADE,
    ADD CONSTRAINT identity_recovery_codes_identity_id_fk FOREIGN KEY (identity_id) REFERENCES identities (id) ON DELETE CASCADE; -- this foreign key constraint was previously misnamed, this is the correct name

ALTER TABLE identity_recovery_tokens
    ADD CONSTRAINT identity_recovery_tokens_ibfk_1 FOREIGN KEY (identity_recovery_address_id) REFERENCES identity_recovery_addresses(id) ON DELETE CASCADE,
    ADD CONSTRAINT identity_recovery_tokens_identity_id_fk_idx FOREIGN KEY (identity_id) REFERENCES identities(id) ON DELETE CASCADE;
