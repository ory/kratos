CREATE TABLE IF NOT EXISTS "session_devices"
(
  "id"         TEXT PRIMARY KEY,
  "ip_address" TEXT DEFAULT '',
  "user_agent" TEXT DEFAULT '',
  "location"   TEXT DEFAULT '',
  "session_id" UUID      NOT NULL,
  "nid"        UUID      NOT NULL,
  "created_at" timestamp NOT NULL,
  "updated_at"  timestamp NOT NULL,
  FOREIGN KEY ("session_id") REFERENCES "sessions" ("id") ON DELETE cascade,
  FOREIGN KEY ("nid") REFERENCES "networks" ("id") ON DELETE cascade,
  CONSTRAINT unique_session_device UNIQUE (nid, session_id, ip_address, user_agent)
);
