CREATE TABLE "selfservice_login_flow_methods" (
"id" TEXT PRIMARY KEY,
"method" TEXT NOT NULL,
"selfservice_login_flow_id" char(36) NOT NULL,
"config" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
FOREIGN KEY (selfservice_login_flow_id) REFERENCES selfservice_login_flow_methods (id) ON DELETE cascade
);
CREATE TABLE "_selfservice_login_flows_tmp" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"active_method" TEXT NOT NULL,
"csrf_token" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"forced" bool NOT NULL DEFAULT 'false',
"type" TEXT NOT NULL DEFAULT 'browser'
);
INSERT INTO "_selfservice_login_flows_tmp" (id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, forced, type) SELECT id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, forced, type FROM "selfservice_login_flows";

DROP TABLE "selfservice_login_flows";
ALTER TABLE "_selfservice_login_flows_tmp" RENAME TO "selfservice_login_flows";
ALTER TABLE "selfservice_login_flows" ADD COLUMN "messages" TEXT;