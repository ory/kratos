CREATE TABLE "session_logs"
(
  "id"         UUID      NOT NULL,
  PRIMARY KEY ("id"),
  "ip_address" VARCHAR(50) DEFAULT '',
  "user_agent" TEXT        DEFAULT '',
  "location"   TEXT        DEFAULT '',
  "session_id" UUID      NOT NULL,
  "created_at" timestamp NOT NULL,
  FOREIGN KEY ("session_id") REFERENCES "sessions" ("id") ON DELETE cascade
);
