ALTER TABLE session_devices
    DROP CONSTRAINT session_devices_pkey,
    ADD PRIMARY KEY (id),
    DROP COLUMN identity_id;
