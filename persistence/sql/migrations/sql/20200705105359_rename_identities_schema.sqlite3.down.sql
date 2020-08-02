ALTER TABLE "identities" RENAME TO "_identities_tmp";
CREATE TABLE "identities" (
"id" TEXT PRIMARY KEY,
"traits_schema_id" TEXT NOT NULL,
"traits" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL
);
INSERT INTO "identities" (id, traits_schema_id, traits, created_at, updated_at) SELECT id, schema_id, traits, created_at, updated_at FROM "_identities_tmp";
DROP TABLE "_identities_tmp";