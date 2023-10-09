CREATE INDEX IF NOT EXISTS "identity_recovery_codes_identity_id_nid_idx" ON "identity_recovery_codes" (identity_id, nid);

CREATE INDEX IF NOT EXISTS "identity_verification_codes_verifiable_address_nid_idx" ON "identity_verification_codes" (identity_verifiable_address_id, nid);

CREATE INDEX IF NOT EXISTS "selfservice_settings_flows_identity_id_nid_idx" ON "selfservice_settings_flows" (identity_id, nid);

CREATE INDEX IF NOT EXISTS "continuity_containers_identity_id_nid_idx" ON "continuity_containers" (identity_id, nid);

CREATE INDEX IF NOT EXISTS "selfservice_recovery_flows_recovered_identity_id_nid_idx" ON "selfservice_recovery_flows" (recovered_identity_id, nid);

CREATE INDEX IF NOT EXISTS "identity_recovery_tokens_identity_id_nid_idx" ON "identity_recovery_tokens" (identity_id, nid);

CREATE INDEX IF NOT EXISTS "identity_recovery_codes_identity_recovery_address_id_nid_idx" ON "identity_recovery_codes" (identity_recovery_address_id, nid);
