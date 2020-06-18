ALTER TABLE "selfservice_settings_requests" ADD COLUMN "state" TEXT NOT NULL DEFAULT 'show_form';
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
"state" TEXT NOT NULL DEFAULT 'show_form'
);
INSERT INTO "selfservice_settings_requests" (id, request_url, issued_at, expires_at, identity_id, created_at, updated_at, active_method, messages, state) SELECT id, request_url, issued_at, expires_at, identity_id, created_at, updated_at, active_method, messages, state FROM "_selfservice_settings_requests_tmp";
DROP TABLE "_selfservice_settings_requests_tmp";