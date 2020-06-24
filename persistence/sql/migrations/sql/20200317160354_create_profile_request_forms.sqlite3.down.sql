DROP TABLE "selfservice_profile_management_request_methods";
ALTER TABLE "selfservice_profile_management_requests" RENAME TO "_selfservice_profile_management_requests_tmp";
CREATE TABLE "selfservice_profile_management_requests" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"identity_id" char(36) NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"update_successful" bool NOT NULL DEFAULT 'false',
FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE NO ACTION ON DELETE CASCADE
);
INSERT INTO "selfservice_profile_management_requests" (id, request_url, issued_at, expires_at, identity_id, created_at, updated_at, update_successful) SELECT id, request_url, issued_at, expires_at, identity_id, created_at, updated_at, update_successful FROM "_selfservice_profile_management_requests_tmp";
DROP TABLE "_selfservice_profile_management_requests_tmp";