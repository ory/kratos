ALTER TABLE identity_recovery_tokens
ADD token_type int NOT NULL DEFAULT 0;

COMMIT TRANSACTION;BEGIN TRANSACTION;

UPDATE identity_recovery_tokens
SET token_type = 1
WHERE selfservice_recovery_flow_id IS NULL;

UPDATE identity_recovery_tokens
SET token_type = 2
WHERE selfservice_recovery_flow_id IS NOT NULL;
