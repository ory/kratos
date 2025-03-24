CREATE TABLE IF NOT EXISTS scim_groups
(
    id              UUID PRIMARY KEY,
    nid             UUID         NOT NULL,
    organization_id UUID         NULL,
    parent_id       UUID         NULL,
    display_name    VARCHAR(255) NOT NULL,
    created_at      TIMESTAMP    NOT NULL,
    updated_at      TIMESTAMP    NOT NULL,

    FOREIGN KEY (nid) REFERENCES networks (id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES scim_groups (id) ON DELETE SET NULL
);

CREATE INDEX scim_groups_id_nid_idx
    ON scim_groups (id, nid);

CREATE INDEX scim_groups_parent_id_nid_idx
    ON scim_groups (parent_id, nid);

CREATE TABLE IF NOT EXISTS scim_groups_members
(
    group_id    UUID NOT NULL,
    identity_id UUID NOT NULL,

    PRIMARY KEY (identity_id, group_id),

    FOREIGN KEY (group_id) REFERENCES scim_groups (id) ON DELETE CASCADE,
    FOREIGN KEY (identity_id) REFERENCES identities (id) ON DELETE CASCADE
);

