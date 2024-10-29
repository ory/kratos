-- THIS IS COCKROACH ONLY
ALTER INDEX identity_credential_identifiers_identifier_nid_type_uq_idx RENAME TO identity_credential_identifiers_identifier_nid_type_uq_idx_deleteme;
CREATE UNIQUE INDEX IF NOT EXISTS identity_credential_identifiers_identifier_nid_type_uq_idx ON identity_credential_identifiers(nid ASC, identity_credential_type_id ASC, identifier ASC);
DROP INDEX IF EXISTS identity_credential_identifiers_identifier_nid_type_uq_idx_deleteme;
--

CREATE INDEX IF NOT EXISTS identity_credential_identifiers_nid_id_idx ON identity_credential_identifiers (nid ASC, id ASC);
CREATE INDEX IF NOT EXISTS identity_credential_identifiers_id_nid_idx ON identity_credential_identifiers (id ASC, nid ASC);
CREATE INDEX IF NOT EXISTS identity_credential_identifiers_nid_identity_credential_id_idx ON identity_credential_identifiers (identity_credential_id ASC, nid ASC);

DROP INDEX IF EXISTS identity_credential_identifiers_identity_credential_id_idx;
