ALTER INDEX identity_credential_identifiers_identifier_nid_type_uq_idx RENAME TO ici_tmp;

CREATE UNIQUE INDEX
  identity_credential_identifiers_identifier_nid_type_uq_idx
  ON identity_credential_identifiers(
    nid ASC,
    identity_credential_type_id ASC,
    identifier ASC
  );

DROP INDEX ici_tmp;
