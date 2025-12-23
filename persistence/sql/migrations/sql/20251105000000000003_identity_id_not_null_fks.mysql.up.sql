ALTER TABLE identity_credential_identifiers
    MODIFY identity_id char(36) NOT NULL,
    ADD CONSTRAINT `identity_credential_identifiers_identity_id_fk` FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE RESTRICT ON DELETE CASCADE;

ALTER TABLE session_devices
    MODIFY identity_id char(36) NOT NULL,
    ADD CONSTRAINT `session_devices_identity_id_fk` FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE RESTRICT ON DELETE CASCADE;
