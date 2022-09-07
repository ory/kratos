ALTER TABLE sessions
  ADD COLUMN client_ip_address STRING,
  ADD COLUMN user_agent        STRING,
  ADD COLUMN geo_location      STRING;
