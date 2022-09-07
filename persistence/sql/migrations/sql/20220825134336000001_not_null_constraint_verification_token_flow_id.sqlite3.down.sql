ALTER TABLE identity_verification_tokens
RENAME TO identity_verification_tokens_;

CREATE TABLE "identity_verification_tokens" (
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

DROP INDEX identity_verification_tokens_id_nid_idx;
DROP INDEX identity_verification_tokens_nid_id_idx;
DROP INDEX identity_verification_tokens_token_nid_used_flow_id_idx;
DROP INDEX identity_verification_tokens_token_uq_idx;
DROP INDEX identity_verification_tokens_verifiable_address_id_idx;
DROP INDEX identity_verification_tokens_verification_flow_id_idx;

CREATE INDEX identity_verification_tokens_id_nid_idx ON identity_verification_tokens (id, nid);
CREATE INDEX identity_verification_tokens_nid_id_idx ON identity_verification_tokens (nid, id);
CREATE INDEX identity_verification_tokens_token_nid_used_idx ON identity_verification_tokens (nid, token, used);
CREATE UNIQUE INDEX "identity_verification_tokens_token_uq_idx" ON "identity_verification_tokens" (token);
CREATE INDEX "identity_verification_tokens_verifiable_address_id_idx" ON "identity_verification_tokens" (identity_verifiable_address_id);
CREATE INDEX "identity_verification_tokens_verification_flow_id_idx" ON "identity_verification_tokens" (selfservice_verification_flow_id);

INSERT INTO identity_verification_tokens
SELECT * FROM identity_verification_tokens_;

DROP TABLE identity_verification_tokens_;
