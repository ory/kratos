ALTER TABLE session_devices ADD COLUMN latitude REAL;
ALTER TABLE session_devices ADD COLUMN longitude REAL;
ALTER TABLE sessions ADD COLUMN impossible_travel INTEGER NOT NULL DEFAULT 0;
-- optimize on: "get the latest device for session" queries
CREATE INDEX idx_session_devices_created
  ON session_devices(session_id, created_at DESC);

