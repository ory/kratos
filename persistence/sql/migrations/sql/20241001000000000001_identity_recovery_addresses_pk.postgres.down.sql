ALTER TABLE identity_recovery_addresses
    DROP CONSTRAINT identity_recovery_addresses_pkey,
    ADD PRIMARY KEY (id);
