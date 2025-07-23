CREATE UNIQUE INDEX IF NOT EXISTS identities_nid_external_id_idx
    ON identities (external_id, nid) USING HASH WHERE external_id IS NOT NULL AND external_id != '';