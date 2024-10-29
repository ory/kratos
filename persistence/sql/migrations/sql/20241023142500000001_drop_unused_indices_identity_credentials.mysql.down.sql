CREATE INDEX identity_credentials_id_nid_idx ON identity_credentials (id ASC, nid ASC);
CREATE INDEX identity_credentials_nid_id_idx ON identity_credentials (nid ASC, id ASC);

DROP INDEX identity_credentials_nid_idx ON identity_credentials;
