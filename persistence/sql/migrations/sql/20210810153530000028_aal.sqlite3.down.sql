CREATE TABLE "_sessions_tmp" (
"id" TEXT PRIMARY KEY,
"issued_at" DATETIME NOT NULL DEFAULT 'CURRENT_TIMESTAMP',
"expires_at" DATETIME NOT NULL,
"authenticated_at" DATETIME NOT NULL,
"identity_id" char(36) NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"token" TEXT,
"active" NUMERIC DEFAULT 'false',
"nid" char(36),
"logout_token" TEXT,
"aal" TEXT NOT NULL DEFAULT 'aal1',
FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE NO ACTION ON DELETE CASCADE
);