DELETE FROM "identity_credential_identifiers"; -- This migration is destructive.
CREATE UNIQUE INDEX "identity_credential_identifiers_identifier_nid_uq_idx" ON "identity_credential_identifiers" (nid, identifier);
