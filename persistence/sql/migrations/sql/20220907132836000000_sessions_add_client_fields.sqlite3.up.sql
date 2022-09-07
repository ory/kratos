ALTER TABLE sessions
  ADD client_ip_address TEXT DEFAULT '';
ALTER TABLE sessions
  ADD user_agent TEXT DEFAULT '';
ALTER TABLE sessions
  ADD geo_location TEXT DEFAULT '';
