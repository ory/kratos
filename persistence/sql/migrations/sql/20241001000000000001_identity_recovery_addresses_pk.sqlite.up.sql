CREATE TABLE "_identity_recovery_addresses_tmp" (
"id" TEXT NOT NULL,
"via" TEXT NOT NULL,
"value" TEXT NOT NULL,
"identity_id" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"nid" TEXT NOT NULL,
PRIMARY KEY (identity_id,id),
FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE RESTRICT ON DELETE CASCADE,
FOREIGN KEY (nid) REFERENCES networks (id) ON UPDATE RESTRICT ON DELETE CASCADE
);

INSERT INTO "_identity_recovery_addresses_tmp"
    ("id", "via", "value", "identity_id", "created_at", "updated_at", "nid")
SELECT 
    "id", "via", "value", "identity_id", "created_at", "updated_at", "nid"
FROM "identity_recovery_addresses";

DROP TABLE "identity_recovery_addresses";
ALTER TABLE "_identity_recovery_addresses_tmp" RENAME TO "identity_recovery_addresses";

CREATE INDEX identity_recovery_addresses_nid_id_idx ON identity_recovery_addresses (nid, id);
CREATE UNIQUE INDEX "identity_recovery_addresses_id_uq_idx" ON "identity_recovery_addresses" (id);
CREATE UNIQUE INDEX "identity_recovery_addresses_status_via_uq_idx" ON "identity_recovery_addresses" (nid, via, value);
CREATE INDEX "identity_recovery_addresses_status_via_idx" ON "identity_recovery_addresses" (nid, via, value);
