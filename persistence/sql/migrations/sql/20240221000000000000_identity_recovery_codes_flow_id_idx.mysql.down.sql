ALTER TABLE `identity_recovery_codes`
    DROP FOREIGN KEY `identity_recovery_codes_identity_id_fk`,
    ADD CONSTRAINT `identity_recovery_tokens_identity_id_fk` FOREIGN KEY (`identity_id`) REFERENCES `identities` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT;

ALTER TABLE `identity_login_codes`
   DROP FOREIGN KEY `identity_login_codes_identity_id_fk`;
