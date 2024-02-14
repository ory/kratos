ALTER TABLE identity_credential_identifiers
    DROP CONSTRAINT identity_credential_identifiers_pkey,
    ADD PRIMARY KEY (id),
    DROP COLUMN identity_id;
