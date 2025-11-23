DROP INDEX IF EXISTS idx_session_devices_created;
ALTER TABLE sessions DROP COLUMN impossible_travel;
ALTER TABLE session_devices DROP COLUMN longitude;
ALTER TABLE session_devices DROP COLUMN latitude;
