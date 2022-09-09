ALTER TABLE identity_recovery_tokens 
ADD CONSTRAINT identity_recovery_tokens_token_type_ck CHECK (token_type = 1 OR token_type = 2);
