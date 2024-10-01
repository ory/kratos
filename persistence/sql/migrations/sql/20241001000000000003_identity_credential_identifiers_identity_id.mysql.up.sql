ALTER TABLE identity_credential_identifiers ADD COLUMN identity_id char(36) NULL;

UPDATE
    identity_credential_identifiers ici
JOIN
    identity_credentials ic
    ON ici.identity_credential_id = ic.id
    AND ici.nid = ic.nid
SET
    ici.identity_id = ic.identity_id 
WHERE
    ici.identity_id IS NULL;

ALTER TABLE identity_credential_identifiers
    MODIFY identity_id char(36) NOT NULL,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY(identity_id, identity_credential_id, id);
