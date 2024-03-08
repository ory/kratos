-- This FK was previously misnamed.
ALTER TABLE `identity_recovery_codes`
    DROP FOREIGN KEY `identity_recovery_tokens_identity_id_fk`,
    ADD CONSTRAINT `identity_recovery_codes_identity_id_fk` FOREIGN KEY (`identity_id`) REFERENCES `identities` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT;

-- Missing FK
ALTER TABLE `identity_login_codes`
    ADD CONSTRAINT `identity_login_codes_identity_id_fk` FOREIGN KEY (`identity_id`) REFERENCES `identities` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT;

-- MySQL has created the remaining indices automatically together with the foreign key constraints.

-- CREATE INDEX identity_login_codes_identity_id_idx ON identity_login_codes (identity_id ASC);
-- CREATE INDEX identity_login_codes_flow_id_idx ON identity_login_codes (selfservice_login_flow_id ASC);
-- CREATE INDEX identity_registration_codes_flow_id_idx ON identity_registration_codes (selfservice_registration_flow_id ASC);
-- CREATE INDEX identity_recovery_codes_flow_id_idx ON identity_recovery_codes (selfservice_recovery_flow_id ASC);
-- CREATE INDEX identity_verification_codes_flow_id_idx ON identity_verification_codes (selfservice_verification_flow_id ASC);
