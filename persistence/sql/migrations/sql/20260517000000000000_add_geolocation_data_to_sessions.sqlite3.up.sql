ALTER TABLE sessions ADD COLUMN last_location_lat REAL;
ALTER TABLE sessions ADD COLUMN last_location_lon REAL;
ALTER TABLE sessions ADD COLUMN last_location_at DATETIME;
ALTER TABLE sessions ADD COLUMN impossible_travel BOOLEAN NOT NULL DEFAULT FALSE;
