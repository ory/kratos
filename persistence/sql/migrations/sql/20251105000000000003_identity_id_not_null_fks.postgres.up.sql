ALTER TABLE identity_credential_identifiers
    ALTER identity_id SET NOT NULL,
    ADD CONSTRAINT "identity_credential_identifiers_identities_id_fk" FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE RESTRICT ON DELETE CASCADE;


ALTER TABLE session_devices
    ALTER identity_id SET NOT NULL,
    ADD CONSTRAINT "session_devices_identities_id_fk" FOREIGN KEY (identity_id) REFERENCES identities (id) ON UPDATE RESTRICT ON DELETE CASCADE;
