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