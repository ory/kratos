CREATE INDEX IF NOT EXISTS identity_credential_identifiers_nid_ici_i_idx
  ON identity_credential_identifiers (nid ASC, identity_credential_id ASC, identifier ASC);

CREATE INDEX IF NOT EXISTS identity_credential_identifiers_identity_credential_id_idx
  ON identity_credential_identifiers (identity_credential_id ASC);
