
DROP INDEX identity_recovery_codes_id_nid_flow_id_idx;

DROP TABLE identity_recovery_codes;

ALTER TABLE selfservice_recovery_flows DROP submit_count;
