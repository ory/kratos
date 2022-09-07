ALTER TABLE sessions
  DROP COLUMN client_ip_address,
  DROP COLUMN user_agent,
  DROP COLUMN geo_location;
