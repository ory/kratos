CREATE TABLE "sessions"
(
  "id"         UUID      NOT NULL,
  PRIMARY KEY ("id"),
  "ip_address" STRING DEFAULT '',
  "user_agent" STRING DEFAULT '',
  "location"   STRING DEFAULT '',
  "session_id" UUID      NOT NULL,
  "created_at" timestamp NOT NULL,
  CONSTRAINT "session_logs_sessions_id_fk" FOREIGN KEY ("session_id") REFERENCES "sessions" ("id") ON DELETE cascade
);
