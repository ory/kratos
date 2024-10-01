CREATE TABLE IF NOT EXISTS "_identity_verifiable_addresses_tmp" (
"id" TEXT PRIMARY KEY,
"status" TEXT NOT NULL,
"via" TEXT NOT NULL,
"verified" bool NOT NULL,
"value" TEXT NOT NULL,
"verified_at" DATETIME,
"identity_id" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"nid" TEXT NOT NULL,
FOREIGN KEY ("identity_id") REFERENCES "identities" ("id") ON UPDATE RESTRICT ON DELETE CASCADE,
FOREIGN KEY ("nid") REFERENCES "networks" ("id") ON UPDATE RESTRICT ON DELETE CASCADE
);

INSERT INTO "_identity_verifiable_addresses_tmp"
    ("id", "status", "via", "verified", "value", "verified_at", "identity_id", "created_at", "updated_at", "nid")
SELECT 
    "id", "status", "via", "verified", "value", "verified_at", "identity_id", "created_at", "updated_at", "nid"
FROM "identity_verifiable_addresses";

DROP TABLE "identity_verifiable_addresses";
ALTER TABLE "_identity_verifiable_addresses_tmp" RENAME TO "identity_verifiable_addresses";

CREATE UNIQUE INDEX IF NOT EXISTS "identity_verifiable_addresses_status_via_uq_idx" ON "identity_verifiable_addresses" (nid, via, value);
CREATE INDEX IF NOT EXISTS "identity_verifiable_addresses_status_via_idx" ON "identity_verifiable_addresses" (nid, via, value);
CREATE INDEX IF NOT EXISTS identity_recovery_addresses_nid_id_idx ON identity_recovery_addresses (nid, id);
