CREATE TABLE "_identity_verification_tokens_tmp" (
"id" TEXT PRIMARY KEY,
"token" TEXT NOT NULL,
"used" bool NOT NULL DEFAULT 'false',
"used_at" DATETIME,
"expires_at" DATETIME NOT NULL,
"issued_at" DATETIME NOT NULL,
"identity_verifiable_address_id" char(36) NOT NULL,
"selfservice_verification_flow_id" char(36),
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
"nid" char(36),
FOREIGN KEY (selfservice_verification_flow_id) REFERENCES selfservice_verification_flows (id) ON UPDATE NO ACTION ON DELETE CASCADE,
FOREIGN KEY (identity_verifiable_address_id) REFERENCES identity_verifiable_addresses (id) ON UPDATE NO ACTION ON DELETE CASCADE
);