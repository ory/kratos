UPDATE identity_recovery_addresses SET value = LOWER(value) WHERE TRUE;
UPDATE identity_verification_addresses SET value = LOWER(value) WHERE TRUE;
