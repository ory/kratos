ALTER TABLE session_devices ADD COLUMN identity_id uuid NULL;

CREATE UNIQUE INDEX session_devices_id_uq_idx ON session_devices(id);

UPDATE
    session_devices sd
SET
    identity_id = s.identity_id
FROM
    sessions s
WHERE
    sd.session_id = s.id
    AND sd.nid = s.nid
    AND sd.identity_id IS NULL;

ALTER TABLE session_devices
    ALTER COLUMN identity_id SET NOT NULL,
    DROP CONSTRAINT session_devices_pkey,
    ADD PRIMARY KEY (session_id, identity_id, id);
