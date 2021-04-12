CREATE TABLE "_identity_credentials_tmp" (
"id" TEXT PRIMARY KEY,
"config" TEXT NOT NULL,
"identity_credential_type_id" char(36) NOT NULL,
"identity_id" char(36) NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE NO ACTION ON DELETE CASCADE,
FOREIGN KEY (identity_credential_type_id) REFERENCES identity_credential_types (id) ON UPDATE NO ACTION ON DELETE CASCADE
);