CREATE TABLE IF NOT EXISTS "_identity_credential_identifiers_tmp" (
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
CREATE INDEX "identity_credential_identifiers_nid_identity_credential_id_idx" ON "identity_credential_identifiers" (identity_credential_id, nid);
CREATE INDEX identity_credential_identifiers_nid_id_idx ON identity_credential_identifiers (nid, id);
CREATE INDEX identity_credential_identifiers_id_nid_idx ON identity_credential_identifiers (id, nid);
CREATE INDEX identity_credential_identifiers_nid_i_ici_idx ON identity_credential_identifiers (nid, identifier, identity_credential_id);
