ALTER TABLE identity_verifiable_addresses
    DROP CONSTRAINT identity_verifiable_addresses_pkey,
    ADD PRIMARY KEY (id);
