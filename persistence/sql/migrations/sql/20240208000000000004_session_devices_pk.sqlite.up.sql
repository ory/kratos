CREATE TABLE IF NOT EXISTS "_session_devices_tmp"
(
  "session_id"  TEXT             NOT NULL,
  "identity_id" TEXT             NOT NULL,
  "id"          TEXT             NOT NULL,
  "ip_address"  VARCHAR(50)  DEFAULT '',
  "user_agent"  VARCHAR(512) DEFAULT '',
  "location"    VARCHAR(512) DEFAULT '',
  "nid"         TEXT             NOT NULL,
  "created_at"  timestamp        NOT NULL,
  "updated_at"  timestamp        NOT NULL,
  PRIMARY KEY (session_id, identity_id, id),
  CONSTRAINT "session_metadata_session_id_fk" FOREIGN KEY ("session_id") REFERENCES "sessions" ("id") ON DELETE cascade,
  CONSTRAINT "session_metadata_identity_id_fk" FOREIGN KEY ("identity_id") REFERENCES "identities" ("id") ON DELETE cascade,
  CONSTRAINT "session_metadata_nid_fk" FOREIGN KEY ("nid") REFERENCES "networks" ("id") ON DELETE cascade,
  CONSTRAINT unique_session_device UNIQUE (nid, session_id, ip_address, user_agent),
  CONSTRAINT "unique_session_device_id" UNIQUE (id)
);

INSERT INTO _session_devices_tmp
    (id, identity_id, session_id, ip_address, user_agent, "location", nid, created_at, updated_at)
SELECT
    sd.id, s.identity_id, sd.session_id, sd.ip_address, sd.user_agent, sd.location, sd.nid, sd.created_at, sd.updated_at
FROM session_devices sd
    JOIN sessions s
        ON s.id = sd.session_id AND s.nid = sd.nid;

DROP TABLE "session_devices";
ALTER TABLE "_session_devices_tmp" RENAME TO "session_devices";
