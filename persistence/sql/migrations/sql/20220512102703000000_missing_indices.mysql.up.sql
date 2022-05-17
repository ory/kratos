CREATE INDEX sessions_identity_id_nid_idx ON sessions (identity_id, nid);

CREATE INDEX identities_id_nid_idx ON identities (id, nid);
CREATE INDEX identities_nid_id_idx ON identities (nid, id);
DROP INDEX identities_nid_idx ON identities;

CREATE INDEX continuity_containers_nid_id_idx ON continuity_containers (nid, id);
CREATE INDEX continuity_containers_id_nid_idx ON continuity_containers (id, nid);
DROP INDEX continuity_containers_nid_idx ON continuity_containers;

CREATE INDEX courier_messages_nid_id_idx ON courier_messages (nid, id);
CREATE INDEX courier_messages_id_nid_idx ON courier_messages (id, nid);
DROP INDEX courier_messages_nid_idx ON courier_messages;

CREATE INDEX identity_credential_identifiers_nid_id_idx ON identity_credential_identifiers (nid, id);
CREATE INDEX identity_credential_identifiers_id_nid_idx ON identity_credential_identifiers (id, nid);
DROP INDEX identity_credential_identifiers_nid_idx ON identity_credential_identifiers;

CREATE INDEX identity_credentials_nid_id_idx ON identity_credentials (nid, id);
CREATE INDEX identity_credentials_id_nid_idx ON identity_credentials (id, nid);
DROP INDEX identity_credentials_nid_idx ON identity_credentials;

CREATE INDEX identity_recovery_addresses_nid_id_idx ON identity_recovery_addresses (nid, id);
CREATE INDEX identity_recovery_addresses_id_nid_idx ON identity_recovery_addresses (id, nid);
DROP INDEX identity_recovery_addresses_nid_idx ON identity_recovery_addresses;

CREATE INDEX identity_recovery_tokens_nid_id_idx ON identity_recovery_tokens (nid, id);
CREATE INDEX identity_recovery_tokens_id_nid_idx ON identity_recovery_tokens (id, nid);
CREATE INDEX identity_recovery_tokens_selfservice_recovery_flow_id_idx ON identity_recovery_tokens (selfservice_recovery_flow_id);
CREATE INDEX identity_recovery_tokens_identity_recovery_address_id_idx ON identity_recovery_tokens (identity_recovery_address_id);
CREATE INDEX identity_recovery_tokens_token_nid_used_idx ON identity_recovery_tokens (nid, token, used);
DROP INDEX identity_recovery_tokens_nid_idx ON identity_recovery_tokens;
DROP INDEX identity_recovery_addresses_code_idx ON identity_recovery_tokens;

CREATE INDEX identity_verifiable_addresses_nid_id_idx ON identity_verifiable_addresses (nid, id);
CREATE INDEX identity_verifiable_addresses_id_nid_idx ON identity_verifiable_addresses (id, nid);
DROP INDEX identity_verifiable_addresses_nid_idx ON identity_verifiable_addresses;

CREATE INDEX identity_verification_tokens_nid_id_idx ON identity_verification_tokens (nid, id);
CREATE INDEX identity_verification_tokens_id_nid_idx ON identity_verification_tokens (id, nid);
CREATE INDEX identity_verification_tokens_token_nid_used_idx ON identity_verification_tokens (nid, token, used);
DROP INDEX identity_verification_tokens_nid_idx ON identity_verification_tokens;
DROP INDEX identity_verification_tokens_token_idx ON identity_verification_tokens;

CREATE INDEX selfservice_login_flows_nid_id_idx ON selfservice_login_flows (nid, id);
CREATE INDEX selfservice_login_flows_id_nid_idx ON selfservice_login_flows (id, nid);
DROP INDEX selfservice_login_flows_nid_idx ON selfservice_login_flows;

CREATE INDEX selfservice_recovery_flows_nid_id_idx ON selfservice_recovery_flows (nid, id);
CREATE INDEX selfservice_recovery_flows_id_nid_idx ON selfservice_recovery_flows (id, nid);
DROP INDEX selfservice_recovery_flows_nid_idx ON selfservice_recovery_flows;

CREATE INDEX selfservice_registration_flows_nid_id_idx ON selfservice_registration_flows (nid, id);
CREATE INDEX selfservice_registration_flows_id_nid_idx ON selfservice_registration_flows (id, nid);
DROP INDEX selfservice_registration_flows_nid_idx ON selfservice_registration_flows;

CREATE INDEX selfservice_settings_flows_nid_id_idx ON selfservice_settings_flows (nid, id);
CREATE INDEX selfservice_settings_flows_id_nid_idx ON selfservice_settings_flows (id, nid);
DROP INDEX selfservice_settings_flows_nid_idx ON selfservice_settings_flows;

CREATE INDEX selfservice_verification_flows_nid_id_idx ON selfservice_verification_flows (nid, id);
CREATE INDEX selfservice_verification_flows_id_nid_idx ON selfservice_verification_flows (id, nid);
DROP INDEX selfservice_verification_flows_nid_idx ON selfservice_verification_flows;

CREATE INDEX sessions_nid_id_identity_id_idx ON sessions (nid, identity_id, id);
CREATE INDEX sessions_id_nid_idx ON sessions (id, nid);
CREATE INDEX sessions_token_nid_idx ON sessions (nid, token);
DROP INDEX sessions_nid_idx ON sessions;
DROP INDEX sessions_token_idx ON sessions;
DROP INDEX sessions_logout_token_idx ON sessions;
