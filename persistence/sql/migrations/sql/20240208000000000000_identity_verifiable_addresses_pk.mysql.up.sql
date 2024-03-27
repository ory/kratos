ALTER TABLE identity_verifiable_addresses
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (identity_id, id);
