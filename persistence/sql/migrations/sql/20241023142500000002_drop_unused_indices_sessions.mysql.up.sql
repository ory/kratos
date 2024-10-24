CREATE INDEX sessions_nid_expires_at_authenticated_at_idx ON sessions (nid ASC, expires_at ASC, authenticated_at DESC); -- Used for listing sessions
CREATE INDEX sessions_identity_id_idx ON sessions (identity_at DESC); -- Used for listing sessions

DROP INDEX sessions_nid_id_identity_id_idx ON sessions;
DROP INDEX sessions_id_nid_idx ON sessions;
DROP INDEX sessions_token_nid_idx ON sessions;
DROP INDEX sessions_identity_id_nid_sorted_idx ON sessions;
DROP INDEX sessions_nid_created_at_id_idx ON sessions;
