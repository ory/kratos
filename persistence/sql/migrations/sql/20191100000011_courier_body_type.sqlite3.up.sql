ALTER TABLE "courier_messages" RENAME TO "_courier_messages_tmp";
CREATE TABLE "courier_messages" (
"id" TEXT PRIMARY KEY,
"type" INTEGER NOT NULL,
"status" INTEGER NOT NULL,
"body" TEXT NOT NULL,
"subject" TEXT NOT NULL,
"recipient" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL
);
INSERT INTO "courier_messages" (id, type, status, body, subject, recipient, created_at, updated_at) SELECT id, type, status, body, subject, recipient, created_at, updated_at FROM "_courier_messages_tmp";
DROP TABLE "_courier_messages_tmp";