CREATE TABLE "session_logs"
(
  "id"         TEXT PRIMARY KEY,
  "ip_address" TEXT DEFAULT '',
  "user_agent" TEXT DEFAULT '',
  "location"   TEXT DEFAULT '',
  "session_id" UUID      NOT NULL,
  "created_at" timestamp NOT NULL,
  FOREIGN KEY ("session_id") REFERENCES "sessions" ("id") ON DELETE cascade
);
