-- For SQLite, we do all operations in a single migration for simplicity.

CREATE TABLE "_identity_credential_identifiers_tmp" (
"id" TEXT PRIMARY KEY,
"identifier" TEXT NOT NULL,
"identity_credential_id" char(36) NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"nid" char(36),
"identity_credential_type_id" char(36) NOT NULL,
FOREIGN KEY (identity_credential_id) REFERENCES identity_credentials (id) ON UPDATE NO ACTION ON DELETE CASCADE
);

INSERT INTO _identity_credential_identifiers_tmp (id, identifier, identity_credential_id, created_at, updated_at, nid, identity_credential_type_id)
    SELECT id, identifier, identity_credential_id, created_at, updated_at, nid, identity_credential_type_id
    FROM identity_credential_identifiers;

DROP TABLE identity_credential_identifiers;
ALTER TABLE "_identity_credential_identifiers_tmp" RENAME TO "identity_credential_identifiers";


CREATE UNIQUE INDEX "identity_credential_identifiers_identifier_nid_type_uq_idx" ON "identity_credential_identifiers" (nid, identity_credential_type_id, identifier);
CREATE INDEX identity_credential_identifiers_nid_i_ici_idx ON "identity_credential_identifiers" (nid, identifier, identity_credential_id);
CREATE INDEX identity_credential_identifiers_ici_nid_i_idx ON "identity_credential_identifiers" (identity_credential_id ASC, nid ASC, identifier ASC);


CREATE TABLE IF NOT EXISTS "_session_devices_tmp"
(
  "id"         UUID PRIMARY KEY NOT NULL,
  "ip_address" VARCHAR(50)  DEFAULT '',
  "user_agent" VARCHAR(512) DEFAULT '',
  "location"   VARCHAR(512) DEFAULT '',
  "nid"        UUID             NOT NULL,
  "session_id" UUID             NOT NULL,
  "created_at" timestamp        NOT NULL,
  "updated_at" timestamp        NOT NULL,
  CONSTRAINT "session_metadata_sessions_id_fk" FOREIGN KEY ("session_id") REFERENCES "sessions" ("id") ON DELETE cascade,
  CONSTRAINT "session_metadata_nid_fk" FOREIGN KEY ("nid") REFERENCES "networks" ("id") ON DELETE cascade,
  CONSTRAINT unique_session_device UNIQUE (nid, session_id, ip_address, user_agent)
);

INSERT INTO "_session_devices_tmp" (id, ip_address, user_agent, location, nid, session_id, created_at, updated_at)
    SELECT id, ip_address, user_agent, location, nid, session_id, created_at, updated_at
    FROM session_devices;

DROP TABLE session_devices;
ALTER TABLE "_session_devices_tmp" RENAME TO "session_devices";

CREATE INDEX session_devices_nid_idx ON session_devices (nid ASC);
CREATE INDEX session_devices_session_id_idx ON session_devices (session_id ASC);
