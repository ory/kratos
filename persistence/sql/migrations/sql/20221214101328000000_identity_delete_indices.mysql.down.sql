-- MySQL requires indexes on foreign keys and referenced keys so that foreign key checks can be fast and not require a table scan.
-- In the referencing table, there must be an index where the foreign key columns are listed as the first columns in the same order.
-- Such an index is created on the referencing table automatically if it does not exist. This index might be silently dropped later
-- if you create another index that can be used to enforce the foreign key constraint.

-- from https://dev.mysql.com/doc/refman/8.0/en/create-table-foreign-keys.html

-- -> The indexes in question already existed. We have to create new ones that are just the foreign key to restore the previous state.

ALTER TABLE identity_recovery_codes ADD INDEX (identity_id);

DROP INDEX identity_recovery_codes_identity_id_nid_idx ON identity_recovery_codes;

ALTER TABLE identity_verification_codes ADD INDEX (identity_verifiable_address_id);

DROP INDEX identity_verification_codes_verifiable_address_nid_idx ON identity_verification_codes;

ALTER TABLE selfservice_settings_flows ADD INDEX (identity_id);

DROP INDEX selfservice_settings_flows_identity_id_nid_idx ON selfservice_settings_flows;

ALTER TABLE continuity_containers ADD INDEX (identity_id);

DROP INDEX continuity_containers_identity_id_nid_idx ON continuity_containers;

ALTER TABLE selfservice_recovery_flows ADD INDEX (recovered_identity_id);

DROP INDEX selfservice_recovery_flows_recovered_identity_id_nid_idx ON selfservice_recovery_flows;

ALTER TABLE identity_recovery_tokens ADD INDEX (identity_id);

DROP INDEX identity_recovery_tokens_identity_id_nid_idx ON identity_recovery_tokens;

ALTER TABLE identity_recovery_codes ADD INDEX (identity_recovery_address_id);

DROP INDEX identity_recovery_codes_identity_recovery_address_id_nid_idx ON identity_recovery_codes;
