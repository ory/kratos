
DROP INDEX identity_recovery_codes_nid_flow_id_idx;
DROP INDEX identity_recovery_codes_id_nid_idx;

DROP TABLE identity_recovery_codes;

ALTER TABLE selfservice_recovery_flows DROP submit_count;

ALTER TABLE selfservice_recovery_flows DROP skip_csrf_check;
