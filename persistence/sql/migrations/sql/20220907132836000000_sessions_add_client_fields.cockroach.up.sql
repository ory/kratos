ALTER TABLE sessions
  ADD COLUMN client_ip_address STRING DEFAULT '',
  ADD COLUMN user_agent        STRING DEFAULT '',
  ADD COLUMN geo_location      STRING DEFAULT '';
