CREATE INDEX IF NOT EXISTS sessions_nid_id_identity_id_idx ON sessions(nid ASC, identity_id ASC, id ASC);
CREATE INDEX IF NOT EXISTS sessions_id_nid_idx ON sessions(id ASC, nid ASC);
CREATE INDEX IF NOT EXISTS sessions_token_nid_idx ON sessions(nid ASC, token ASC);
CREATE INDEX IF NOT EXISTS sessions_identity_id_nid_sorted_idx ON sessions(identity_id ASC, nid ASC, authenticated_at DESC);
CREATE INDEX IF NOT EXISTS sessions_nid_created_at_id_idx ON sessions(nid ASC, created_at DESC, id ASC);

DROP INDEX IF EXISTS sessions_list_idx;
DROP INDEX IF EXISTS sessions_list_active_idx;
DROP INDEX IF EXISTS sessions_list_identity_idx;
