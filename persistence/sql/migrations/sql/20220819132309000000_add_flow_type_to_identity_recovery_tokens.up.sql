ALTER TABLE identity_recovery_tokens
ADD token_type int;

UPDATE identity_recovery_tokens
SET token_type = 1
WHERE selfservice_recovery_flow_id IS NULL;
