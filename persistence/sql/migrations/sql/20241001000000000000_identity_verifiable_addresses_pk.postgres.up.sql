CREATE UNIQUE INDEX identity_verifiable_addresses_id_uq_idx ON identity_verifiable_addresses (id);

ALTER TABLE identity_verification_codes
    DROP CONSTRAINT identity_verification_codes_identity_verifiable_addresses_id_fk;

ALTER TABLE identity_verification_tokens
    DROP CONSTRAINT identity_verification_tokens_identity_verifiable_address_i_fkey;

ALTER TABLE identity_verifiable_addresses
    DROP CONSTRAINT identity_verifiable_addresses_pkey,
    ADD PRIMARY KEY (identity_id, id);

ALTER TABLE identity_verification_codes
    ADD CONSTRAINT identity_verification_codes_identity_verifiable_addresses_id_fk FOREIGN KEY (identity_verifiable_address_id) REFERENCES identity_verifiable_addresses(id) ON DELETE CASCADE;

ALTER TABLE identity_verification_tokens
    ADD CONSTRAINT identity_verification_tokens_identity_verifiable_address_i_fkey FOREIGN KEY (identity_verifiable_address_id) REFERENCES identity_verifiable_addresses(id) ON DELETE CASCADE;
