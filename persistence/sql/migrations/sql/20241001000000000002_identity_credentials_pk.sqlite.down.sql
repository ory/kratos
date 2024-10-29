CREATE TABLE "_identity_credentials_tmp" (
"id" TEXT PRIMARY KEY,
"config" TEXT NOT NULL,
"identity_credential_type_id" char(36) NOT NULL,
"identity_id" char(36) NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"nid" char(36), version INT NOT NULL DEFAULT '0',
FOREIGN KEY (identity_credential_type_id) REFERENCES identity_credential_types (id) ON UPDATE NO ACTION ON DELETE CASCADE,
FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE NO ACTION ON DELETE CASCADE
);

INSERT INTO "_identity_credentials_tmp"
    ("id", "config", "identity_credential_type_id", "identity_id", "created_at", "updated_at", "nid", "version")
SELECT 
    "id", "config", "identity_credential_type_id", "identity_id", "created_at", "updated_at", "nid", "version"
FROM "identity_credentials";

DROP TABLE "identity_credentials";
ALTER TABLE "_identity_credentials_tmp" RENAME TO "identity_credentials";

CREATE INDEX identity_credentials_nid_id_idx ON identity_credentials (nid, id);
CREATE INDEX identity_credentials_id_nid_idx ON identity_credentials (id, nid);
CREATE UNIQUE INDEX identity_credentials_id_uq_idx ON identity_credentials (id);
