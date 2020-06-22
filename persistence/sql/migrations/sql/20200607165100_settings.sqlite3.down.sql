ALTER TABLE "selfservice_settings_requests" RENAME TO "_selfservice_settings_requests_tmp";
CREATE TABLE "selfservice_settings_requests" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"identity_id" char(36) NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"active_method" TEXT,
"messages" TEXT,
FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE NO ACTION ON DELETE CASCADE
);
INSERT INTO "selfservice_settings_requests" (id, request_url, issued_at, expires_at, identity_id, created_at, updated_at, active_method, messages) SELECT id, request_url, issued_at, expires_at, identity_id, created_at, updated_at, active_method, messages FROM "_selfservice_settings_requests_tmp";
DROP TABLE "_selfservice_settings_requests_tmp";
ALTER TABLE "selfservice_settings_requests" ADD COLUMN "update_successful" bool NOT NULL DEFAULT 'false';