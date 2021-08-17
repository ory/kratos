ALTER TABLE "identity_credential_identifiers"
  ADD CONSTRAINT "identity_credential_identifiers_type_id_fk_idx" FOREIGN KEY ("identity_credential_type_id") REFERENCES "identity_credential_types" ("id") ON UPDATE RESTRICT ON DELETE CASCADE;
