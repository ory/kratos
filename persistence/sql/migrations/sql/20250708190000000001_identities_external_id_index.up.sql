CREATE UNIQUE INDEX IF NOT EXISTS identities_nid_external_id_idx
    ON identities (nid, external_id) WHERE external_id IS NOT NULL;