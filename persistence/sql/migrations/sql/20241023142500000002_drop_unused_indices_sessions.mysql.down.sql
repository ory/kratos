CREATE INDEX sessions_nid_id_identity_id_idx ON sessions(nid ASC, identity_id ASC, id ASC);
CREATE INDEX sessions_id_nid_idx ON sessions(id ASC, nid ASC);
CREATE INDEX sessions_token_nid_idx ON sessions(nid ASC, token ASC);
CREATE INDEX sessions_identity_id_nid_sorted_idx ON sessions(identity_id ASC, nid ASC, authenticated_at DESC);
CREATE INDEX sessions_nid_created_at_id_idx ON sessions(nid ASC, created_at DESC, id ASC);

DROP INDEX sessions_list_idx ON sessions;
DROP INDEX sessions_list_active_idx ON sessions;
DROP INDEX sessions_list_identity_idx ON sessions;
