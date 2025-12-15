ALTER TABLE identity_credential_identifiers
    DROP CONSTRAINT identity_credential_identifiers_identities_id_fk,
    ALTER identity_id DROP NOT NULL;

ALTER TABLE session_devices
    DROP CONSTRAINT session_devices_identities_id_fk,
    ALTER identity_id DROP NOT NULL;
