INSERT INTO "_identity_credential_identifiers_tmp" (id, identifier, identity_credential_id, created_at, updated_at, nid,
                                                    identity_credential_type_id)
SELECT id, identifier, identity_credential_id, created_at, updated_at, nid, identity_credential_type_id
FROM "identity_credential_identifiers";
