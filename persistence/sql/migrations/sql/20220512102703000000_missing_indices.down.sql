CREATE INDEX sessions_nid_idx ON sessions (id, nid);
CREATE INDEX sessions_token_idx ON sessions (token);
CREATE INDEX sessions_logout_token_idx ON sessions (logout_token);

CREATE INDEX identities_nid_idx ON identities (id, nid);

CREATE INDEX continuity_containers_nid_idx ON continuity_containers (id, nid);

CREATE INDEX courier_messages_nid_idx ON courier_messages (id, nid);

CREATE INDEX identity_credential_identifiers_nid_idx ON identity_credential_identifiers (id, nid);

CREATE INDEX identity_credentials_nid_idx ON identity_credentials (id, nid);

CREATE INDEX identity_recovery_addresses_nid_idx ON identity_recovery_addresses (id, nid);

CREATE INDEX identity_recovery_tokens_nid_idx ON identity_recovery_tokens (id, nid);
CREATE INDEX identity_recovery_addresses_code_idx ON identity_recovery_tokens (token);

CREATE INDEX identity_verifiable_addresses_nid_idx ON identity_verifiable_addresses (id, nid);

CREATE INDEX identity_verification_tokens_nid_idx ON identity_verification_tokens (id, nid);
CREATE INDEX identity_verification_tokens_token_idx ON identity_verification_tokens (token);

CREATE INDEX selfservice_login_flows_nid_idx ON selfservice_login_flows (id,nid);
CREATE INDEX selfservice_recovery_flows_nid_idx ON selfservice_recovery_flows (id,nid);
CREATE INDEX selfservice_registration_flows_nid_idx ON selfservice_registration_flows (id,nid);
CREATE INDEX selfservice_settings_flows_nid_idx ON selfservice_settings_flows (id,nid);
CREATE INDEX selfservice_verification_flows_nid_idx ON selfservice_verification_flows (id,nid);

DROP INDEX sessions_identity_id_nid_idx;
DROP INDEX sessions_nid_id_identity_id_idx;
DROP INDEX sessions_id_nid_idx;
DROP INDEX sessions_token_nid_idx;

DROP INDEX identities_id_nid_idx;
DROP INDEX identities_nid_id_idx;
DROP INDEX continuity_containers_nid_id_idx;
DROP INDEX continuity_containers_id_nid_idx;
DROP INDEX courier_messages_nid_id_idx;
DROP INDEX courier_messages_id_nid_idx;
DROP INDEX identity_credential_identifiers_nid_id_idx;
DROP INDEX identity_credential_identifiers_id_nid_idx;
DROP INDEX identity_credentials_nid_id_idx;
DROP INDEX identity_credentials_id_nid_idx;
DROP INDEX identity_recovery_addresses_nid_id_idx;
DROP INDEX identity_recovery_addresses_id_nid_idx;
DROP INDEX identity_recovery_tokens_nid_id_idx;
DROP INDEX identity_recovery_tokens_id_nid_idx;
DROP INDEX identity_recovery_tokens_selfservice_recovery_flow_id_idx;
DROP INDEX identity_recovery_tokens_identity_recovery_address_id_idx;
DROP INDEX identity_verification_tokens_nid_id_idx;
DROP INDEX identity_verification_tokens_id_nid_idx;
DROP INDEX identity_verification_tokens_token_nid_used_idx;
DROP INDEX selfservice_login_flows_nid_id_idx;
DROP INDEX selfservice_login_flows_id_nid_idx;
DROP INDEX selfservice_recovery_flows_nid_id_idx;
DROP INDEX selfservice_recovery_flows_id_nid_idx;
DROP INDEX selfservice_registration_flows_nid_id_idx;
DROP INDEX selfservice_registration_flows_id_nid_idx;
DROP INDEX selfservice_settings_flows_nid_id_idx;
DROP INDEX selfservice_settings_flows_id_nid_idx;
DROP INDEX selfservice_verification_flows_nid_id_idx;
DROP INDEX selfservice_verification_flows_id_nid_idx;
