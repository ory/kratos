ALTER TABLE identity_recovery_codes
    DROP FOREIGN KEY identity_recovery_codes_identity_recovery_addresses_id_fk,
    DROP FOREIGN KEY identity_recovery_codes_identity_id_fk;

ALTER TABLE identity_recovery_tokens
    DROP FOREIGN KEY identity_recovery_tokens_ibfk_1,
    DROP FOREIGN KEY identity_recovery_tokens_identity_id_fk_idx;

ALTER TABLE identity_recovery_addresses
    DROP FOREIGN KEY identity_recovery_addresses_ibfk_1;

ALTER TABLE identity_recovery_addresses
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (id),
    ADD CONSTRAINT identity_recovery_addresses_ibfk_1 FOREIGN KEY (identity_id) REFERENCES identities(id) ON DELETE CASCADE;

ALTER TABLE identity_recovery_codes
    ADD CONSTRAINT identity_recovery_codes_identity_recovery_addresses_id_fk FOREIGN KEY (identity_recovery_address_id) REFERENCES identity_recovery_addresses (id) ON DELETE CASCADE,
    ADD CONSTRAINT identity_recovery_tokens_identity_id_fk FOREIGN KEY (identity_id) REFERENCES identities (id) ON DELETE CASCADE; -- this foreign key constraint was previously misnamed and this down-migration restores the incorrect name


ALTER TABLE identity_recovery_tokens
    ADD CONSTRAINT identity_recovery_tokens_ibfk_1 FOREIGN KEY (identity_recovery_address_id) REFERENCES identity_recovery_addresses(id) ON DELETE CASCADE,
    ADD CONSTRAINT identity_recovery_tokens_identity_id_fk_idx FOREIGN KEY (identity_id) REFERENCES identities(id) ON DELETE CASCADE;
