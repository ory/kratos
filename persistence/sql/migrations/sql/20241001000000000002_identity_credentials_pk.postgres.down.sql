ALTER TABLE identity_credentials
    DROP CONSTRAINT identity_credentials_pkey,
    ADD PRIMARY KEY (id);
