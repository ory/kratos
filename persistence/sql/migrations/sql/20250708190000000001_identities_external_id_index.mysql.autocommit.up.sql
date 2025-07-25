CREATE UNIQUE INDEX identities_nid_external_id_idx
    ON identities (nid, external_id);