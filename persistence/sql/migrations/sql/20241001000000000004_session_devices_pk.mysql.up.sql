ALTER TABLE session_devices ADD COLUMN identity_id char(36) NULL;

UPDATE
    session_devices sd
JOIN
    sessions s ON sd.session_id = s.id
SET
    sd.identity_id = s.identity_id 
WHERE
    sd.identity_id IS NULL;

ALTER TABLE session_devices
    MODIFY identity_id char(36) NOT NULL,
    DROP FOREIGN KEY session_devices_ibfk_1,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY(session_id(36), identity_id(36), id(36));
ALTER TABLE session_devices
    ADD CONSTRAINT session_devices_ibfk_1 FOREIGN KEY (nid,identity_id,session_id) REFERENCES sessions(nid,identity_id,id) ON DELETE CASCADE;
