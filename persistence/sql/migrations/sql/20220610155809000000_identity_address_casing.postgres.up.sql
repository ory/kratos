UPDATE identity_recovery_addresses SET value = LOWER(value) WHERE TRUE;
UPDATE identity_verifiable_addresses SET value = LOWER(value) WHERE TRUE;
