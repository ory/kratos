ALTER TABLE identity_credentials
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (identity_id(36), id(36));
