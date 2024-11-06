CREATE INDEX IF NOT EXISTS identity_credential_identifiers_ici_nid_i_idx
    ON identity_credential_identifiers (identity_credential_id ASC, nid ASC, identifier ASC);

DROP INDEX IF EXISTS identity_credential_identifiers_identity_credential_id_idx;
