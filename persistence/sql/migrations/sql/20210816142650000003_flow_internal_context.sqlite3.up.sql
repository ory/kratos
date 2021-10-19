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
"type" TEXT NOT NULL DEFAULT 'browser',
"ui" TEXT,
"nid" char(36),
"requested_aal" TEXT NOT NULL DEFAULT 'aal1',
"internal_context" TEXT NOT NULL
);