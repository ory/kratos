CREATE INDEX identity_credential_identifiers_nid_ici_idx
    ON identity_credentials_identifiers (nid ASC, identity_credential_id ASC) STORING (identifier);
