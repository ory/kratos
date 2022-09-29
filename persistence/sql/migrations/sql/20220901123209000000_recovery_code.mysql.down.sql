DROP TABLE identity_recovery_codes;

ALTER TABLE selfservice_recovery_flows DROP submit_count;

ALTER TABLE selfservice_recovery_flows DROP skip_csrf_check;
