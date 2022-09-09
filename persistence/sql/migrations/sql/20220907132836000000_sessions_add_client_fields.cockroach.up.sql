CREATE TABLE "session_metadata"
(
  "id"         UUID      NOT NULL,
  PRIMARY KEY ("id"),
  "ip_address" STRING DEFAULT '',
  "user_agent" STRING DEFAULT '',
  "location"   STRING DEFAULT '',
  "session_id" UUID      NOT NULL,
  "nid"        UUID      NOT NULL,
  "created_at" timestamp NOT NULL,
  "last_seen"  timestamp NOT NULL,
  CONSTRAINT "session_metadata_sessions_id_fk" FOREIGN KEY ("session_id") REFERENCES "sessions" ("id") ON DELETE cascade,
  CONSTRAINT "session_metadata_nid_fk" FOREIGN KEY ("nid") REFERENCES "networks" ("id") ON DELETE cascade
);
