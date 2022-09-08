ALTER TABLE sessions
  ADD COLUMN client_ip_address VARCHAR(50) DEFAULT '',
  ADD COLUMN user_agent VARCHAR(255) DEFAULT '' AFTER client_ip_address,
  ADD COLUMN geo_location VARCHAR(255) DEFAULT '' AFTER user_agent;
