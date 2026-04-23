DELETE FROM identity_verification_tokens WHERE identity_verifiable_address_id IS NULL;
ALTER TABLE identity_verification_tokens MODIFY identity_verifiable_address_id CHAR(36) NOT NULL;
