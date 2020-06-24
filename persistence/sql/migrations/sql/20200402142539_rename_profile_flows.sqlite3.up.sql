ALTER TABLE "selfservice_profile_management_request_methods" RENAME TO "_selfservice_profile_management_request_methods_tmp";
CREATE TABLE "selfservice_profile_management_request_methods" (
"id" TEXT PRIMARY KEY,
"method" TEXT NOT NULL,
"selfservice_settings_request_id" char(36) NOT NULL,
"config" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL
);
INSERT INTO "selfservice_profile_management_request_methods" (id, method, selfservice_settings_request_id, config, created_at, updated_at) SELECT id, method, selfservice_profile_management_request_id, config, created_at, updated_at FROM "_selfservice_profile_management_request_methods_tmp";
DROP TABLE "_selfservice_profile_management_request_methods_tmp";
ALTER TABLE "selfservice_profile_management_request_methods" RENAME TO "selfservice_settings_request_methods";
ALTER TABLE "selfservice_profile_management_requests" RENAME TO "selfservice_settings_requests";