ALTER TABLE session_devices
    DROP FOREIGN KEY session_devices_ibfk_1,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY(id),
    DROP COLUMN identity_id;
ALTER TABLE session_devices
    ADD CONSTRAINT session_devices_ibfk_1 FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE;
