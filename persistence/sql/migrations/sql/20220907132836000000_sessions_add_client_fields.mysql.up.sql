ALTER TABLE sessions
  ADD COLUMN client_ip_address VARCHAR(50),
  ADD COLUMN user_agent VARCHAR(255) AFTER client_ip_address,
  ADD COLUMN geo_location VARCHAR(255) AFTER user_agent;
