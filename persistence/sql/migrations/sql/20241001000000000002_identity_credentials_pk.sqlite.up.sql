CREATE TABLE "_identity_credentials_tmp" (
"id" TEXT NOT NULL,
"config" TEXT NOT NULL,
"identity_credential_type_id" TEXT NOT NULL,
"identity_id" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"nid" TEXT NOT NULL,
"version" INT NOT NULL DEFAULT '0',
PRIMARY KEY (identity_id,id),
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
CREATE UNIQUE INDEX identity_credentials_id_uq_idx ON identity_credentials (id);
