ALTER TABLE identity_credential_identifiers ADD COLUMN identity_id UUID NULL;

CREATE UNIQUE INDEX identity_credential_identifiers_id_uq_idx ON identity_credential_identifiers(id);

UPDATE
    identity_credential_identifiers ici
SET
    identity_id = ic.identity_id
FROM
    identity_credentials ic
WHERE
    ici.identity_credential_id = ic.id
    AND ici.nid = ic.nid
    AND ici.identity_id IS NULL;

ALTER TABLE identity_credential_identifiers
    ALTER identity_id SET NOT NULL,
    DROP CONSTRAINT identity_credential_identifiers_pkey,
    ADD PRIMARY KEY(identity_id, identity_credential_id, id);
