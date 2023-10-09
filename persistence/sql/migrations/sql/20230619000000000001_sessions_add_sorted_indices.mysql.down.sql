CREATE INDEX sessions_identity_id_nid_idx
  ON sessions (identity_id, nid);

DROP INDEX sessions_identity_id_nid_sorted_idx
  ON sessions;
