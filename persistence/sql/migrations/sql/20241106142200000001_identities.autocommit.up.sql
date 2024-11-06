CREATE INDEX IF NOT EXISTS identity_credential_identifiers_nid_ici_idx
    ON identity_credential_identifiers (nid ASC, identity_credential_id ASC);
