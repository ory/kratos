ALTER TABLE session_devices DROP FOREIGN KEY session_devices_ibfk_2;
ALTER TABLE session_devices DROP INDEX unique_session_device;
ALTER TABLE session_devices ADD CONSTRAINT session_devices_ibfk_2 FOREIGN KEY (nid) REFERENCES networks(id) ON DELETE CASCADE;
