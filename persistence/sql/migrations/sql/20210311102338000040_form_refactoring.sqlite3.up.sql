CREATE TABLE "_selfservice_recovery_flows_tmp" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"active_method" TEXT,
"csrf_token" TEXT NOT NULL,
"state" TEXT NOT NULL,
"recovered_identity_id" char(36),
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"type" TEXT NOT NULL DEFAULT 'browser',
"ui" TEXT,
FOREIGN KEY (recovered_identity_id) REFERENCES identities (id) ON UPDATE NO ACTION ON DELETE CASCADE
);