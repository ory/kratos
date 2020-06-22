ALTER TABLE "selfservice_verification_requests" RENAME TO "_selfservice_verification_requests_tmp";
CREATE TABLE "selfservice_verification_requests" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"form" TEXT NOT NULL,
"via" TEXT NOT NULL,
"csrf_token" TEXT NOT NULL,
"success" bool NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL
);
INSERT INTO "selfservice_verification_requests" (id, request_url, issued_at, expires_at, form, via, csrf_token, success, created_at, updated_at) SELECT id, request_url, issued_at, expires_at, form, via, csrf_token, success, created_at, updated_at FROM "_selfservice_verification_requests_tmp";
DROP TABLE "_selfservice_verification_requests_tmp";
ALTER TABLE "selfservice_login_requests" RENAME TO "_selfservice_login_requests_tmp";
CREATE TABLE "selfservice_login_requests" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"active_method" TEXT NOT NULL,
"csrf_token" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"forced" bool NOT NULL DEFAULT 'false'
);
INSERT INTO "selfservice_login_requests" (id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, forced) SELECT id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, forced FROM "_selfservice_login_requests_tmp";
DROP TABLE "_selfservice_login_requests_tmp";
ALTER TABLE "selfservice_registration_requests" RENAME TO "_selfservice_registration_requests_tmp";
CREATE TABLE "selfservice_registration_requests" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"active_method" TEXT NOT NULL,
"csrf_token" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL
);
INSERT INTO "selfservice_registration_requests" (id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at) SELECT id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at FROM "_selfservice_registration_requests_tmp";
DROP TABLE "_selfservice_registration_requests_tmp";