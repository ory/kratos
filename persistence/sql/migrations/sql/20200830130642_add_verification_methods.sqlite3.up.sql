ALTER TABLE "selfservice_verification_flows" ADD COLUMN "state" TEXT NOT NULL DEFAULT 'show_form';
UPDATE selfservice_verification_flows SET state='passed_challenge' WHERE success IS TRUE;
CREATE TABLE "selfservice_verification_flow_methods" (
"id" TEXT PRIMARY KEY,
"method" TEXT NOT NULL,
"selfservice_verification_flow_id" char(36) NOT NULL,
"config" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL
);
ALTER TABLE "selfservice_verification_flows" ADD COLUMN "active_method" TEXT;
INSERT INTO selfservice_verification_flow_methods (id, method, selfservice_verification_flow_id, config, created_at, updated_at) SELECT id, 'link', id, form, created_at, updated_at FROM selfservice_verification_flows;
CREATE TABLE "_selfservice_verification_flows_tmp" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"via" TEXT NOT NULL,
"csrf_token" TEXT NOT NULL,
"success" bool NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"messages" TEXT,
"type" TEXT NOT NULL DEFAULT 'browser',
"state" TEXT NOT NULL DEFAULT 'show_form',
"active_method" TEXT
);
INSERT INTO "_selfservice_verification_flows_tmp" (id, request_url, issued_at, expires_at, via, csrf_token, success, created_at, updated_at, messages, type, state, active_method) SELECT id, request_url, issued_at, expires_at, via, csrf_token, success, created_at, updated_at, messages, type, state, active_method FROM "selfservice_verification_flows";

DROP TABLE "selfservice_verification_flows";
ALTER TABLE "_selfservice_verification_flows_tmp" RENAME TO "selfservice_verification_flows";
CREATE TABLE "_selfservice_verification_flows_tmp" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"csrf_token" TEXT NOT NULL,
"success" bool NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"messages" TEXT,
"type" TEXT NOT NULL DEFAULT 'browser',
"state" TEXT NOT NULL DEFAULT 'show_form',
"active_method" TEXT
);
INSERT INTO "_selfservice_verification_flows_tmp" (id, request_url, issued_at, expires_at, csrf_token, success, created_at, updated_at, messages, type, state, active_method) SELECT id, request_url, issued_at, expires_at, csrf_token, success, created_at, updated_at, messages, type, state, active_method FROM "selfservice_verification_flows";

DROP TABLE "selfservice_verification_flows";
ALTER TABLE "_selfservice_verification_flows_tmp" RENAME TO "selfservice_verification_flows";
CREATE TABLE "_selfservice_verification_flows_tmp" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"csrf_token" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"messages" TEXT,
"type" TEXT NOT NULL DEFAULT 'browser',
"state" TEXT NOT NULL DEFAULT 'show_form',
"active_method" TEXT
);
INSERT INTO "_selfservice_verification_flows_tmp" (id, request_url, issued_at, expires_at, csrf_token, created_at, updated_at, messages, type, state, active_method) SELECT id, request_url, issued_at, expires_at, csrf_token, created_at, updated_at, messages, type, state, active_method FROM "selfservice_verification_flows";

DROP TABLE "selfservice_verification_flows";
ALTER TABLE "_selfservice_verification_flows_tmp" RENAME TO "selfservice_verification_flows";