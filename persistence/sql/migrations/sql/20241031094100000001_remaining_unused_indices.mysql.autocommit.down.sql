CREATE INDEX session_devices_id_nid_idx ON session_devices (nid ASC, id ASC); -- the original index is id, nid - but then we can't drop session_devices_nid_idx
CREATE INDEX session_devices_session_id_nid_idx ON session_devices (session_id ASC, nid ASC);
DROP INDEX session_devices_nid_idx ON session_devices;
DROP INDEX session_devices_session_id_idx ON session_devices;

CREATE INDEX session_token_exchanges_nid_code_idx ON session_token_exchanges (init_code ASC, nid ASC);
CREATE INDEX session_token_exchanges_nid_flow_id_idx ON session_token_exchanges (flow_id ASC, nid ASC);
DROP INDEX session_token_exchanges_flow_id_nid_init_code_idx ON session_token_exchanges;
DROP INDEX session_token_exchanges_nid_init_code_idx ON session_token_exchanges;

CREATE INDEX courier_messages_status_idx ON courier_messages (status ASC);
CREATE INDEX courier_messages_nid_id_idx ON courier_messages (nid ASC, id ASC);
CREATE INDEX courier_messages_id_nid_idx ON courier_messages (id ASC, nid ASC);
CREATE INDEX courier_messages_nid_created_at_id_idx ON courier_messages (nid ASC, created_at DESC);
DROP INDEX courier_messages_status_id_idx ON courier_messages;
DROP INDEX courier_messages_nid_id_created_at_idx ON courier_messages;

CREATE INDEX continuity_containers_nid_id_idx ON continuity_containers (nid ASC, id ASC);
CREATE INDEX continuity_containers_id_nid_idx ON continuity_containers (id ASC, nid ASC);
CREATE INDEX continuity_containers_identity_id_nid_idx ON continuity_containers (identity_id ASC, nid ASC);
DROP INDEX continuity_containers_identity_id_idx ON continuity_containers;
DROP INDEX continuity_containers_nid_idx ON continuity_containers;

CREATE INDEX identity_verification_codes_nid_flow_id_idx ON identity_verification_codes (nid ASC, selfservice_verification_flow_id ASC);
CREATE INDEX identity_verification_codes_id_nid_idx ON identity_verification_codes (id ASC, nid ASC);
CREATE INDEX identity_verification_codes_verifiable_address_nid_idx ON identity_verification_codes (identity_verifiable_address_id ASC, nid ASC);
DROP INDEX identity_verification_codes_verifiable_address_idx ON identity_verification_codes;
DROP INDEX identity_verification_codes_nid_idx ON identity_verification_codes;

CREATE INDEX identity_verification_tokens_nid_id_idx ON identity_verification_tokens (nid ASC, id ASC);
CREATE INDEX identity_verification_tokens_id_nid_idx ON identity_verification_tokens (id ASC, nid ASC);
CREATE INDEX identity_verification_tokens_token_nid_used_flow_id_idx ON identity_verification_tokens (nid ASC, token ASC, used ASC, selfservice_verification_flow_id ASC);
DROP INDEX identity_verification_tokens_nid_idx ON identity_verification_tokens;

CREATE INDEX identity_registration_codes_nid_flow_id_idx ON identity_registration_codes (nid ASC, selfservice_registration_flow_id ASC);
CREATE INDEX identity_registration_codes_id_nid_idx ON identity_registration_codes (id ASC, nid ASC);
DROP INDEX identity_registration_codes_nid_idx ON identity_registration_codes;

CREATE INDEX identity_recovery_tokens_nid_id_idx ON identity_recovery_tokens (nid ASC, id ASC);
CREATE INDEX identity_recovery_tokens_id_nid_idx ON identity_recovery_tokens (id ASC, nid ASC);
CREATE INDEX identity_recovery_tokens_token_nid_used_idx ON identity_recovery_tokens (nid ASC, token ASC, used ASC);
CREATE INDEX identity_recovery_tokens_identity_id_nid_idx ON identity_recovery_tokens (identity_id ASC, nid ASC);
DROP INDEX identity_recovery_tokens_identity_id_idx ON identity_recovery_tokens;
DROP INDEX identity_recovery_tokens_nid_idx ON identity_recovery_tokens;

CREATE INDEX identity_recovery_codes_nid_flow_id_idx ON identity_recovery_codes (nid ASC, selfservice_recovery_flow_id ASC);
CREATE INDEX identity_recovery_codes_id_nid_idx ON identity_recovery_codes (id ASC, nid ASC);
CREATE INDEX identity_recovery_codes_identity_id_nid_idx ON identity_recovery_codes (identity_id ASC, nid ASC);
CREATE INDEX identity_recovery_codes_identity_recovery_address_id_nid_idx ON identity_recovery_codes (identity_recovery_address_id ASC, nid ASC);
DROP INDEX identity_recovery_codes_address_id_idx ON identity_recovery_codes;
DROP INDEX identity_recovery_codes_identity_id_idx ON identity_recovery_codes;
DROP INDEX identity_recovery_codes_nid_idx ON identity_recovery_codes;

CREATE INDEX identity_login_codes_nid_flow_id_idx ON identity_login_codes (nid ASC, selfservice_login_flow_id ASC);
CREATE INDEX identity_login_codes_id_nid_idx ON identity_login_codes (id ASC, nid ASC);
DROP INDEX identity_login_codes_nid_idx ON identity_login_codes;
