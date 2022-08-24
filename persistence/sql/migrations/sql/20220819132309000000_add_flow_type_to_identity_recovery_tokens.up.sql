ALTER TABLE identity_recovery_tokens
ADD token_type int NOT NULL DEFAULT 0;

UPDATE identity_recovery_tokens
SET token_type = 1
WHERE selfservice_recovery_flow_id IS NULL;

UPDATE identity_recovery_tokens
SET token_type = 2
WHERE selfservice_recovery_flow_id IS NOT NULL;

ALTER TABLE identity_recovery_tokens ADD CONSTRAINT identity_recovery_tokens_token_type_ck CHECK (token_type = 1 OR token_type = 2);
