ALTER TABLE sessions
  ADD COLUMN last_location_lat DECIMAL(10, 8),
  ADD COLUMN last_location_lon DECIMAL(11, 8),
  ADD COLUMN last_location_at TIMESTAMP NULL DEFAULT NULL,
  ADD COLUMN impossible_travel BOOLEAN NOT NULL DEFAULT FALSE;
