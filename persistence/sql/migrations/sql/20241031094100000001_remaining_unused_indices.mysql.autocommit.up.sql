CREATE INDEX session_devices_nid_idx ON session_devices (nid ASC);
CREATE INDEX session_devices_session_id_idx ON session_devices (session_id ASC);
-- DROP INDEXsession_devices_id_nid_idx ON session_devices;
-- DROP INDEXsession_devices_session_id_nid_idx ON session_devices;

CREATE INDEX session_token_exchanges_flow_id_nid_init_code_idx ON session_token_exchanges (flow_id ASC, nid ASC, init_code ASC);
CREATE INDEX session_token_exchanges_nid_init_code_idx ON session_token_exchanges (nid ASC, init_code ASC);
-- DROP INDEXsession_token_exchanges_nid_code_idx ON session_token_exchanges;
-- DROP INDEXsession_token_exchanges_nid_flow_id_idx ON session_token_exchanges;

CREATE INDEX courier_messages_status_id_idx ON courier_messages (status ASC, id ASC);
CREATE INDEX courier_messages_nid_id_created_at_idx ON courier_messages (nid ASC, id ASC, created_at DESC);
-- DROP INDEXcourier_messages_status_idx ON courier_messages;
-- DROP INDEXcourier_messages_nid_id_idx ON courier_messages;
-- DROP INDEXcourier_messages_id_nid_idx ON courier_messages;
-- DROP INDEXcourier_messages_nid_created_at_id_idx ON courier_messages;

CREATE INDEX continuity_containers_identity_id_idx ON continuity_containers (identity_id ASC);
CREATE INDEX continuity_containers_nid_idx ON continuity_containers (nid ASC);
-- DROP INDEXcontinuity_containers_nid_id_idx ON continuity_containers;
-- DROP INDEXcontinuity_containers_id_nid_idx ON continuity_containers;
-- DROP INDEXcontinuity_containers_identity_id_nid_idx ON continuity_containers;

CREATE INDEX identity_verification_codes_verifiable_address_idx ON identity_verification_codes (identity_verifiable_address_id ASC);
CREATE INDEX identity_verification_codes_nid_idx ON identity_verification_codes (nid ASC);
-- DROP INDEXidentity_verification_codes_nid_flow_id_idx ON identity_verification_codes;
-- DROP INDEXidentity_verification_codes_id_nid_idx ON identity_verification_codes;
-- DROP INDEXidentity_verification_codes_verifiable_address_nid_idx ON identity_verification_codes;

CREATE INDEX identity_verification_tokens_nid_idx ON identity_verification_tokens (nid ASC);
-- DROP INDEXidentity_verification_tokens_nid_id_idx ON identity_verification_tokens;
-- DROP INDEXidentity_verification_tokens_id_nid_idx ON identity_verification_tokens;
-- DROP INDEXidentity_verification_tokens_token_nid_used_flow_id_idx ON identity_verification_tokens;

CREATE INDEX identity_registration_codes_nid_idx ON identity_registration_codes (nid ASC);
-- DROP INDEXidentity_registration_codes_nid_flow_id_idx ON identity_registration_codes;
-- DROP INDEXidentity_registration_codes_id_nid_idx ON identity_registration_codes;

CREATE INDEX identity_recovery_tokens_identity_id_idx ON identity_recovery_tokens (identity_id ASC);
CREATE INDEX identity_recovery_tokens_nid_idx ON identity_recovery_tokens (nid ASC);
-- DROP INDEXidentity_recovery_tokens_nid_id_idx ON identity_recovery_tokens;
-- DROP INDEXidentity_recovery_tokens_id_nid_idx ON identity_recovery_tokens;
-- DROP INDEXidentity_recovery_tokens_token_nid_used_idx ON identity_recovery_tokens;
-- DROP INDEXidentity_recovery_tokens_identity_id_nid_idx ON identity_recovery_tokens;

CREATE INDEX identity_recovery_codes_address_id_idx ON identity_recovery_codes (identity_recovery_address_id ASC);
CREATE INDEX identity_recovery_codes_identity_id_idx ON identity_recovery_codes (identity_id ASC);
CREATE INDEX identity_recovery_codes_nid_idx ON identity_recovery_codes (nid ASC);
-- DROP INDEXidentity_recovery_codes_nid_flow_id_idx ON identity_recovery_codes;
-- DROP INDEXidentity_recovery_codes_id_nid_idx ON identity_recovery_codes;
-- DROP INDEXidentity_recovery_codes_identity_id_nid_idx ON identity_recovery_codes;
-- DROP INDEXidentity_recovery_codes_identity_recovery_address_id_nid_idx ON identity_recovery_codes;

CREATE INDEX identity_login_codes_nid_idx ON identity_login_codes (nid ASC);
-- DROP INDEXidentity_login_codes_nid_flow_id_idx ON identity_login_codes;
-- DROP INDEXidentity_login_codes_id_nid_idx ON identity_login_codes;
