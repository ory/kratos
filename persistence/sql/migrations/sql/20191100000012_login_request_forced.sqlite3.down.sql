ALTER TABLE "selfservice_login_requests" RENAME TO "_selfservice_login_requests_tmp";
CREATE TABLE "selfservice_login_requests" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"active_method" TEXT NOT NULL,
"csrf_token" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL
);
INSERT INTO "selfservice_login_requests" (id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at) SELECT id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at FROM "_selfservice_login_requests_tmp";
DROP TABLE "_selfservice_login_requests_tmp";