CREATE UNIQUE INDEX identity_credentials_id_uq_idx ON identity_credentials(id);

ALTER TABLE identity_credential_identifiers
    DROP CONSTRAINT identity_credential_identifiers_identity_credential_id_fkey;

ALTER TABLE identity_credentials
    DROP CONSTRAINT identity_credentials_pkey,
    ADD PRIMARY KEY (identity_id, id);

ALTER TABLE identity_credential_identifiers
    ADD CONSTRAINT identity_credential_identifiers_identity_credential_id_fkey FOREIGN KEY (identity_credential_id) REFERENCES identity_credentials(id) ON DELETE CASCADE;
