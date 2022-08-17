
DROP INDEX identity_recovery_codes_nid_idx;

DROP TABLE identity_recovery_codes;

ALTER TABLE selfservice_recovery_flows DROP submit_count;
