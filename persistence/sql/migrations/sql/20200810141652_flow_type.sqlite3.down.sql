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
"forced" bool NOT NULL DEFAULT 'false',
"messages" TEXT
);
INSERT INTO "selfservice_login_requests" (id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, forced, messages) SELECT id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, forced, messages FROM "_selfservice_login_requests_tmp";
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
"updated_at" DATETIME NOT NULL,
"messages" TEXT
);
INSERT INTO "selfservice_registration_requests" (id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, messages) SELECT id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, messages FROM "_selfservice_registration_requests_tmp";
DROP TABLE "_selfservice_registration_requests_tmp";
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
"state" TEXT NOT NULL DEFAULT 'show_form',
FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE NO ACTION ON DELETE CASCADE
);
INSERT INTO "selfservice_settings_requests" (id, request_url, issued_at, expires_at, identity_id, created_at, updated_at, active_method, messages, state) SELECT id, request_url, issued_at, expires_at, identity_id, created_at, updated_at, active_method, messages, state FROM "_selfservice_settings_requests_tmp";
DROP TABLE "_selfservice_settings_requests_tmp";
ALTER TABLE "selfservice_recovery_requests" RENAME TO "_selfservice_recovery_requests_tmp";
CREATE TABLE "selfservice_recovery_requests" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"messages" TEXT,
"active_method" TEXT,
"csrf_token" TEXT NOT NULL,
"state" TEXT NOT NULL,
"recovered_identity_id" char(36),
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
FOREIGN KEY (recovered_identity_id) REFERENCES identities (id) ON UPDATE NO ACTION ON DELETE CASCADE
);
INSERT INTO "selfservice_recovery_requests" (id, request_url, issued_at, expires_at, messages, active_method, csrf_token, state, recovered_identity_id, created_at, updated_at) SELECT id, request_url, issued_at, expires_at, messages, active_method, csrf_token, state, recovered_identity_id, created_at, updated_at FROM "_selfservice_recovery_requests_tmp";
DROP TABLE "_selfservice_recovery_requests_tmp";
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
"updated_at" DATETIME NOT NULL,
"messages" TEXT
);
INSERT INTO "selfservice_verification_requests" (id, request_url, issued_at, expires_at, form, via, csrf_token, success, created_at, updated_at, messages) SELECT id, request_url, issued_at, expires_at, form, via, csrf_token, success, created_at, updated_at, messages FROM "_selfservice_verification_requests_tmp";
DROP TABLE "_selfservice_verification_requests_tmp";