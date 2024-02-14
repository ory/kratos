ALTER TABLE identity_verifiable_addresses
    DROP FOREIGN KEY identity_verifiable_addresses_ibfk_1;

ALTER TABLE identity_verifiable_addresses
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (id),
    ADD CONSTRAINT identity_verifiable_addresses_ibfk_1 FOREIGN KEY (identity_id) REFERENCES identities(id) ON DELETE CASCADE;
