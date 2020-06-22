ALTER TABLE "selfservice_errors" RENAME TO "_selfservice_errors_tmp";
CREATE TABLE "selfservice_errors" (
"id" TEXT PRIMARY KEY,
"errors" TEXT NOT NULL,
"seen_at" DATETIME,
"was_seen" bool NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL
);
INSERT INTO "selfservice_errors" (id, errors, seen_at, was_seen, created_at, updated_at) SELECT id, errors, seen_at, was_seen, created_at, updated_at FROM "_selfservice_errors_tmp";
DROP TABLE "_selfservice_errors_tmp";