ALTER TABLE identity_verifiable_addresses
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (identity_id, id),
    ADD UNIQUE KEY identity_verifiable_addresses_id_uq_idx (id);
