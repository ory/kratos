CREATE TABLE "_identities_tmp" (
"id" TEXT PRIMARY KEY,
"schema_id" TEXT NOT NULL,
"traits" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"nid" char(36)
);