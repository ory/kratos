ALTER TABLE sessions
  ADD COLUMN client_ip_address VARCHAR(50),
  ADD COLUMN user_agent TEXT,
  ADD COLUMN geo_location TEXT;
