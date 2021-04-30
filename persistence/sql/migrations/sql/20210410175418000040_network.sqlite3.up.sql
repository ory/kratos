CREATE TABLE "_continuity_containers_tmp" (
"id" TEXT PRIMARY KEY,
"identity_id" char(36),
"name" TEXT NOT NULL,
"payload" TEXT,
"expires_at" DATETIME NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"nid" char(36),
FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE NO ACTION ON DELETE CASCADE
);