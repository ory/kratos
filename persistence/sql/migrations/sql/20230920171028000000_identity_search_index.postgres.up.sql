CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS btree_gin;

CREATE INDEX identity_credential_identifiers_nid_identifier_gin ON identity_credential_identifiers USING GIN (nid, identifier gin_trgm_ops);
