CREATE UNIQUE INDEX IF NOT EXISTS unique_session_device ON session_devices (nid, session_id, ip_address, user_agent);
