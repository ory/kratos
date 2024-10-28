CREATE INDEX IF NOT EXISTS identity_credentials_identity_id_idx ON identity_credentials (identity_id ASC);
CREATE INDEX IF NOT EXISTS identity_credentials_nid_idx ON identity_credentials (nid ASC);

DROP INDEX IF EXISTS identity_credentials_id_nid_idx;
DROP INDEX IF EXISTS identity_credentials_nid_id_idx;
DROP INDEX IF EXISTS identity_credentials_nid_identity_id_idx;
