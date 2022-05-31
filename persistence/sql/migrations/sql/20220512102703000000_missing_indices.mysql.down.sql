-- This file has a couple more indexes added which MySQL needs for its FK constraints. Other
-- databases generate those indices automatically.
CREATE INDEX sessions_nid_idx ON sessions (id, nid);
CREATE INDEX sessions_nid_mysqlfk_idx ON sessions (nid);

CREATE INDEX sessions_token_idx ON sessions (token);
CREATE INDEX sessions_mysql_identity_id_idx ON sessions (identity_id);
CREATE INDEX sessions_logout_token_idx ON sessions (logout_token);

CREATE INDEX identities_nid_idx ON identities (id, nid);
CREATE INDEX identities_nid_mysqlfk_idx ON identities (nid);

CREATE INDEX continuity_containers_nid_idx ON continuity_containers (id, nid);
CREATE INDEX continuity_containers_mysqlfk_idx ON continuity_containers (nid);

CREATE INDEX courier_messages_nid_idx ON courier_messages (id, nid);
CREATE INDEX courier_messages_mysqlfk_idx ON courier_messages (nid);

CREATE INDEX identity_credential_identifiers_nid_idx ON identity_credential_identifiers (id, nid);
CREATE INDEX identity_credential_identifiers_mysqlfk_idx ON identity_credential_identifiers (nid);

CREATE INDEX identity_credentials_nid_idx ON identity_credentials (id, nid);
CREATE INDEX identity_credentials_mysqlfk_idx ON identity_credentials (nid);

CREATE INDEX identity_recovery_addresses_nid_idx ON identity_recovery_addresses (id, nid);
CREATE INDEX identity_recovery_addresses_nid_mysqlfk_idx ON identity_recovery_addresses (nid);

CREATE INDEX identity_recovery_tokens_nid_idx ON identity_recovery_tokens (id, nid);
CREATE INDEX identity_recovery_tokens_nid_mysqlfk_idx ON identity_recovery_tokens (nid);
CREATE INDEX identity_recovery_addresses_code_idx ON identity_recovery_tokens (token);
CREATE INDEX identity_recovery_tokens_srf_id_mysqlfk_idx ON identity_recovery_tokens (selfservice_recovery_flow_id);
CREATE INDEX identity_recovery_tokens_ira_id_mysqlfk_idx ON identity_recovery_tokens (identity_recovery_address_id);

CREATE INDEX identity_verifiable_addresses_nid_idx ON identity_verifiable_addresses (id, nid);
CREATE INDEX identity_verifiable_addresses_nid_mysqlfk_idx ON identity_verifiable_addresses (nid);

CREATE INDEX identity_verification_tokens_nid_idx ON identity_verification_tokens (id, nid);
CREATE INDEX identity_verification_tokens_nid_mysqlfk_idx ON identity_verification_tokens (nid);
CREATE INDEX identity_verification_tokens_token_idx ON identity_verification_tokens (token);

CREATE INDEX selfservice_login_flows_nid_idx ON selfservice_login_flows (id, nid);
CREATE INDEX selfservice_login_flows_nid_mysqlfk_idx ON selfservice_login_flows (nid);

CREATE INDEX selfservice_recovery_flows_nid_idx ON selfservice_recovery_flows (id, nid);
CREATE INDEX selfservice_recovery_flows_nid_mysqlfk_idx ON selfservice_recovery_flows (nid);

CREATE INDEX selfservice_registration_flows_nid_idx ON selfservice_registration_flows (id, nid);
CREATE INDEX selfservice_registration_flows_nid_mysqlfk_idx ON selfservice_registration_flows (nid);

CREATE INDEX selfservice_settings_flows_nid_idx ON selfservice_settings_flows (id, nid);
CREATE INDEX selfservice_settings_flows_nid_mysqlfk_idx ON selfservice_settings_flows (nid);

CREATE INDEX selfservice_verification_flows_nid_idx ON selfservice_verification_flows (id, nid);
CREATE INDEX selfservice_verification_flows_nid_mysqlfk_idx ON selfservice_verification_flows (nid);


DROP INDEX sessions_nid_id_identity_id_idx ON sessions;
DROP INDEX sessions_id_nid_idx ON sessions;
DROP INDEX sessions_token_nid_idx ON sessions;

DROP INDEX sessions_identity_id_nid_idx ON sessions;
DROP INDEX identities_id_nid_idx ON identities;
DROP INDEX identities_nid_id_idx ON identities;
DROP INDEX continuity_containers_nid_id_idx ON continuity_containers;
DROP INDEX continuity_containers_id_nid_idx ON continuity_containers;
DROP INDEX courier_messages_nid_id_idx ON courier_messages;
DROP INDEX courier_messages_id_nid_idx ON courier_messages;
DROP INDEX identity_credential_identifiers_nid_id_idx ON identity_credential_identifiers;
DROP INDEX identity_credential_identifiers_id_nid_idx ON identity_credential_identifiers;
DROP INDEX identity_credentials_nid_id_idx ON identity_credentials;
DROP INDEX identity_credentials_id_nid_idx ON identity_credentials;
DROP INDEX identity_recovery_addresses_nid_id_idx ON identity_recovery_addresses;
DROP INDEX identity_recovery_addresses_id_nid_idx ON identity_recovery_addresses;
DROP INDEX identity_recovery_tokens_nid_id_idx ON identity_recovery_tokens;
DROP INDEX identity_recovery_tokens_id_nid_idx ON identity_recovery_tokens;
DROP INDEX identity_recovery_tokens_selfservice_recovery_flow_id_idx ON identity_recovery_tokens;
DROP INDEX identity_recovery_tokens_identity_recovery_address_id_idx ON identity_recovery_tokens;
DROP INDEX identity_verification_tokens_nid_id_idx ON identity_verification_tokens;
DROP INDEX identity_verification_tokens_id_nid_idx ON identity_verification_tokens;
DROP INDEX identity_verification_tokens_token_nid_used_idx ON identity_verification_tokens;
DROP INDEX selfservice_login_flows_nid_id_idx ON selfservice_login_flows;
DROP INDEX selfservice_login_flows_id_nid_idx ON selfservice_login_flows;
DROP INDEX selfservice_recovery_flows_nid_id_idx ON selfservice_recovery_flows;
DROP INDEX selfservice_recovery_flows_id_nid_idx ON selfservice_recovery_flows;
DROP INDEX selfservice_registration_flows_nid_id_idx ON selfservice_registration_flows;
DROP INDEX selfservice_registration_flows_id_nid_idx ON selfservice_registration_flows;
DROP INDEX selfservice_settings_flows_nid_id_idx ON selfservice_settings_flows;
DROP INDEX selfservice_settings_flows_id_nid_idx ON selfservice_settings_flows;
DROP INDEX selfservice_verification_flows_nid_id_idx ON selfservice_verification_flows;
DROP INDEX selfservice_verification_flows_id_nid_idx ON selfservice_verification_flows;
