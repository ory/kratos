CREATE INDEX sessions_identity_id_nid_sorted_idx
  ON sessions (identity_id, nid, authenticated_at DESC);

DROP INDEX sessions_identity_id_nid_idx
  ON sessions;
