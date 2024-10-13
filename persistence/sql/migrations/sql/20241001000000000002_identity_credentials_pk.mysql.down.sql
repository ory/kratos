ALTER TABLE identity_credentials
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (id(36));
