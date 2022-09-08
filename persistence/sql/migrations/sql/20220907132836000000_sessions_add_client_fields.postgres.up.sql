ALTER TABLE sessions
  ADD COLUMN client_ip_address VARCHAR(50) DEFAULT '',
  ADD COLUMN user_agent TEXT DEFAULT '',
  ADD COLUMN geo_location TEXT DEFAULT '';
