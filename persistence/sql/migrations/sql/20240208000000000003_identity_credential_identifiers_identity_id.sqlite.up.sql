CREATE TABLE "_identity_credential_identifiers_tmp" (
"id" TEXT NOT NULL,
"identifier" TEXT NOT NULL,
"identity_credential_id" TEXT NOT NULL,
"identity_id" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"nid" TEXT NOT NULL,
"identity_credential_type_id" TEXT NOT NULL,
PRIMARY KEY (identity_id, identity_credential_id, id),
FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE RESTRICT ON DELETE CASCADE,
FOREIGN KEY (identity_credential_id) REFERENCES identity_credentials (id) ON UPDATE RESTRICT ON DELETE CASCADE,
FOREIGN KEY (nid) REFERENCES networks (id) ON UPDATE RESTRICT ON DELETE CASCADE
);


INSERT INTO _identity_credential_identifiers_tmp (id, identifier, identity_credential_id, created_at, updated_at, nid, identity_credential_type_id, identity_id)
    SELECT ici.id, ici.identifier, ici.identity_credential_id, ici.created_at, ici.updated_at, ici.nid, ici.identity_credential_type_id, ic.identity_id
    FROM identity_credential_identifiers ici
        INNER JOIN identity_credentials ic ON ici.identity_credential_id = ic.id AND ici.nid = ic.nid;

DROP TABLE identity_credential_identifiers;
ALTER TABLE "_identity_credential_identifiers_tmp" RENAME TO "identity_credential_identifiers";

CREATE UNIQUE INDEX "identity_credential_identifiers_identifier_nid_type_uq_idx" ON "identity_credential_identifiers" (nid, identity_credential_type_id, identifier);
CREATE INDEX "identity_credential_identifiers_nid_identity_credential_id_idx" ON "identity_credential_identifiers" (identity_credential_id, nid);
CREATE UNIQUE INDEX "identity_credential_identifiers_id_uq_idx" ON "identity_credential_identifiers" (id);
CREATE INDEX identity_credential_identifiers_nid_id_idx ON identity_credential_identifiers (nid, id);
CREATE INDEX identity_credential_identifiers_id_nid_idx ON identity_credential_identifiers (id, nid);
CREATE INDEX identity_credential_identifiers_nid_i_ici_idx ON identity_credential_identifiers (nid, identifier, identity_credential_id);
