-- MySQL requires indexes on foreign keys and referenced keys so that foreign key checks can be fast and not require a table scan.
-- In the referencing table, there must be an index where the foreign key columns are listed as the first columns in the same order.
-- Such an index is created on the referencing table automatically if it does not exist. This index might be silently dropped later
-- if you create another index that can be used to enforce the foreign key constraint.

-- from https://dev.mysql.com/doc/refman/8.0/en/create-table-foreign-keys.html

-- -> We create new indexes to be consistent with the other databases. However, dropping those will be a bit different.

CREATE INDEX identity_recovery_codes_identity_id_nid_idx ON identity_recovery_codes (identity_id, nid);

CREATE INDEX identity_verification_codes_verifiable_address_nid_idx ON identity_verification_codes (identity_verifiable_address_id, nid);

CREATE INDEX selfservice_settings_flows_identity_id_nid_idx ON selfservice_settings_flows (identity_id, nid);

CREATE INDEX continuity_containers_identity_id_nid_idx ON continuity_containers (identity_id, nid);

CREATE INDEX selfservice_recovery_flows_recovered_identity_id_nid_idx ON selfservice_recovery_flows (recovered_identity_id, nid);

CREATE INDEX identity_recovery_tokens_identity_id_nid_idx ON identity_recovery_tokens (identity_id, nid);

CREATE INDEX identity_recovery_codes_identity_recovery_address_id_nid_idx ON identity_recovery_codes (identity_recovery_address_id, nid);
