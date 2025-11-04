ALTER TABLE identity_credential_identifiers
    DROP CONSTRAINT `identity_credential_identifiers_identity_id_fk`,
    MODIFY identity_id char(36) NULL;

ALTER TABLE session_devices
    DROP CONSTRAINT `session_devices_identity_id_fk`,
    MODIFY identity_id char(36) NULL;
