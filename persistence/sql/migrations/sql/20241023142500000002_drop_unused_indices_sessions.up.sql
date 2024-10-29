CREATE INDEX IF NOT EXISTS sessions_list_idx ON sessions (nid ASC, created_at DESC, id ASC);
CREATE INDEX IF NOT EXISTS sessions_list_active_idx ON sessions (nid ASC, expires_at ASC, active ASC, created_at DESC, id ASC);
CREATE INDEX IF NOT EXISTS sessions_list_identity_idx ON sessions (identity_id ASC, nid ASC, created_at DESC);

DROP INDEX IF EXISTS sessions_nid_id_identity_id_idx;
DROP INDEX IF EXISTS sessions_id_nid_idx;
DROP INDEX IF EXISTS sessions_token_nid_idx;
DROP INDEX IF EXISTS sessions_identity_id_nid_sorted_idx;
DROP INDEX IF EXISTS sessions_nid_created_at_id_idx;
