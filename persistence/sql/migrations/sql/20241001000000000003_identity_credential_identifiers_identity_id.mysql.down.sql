ALTER TABLE identity_credential_identifiers
    DROP PRIMARY KEY,
    ADD PRIMARY KEY(id);
