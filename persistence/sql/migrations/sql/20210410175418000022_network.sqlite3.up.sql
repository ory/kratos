CREATE TABLE "_selfservice_settings_flows_tmp" (
"id" TEXT PRIMARY KEY,
"request_url" TEXT NOT NULL,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"identity_id" char(36) NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"active_method" TEXT,
"state" TEXT NOT NULL DEFAULT 'show_form',
"type" TEXT NOT NULL DEFAULT 'browser',
"ui" TEXT,
"nid" char(36),
FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE NO ACTION ON DELETE CASCADE
);