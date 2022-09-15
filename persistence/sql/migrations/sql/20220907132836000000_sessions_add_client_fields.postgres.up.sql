CREATE TABLE "session_devices"
(
  "id"         UUID      NOT NULL,
  PRIMARY KEY ("id"),
  "ip_address" VARCHAR(50) DEFAULT '',
  "user_agent" TEXT        DEFAULT '',
  "location"   TEXT        DEFAULT '',
  "nid"        UUID      NOT NULL,
  "session_id" UUID      NOT NULL,
  "created_at" timestamp NOT NULL,
  "updated_at"  timestamp NOT NULL,
  FOREIGN KEY ("session_id") REFERENCES "sessions" ("id") ON DELETE cascade,
  FOREIGN KEY ("nid") REFERENCES "networks" ("id") ON DELETE cascade,
  CONSTRAINT unique_session_device UNIQUE (nid, session_id, ip_address, user_agent)
);
