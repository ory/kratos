CREATE INDEX identity_credentials_identity_id_idx ON identity_credentials (identity_id ASC);
CREATE INDEX identity_credentials_nid_idx ON identity_credentials (nid ASC);

DROP INDEX identity_credentials_id_nid_idx ON identity_credentials;
DROP INDEX identity_credentials_nid_id_idx ON identity_credentials;
DROP INDEX identity_credentials_nid_identity_id_idx ON identity_credentials;
