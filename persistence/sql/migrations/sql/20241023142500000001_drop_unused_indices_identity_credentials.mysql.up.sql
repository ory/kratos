CREATE INDEX identity_credentials_nid_identity_id_v2_idx ON identity_credentials (nid ASC, identity_id ASC);

DROP INDEX identity_credentials_nid_id_idx ON identity_credentials;
DROP INDEX identity_credentials_id_nid_idx ON identity_credentials;
DROP INDEX identity_credentials_nid_identity_id_idx ON identity_credentials;
