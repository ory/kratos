CREATE TABLE "selfservice_profile_management_request_methods" (
"id" TEXT PRIMARY KEY,
"method" TEXT NOT NULL,
"selfservice_profile_management_request_id" char(36) NOT NULL,
"config" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL
);
ALTER TABLE "selfservice_profile_management_requests" ADD COLUMN "active_method" TEXT;
INSERT INTO selfservice_profile_management_request_methods (id, method, selfservice_profile_management_request_id, config) SELECT id, 'traits', id, form FROM selfservice_profile_management_requests;
ALTER TABLE "selfservice_profile_management_requests" RENAME TO "_selfservice_profile_management_requests_tmp";
CREATE TABLE "selfservice_profile_management_requests" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"update_successful" bool NOT NULL,
"identity_id" char(36) NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"active_method" TEXT
);
INSERT INTO "selfservice_profile_management_requests" (id, request_url, issued_at, expires_at, update_successful, identity_id, created_at, updated_at, active_method) SELECT id, request_url, issued_at, expires_at, update_successful, identity_id, created_at, updated_at, active_method FROM "_selfservice_profile_management_requests_tmp";
DROP TABLE "_selfservice_profile_management_requests_tmp";