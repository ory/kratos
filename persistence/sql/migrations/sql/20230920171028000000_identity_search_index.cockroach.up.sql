CREATE INDEX IF NOT EXISTS identity_credential_identifiers_nid_identifier_gin ON identity_credential_identifiers USING GIN (nid, identifier gin_trgm_ops);
